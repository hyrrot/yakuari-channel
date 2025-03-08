package converter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hyrrot/ymm4-relativizer/internal/model"
)

func TestIntegrationWithRealFile(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ymmp-integration-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test cases
	tests := []struct {
		name        string
		config      RelativizerConfig
		checkOutput func(t *testing.T, outputPath string, assetsDir string)
	}{
		{
			name: "Relativize real YMMP file (full mode)",
			config: RelativizerConfig{
				InputPath:     "../../.work/movie-new-remove-voicecache.ymmp",
				OutputDir:     filepath.Join(tmpDir, "full"),
				AssetsDir:     "assets",
				DirectoryMode: "full",
				SkipMissing:   true,
			},
			checkOutput: func(t *testing.T, outputPath string, assetsDir string) {
				// Check if output file exists
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("Output file was not created: %s", outputPath)
					return
				}

				// Check if file can be parsed as JSON
				data, err := os.ReadFile(outputPath)
				if err != nil {
					t.Errorf("Failed to read output file: %v", err)
					return
				}

				var result map[string]interface{}
				if err := json.Unmarshal(data, &result); err != nil {
					t.Errorf("Failed to parse output file as JSON: %v", err)
					return
				}

				// Check if root FilePath is null
				if result["FilePath"] != nil {
					t.Error("Root FilePath is not null")
				}

				// Check if assets directory was created
				if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
					t.Error("Assets directory was not created")
				}
			},
		},
		{
			name: "Relativize real YMMP file (flat mode)",
			config: RelativizerConfig{
				InputPath:     "../../.work/movie-new-remove-voicecache.ymmp",
				OutputDir:     filepath.Join(tmpDir, "flat"),
				AssetsDir:     "assets",
				DirectoryMode: "flat",
				SkipMissing:   true,
			},
			checkOutput: func(t *testing.T, outputPath string, assetsDir string) {
				// Check if output file exists
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("Output file was not created: %s", outputPath)
					return
				}

				// Check if file can be parsed as JSON
				data, err := os.ReadFile(outputPath)
				if err != nil {
					t.Errorf("Failed to read output file: %v", err)
					return
				}

				var result map[string]interface{}
				if err := json.Unmarshal(data, &result); err != nil {
					t.Errorf("Failed to parse output file as JSON: %v", err)
					return
				}

				// Check assets directory contents
				files, err := os.ReadDir(assetsDir)
				if err != nil {
					t.Errorf("Failed to read assets directory: %v", err)
					return
				}

				// In flat mode, all files should be directly in the assets directory
				for _, file := range files {
					if file.IsDir() {
						t.Errorf("Found subdirectory in flat mode: %s", file.Name())
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run relativization
			if err := Relativize(tt.config); err != nil {
				t.Fatalf("Relativization failed: %v", err)
			}

			// Build output paths
			baseName := filepath.Base(tt.config.InputPath)
			outputName := baseName[:len(baseName)-len(filepath.Ext(baseName))] + ".ymmpr"
			outputPath := filepath.Join(tt.config.OutputDir, outputName)
			assetsDir := filepath.Join(tt.config.OutputDir, tt.config.AssetsDir)

			// Verify output
			tt.checkOutput(t, outputPath, assetsDir)

			// Test absolutization
			absolutizeConfig := AbsolutizerConfig{
				InputPath:   outputPath,
				OutputDir:   filepath.Join(tt.config.OutputDir, "abs"),
				SkipMissing: true,
			}

			// Run absolutization
			if err := Absolutize(absolutizeConfig); err != nil {
				t.Fatalf("Absolutization failed: %v", err)
			}

			// Build absolutized output paths
			absOutputName := outputName[:len(outputName)-len(filepath.Ext(outputName))] + ".ymmp"
			absOutputPath := filepath.Join(absolutizeConfig.OutputDir, absOutputName)

			// Check if absolutized output file exists
			if _, err := os.Stat(absOutputPath); os.IsNotExist(err) {
				t.Errorf("Absolutized output file was not created: %s", absOutputPath)
			}

			// Check if file can be parsed as JSON
			data, err := os.ReadFile(absOutputPath)
			if err != nil {
				t.Errorf("Failed to read absolutized output file: %v", err)
				return
			}

			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Errorf("Failed to parse absolutized output file as JSON: %v", err)
				return
			}

			// Check if root FilePath is an absolute path
			if filePath, ok := result["FilePath"].(string); !ok {
				t.Error("Root FilePath is not a string")
			} else if !filepath.IsAbs(filePath) {
				t.Error("Root FilePath is not an absolute path")
			}
		})
	}
}

func TestIntegrationEdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ymm4-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		mode     string
		validate func(t *testing.T, outputPath string)
	}{
		{
			name: "Empty YMMP file",
			setup: func(t *testing.T) string {
				path := filepath.Join(tempDir, "empty.ymmp")
				if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			mode: "full",
			validate: func(t *testing.T, outputPath string) {
				content, err := os.ReadFile(outputPath)
				if err != nil {
					t.Fatal(err)
				}
				var data map[string]interface{}
				if err := json.Unmarshal(content, &data); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "YMMP with non-existent files",
			setup: func(t *testing.T) string {
				filePath := "non_existent.wav"
				ymmp := struct {
					FilePath *string `json:"FilePath"`
					Items    []struct {
						FilePath *string `json:"FilePath"`
					} `json:"Items"`
				}{
					FilePath: &filePath,
					Items: []struct {
						FilePath *string `json:"FilePath"`
					}{
						{FilePath: &filePath},
					},
				}
				data, err := json.Marshal(ymmp)
				if err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(tempDir, "non_existent.ymmp")
				if err := os.WriteFile(path, data, 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			mode: "flat",
			validate: func(t *testing.T, outputPath string) {
				content, err := os.ReadFile(outputPath)
				if err != nil {
					t.Fatal(err)
				}
				var data map[string]interface{}
				if err := json.Unmarshal(content, &data); err != nil {
					t.Fatal(err)
				}
				filePath, ok := data["FilePath"].(string)
				if !ok || !strings.Contains(filePath, "-non_existent.wav") {
					t.Error("expected flattened path for non-existent file")
				}
			},
		},
		{
			name: "YMMP with special characters in paths",
			setup: func(t *testing.T) string {
				specialPath := filepath.Join(tempDir, "特殊な名前")
				if err := os.MkdirAll(specialPath, 0755); err != nil {
					t.Fatal(err)
				}
				ymmp := &YMMP{
					FilePath: filepath.Join(specialPath, "テスト.wav"),
					Items: []Item{
						{FilePath: filepath.Join(specialPath, "🎮.wav")},
					},
				}
				data, err := json.Marshal(ymmp)
				if err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(tempDir, "special.ymmp")
				if err := os.WriteFile(path, data, 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			mode: "full",
			validate: func(t *testing.T, outputPath string) {
				content, err := os.ReadFile(outputPath)
				if err != nil {
					t.Fatal(err)
				}
				var ymmp YMMP
				if err := json.Unmarshal(content, &ymmp); err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(ymmp.FilePath, "特殊な名前") {
					t.Error("expected path to contain Japanese characters")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := tt.setup(t)
			outputDir := filepath.Join(tempDir, "output")
			
			err := Relativize(RelativizerConfig{
				InputPath:     inputPath,
				OutputDir:     outputDir,
				DirectoryMode: tt.mode,
			})
			if err != nil {
				t.Fatal(err)
			}
			
			outputPath := filepath.Join(outputDir, filepath.Base(inputPath))
			tt.validate(t, outputPath)
		})
	}
} 
