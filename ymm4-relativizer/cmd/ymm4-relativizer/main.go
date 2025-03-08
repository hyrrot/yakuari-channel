package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyrrot/ymm4-relativizer/internal/converter"
)

type Config struct {
	Mode            string
	InputPath       string
	OutputDir       string
	AssetsDir       string
	DirectoryMode   string
	SkipMissing     bool
	DirectoryLevels int
}

func main() {
	config := parseFlags()
	
	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() *Config {
	config := &Config{}
	
	flag.StringVar(&config.Mode, "mode", "relativize", "Conversion mode (relativize/absolutize)")
	flag.StringVar(&config.InputPath, "input", "", "Input file or directory path")
	flag.StringVar(&config.OutputDir, "output", "", "Output directory")
	flag.StringVar(&config.AssetsDir, "assets", "assets", "Assets directory name")
	flag.StringVar(&config.DirectoryMode, "dirmode", "full", "Directory mode (full/partial/flat)")
	flag.IntVar(&config.DirectoryLevels, "levels", 2, "Number of directory levels to keep")
	flag.BoolVar(&config.SkipMissing, "skip-missing", false, "Skip missing files")
	
	flag.Parse()
	
	if config.InputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: Input path is required")
		flag.Usage()
		os.Exit(1)
	}

	if config.OutputDir == "" {
		fmt.Fprintln(os.Stderr, "Error: Output directory is required")
		flag.Usage()
		os.Exit(1)
	}

	if config.DirectoryMode != "full" && config.DirectoryMode != "partial" && config.DirectoryMode != "flat" {
		fmt.Fprintln(os.Stderr, "Error: Invalid directory mode. Must be one of: full, partial, or flat")
		flag.Usage()
		os.Exit(1)
	}

	if config.DirectoryMode == "partial" && config.DirectoryLevels < 1 {
		fmt.Fprintln(os.Stderr, "Error: Directory levels must be greater than 0")
		flag.Usage()
		os.Exit(1)
	}
	
	return config
}

func run(config *Config) error {
	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", config.OutputDir, err)
	}

	switch config.Mode {
	case "relativize":
		if err := converter.Relativize(converter.RelativizerConfig{
			InputPath:       config.InputPath,
			OutputDir:       config.OutputDir,
			AssetsDir:       config.AssetsDir,
			DirectoryMode:   config.DirectoryMode,
			SkipMissing:     config.SkipMissing,
			DirectoryLevels: config.DirectoryLevels,
		}); err != nil {
			return fmt.Errorf("relativization failed: %w", err)
		}

	case "absolutize":
		if err := converter.Absolutize(converter.AbsolutizerConfig{
			InputPath:   config.InputPath,
			OutputDir:   config.OutputDir,
			SkipMissing: config.SkipMissing,
		}); err != nil {
			return fmt.Errorf("absolutization failed: %w", err)
		}

	default:
		return fmt.Errorf("invalid mode: %s", config.Mode)
	}

	return nil
} 