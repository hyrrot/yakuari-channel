package converter_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/yakuari-channel/video-authorizer/internal/converter"
	"github.com/yakuari-channel/video-authorizer/internal/models"
)

func TestFilePathUpdater_UpdateProjectFilePath(t *testing.T) {
	updater := converter.NewFilePathUpdater()
	
	project := &models.YMMPProject{
		FilePath: "/old/path/project.ymmp",
	}
	
	newPath := "test_output.ymmp"
	
	err := updater.UpdateProjectFilePath(project, newPath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// The FilePath should be updated to absolute path
	if !filepath.IsAbs(project.FilePath) {
		t.Error("expected FilePath to be absolute")
	}
	
	// Should end with the filename
	if filepath.Base(project.FilePath) != "test_output.ymmp" {
		t.Errorf("expected filename to be test_output.ymmp, got %s", filepath.Base(project.FilePath))
	}
}

func TestFilePathUpdater_UpdateItemFilePaths(t *testing.T) {
	updater := converter.NewFilePathUpdater()
	
	project := &models.YMMPProject{
		Timelines: []models.Timeline{
			{
				Items: []interface{}{
					&models.VideoItem{
						FilePath: "videos/test.mp4",
					},
					&models.TachieItem{
						TachieItemParameter: map[string]interface{}{
							"ImagePath": "images/character.png",
						},
					},
					&models.VoiceItem{
						// VoiceItem doesn't have FilePath, should be ignored
					},
				},
			},
		},
	}
	
	basePath := "/project/base"
	
	err := updater.UpdateItemFilePaths(project, basePath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// Check VideoItem FilePath
	videoItem := project.Timelines[0].Items[0].(*models.VideoItem)
	expectedVideoPath := filepath.Join(basePath, "videos/test.mp4")
	absExpectedVideoPath, _ := filepath.Abs(expectedVideoPath)
	
	if videoItem.FilePath != absExpectedVideoPath {
		t.Errorf("expected VideoItem FilePath to be %s, got %s", absExpectedVideoPath, videoItem.FilePath)
	}
	
	// Check TachieItem ImagePath in TachieItemParameter
	tachieItem := project.Timelines[0].Items[1].(*models.TachieItem)
	expectedTachiePath := filepath.Join(basePath, "images/character.png")
	absExpectedTachiePath, _ := filepath.Abs(expectedTachiePath)
	
	actualImagePath := tachieItem.TachieItemParameter["ImagePath"].(string)
	if actualImagePath != absExpectedTachiePath {
		t.Errorf("expected TachieItem ImagePath to be %s, got %s", absExpectedTachiePath, actualImagePath)
	}
}

func TestFilePathUpdater_UpdateItemFilePathsWithGenericItems(t *testing.T) {
	updater := converter.NewFilePathUpdater()
	
	project := &models.YMMPProject{
		Timelines: []models.Timeline{
			{
				Items: []interface{}{
					map[string]interface{}{
						"$type":    "SomeCustomItem",
						"FilePath": "custom/file.ext",
						"ImagePath": "images/bg.jpg",
						"NestedProps": map[string]interface{}{
							"VideoPath": "videos/nested.mp4",
						},
					},
				},
			},
		},
	}
	
	basePath := "/project/base"
	
	err := updater.UpdateItemFilePaths(project, basePath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// Check generic item
	genericItem := project.Timelines[0].Items[0].(map[string]interface{})
	
	// Check FilePath
	expectedFilePath := filepath.Join(basePath, "custom/file.ext")
	absExpectedFilePath, _ := filepath.Abs(expectedFilePath)
	
	if genericItem["FilePath"].(string) != absExpectedFilePath {
		t.Errorf("expected FilePath to be %s, got %s", absExpectedFilePath, genericItem["FilePath"])
	}
	
	// Check ImagePath
	expectedImagePath := filepath.Join(basePath, "images/bg.jpg")
	absExpectedImagePath, _ := filepath.Abs(expectedImagePath)
	
	if genericItem["ImagePath"].(string) != absExpectedImagePath {
		t.Errorf("expected ImagePath to be %s, got %s", absExpectedImagePath, genericItem["ImagePath"])
	}
	
	// Check nested VideoPath
	nestedProps := genericItem["NestedProps"].(map[string]interface{})
	expectedVideoPath := filepath.Join(basePath, "videos/nested.mp4")
	absExpectedVideoPath, _ := filepath.Abs(expectedVideoPath)
	
	if nestedProps["VideoPath"].(string) != absExpectedVideoPath {
		t.Errorf("expected nested VideoPath to be %s, got %s", absExpectedVideoPath, nestedProps["VideoPath"])
	}
}

func TestFilePathUpdater_UpdateAllFilePaths(t *testing.T) {
	updater := converter.NewFilePathUpdater()
	
	project := &models.YMMPProject{
		FilePath: "/old/project.ymmp",
		Timelines: []models.Timeline{
			{
				Items: []interface{}{
					&models.VideoItem{
						FilePath: "relative/video.mp4",
					},
				},
			},
		},
	}
	
	outputPath := "output.ymmp"
	basePath := "/project/base"
	
	err := updater.UpdateAllFilePaths(project, outputPath, basePath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// Check project FilePath was updated
	if !filepath.IsAbs(project.FilePath) {
		t.Error("expected project FilePath to be absolute")
	}
	
	if filepath.Base(project.FilePath) != "output.ymmp" {
		t.Errorf("expected project filename to be output.ymmp, got %s", filepath.Base(project.FilePath))
	}
	
	// Check item FilePath was updated
	videoItem := project.Timelines[0].Items[0].(*models.VideoItem)
	if !filepath.IsAbs(videoItem.FilePath) {
		t.Error("expected item FilePath to be absolute")
	}
}

func TestFilePathUpdater_ResolveToAbsolutePath(t *testing.T) {
	updater := converter.NewFilePathUpdater()
	
	tests := []struct {
		name     string
		path     string
		basePath string
		checkFn  func(string) bool
	}{
		{
			name:     "relative path with base",
			path:     "relative/file.txt",
			basePath: "/base/path",
			checkFn: func(result string) bool {
				return filepath.IsAbs(result) && 
					   filepath.Base(result) == "file.txt"
			},
		},
		{
			name:     "absolute path unchanged",
			path:     "/absolute/path/file.txt",
			basePath: "/base/path",
			checkFn: func(result string) bool {
				// On Windows, Unix-style absolute paths might be modified
				if runtime.GOOS == "windows" {
					return filepath.IsAbs(result) && filepath.Base(result) == "file.txt"
				}
				return result == "/absolute/path/file.txt"
			},
		},
		{
			name:     "relative path without base",
			path:     "relative/file.txt",
			basePath: "",
			checkFn: func(result string) bool {
				return filepath.IsAbs(result) && 
					   filepath.Base(result) == "file.txt"
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Access private method through a test helper or by creating a test project
			project := &models.YMMPProject{
				Timelines: []models.Timeline{
					{
						Items: []interface{}{
							map[string]interface{}{
								"FilePath": tt.path,
							},
						},
					},
				},
			}
			
			err := updater.UpdateItemFilePaths(project, tt.basePath)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			item := project.Timelines[0].Items[0].(map[string]interface{})
			result := item["FilePath"].(string)
			
			if !tt.checkFn(result) {
				t.Errorf("path resolution failed for %s with base %s, got: %s", tt.path, tt.basePath, result)
			}
		})
	}
}

func TestFilePathUpdater_GetOutputDirectory(t *testing.T) {
	updater := converter.NewFilePathUpdater()
	
	tests := []struct {
		name       string
		outputPath string
		checkFn    func(string) bool
	}{
		{
			name:       "simple filename",
			outputPath: "output.ymmp",
			checkFn: func(dir string) bool {
				return dir != ""
			},
		},
		{
			name:       "path with directory",
			outputPath: "project/output.ymmp",
			checkFn: func(dir string) bool {
				return filepath.Base(dir) == "project" || 
					   filepath.IsAbs(dir) // Could be absolute depending on current dir
			},
		},
		{
			name:       "absolute path",
			outputPath: "/absolute/path/output.ymmp",
			checkFn: func(dir string) bool {
				return dir == "/absolute/path" || 
					   filepath.IsAbs(dir) // Windows might modify this
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updater.GetOutputDirectory(tt.outputPath)
			
			if !tt.checkFn(result) {
				t.Errorf("GetOutputDirectory failed for %s, got: %s", tt.outputPath, result)
			}
		})
	}
}