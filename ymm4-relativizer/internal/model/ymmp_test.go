package model

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
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
			result, err := ParseYMMP([]byte(tt.input))
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

func TestParseYMMPEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		validate    func(*testing.T, *YMMP)
	}{
		{
			name: "Empty JSON object",
			input: `{}`,
			expectError: false,
			validate: func(t *testing.T, ymmp *YMMP) {
				if ymmp == nil {
					t.Error("expected non-nil YMMP object")
				}
			},
		},
		{
			name: "FilePath with null values in nested objects",
			input: `{
				"FilePath": "test.wav",
				"Items": [
					{
						"FilePath": null
					},
					{
						"FilePath": "audio.wav",
						"Items": [
							{
								"FilePath": null
							}
						]
					}
				]
			}`,
			expectError: false,
			validate: func(t *testing.T, ymmp *YMMP) {
				if ymmp.FilePath == nil || *ymmp.FilePath != "test.wav" {
					t.Error("expected root FilePath to be 'test.wav'")
				}
			},
		},
		{
			name: "Malformed JSON with unexpected end",
			input: `{
				"FilePath": "test.wav",
				"Items": [
					{`,
			expectError: true,
		},
		{
			name: "JSON with very large nested structure",
			input: generateDeepNestedJSON(100),
			expectError: false,
			validate: func(t *testing.T, ymmp *YMMP) {
				paths := ymmp.FindAllFilePaths()
				if len(paths) != 100 {
					t.Errorf("expected 100 file paths, got %d", len(paths))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ymmp, err := ParseYMMP([]byte(tt.input))
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tt.validate != nil {
				tt.validate(t, ymmp)
			}
		})
	}
}

// ヘルパー関数：深いネストのJSONを生成
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