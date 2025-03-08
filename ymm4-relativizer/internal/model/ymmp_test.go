package model_test

import (
	"encoding/json"
	"testing"
	
	"github.com/hyrrot/ymm4-relativizer/internal/model"
)

func TestParseYMMP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(*testing.T, *model.YMMP)
	}{
		{
			name: "基本的なYMMPファイル",
			input: `{
				"FilePath": "test.wav",
				"Content": {
					"FilePath": "asset.png"
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, ymmp *model.YMMP) {
				paths := ymmp.FindAllFilePaths()
				foundRoot := false
				foundAsset := false
				
				for _, p := range paths {
					if p.IsRoot && p.Path == "test.wav" {
						foundRoot = true
					}
					if !p.IsRoot && p.Path == "asset.png" {
						foundAsset = true
					}
				}
				
				if !foundRoot {
					t.Error("ルートのFilePathが見つかりません")
				}
				if !foundAsset {
					t.Error("アセットのFilePathが見つかりません")
				}
			},
		},
		{
			name: "nullのFilePath",
			input: `{
				"FilePath": null,
				"Content": {
					"FilePath": "asset.png"
				}
			}`,
			wantErr: false,
			validate: func(t *testing.T, ymmp *model.YMMP) {
				paths := ymmp.FindAllFilePaths()
				foundAsset := false
				
				for _, p := range paths {
					if !p.IsRoot && p.Path == "asset.png" {
						foundAsset = true
					}
				}
				
				if !foundAsset {
					t.Error("アセットのFilePathが見つかりません")
				}
			},
		},
		{
			name: "大規模なネスト構造",
			input: `{
				"FilePath": "root.wav",
				"Items": [
					{
						"FilePath": "item1.wav",
						"Items": [
							{
								"FilePath": "item2.wav"
							}
						]
					},
					{
						"FilePath": "item3.wav"
					}
				]
			}`,
			wantErr: false,
			validate: func(t *testing.T, ymmp *model.YMMP) {
				paths := ymmp.FindAllFilePaths()
				expected := map[string]bool{
					"root.wav": false,
					"item1.wav": false,
					"item2.wav": false,
					"item3.wav": false,
				}
				
				for _, p := range paths {
					if p.IsRoot && p.Path == "root.wav" {
						expected["root.wav"] = true
					}
					if !p.IsRoot {
						if _, ok := expected[p.Path]; ok {
							expected[p.Path] = true
						}
					}
				}
				
				for path, found := range expected {
					if !found {
						t.Errorf("パス '%s' が見つかりません", path)
					}
				}
			},
		},
		{
			name:     "不正なJSON",
			input:    `{"FilePath": }`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ymmp, err := model.ParseYMMP([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseYMMP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, ymmp)
			}
		})
	}
}

func TestUpdateFilePaths(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		updateFn func(string, bool) string
		validate func(*testing.T, *model.YMMP)
	}{
		{
			name: "全てのパスを更新",
			input: `{
				"FilePath": "test.wav",
				"Content": {
					"FilePath": "asset.png"
				}
			}`,
			updateFn: func(path string, isRoot bool) string {
				if isRoot {
					return ""
				}
				return "assets/" + path
			},
			validate: func(t *testing.T, ymmp *model.YMMP) {
				paths := ymmp.FindAllFilePaths()
				foundAsset := false
				
				for _, p := range paths {
					if !p.IsRoot && p.Path == "assets/asset.png" {
						foundAsset = true
					}
				}
				
				if !foundAsset {
					t.Error("更新後のアセットパスが見つかりません")
				}
				
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
				
				if data["FilePath"] != nil {
					t.Error("ルートのFilePathはnullであるべきです")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ymmp, err := model.ParseYMMP([]byte(tt.input))
			if err != nil {
				t.Fatalf("ParseYMMPが失敗しました: %v", err)
			}
			
			ymmp.UpdateFilePaths(tt.updateFn)
			
			if tt.validate != nil {
				tt.validate(t, ymmp)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name: "基本的なYMMPのJSON変換",
			input: `{
				"FilePath": "test.ymmp",
				"Content": {
					"FilePath": "asset.png"
				}
			}`,
			expected: `{
				"Content": {
					"FilePath": "asset.png"
				},
				"FilePath": "test.ymmp"
			}`,
			wantErr: false,
		},
		{
			name: "nullのFilePathを持つYMMPのJSON変換",
			input: `{
				"FilePath": null,
				"Content": {
					"FilePath": "asset.png"
				}
			}`,
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
			ymmp, err := model.ParseYMMP([]byte(tt.input))
			if err != nil {
				t.Fatalf("ParseYMMPが失敗しました: %v", err)
			}
			
			result, err := ymmp.ToJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				var expectedData, actualData map[string]interface{}
				if err := json.Unmarshal([]byte(tt.expected), &expectedData); err != nil {
					t.Fatalf("期待値のJSONパースに失敗: %v", err)
				}
				if err := json.Unmarshal(result, &actualData); err != nil {
					t.Fatalf("実際の出力のJSONパースに失敗: %v", err)
				}
				
				if expectedData["FilePath"] != actualData["FilePath"] {
					t.Errorf("FilePathが一致しません: got %v, want %v", 
						actualData["FilePath"], expectedData["FilePath"])
				}
				
				expectedContent := expectedData["Content"].(map[string]interface{})
				actualContent := actualData["Content"].(map[string]interface{})
				if expectedContent["FilePath"] != actualContent["FilePath"] {
					t.Errorf("Content.FilePathが一致しません: got %v, want %v",
						actualContent["FilePath"], expectedContent["FilePath"])
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
		validate    func(*testing.T, *model.YMMP)
	}{
		{
			name: "空のJSONオブジェクト",
			input: `{}`,
			expectError: false,
			validate: func(t *testing.T, ymmp *model.YMMP) {
				if ymmp == nil {
					t.Error("YMMPオブジェクトがnilです")
				}
			},
		},
		{
			name: "大規模なネスト構造",
			input: `{
				"FilePath": "root.wav",
				"Items": [
					{
						"FilePath": "item1.wav",
						"Items": [
							{
								"FilePath": "item2.wav"
							},
							{
								"FilePath": "item3.wav"
							}
						]
					},
					{
						"FilePath": "item4.wav",
						"Items": [
							{
								"FilePath": "item5.wav"
							}
						]
					}
				]
			}`,
			expectError: false,
			validate: func(t *testing.T, ymmp *model.YMMP) {
				paths := ymmp.FindAllFilePaths()
				expected := map[string]bool{
					"root.wav": false,
					"item1.wav": false,
					"item2.wav": false,
					"item3.wav": false,
					"item4.wav": false,
					"item5.wav": false,
				}
				
				for _, p := range paths {
					if p.IsRoot && p.Path == "root.wav" {
						expected["root.wav"] = true
					}
					if !p.IsRoot {
						if _, ok := expected[p.Path]; ok {
							expected[p.Path] = true
						}
					}
				}
				
				for path, found := range expected {
					if !found {
						t.Errorf("パス '%s' が見つかりません", path)
					}
				}
			},
		},
		{
			name: "不正なJSON形式",
			input: `{
				"FilePath": "test.wav",
				"Items": [
					{`,
			expectError: true,
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