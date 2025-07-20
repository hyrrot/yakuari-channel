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
	rootCmd := &cobra.Command{
		Use:   "ymmps-authorizer",
		Short: "Convert YMMPS scenario files to YMMP project files",
		Long: `ymmps-authorizer is a tool that converts YMMPS (YukkuriMovieMaker Project Scenario) 
files to YMMP (YukkuriMovieMaker Project) files using templates.`,
		SilenceUsage: true,
	}

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