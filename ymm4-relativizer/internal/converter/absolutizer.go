package converter

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/hyrrot/ymm4-relativizer/internal/model"
	"github.com/hyrrot/ymm4-relativizer/internal/utils"
)

type AbsolutizerConfig struct {
	InputPath   string
	OutputDir   string
	SkipMissing bool
}

func Absolutize(config AbsolutizerConfig) error {
	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Read input file
	data, err := os.ReadFile(config.InputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Parse YMMP file
	ymmp, err := model.ParseYMMP(data)
	if err != nil {
		return fmt.Errorf("failed to parse YMMP file: %w", err)
	}

	// Get input file directory path
	inputDir := filepath.Dir(config.InputPath)

	// Build output file path
	outName := filepath.Base(config.InputPath)
	outName = outName[:len(outName)-len(filepath.Ext(outName))] + ".ymmp"
	outPath := filepath.Join(config.OutputDir, outName)
	absOutPath, err := filepath.Abs(outPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute output path: %w", err)
	}

	// Process file paths
	processPath := func(path string, isRoot bool) string {
		if isRoot {
			// Set root FilePath to output file's absolute path
			return absOutPath
		}

		if path == "" {
			return path
		}

		// Convert relative path to absolute path
		absPath := filepath.Join(inputDir, path)
		absPath, err := filepath.Abs(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to get absolute path for %s: %v\n", path, err)
			return path
		}
		
		// Check if file exists
		if !utils.FileExists(absPath) {
			if config.SkipMissing {
				return path
			}
			fmt.Fprintf(os.Stderr, "Warning: File does not exist %s\n", absPath)
			return path
		}

		return absPath
	}

	// Update FilePaths
	ymmp.UpdateFilePaths(processPath)

	// Save result
	output, err := ymmp.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to generate JSON: %w", err)
	}

	if err := os.WriteFile(outPath, output, 0644); err != nil {
		return fmt.Errorf("failed to save output file: %w", err)
	}

	return nil
} 