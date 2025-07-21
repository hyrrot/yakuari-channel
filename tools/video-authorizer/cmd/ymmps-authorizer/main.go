package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/yakuari-channel/video-authorizer/internal/converter"
	"github.com/yakuari-channel/video-authorizer/internal/parser"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		log.Fatal(err)
	}
}

func newRootCmd() *cobra.Command {
	var (
		templateFile  string
		outputPath    string
		validateOnly  bool
		dryRun        bool
		verbose       bool
		basePath      string
	)

	rootCmd := &cobra.Command{
		Use:   "ymmps-authorizer [flags] <YMMPS_FILE_PATH>",
		Short: "Convert YMMPS scenario files to YMMP project files",
		Long: `ymmps-authorizer is a tool that converts YMMPS (YukkuriMovieMaker Project Scenario) 
files to YMMP (YukkuriMovieMaker Project) files using templates.`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			scenarioPath := args[0]

			// Validate-only mode
			if validateOnly {
				return runValidateOnly(scenarioPath, verbose)
			}

			// Template file is required for conversion
			if templateFile == "" {
				return fmt.Errorf("--template-file is required")
			}

			// Dry-run mode
			if dryRun {
				return runDryRun(scenarioPath, templateFile, basePath, verbose)
			}

			// Normal conversion
			return runConvert(scenarioPath, templateFile, outputPath, basePath, verbose)
		},
	}

	// Add flags according to requirements
	rootCmd.Flags().StringVar(&templateFile, "template-file", "", "Template YMMP file path (required)")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output YMMP file path (default: <scenario>.ymmp)")
	rootCmd.Flags().BoolVar(&validateOnly, "validate-only", false, "Validate YMMPS file without conversion")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	rootCmd.Flags().StringVar(&basePath, "base-path", "", "Base path for resolving relative paths")

	// Keep the old convert command for backward compatibility
	rootCmd.AddCommand(newConvertCmd())
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

