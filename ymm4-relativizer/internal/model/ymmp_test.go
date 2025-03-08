package model_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	
	"github.com/hyrrot/ymm4-relativizer/internal/model"
)

// まず構造体を定義
type Item struct {
	FilePath  *string `json:"FilePath"`
	SubItems  []Item  `json:"SubItems,omitempty"`
}

type YMMP struct {
	FilePath *string `json:"FilePath"`
	Items    []Item  `json:"Items,omitempty"`
}

func TestParseYMMP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *YMMP
		wantErr  bool
	}{
		{
			name: "Basic YMMP file",
			input: `{
				"FilePath": "C:\\test\\file.ymmp",
				"Content": {
					"FilePath": "C:\\test\\asset.png"
				}
			}`,
			expected: &YMMP{
				RootFilePath: "C:\\test\\file.ymmp",
				Content: map[string]interface{}{
					"Content": map[string]interface{}{
						"FilePath": "C:\\test\\asset.png",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "YMMP file with null FilePath",
			input: `{
				"FilePath": null,
				"Content": {
					"FilePath": "asset.png"
				}
			}`,
			expected: &YMMP{
				RootFilePath: nil,
				Content: map[string]interface{}{
					"Content": map[string]interface{}{
						"FilePath": "asset.png",
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "Invalid JSON",
			input:    `{"FilePath": }`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := model.ParseYMMP([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseYMMP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseYMMP() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUpdateFilePaths(t *testing.T) {
	tests := []struct {
		name     string
		ymmp     *YMMP
		updateFn func(string, bool) string
		expected *YMMP
	}{
		{
			name: "Update all paths",
			ymmp: &YMMP{
				RootFilePath: "C:\\test\\file.ymmp",
				Content: map[string]interface{}{
					"Content": map[string]interface{}{
						"FilePath": "C:\\test\\asset.png",
					},
				},
			},
			updateFn: func(path string, isRoot bool) string {
				if isRoot {
					return ""
				}
				return "assets/" + filepath.Base(path)
			},
			expected: &YMMP{
				RootFilePath: nil,
				Content: map[string]interface{}{
					"Content": map[string]interface{}{
						"FilePath": "assets/asset.png",
					},
				},
			},
		},
		{
			name: "Handle null FilePath",
			ymmp: &YMMP{
				RootFilePath: nil,
				Content: map[string]interface{}{
					"Content": map[string]interface{}{
						"FilePath": nil,
					},
				},
			},
			updateFn: func(path string, isRoot bool) string {
				return "new/path"
			},
			expected: &YMMP{
				RootFilePath: "new/path",
				Content: map[string]interface{}{
					"Content": map[string]interface{}{
						"FilePath": "new/path",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ymmp.UpdateFilePaths(tt.updateFn)
			if !reflect.DeepEqual(tt.ymmp, tt.expected) {
				t.Errorf("UpdateFilePaths() result = %v, want %v", tt.ymmp, tt.expected)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	tests := []struct {
		name     string
		ymmp     *YMMP
		expected string
		wantErr  bool
	}{
		{
			name: "Basic YMMP to JSON",
			ymmp: &YMMP{
				RootFilePath: "test.ymmp",
				Content: map[string]interface{}{
					"Content": map[string]interface{}{
						"FilePath": "asset.png",
					},
				},
			},
			expected: `{
  "Content": {
    "FilePath": "asset.png"
  },
  "FilePath": "test.ymmp"
}`,
			wantErr: false,
		},
		{
			name: "YMMP with null FilePath to JSON",
			ymmp: &YMMP{
				RootFilePath: nil,
				Content: map[string]interface{}{
					"Content": map[string]interface{}{
						"FilePath": "asset.png",
					},
				},
			},
			expected: `{
  "Content": {
    "FilePath": "asset.png"
  },
  "FilePath": null
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.ymmp.ToJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Compare JSON after normalizing whitespace
				var expected, actual interface{}
				if err := json.Unmarshal([]byte(tt.expected), &expected); err != nil {
					t.Fatalf("Failed to parse expected JSON: %v", err)
				}
				if err := json.Unmarshal(result, &actual); err != nil {
					t.Fatalf("Failed to parse actual JSON: %v", err)
				}
				if !reflect.DeepEqual(actual, expected) {
					t.Errorf("ToJSON() = %v, want %v", string(result), tt.expected)
				}
			}
		})
	}
}

// generateDeepNestedJSONはテスト用の深いネストを持つJSONを生成します
func generateDeepNestedJSON(depth int) string {
	var sb strings.Builder
	sb.WriteString(`{"FilePath": "root.wav", "Items": [`)
	
	for i := 0; i < depth; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		fmt.Fprintf(&sb, `{"FilePath": "level_%d.wav"}`, i)
	}
	
	sb.WriteString("]}")
	return sb.String()
}

func TestParseYMMPBasic(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		validate    func(*testing.T, *model.YMMP)
	}{
		{
			name: "空のJSONオブジェクト",
			input: `{}`,
			expectError: false,
			validate: func(t *testing.T, ymmp *model.YMMP) {
				output, err := ymmp.ToJSON()
				if err != nil {
					t.Errorf("ToJSONが失敗しました: %v", err)
					return
				}
				
				var data map[string]interface{}
				if err := json.Unmarshal(output, &data); err != nil {
					t.Errorf("出力のJSONパースに失敗しました: %v", err)
					return
				}
				
				filePath, exists := data["FilePath"]
				if !exists {
					t.Error("FilePathフィールドが存在しません")
				}
				if filePath != nil {
					t.Error("FilePathフィールドはnullであるべきです")
				}
			},
		},
		{
			name: "基本的なYMMPオブジェクト",
			input: `{
				"FilePath": "test.wav",
				"Items": [
					{
						"FilePath": "audio1.wav"
					},
					{
						"FilePath": "audio2.wav"
					}
				]
			}`,
			expectError: false,
			validate: func(t *testing.T, ymmp *model.YMMP) {
				paths := ymmp.FindAllFilePaths()
				expectedPaths := map[string]bool{
					"test.wav": false,
					"audio1.wav": false,
					"audio2.wav": false,
				}
				
				for _, p := range paths {
					if p.IsRoot && p.Path == "test.wav" {
						expectedPaths["test.wav"] = true
					}
					if !p.IsRoot {
						expectedPaths[p.Path] = true
					}
				}
				
				for path, found := range expectedPaths {
					if !found {
						t.Errorf("パス '%s' が見つかりません", path)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ymmp, err := model.ParseYMMP([]byte(tt.input))
			if tt.expectError {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				}
				return
			}
			if err != nil {
				t.Errorf("予期せぬエラー: %v", err)
				return
			}
			if tt.validate != nil {
				tt.validate(t, ymmp)
			}
		})
	}
}

func TestParseYMMPEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		validate    func(*testing.T, *model.YMMP)
	}{
		{
			name: "不正なJSON形式",
			input: `{
				"FilePath": "test.wav",
				"Items": [
					{`,
			expectError: true,
		},
		{
			name: "大規模なネスト構造を持つJSON",
			input: generateDeepNestedJSON(100),
			expectError: false,
			validate: func(t *testing.T, ymmp *model.YMMP) {
				paths := ymmp.FindAllFilePaths()
				if len(paths) != 101 {
					t.Errorf("期待されるファイルパス数は101ですが、%d個見つかりました", len(paths))
				}
				
				foundRoot := false
				for _, p := range paths {
					if p.IsRoot && p.Path == "root.wav" {
						foundRoot = true
						break
					}
				}
				if !foundRoot {
					t.Error("ルートのFilePath 'root.wav'が見つかりません")
				}
			},
		},
		{
			name: "nullのFilePathを含むオブジェクト",
			input: `{
				"FilePath": null,
				"Items": [
					{
						"FilePath": null
					},
					{
						"FilePath": "audio.wav"
					}
				]
			}`,
			expectError: false,
			validate: func(t *testing.T, ymmp *model.YMMP) {
				paths := ymmp.FindAllFilePaths()
				foundAudio := false
				
				for _, p := range paths {
					if !p.IsRoot && p.Path == "audio.wav" {
						foundAudio = true
					}
				}
				
				if !foundAudio {
					t.Error("ネストされたFilePath 'audio.wav'が見つかりません")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ymmp, err := model.ParseYMMP([]byte(tt.input))
			if tt.expectError {
				if err == nil {
					t.Error("エラーが期待されましたが、nilが返されました")
				}
				return
			}
			if err != nil {
				t.Errorf("予期せぬエラー: %v", err)
				return
			}
			if tt.validate != nil {
				tt.validate(t, ymmp)
			}
		})
	}
} 