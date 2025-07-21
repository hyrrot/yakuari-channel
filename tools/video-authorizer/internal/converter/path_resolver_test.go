package converter_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/yakuari-channel/video-authorizer/internal/converter"
)

func TestPathResolver(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		input    string
		checkFn  func(string) bool
	}{
		{
			name:     "relative path resolved",
			basePath: "/project/base",
			input:    "assets/video.mp4",
			checkFn: func(result string) bool {
				expected := filepath.Join("/project/base", "assets/video.mp4")
				return result == expected
			},
		},
		{
			name:     "empty base path returns input",
			basePath: "",
			input:    "assets/video.mp4",
			checkFn: func(result string) bool {
				return result == "assets/video.mp4"
			},
		},
	}

	// Platform-specific tests
	if runtime.GOOS == "windows" {
		tests = append(tests, []struct {
			name     string
			basePath string
			input    string
			checkFn  func(string) bool
		}{
			{
				name:     "windows absolute path unchanged",
				basePath: "C:\\project\\base",
				input:    "D:\\media\\video.mp4",
				checkFn: func(result string) bool {
					return result == "D:\\media\\video.mp4"
				},
			},
			{
				name:     "windows relative path resolved",
				basePath: "C:\\project\\base",
				input:    "assets\\video.mp4",
				checkFn: func(result string) bool {
					expected := filepath.Join("C:\\project\\base", "assets\\video.mp4")
					return result == expected
				},
			},
		}...)
	} else {
		tests = append(tests, []struct {
			name     string
			basePath string
			input    string
			checkFn  func(string) bool
		}{
			{
				name:     "unix absolute path unchanged",
				basePath: "/project/base",
				input:    "/absolute/path/file.mp4",
				checkFn: func(result string) bool {
					return result == "/absolute/path/file.mp4"
				},
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := converter.NewPathResolver(tt.basePath)
			result := resolver.ResolvePath(tt.input)
			if !tt.checkFn(result) {
				t.Errorf("unexpected result: %s", result)
			}
		})
	}
}

func TestResolveItemProperties(t *testing.T) {
	basePath := filepath.Join("project", "base")
	resolver := converter.NewPathResolver(basePath)

	properties := map[string]interface{}{
		"FilePath": "assets/video.mp4",
		"Text":     "This is not a path",
		"ImageSettings": map[string]interface{}{
			"ImagePath": "images/background.png",
			"Width":     1920,
			"Height":    1080,
		},
		"PlayList": []interface{}{
			"sounds/bgm1.mp3",
			"sounds/bgm2.mp3",
		},
	}

	resolved := resolver.ResolveItemProperties(properties)

	// Check FilePath was resolved
	expectedFilePath := filepath.Join(basePath, "assets", "video.mp4")
	if resolved["FilePath"] != expectedFilePath {
		t.Errorf("FilePath not resolved correctly: expected %s, got %v", expectedFilePath, resolved["FilePath"])
	}

	// Check Text was not modified
	if resolved["Text"] != "This is not a path" {
		t.Errorf("Text was incorrectly modified: %v", resolved["Text"])
	}

	// Check nested ImagePath was resolved
	imageSettings := resolved["ImageSettings"].(map[string]interface{})
	expectedImagePath := filepath.Join(basePath, "images", "background.png")
	if imageSettings["ImagePath"] != expectedImagePath {
		t.Errorf("Nested ImagePath not resolved correctly: expected %s, got %v", expectedImagePath, imageSettings["ImagePath"])
	}

	// Check array items were resolved
	playList := resolved["PlayList"].([]interface{})
	expectedBGM1 := filepath.Join(basePath, "sounds", "bgm1.mp3")
	if playList[0] != expectedBGM1 {
		t.Errorf("Array item not resolved correctly: expected %s, got %v", expectedBGM1, playList[0])
	}
}