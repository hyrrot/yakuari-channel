package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/user/ymmp-compiler/internal/core/converter"
	"github.com/user/ymmp-compiler/internal/core/parser"
	"github.com/user/ymmp-compiler/internal/core/validator"
	"github.com/user/ymmp-compiler/internal/plugins"
	"github.com/user/ymmp-compiler/internal/plugins/items"
)

func main() {
	var (
		input    = flag.String("i", "", "input YAML file path")
		output   = flag.String("o", "", "output YMMP file path")
		basePath = flag.String("base-path", "", "base directory for resolving relative paths (default: directory of input file)")
		help     = flag.Bool("help", false, "show help")
	)
	
	flag.Parse()
	
	if *help {
		printHelp()
		return
	}
	
	if *input == "" || *output == "" {
		fmt.Fprintf(os.Stderr, "Error: both -i and -o flags are required\n")
		printHelp()
		os.Exit(1)
	}
	
	fmt.Printf("Converting %s to %s...\n", *input, *output)
	
	// Validate input file exists
	if _, err := os.Stat(*input); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Input file '%s' does not exist\n", *input)
		os.Exit(1)
	}
	
	// Determine base path
	resolvedBasePath := *basePath
	if resolvedBasePath == "" {
		// Use the directory of the input file as the default base path
		resolvedBasePath = filepath.Dir(*input)
	}
	
	// Perform the conversion
	err := ConvertAndWriteYMMPFile(*input, *output, resolvedBasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Conversion failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Conversion completed successfully")
}

// ConvertAndWriteYMMPFile converts a SYMMP file and writes it to a YMMP file
func ConvertAndWriteYMMPFile(inputPath, outputPath, basePath string) error {
	// Open the input file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer file.Close()

	// Create plugin registry
	registry := plugins.NewPluginRegistry()
	registry.RegisterItemPlugin(items.NewImagePlugin())
	registry.RegisterItemPlugin(items.NewVoicePlugin())
	registry.RegisterItemPlugin(items.NewAudioPlugin())
	registry.RegisterItemPlugin(items.NewVideoPlugin())
	registry.RegisterItemPlugin(items.NewTachiePlugin())

	// Parse the SYMMP file
	episode, err := parser.ParseSYMMPWithPlugins(file, registry)
	if err != nil {
		return fmt.Errorf("failed to parse SYMMP file: %v", err)
	}

	// Validate the episode
	if err := validator.Validate(episode); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	// Calculate timeline
	if err := converter.CalculateTimeline(episode); err != nil {
		return fmt.Errorf("failed to calculate timeline: %v", err)
	}

	// Convert to YMMP format
	ymmProject, err := converter.ConvertToYMMP(episode, basePath)
	if err != nil {
		return fmt.Errorf("failed to convert to YMMP: %v", err)
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(ymmProject, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal YMMP data: %v", err)
	}

	// Write to output file
	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	return nil
}

func printHelp() {
	fmt.Println("ymmp-compiler - Convert SYMMP files to YMMP format")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  ymmp-compiler -i <input.symmp> -o <output.ymmp>")
	fmt.Println("")
	fmt.Println("Required flags:")
	fmt.Println("  -i, --input <path>      input YAML file path")
	fmt.Println("  -o, --output <path>     output YMMP file path")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --base-path <path>      base directory for resolving relative paths")
	fmt.Println("  --help                  show this help message")
}