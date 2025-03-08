package converter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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
		setup    func(t *testing.T, tempDir string) (string, string) // 入力パスと出力ディレクトリを返す
		mode     string
		validate func(t *testing.T, outputPath string)
	}{
		{
			name: "Empty YMMP file",
			setup: func(t *testing.T, tempDir string) (string, string) {
				inputDir := filepath.Join(tempDir, "input")
				outputDir := filepath.Join(tempDir, "output")
				
				// 入力ディレクトリと出力ディレクトリを作成
				if err := os.MkdirAll(inputDir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatal(err)
				}
				
				// 入力ファイルを作成
				inputPath := filepath.Join(inputDir, "empty.ymmp")
				if err := os.WriteFile(inputPath, []byte("{}"), 0644); err != nil {
					t.Fatal(err)
				}
				
				return inputPath, outputDir
			},
			mode: "full",
			validate: func(t *testing.T, outputPath string) {
				// .ymmpr拡張子で出力ファイルを確認
				outputPath = outputPath[:len(outputPath)-len(filepath.Ext(outputPath))] + ".ymmpr"
				
				// 出力ファイルが存在することを確認
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Fatalf("output file not found: %s", outputPath)
				}
				
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
			setup: func(t *testing.T, tempDir string) (string, string) {
				inputDir := filepath.Join(tempDir, "input2")
				outputDir := filepath.Join(tempDir, "output2")
				
				// 入力ディレクトリと出力ディレクトリを作成
				if err := os.MkdirAll(inputDir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					t.Fatal(err)
				}
				
				data := map[string]interface{}{
					"FilePath": "non_existent.wav",
					"Items": []map[string]interface{}{
						{
							"FilePath": "non_existent.wav",
						},
					},
				}
				jsonData, err := json.Marshal(data)
				if err != nil {
					t.Fatal(err)
				}
				
				// 入力ファイルを作成
				inputPath := filepath.Join(inputDir, "non_existent.ymmp")
				if err := os.WriteFile(inputPath, jsonData, 0644); err != nil {
					t.Fatal(err)
				}
				
				return inputPath, outputDir
			},
			mode: "flat",
			validate: func(t *testing.T, outputPath string) {
				// .ymmpr拡張子で出力ファイルを確認
				outputPath = outputPath[:len(outputPath)-len(filepath.Ext(outputPath))] + ".ymmpr"
				
				// 出力ファイルが存在することを確認
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Fatalf("output file not found: %s", outputPath)
				}
				
				content, err := os.ReadFile(outputPath)
				if err != nil {
					t.Fatal(err)
				}
				var data map[string]interface{}
				if err := json.Unmarshal(content, &data); err != nil {
					t.Fatal(err)
				}

				// ルートのFilePathがnullであることを確認
				if data["FilePath"] != nil {
					t.Error("expected root FilePath to be null")
				}

				// Itemsの中のFilePathが変更されていないことを確認
				if items, ok := data["Items"].([]interface{}); ok {
					if len(items) > 0 {
						if item, ok := items[0].(map[string]interface{}); ok {
							if filePath, ok := item["FilePath"].(string); ok {
								if filePath != "non_existent.wav" {
									t.Errorf("expected non-existent file path to remain unchanged, got %q", filePath)
								}
							}
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setupで入力パスと出力ディレクトリを取得
			inputPath, outputDir := tt.setup(t, tempDir)
			
			// Relativize関数を実行
			err := Relativize(RelativizerConfig{
				InputPath:     inputPath,
				OutputDir:     outputDir,
				DirectoryMode: tt.mode,
			})
			if err != nil {
				t.Fatal(err)
			}
			
			// 出力ファイルのパスを構築
			outputPath := filepath.Join(outputDir, filepath.Base(inputPath))
			
			// 検証を実行
			tt.validate(t, outputPath)
		})
	}
} 
