package items

import (
	"testing"
	
	"github.com/user/ymmp-compiler/internal/models"
)

func TestTachiePluginGetType(t *testing.T) {
	// Red: This test will fail because TachiePlugin doesn't exist yet
	plugin := NewTachiePlugin()
	
	if plugin.GetType() != "tachie" {
		t.Errorf("Expected type 'tachie', got %s", plugin.GetType())
	}
}

func TestTachiePluginParseYAML(t *testing.T) {
	// Red: This test will fail because TachiePlugin doesn't exist yet
	plugin := NewTachiePlugin()
	
	yamlData := map[string]interface{}{
		"item":     "ずんだもん",
		"x":        640,
		"y":        360,
		"z":        15,
		"zoom":     0.8,
		"rotation": 5.0,
		"alpha":    230,
		"fade_in":  0.5,
		"fade_out": 1.0,
		"length":   10.0,
	}
	
	item, err := plugin.ParseYAML(yamlData)
	if err != nil {
		t.Errorf("ParseYAML failed: %v", err)
	}
	
	tachieItem, ok := item.(*TachieItem)
	if !ok {
		t.Error("Expected item to be *TachieItem")
	}
	
	if tachieItem.Item != "ずんだもん" {
		t.Errorf("Expected item 'ずんだもん', got %s", tachieItem.Item)
	}
	
	if tachieItem.X != 640 {
		t.Errorf("Expected X=640, got %d", tachieItem.X)
	}
	
	if tachieItem.Y != 360 {
		t.Errorf("Expected Y=360, got %d", tachieItem.Y)
	}
	
	if tachieItem.Z != 15 {
		t.Errorf("Expected Z=15, got %d", tachieItem.Z)
	}
	
	if tachieItem.Zoom != 0.8 {
		t.Errorf("Expected Zoom=0.8, got %f", tachieItem.Zoom)
	}
	
	if tachieItem.Rotation != 5.0 {
		t.Errorf("Expected Rotation=5.0, got %f", tachieItem.Rotation)
	}
	
	if tachieItem.Alpha != 230 {
		t.Errorf("Expected Alpha=230, got %d", tachieItem.Alpha)
	}
	
	if tachieItem.FadeIn != 0.5 {
		t.Errorf("Expected FadeIn=0.5, got %f", tachieItem.FadeIn)
	}
	
	if tachieItem.FadeOut != 1.0 {
		t.Errorf("Expected FadeOut=1.0, got %f", tachieItem.FadeOut)
	}
	
	expectedLength := models.LengthSpec{Type: models.LengthTypeNumeric, Value: 10.0}
	if tachieItem.GetLength().Type != expectedLength.Type || tachieItem.GetLength().Value != expectedLength.Value {
		t.Errorf("Expected length %+v, got %+v", expectedLength, tachieItem.GetLength())
	}
}

func TestTachiePluginParseYAMLWithDefaults(t *testing.T) {
	// Red: This test will fail because TachiePlugin doesn't exist yet
	plugin := NewTachiePlugin()
	
	yamlData := map[string]interface{}{
		"item":   "きりたん",
		"length": 5.0,
	}
	
	item, err := plugin.ParseYAML(yamlData)
	if err != nil {
		t.Errorf("ParseYAML failed: %v", err)
	}
	
	tachieItem, ok := item.(*TachieItem)
	if !ok {
		t.Error("Expected item to be *TachieItem")
	}
	
	// Check default values
	if tachieItem.X != 0 {
		t.Errorf("Expected X=0 (default), got %d", tachieItem.X)
	}
	
	if tachieItem.Y != 0 {
		t.Errorf("Expected Y=0 (default), got %d", tachieItem.Y)
	}
	
	if tachieItem.Z != 0 {
		t.Errorf("Expected Z=0 (default), got %d", tachieItem.Z)
	}
	
	if tachieItem.Zoom != 1.0 {
		t.Errorf("Expected Zoom=1.0 (default), got %f", tachieItem.Zoom)
	}
	
	if tachieItem.Rotation != 0.0 {
		t.Errorf("Expected Rotation=0.0 (default), got %f", tachieItem.Rotation)
	}
	
	if tachieItem.Alpha != 255 {
		t.Errorf("Expected Alpha=255 (default), got %d", tachieItem.Alpha)
	}
	
	if tachieItem.FadeIn != 0.0 {
		t.Errorf("Expected FadeIn=0.0 (default), got %f", tachieItem.FadeIn)
	}
	
	if tachieItem.FadeOut != 0.0 {
		t.Errorf("Expected FadeOut=0.0 (default), got %f", tachieItem.FadeOut)
	}
}