func newConvertCmd() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "convert <scenario.ymmps> <template.ymmp>",
		Short: "Convert YMMPS scenario to YMMP project",
		Long: `Convert a YMMPS scenario file to a YMMP project file using a template.
		
The template YMMP file should contain pre-configured items that will be referenced
by the YMMPS scenario file.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			scenarioPath := args[0]
			templatePath := args[1]

			// Parse YMMPS file
			ymmpsParser := parser.NewYMMPSParser()
			ymmpsDoc, err := ymmpsParser.ParseFile(scenarioPath)
			if err != nil {
				return fmt.Errorf("failed to parse YMMPS file: %w", err)
			}

			// Parse template YMMP file
			ymmpParser := parser.NewYMMPParser()
			templateProject, err := ymmpParser.ParseFile(templatePath)
			if err != nil {
				return fmt.Errorf("failed to parse template YMMP file: %w", err)
			}

			// Convert YMMPS to YMMP
			conv := converter.NewConverter()
			project, err := conv.Convert(ymmpsDoc, templateProject)
			if err != nil {
				return fmt.Errorf("failed to convert: %w", err)
			}

			// Determine output path
			if outputPath == "" {
				outputPath = scenarioPath[:len(scenarioPath)-6] + ".ymmp"
			}

			// Write output
			output, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer output.Close()

			encoder := json.NewEncoder(output)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(project); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}

			fmt.Printf("Successfully converted %s to %s\n", scenarioPath, outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output YMMP file path (default: <scenario>.ymmp)")

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ymmps-authorizer %s (%s) built at %s\n", version, commit, date)
		},
	}
}

// runValidateOnly validates the YMMPS file without conversion
func runValidateOnly(scenarioPath string, verbose bool) error {
	if verbose {
		fmt.Printf("Validating YMMPS file: %s\n", scenarioPath)
	}

	// Parse and validate YMMPS file
	ymmpsParser := parser.NewYMMPSParser()
	ymmpsDoc, err := ymmpsParser.ParseFile(scenarioPath)
	if err != nil {
		return fmt.Errorf("failed to parse YMMPS file: %w", err)
	}

	// Validate the document
	if err := ymmpsDoc.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if verbose {
		fmt.Printf("Validation successful. Found %d sequences.\n", len(ymmpsDoc.Sequences))
	} else {
		fmt.Println("Validation successful.")
	}

	return nil
}

// runDryRun shows what would be done without making changes
func runDryRun(scenarioPath, templateFile, basePath string, verbose bool) error {
	fmt.Println("=== DRY RUN MODE ===")
	fmt.Printf("Scenario file: %s\n", scenarioPath)
	fmt.Printf("Template file: %s\n", templateFile)
	if basePath != "" {
		fmt.Printf("Base path: %s\n", basePath)
	}

	// Parse YMMPS file
	ymmpsParser := parser.NewYMMPSParser()
	ymmpsDoc, err := ymmpsParser.ParseFile(scenarioPath)
	if err != nil {
		return fmt.Errorf("failed to parse YMMPS file: %w", err)
	}

	// Validate
	if err := ymmpsDoc.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Parse template
	ymmpParser := parser.NewYMMPParser()
	_, err = ymmpParser.ParseFile(templateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template YMMP file: %w", err)
	}

	// Report what would be done
	fmt.Printf("\nWould process:\n")
	fmt.Printf("- %d sequences\n", len(ymmpsDoc.Sequences))
	
	totalScenes := 0
	totalShots := 0
	totalItems := 0
	
	for _, seq := range ymmpsDoc.Sequences {
		totalScenes += len(seq.Scenes)
		for _, scene := range seq.Scenes {
			totalShots += len(scene.Shots)
			for _, shot := range scene.Shots {
				totalItems += len(shot.Items)
			}
		}
	}
	
	fmt.Printf("- %d scenes\n", totalScenes)
	fmt.Printf("- %d shots\n", totalShots)
	fmt.Printf("- %d items\n", totalItems)

	outputPath := scenarioPath[:len(scenarioPath)-6] + ".ymmp"
	fmt.Printf("\nWould write output to: %s\n", outputPath)

	return nil
}

// runConvert performs the actual conversion
func runConvert(scenarioPath, templateFile, outputPath, basePath string, verbose bool) error {
	if verbose {
		fmt.Printf("Converting %s using template %s\n", scenarioPath, templateFile)
		if basePath != "" {
			fmt.Printf("Base path: %s\n", basePath)
		}
	}

	// Parse YMMPS file
	ymmpsParser := parser.NewYMMPSParser()
	ymmpsDoc, err := ymmpsParser.ParseFile(scenarioPath)
	if err != nil {
		return fmt.Errorf("failed to parse YMMPS file: %w", err)
	}

	// Parse template YMMP file
	ymmpParser := parser.NewYMMPParser()
	templateProject, err := ymmpParser.ParseFile(templateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template YMMP file: %w", err)
	}

	// Convert YMMPS to YMMP
	var conv *converter.Converter
	if basePath != "" {
		conv = converter.NewConverterWithBasePath(basePath)
		if verbose {
			fmt.Printf("Using base path for relative paths: %s\n", basePath)
		}
	} else {
		conv = converter.NewConverter()
	}
	
	// Determine output path first
	if outputPath == "" {
		outputPath = scenarioPath[:len(scenarioPath)-6] + ".ymmp"
	}
	
	project, err := conv.ConvertWithOutput(ymmpsDoc, templateProject, outputPath)
	if err != nil {
		return fmt.Errorf("failed to convert: %w", err)
	}

	// Output path was already determined above

	// Write output
	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(project); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if verbose {
		fmt.Printf("Successfully converted %s to %s\n", scenarioPath, outputPath)
	} else {
		fmt.Printf("Conversion successful: %s\n", outputPath)
	}
	
	return nil
}