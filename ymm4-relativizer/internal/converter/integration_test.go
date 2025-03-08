package converter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestIntegrationWithRealFile(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "ymmp-integration-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// テストケース
	tests := []struct {
		name        string
		config      RelativizerConfig
		checkOutput func(t *testing.T, outputPath string, assetsDir string)
	}{
		{
			name: "実際のYMMPファイルの相対化（フルモード）",
			config: RelativizerConfig{
				InputPath:     "../../.work/movie-new-remove-voicecache.ymmp",
				OutputDir:     filepath.Join(tmpDir, "full"),
				AssetsDir:     "assets",
				DirectoryMode: "full",
				SkipMissing:   true,
			},
			checkOutput: func(t *testing.T, outputPath string, assetsDir string) {
				// 出力ファイルの存在確認
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("出力ファイルが作成されていません: %s", outputPath)
					return
				}

				// JSONとして読み込めることを確認
				data, err := os.ReadFile(outputPath)
				if err != nil {
					t.Errorf("出力ファイルの読み込みに失敗: %v", err)
					return
				}

				var result map[string]interface{}
				if err := json.Unmarshal(data, &result); err != nil {
					t.Errorf("出力ファイルのJSONパースに失敗: %v", err)
					return
				}

				// ルートのFilePathがnullになっていることを確認
				if result["FilePath"] != nil {
					t.Error("ルートのFilePathがnullになっていません")
				}

				// アセットディレクトリが作成されていることを確認
				if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
					t.Error("アセットディレクトリが作成されていません")
				}
			},
		},
		{
			name: "実際のYMMPファイルの相対化（フラットモード）",
			config: RelativizerConfig{
				InputPath:     "../../.work/movie-new-remove-voicecache.ymmp",
				OutputDir:     filepath.Join(tmpDir, "flat"),
				AssetsDir:     "assets",
				DirectoryMode: "flat",
				SkipMissing:   true,
			},
			checkOutput: func(t *testing.T, outputPath string, assetsDir string) {
				// 出力ファイルの存在確認
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("出力ファイルが作成されていません: %s", outputPath)
					return
				}

				// JSONとして読み込めることを確認
				data, err := os.ReadFile(outputPath)
				if err != nil {
					t.Errorf("出力ファイルの読み込みに失敗: %v", err)
					return
				}

				var result map[string]interface{}
				if err := json.Unmarshal(data, &result); err != nil {
					t.Errorf("出力ファイルのJSONパースに失敗: %v", err)
					return
				}

				// アセットディレクトリの内容を確認
				files, err := os.ReadDir(assetsDir)
				if err != nil {
					t.Errorf("アセットディレクトリの読み取りに失敗: %v", err)
					return
				}

				// フラットモードでは全てのファイルが直接assetsディレクトリに配置されているはず
				for _, file := range files {
					if file.IsDir() {
						t.Errorf("フラットモードなのにサブディレクトリが存在します: %s", file.Name())
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 相対化処理の実行
			if err := Relativize(tt.config); err != nil {
				t.Fatalf("相対化処理に失敗: %v", err)
			}

			// 出力ファイルのパスを構築
			baseName := filepath.Base(tt.config.InputPath)
			outputName := baseName[:len(baseName)-len(filepath.Ext(baseName))] + ".ymmpr"
			outputPath := filepath.Join(tt.config.OutputDir, outputName)
			assetsDir := filepath.Join(tt.config.OutputDir, tt.config.AssetsDir)

			// 出力の検証
			tt.checkOutput(t, outputPath, assetsDir)

			// 絶対化のテスト
			absolutizeConfig := AbsolutizerConfig{
				InputPath:   outputPath,
				OutputDir:   filepath.Join(tt.config.OutputDir, "abs"),
				SkipMissing: true,
			}

			// 絶対化処理の実行
			if err := Absolutize(absolutizeConfig); err != nil {
				t.Fatalf("絶対化処理に失敗: %v", err)
			}

			// 絶対化された出力ファイルのパスを構築
			absOutputName := outputName[:len(outputName)-len(filepath.Ext(outputName))] + ".ymmp"
			absOutputPath := filepath.Join(absolutizeConfig.OutputDir, absOutputName)

			// 絶対化された出力ファイルの存在確認
			if _, err := os.Stat(absOutputPath); os.IsNotExist(err) {
				t.Errorf("絶対化された出力ファイルが作成されていません: %s", absOutputPath)
			}

			// JSONとして読み込めることを確認
			data, err := os.ReadFile(absOutputPath)
			if err != nil {
				t.Errorf("絶対化された出力ファイルの読み込みに失敗: %v", err)
				return
			}

			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Errorf("絶対化された出力ファイルのJSONパースに失敗: %v", err)
				return
			}

			// ルートのFilePathが絶対パスになっていることを確認
			if filePath, ok := result["FilePath"].(string); !ok {
				t.Error("ルートのFilePathが文字列ではありません")
			} else if !filepath.IsAbs(filePath) {
				t.Errorf("ルートのFilePathが絶対パスになっていません: %s", filePath)
			} else {
				// 出力ファイルの絶対パスと一致することを確認
				expectedPath, err := filepath.Abs(absOutputPath)
				if err != nil {
					t.Errorf("期待するパスの絶対パス化に失敗: %v", err)
				} else if filePath != expectedPath {
					t.Errorf("ルートのFilePathが期待する値と異なります。\n期待: %s\n実際: %s", expectedPath, filePath)
				}
			}
		})
	}
} 