func TestTachiePluginParseYAMLMissingItem(t *testing.T) {
	// Red: This test will fail because TachiePlugin doesn't exist yet
	plugin := NewTachiePlugin()
	
	yamlData := map[string]interface{}{
		"x":      100,
		"length": 5.0,
		// Missing item
	}
	
	_, err := plugin.ParseYAML(yamlData)
	if err == nil {
		t.Error("Expected error when item is missing")
	}
}

func TestTachiePluginConvertToYMMP(t *testing.T) {
	// Red: This test will fail because TachiePlugin doesn't exist yet
	plugin := NewTachiePlugin()
	
	tachieItem := &TachieItem{
		Item:     "ずんだもん",
		X:        320,
		Y:        240,
		Z:        5,
		Zoom:     1.2,
		Rotation: 10.0,
		Alpha:    200,
		FadeIn:   0.3,
		FadeOut:  0.7,
		length:   models.LengthSpec{Type: models.LengthTypeNumeric, Value: 6.0},
	}
	
	basePath := "/base/path"
	ymmItem, err := plugin.ConvertToYMMP(tachieItem, basePath)
	if err != nil {
		t.Errorf("ConvertToYMMP failed: %v", err)
	}
	
	if ymmItem.Type != "YMM.Parts.TachiePart" {
		t.Errorf("Expected type 'YMM.Parts.TachiePart', got %s", ymmItem.Type)
	}
	
	if ymmItem.FilePath != "" {
		t.Errorf("Expected empty FilePath for tachie, got %s", ymmItem.FilePath)
	}
	
	if ymmItem.Length != 360 { // 6.0 seconds * 60 FPS
		t.Errorf("Expected length 360, got %d", ymmItem.Length)
	}
	
	// Check extended properties
	if ymmItem.Extended["Item"] != "ずんだもん" {
		t.Errorf("Expected Item='ずんだもん', got %v", ymmItem.Extended["Item"])
	}
	
	if ymmItem.Extended["X"] != 320 {
		t.Errorf("Expected X=320, got %v", ymmItem.Extended["X"])
	}
	
	if ymmItem.Extended["Y"] != 240 {
		t.Errorf("Expected Y=240, got %v", ymmItem.Extended["Y"])
	}
	
	if ymmItem.Extended["Z"] != 5 {
		t.Errorf("Expected Z=5, got %v", ymmItem.Extended["Z"])
	}
	
	if ymmItem.Extended["Zoom"] != 1.2 {
		t.Errorf("Expected Zoom=1.2, got %v", ymmItem.Extended["Zoom"])
	}
	
	if ymmItem.Extended["Rotation"] != 10.0 {
		t.Errorf("Expected Rotation=10.0, got %v", ymmItem.Extended["Rotation"])
	}
	
	if ymmItem.Extended["Alpha"] != 200 {
		t.Errorf("Expected Alpha=200, got %v", ymmItem.Extended["Alpha"])
	}
	
	if ymmItem.Extended["FadeIn"] != 0.3 {
		t.Errorf("Expected FadeIn=0.3, got %v", ymmItem.Extended["FadeIn"])
	}
	
	if ymmItem.Extended["FadeOut"] != 0.7 {
		t.Errorf("Expected FadeOut=0.7, got %v", ymmItem.Extended["FadeOut"])
	}
}