package model

import (
	"encoding/json"
)

// YMMP はYMMPファイルの基本構造を表します
type YMMP struct {
	RootFilePath interface{}            `json:"FilePath"` // stringまたはnull
	Content      map[string]interface{} `json:"-"`
}

// FilePathUpdate は FilePath の更新情報を保持します
type FilePathUpdate struct {
	Path     string
	IsRoot   bool
	Original string
}

// ParseYMMP はJSONデータからYMMPオブジェクトを生成します
func ParseYMMP(data []byte) (*YMMP, error) {
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, err
	}

	ymmp := &YMMP{
		Content: make(map[string]interface{}),
	}

	// ルートのFilePathを取得
	if filePath, exists := rawData["FilePath"]; exists {
		ymmp.RootFilePath = filePath
		delete(rawData, "FilePath")
	}

	// その他のコンテンツをコピー
	for k, v := range rawData {
		ymmp.Content[k] = v
	}

	return ymmp, nil
}

// FindAllFilePaths は全てのFilePathフィールドを再帰的に検索します
func (y *YMMP) FindAllFilePaths() []FilePathUpdate {
	paths := []FilePathUpdate{}

	// ルートのFilePathを追加
	if str, ok := y.RootFilePath.(string); ok && str != "" {
		paths = append(paths, FilePathUpdate{
			Path:     str,
			IsRoot:   true,
			Original: str,
		})
	}

	// 他のFilePathを再帰的に検索
	findPaths(y.Content, &paths, false)

	return paths
}

// findPaths は再帰的にFilePathフィールドを検索します
func findPaths(data interface{}, paths *[]FilePathUpdate, isRoot bool) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if key == "FilePath" {
				if str, ok := value.(string); ok && str != "" {
					*paths = append(*paths, FilePathUpdate{
						Path:     str,
						IsRoot:   isRoot,
						Original: str,
					})
				}
			} else {
				findPaths(value, paths, false)
			}
		}
	case []interface{}:
		for _, item := range v {
			findPaths(item, paths, false)
		}
	}
}

// UpdateFilePaths は全てのFilePathフィールドを更新します
func (y *YMMP) UpdateFilePaths(updateFunc func(string, bool) string) {
	// ルートのFilePathを更新
	if str, ok := y.RootFilePath.(string); ok {
		newPath := updateFunc(str, true)
		if newPath == "" {
			y.RootFilePath = nil
		} else {
			y.RootFilePath = newPath
		}
	} else {
		// FilePathがnullまたは文字列以外の場合、新しいパスを生成
		newPath := updateFunc("", true)
		if newPath != "" {
			y.RootFilePath = newPath
		} else {
			y.RootFilePath = nil
		}
	}

	// 他のFilePathを再帰的に更新
	y.Content = updatePathsRecursive(y.Content, updateFunc).(map[string]interface{})
}

// updatePathsRecursive は再帰的にFilePathフィールドを更新します
func updatePathsRecursive(data interface{}, updateFunc func(string, bool) string) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			if key == "FilePath" {
				if str, ok := value.(string); ok {
					newPath := updateFunc(str, false)
					if newPath == "" {
						result[key] = nil
					} else {
						result[key] = newPath
					}
				} else {
					// FilePathがnullまたは文字列以外の場合
					newPath := updateFunc("", false)
					if newPath != "" {
						result[key] = newPath
					} else {
						result[key] = nil
					}
				}
			} else {
				result[key] = updatePathsRecursive(value, updateFunc)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = updatePathsRecursive(item, updateFunc)
		}
		return result
	default:
		return v
	}
}

// ToJSON はYMMPオブジェクトをJSON形式に変換します
func (y *YMMP) ToJSON() ([]byte, error) {
	output := make(map[string]interface{})
	
	// コンテンツをコピー
	for k, v := range y.Content {
		output[k] = v
	}
	
	// ルートのFilePathを設定
	output["FilePath"] = y.RootFilePath

	return json.MarshalIndent(output, "", "  ")
} 
} 