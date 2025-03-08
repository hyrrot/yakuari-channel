package converter

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/hyrrot/ymm4-relativizer/internal/model"
	"github.com/hyrrot/ymm4-relativizer/internal/utils"
)

type RelativizerConfig struct {
	InputPath       string
	OutputDir       string
	AssetsDir       string
	DirectoryMode   string
	SkipMissing     bool
	DirectoryLevels int
}

func Relativize(config RelativizerConfig) error {
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

	// Create assets directory path
	assetsDir := filepath.Join(config.OutputDir, config.AssetsDir)

	// Process file paths
	processPath := func(path string, isRoot bool) string {
		if path == "" {
			return path
		}

		// Set root FilePath to null
		if isRoot {
			return ""
		}

		// Check if file exists
		if !utils.FileExists(path) {
			if config.SkipMissing {
				return path
			}
			return path
		}

		// Generate new path
		newPath := utils.ProcessPathByMode(path, config.DirectoryMode, config.DirectoryLevels)
		dstPath := filepath.Join(assetsDir, newPath)

		// Copy file
		if err := utils.CopyFile(path, dstPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to copy file %s: %v\n", path, err)
			return path
		}

		// Convert to relative path
		rel, err := filepath.Rel(config.OutputDir, dstPath)
		if err != nil {
			return path
		}
		return filepath.ToSlash(rel)
	}

	// Update FilePaths
	ymmp.UpdateFilePaths(processPath)

	// Save result
	output, err := ymmp.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to generate JSON: %w", err)
	}

	// Save output file
	outPath := filepath.Join(config.OutputDir, filepath.Base(config.InputPath[:len(config.InputPath)-len(filepath.Ext(config.InputPath))]+".ymmpr"))
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	if err := os.WriteFile(outPath, output, 0644); err != nil {
		return fmt.Errorf("failed to save output file: %w", err)
	}

	return nil
} 