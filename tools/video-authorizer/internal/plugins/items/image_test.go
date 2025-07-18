package items

import (
	"strings"
	"testing"
	
	"github.com/user/ymmp-compiler/internal/models"
)

func TestImagePluginGetType(t *testing.T) {
	// Red: This test will fail because ImagePlugin doesn't exist yet
	plugin := NewImagePlugin()
	
	if plugin.GetType() != "image" {
		t.Errorf("Expected type 'image', got %s", plugin.GetType())
	}
}

func TestImagePluginParseYAML(t *testing.T) {
	// Red: This test will fail because ImagePlugin doesn't exist yet
	plugin := NewImagePlugin()
	
	yamlData := map[string]interface{}{
		"file_path": "./assets/image.png",
		"x":         160,
		"y":         90,
		"zoom":      1.5,
		"length":    5.0,
	}
	
	item, err := plugin.ParseYAML(yamlData)
	if err != nil {
		t.Errorf("ParseYAML failed: %v", err)
	}
	
	imageItem, ok := item.(*ImageItem)
	if !ok {
		t.Error("Expected item to be *ImageItem")
	}
	
	if imageItem.FilePath != "./assets/image.png" {
		t.Errorf("Expected file_path './assets/image.png', got %s", imageItem.FilePath)
	}
	
	if imageItem.X != 160 {
		t.Errorf("Expected X=160, got %d", imageItem.X)
	}
	
	if imageItem.Y != 90 {
		t.Errorf("Expected Y=90, got %d", imageItem.Y)
	}
	
	if imageItem.Zoom != 1.5 {
		t.Errorf("Expected Zoom=1.5, got %f", imageItem.Zoom)
	}
	
	expectedLength := models.LengthSpec{Type: models.LengthTypeNumeric, Value: 5.0}
	if imageItem.GetLength().Type != expectedLength.Type || imageItem.GetLength().Value != expectedLength.Value {
		t.Errorf("Expected length %+v, got %+v", expectedLength, imageItem.GetLength())
	}
}

func TestImagePluginParseYAMLMissingFilePath(t *testing.T) {
	// Red: This test will fail because ImagePlugin doesn't exist yet
	plugin := NewImagePlugin()
	
	yamlData := map[string]interface{}{
		"x":      160,
		"length": 5.0,
		// Missing file_path
	}
	
	_, err := plugin.ParseYAML(yamlData)
	if err == nil {
		t.Error("Expected error when file_path is missing")
	}
}

func TestImagePluginConvertToYMMP(t *testing.T) {
	// Red: This test will fail because ImagePlugin doesn't exist yet
	plugin := NewImagePlugin()
	
	imageItem := &ImageItem{
		FilePath: "./assets/image.png",
		X:        160,
		Y:        90,
		Z:        10,
		Zoom:     1.5,
		Rotation: 0,
		Alpha:    255,
		length:   models.LengthSpec{Type: models.LengthTypeNumeric, Value: 5.0},
	}
	
	basePath := "/base/path"
	ymmItem, err := plugin.ConvertToYMMP(imageItem, basePath)
	if err != nil {
		t.Errorf("ConvertToYMMP failed: %v", err)
	}
	
	if ymmItem.Type != "YMM.Parts.ImagePart" {
		t.Errorf("Expected type 'YMM.Parts.ImagePart', got %s", ymmItem.Type)
	}
	
	// Check that the file path is properly joined (platform-independent)
	if !strings.HasSuffix(ymmItem.FilePath, "assets/image.png") && !strings.HasSuffix(ymmItem.FilePath, "assets\\image.png") {
		t.Errorf("Expected file path to contain 'assets/image.png', got %s", ymmItem.FilePath)
	}
	
	if ymmItem.Length != 300 { // 5.0 seconds * 60 FPS
		t.Errorf("Expected length 300, got %d", ymmItem.Length)
	}
	
	// Check extended properties
	if ymmItem.Extended["X"] != 160 {
		t.Errorf("Expected X=160, got %v", ymmItem.Extended["X"])
	}
	
	if ymmItem.Extended["Y"] != 90 {
		t.Errorf("Expected Y=90, got %v", ymmItem.Extended["Y"])
	}
	
	if ymmItem.Extended["Zoom"] != 1.5 {
		t.Errorf("Expected Zoom=1.5, got %v", ymmItem.Extended["Zoom"])
	}
}