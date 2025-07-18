package items

import (
	"strings"
	"testing"
	
	"github.com/user/ymmp-compiler/internal/models"
)

func TestVideoPluginGetType(t *testing.T) {
	// Red: This test will fail because VideoPlugin doesn't exist yet
	plugin := NewVideoPlugin()
	
	if plugin.GetType() != "video" {
		t.Errorf("Expected type 'video', got %s", plugin.GetType())
	}
}

func TestVideoPluginParseYAML(t *testing.T) {
	// Red: This test will fail because VideoPlugin doesn't exist yet
	plugin := NewVideoPlugin()
	
	yamlData := map[string]interface{}{
		"file_path": "./assets/clip.mp4",
		"x":         320,
		"y":         180,
		"z":         5,
		"zoom":      1.2,
		"rotation":  15.0,
		"alpha":     200,
		"length":    "auto",
	}
	
	item, err := plugin.ParseYAML(yamlData)
	if err != nil {
		t.Errorf("ParseYAML failed: %v", err)
	}
	
	videoItem, ok := item.(*VideoItem)
	if !ok {
		t.Error("Expected item to be *VideoItem")
	}
	
	if videoItem.FilePath != "./assets/clip.mp4" {
		t.Errorf("Expected file_path './assets/clip.mp4', got %s", videoItem.FilePath)
	}
	
	if videoItem.X != 320 {
		t.Errorf("Expected X=320, got %d", videoItem.X)
	}
	
	if videoItem.Y != 180 {
		t.Errorf("Expected Y=180, got %d", videoItem.Y)
	}
	
	if videoItem.Z != 5 {
		t.Errorf("Expected Z=5, got %d", videoItem.Z)
	}
	
	if videoItem.Zoom != 1.2 {
		t.Errorf("Expected Zoom=1.2, got %f", videoItem.Zoom)
	}
	
	if videoItem.Rotation != 15.0 {
		t.Errorf("Expected Rotation=15.0, got %f", videoItem.Rotation)
	}
	
	if videoItem.Alpha != 200 {
		t.Errorf("Expected Alpha=200, got %d", videoItem.Alpha)
	}
	
	expectedLength := models.LengthSpec{Type: models.LengthTypeAuto}
	if videoItem.GetLength().Type != expectedLength.Type {
		t.Errorf("Expected length type %s, got %s", expectedLength.Type, videoItem.GetLength().Type)
	}
}

func TestVideoPluginParseYAMLWithDefaults(t *testing.T) {
	// Red: This test will fail because VideoPlugin doesn't exist yet
	plugin := NewVideoPlugin()
	
	yamlData := map[string]interface{}{
		"file_path": "./assets/video.avi",
		"length":    25.0,
	}
	
	item, err := plugin.ParseYAML(yamlData)
	if err != nil {
		t.Errorf("ParseYAML failed: %v", err)
	}
	
	videoItem, ok := item.(*VideoItem)
	if !ok {
		t.Error("Expected item to be *VideoItem")
	}
	
	// Check default values
	if videoItem.X != 0 {
		t.Errorf("Expected X=0 (default), got %d", videoItem.X)
	}
	
	if videoItem.Y != 0 {
		t.Errorf("Expected Y=0 (default), got %d", videoItem.Y)
	}
	
	if videoItem.Z != 0 {
		t.Errorf("Expected Z=0 (default), got %d", videoItem.Z)
	}
	
	if videoItem.Zoom != 1.0 {
		t.Errorf("Expected Zoom=1.0 (default), got %f", videoItem.Zoom)
	}
	
	if videoItem.Rotation != 0.0 {
		t.Errorf("Expected Rotation=0.0 (default), got %f", videoItem.Rotation)
	}
	
	if videoItem.Alpha != 255 {
		t.Errorf("Expected Alpha=255 (default), got %d", videoItem.Alpha)
	}
}

func TestVideoPluginParseYAMLMissingFilePath(t *testing.T) {
	// Red: This test will fail because VideoPlugin doesn't exist yet
	plugin := NewVideoPlugin()
	
	yamlData := map[string]interface{}{
		"x":      100,
		"length": "auto",
		// Missing file_path
	}
	
	_, err := plugin.ParseYAML(yamlData)
	if err == nil {
		t.Error("Expected error when file_path is missing")
	}
}

func TestVideoPluginConvertToYMMP(t *testing.T) {
	// Red: This test will fail because VideoPlugin doesn't exist yet
	plugin := NewVideoPlugin()
	
	videoItem := &VideoItem{
		FilePath: "./assets/movie.mp4",
		X:        160,
		Y:        90,
		Z:        10,
		Zoom:     1.5,
		Rotation: 45.0,
		Alpha:    180,
		length:   models.LengthSpec{Type: models.LengthTypeNumeric, Value: 8.0},
	}
	
	basePath := "/base/path"
	ymmItem, err := plugin.ConvertToYMMP(videoItem, basePath)
	if err != nil {
		t.Errorf("ConvertToYMMP failed: %v", err)
	}
	
	if ymmItem.Type != "YMM.Parts.VideoPart" {
		t.Errorf("Expected type 'YMM.Parts.VideoPart', got %s", ymmItem.Type)
	}
	
	// Check that the file path is properly joined (platform-independent)
	if !strings.HasSuffix(ymmItem.FilePath, "assets/movie.mp4") && !strings.HasSuffix(ymmItem.FilePath, "assets\\movie.mp4") {
		t.Errorf("Expected file path to contain 'assets/movie.mp4', got %s", ymmItem.FilePath)
	}
	
	if ymmItem.Length != 480 { // 8.0 seconds * 60 FPS
		t.Errorf("Expected length 480, got %d", ymmItem.Length)
	}
	
	// Check extended properties
	if ymmItem.Extended["X"] != 160 {
		t.Errorf("Expected X=160, got %v", ymmItem.Extended["X"])
	}
	
	if ymmItem.Extended["Y"] != 90 {
		t.Errorf("Expected Y=90, got %v", ymmItem.Extended["Y"])
	}
	
	if ymmItem.Extended["Z"] != 10 {
		t.Errorf("Expected Z=10, got %v", ymmItem.Extended["Z"])
	}
	
	if ymmItem.Extended["Zoom"] != 1.5 {
		t.Errorf("Expected Zoom=1.5, got %v", ymmItem.Extended["Zoom"])
	}
	
	if ymmItem.Extended["Rotation"] != 45.0 {
		t.Errorf("Expected Rotation=45.0, got %v", ymmItem.Extended["Rotation"])
	}
	
	if ymmItem.Extended["Alpha"] != 180 {
		t.Errorf("Expected Alpha=180, got %v", ymmItem.Extended["Alpha"])
	}
}

func TestVideoPluginConvertToYMMPWithAutoLength(t *testing.T) {
	// Red: This test will fail because VideoPlugin doesn't exist yet
	plugin := NewVideoPlugin()
	
	videoItem := &VideoItem{
		FilePath:         "./assets/clip.mov",
		X:                0,
		Y:                0,
		Z:                0,
		Zoom:             1.0,
		Rotation:         0.0,
		Alpha:            255,
		length:           models.LengthSpec{Type: models.LengthTypeAuto},
		calculatedLength: 12.75, // Simulated calculated length
	}
	
	basePath := "/base/path"
	ymmItem, err := plugin.ConvertToYMMP(videoItem, basePath)
	if err != nil {
		t.Errorf("ConvertToYMMP failed: %v", err)
	}
	
	if ymmItem.Length != 765 { // 12.75 seconds * 60 FPS
		t.Errorf("Expected length 765, got %d", ymmItem.Length)
	}
}