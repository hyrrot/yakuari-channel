package model

import (
	"encoding/json"
)

// YMMP represents the basic structure of a YMMP file
type YMMP struct {
	RootFilePath interface{}            `json:"FilePath"` // string or null
	Content      map[string]interface{} `json:"-"`
}

// FilePathUpdate holds information about FilePath updates
type FilePathUpdate struct {
	Path     string
	IsRoot   bool
	Original string
}

// ParseYMMP generates a YMMP object from JSON data
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

// FindAllFilePaths recursively searches for all FilePath fields
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

// findPaths recursively searches for FilePath fields
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

// UpdateFilePaths updates all FilePath fields
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

// updatePathsRecursive recursively updates FilePath fields
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

// ToJSON converts the YMMP object to JSON format
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