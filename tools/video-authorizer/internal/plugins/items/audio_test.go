package items

import (
	"strings"
	"testing"
	
	"github.com/user/ymmp-compiler/internal/models"
)

func TestAudioPluginGetType(t *testing.T) {
	// Red: This test will fail because AudioPlugin doesn't exist yet
	plugin := NewAudioPlugin()
	
	if plugin.GetType() != "audio" {
		t.Errorf("Expected type 'audio', got %s", plugin.GetType())
	}
}

func TestAudioPluginParseYAML(t *testing.T) {
	// Red: This test will fail because AudioPlugin doesn't exist yet
	plugin := NewAudioPlugin()
	
	yamlData := map[string]interface{}{
		"file_path": "./assets/background.mp3",
		"volume":    0.7,
		"length":    "auto",
	}
	
	item, err := plugin.ParseYAML(yamlData)
	if err != nil {
		t.Errorf("ParseYAML failed: %v", err)
	}
	
	audioItem, ok := item.(*AudioItem)
	if !ok {
		t.Error("Expected item to be *AudioItem")
	}
	
	if audioItem.FilePath != "./assets/background.mp3" {
		t.Errorf("Expected file_path './assets/background.mp3', got %s", audioItem.FilePath)
	}
	
	if audioItem.Volume != 0.7 {
		t.Errorf("Expected volume 0.7, got %f", audioItem.Volume)
	}
	
	expectedLength := models.LengthSpec{Type: models.LengthTypeAuto}
	if audioItem.GetLength().Type != expectedLength.Type {
		t.Errorf("Expected length type %s, got %s", expectedLength.Type, audioItem.GetLength().Type)
	}
}

func TestAudioPluginParseYAMLWithNumericLength(t *testing.T) {
	// Red: This test will fail because AudioPlugin doesn't exist yet
	plugin := NewAudioPlugin()
	
	yamlData := map[string]interface{}{
		"file_path": "./assets/music.wav",
		"volume":    1.0,
		"length":    15.5,
	}
	
	item, err := plugin.ParseYAML(yamlData)
	if err != nil {
		t.Errorf("ParseYAML failed: %v", err)
	}
	
	audioItem, ok := item.(*AudioItem)
	if !ok {
		t.Error("Expected item to be *AudioItem")
	}
	
	expectedLength := models.LengthSpec{Type: models.LengthTypeNumeric, Value: 15.5}
	if audioItem.GetLength().Type != expectedLength.Type || audioItem.GetLength().Value != expectedLength.Value {
		t.Errorf("Expected length %+v, got %+v", expectedLength, audioItem.GetLength())
	}
}

func TestAudioPluginParseYAMLMissingFilePath(t *testing.T) {
	// Red: This test will fail because AudioPlugin doesn't exist yet
	plugin := NewAudioPlugin()
	
	yamlData := map[string]interface{}{
		"volume": 0.8,
		"length": "auto",
		// Missing file_path
	}
	
	_, err := plugin.ParseYAML(yamlData)
	if err == nil {
		t.Error("Expected error when file_path is missing")
	}
}

func TestAudioPluginParseYAMLWithDefaults(t *testing.T) {
	// Red: This test will fail because AudioPlugin doesn't exist yet
	plugin := NewAudioPlugin()
	
	yamlData := map[string]interface{}{
		"file_path": "./assets/sound.ogg",
		"length":    "auto",
		// Volume should default to 1.0
	}
	
	item, err := plugin.ParseYAML(yamlData)
	if err != nil {
		t.Errorf("ParseYAML failed: %v", err)
	}
	
	audioItem, ok := item.(*AudioItem)
	if !ok {
		t.Error("Expected item to be *AudioItem")
	}
	
	if audioItem.Volume != 1.0 {
		t.Errorf("Expected volume 1.0 (default), got %f", audioItem.Volume)
	}
}

func TestAudioPluginConvertToYMMP(t *testing.T) {
	// Red: This test will fail because AudioPlugin doesn't exist yet
	plugin := NewAudioPlugin()
	
	audioItem := &AudioItem{
		FilePath: "./assets/music.mp3",
		Volume:   0.8,
		length:   models.LengthSpec{Type: models.LengthTypeNumeric, Value: 10.0},
	}
	
	basePath := "/base/path"
	ymmItem, err := plugin.ConvertToYMMP(audioItem, basePath)
	if err != nil {
		t.Errorf("ConvertToYMMP failed: %v", err)
	}
	
	if ymmItem.Type != "YMM.Parts.AudioPart" {
		t.Errorf("Expected type 'YMM.Parts.AudioPart', got %s", ymmItem.Type)
	}
	
	// Check that the file path is properly joined (platform-independent)
	if !strings.HasSuffix(ymmItem.FilePath, "assets/music.mp3") && !strings.HasSuffix(ymmItem.FilePath, "assets\\music.mp3") {
		t.Errorf("Expected file path to contain 'assets/music.mp3', got %s", ymmItem.FilePath)
	}
	
	if ymmItem.Length != 600 { // 10.0 seconds * 60 FPS
		t.Errorf("Expected length 600, got %d", ymmItem.Length)
	}
	
	// Check extended properties
	if ymmItem.Extended["Volume"] != 0.8 {
		t.Errorf("Expected Volume=0.8, got %v", ymmItem.Extended["Volume"])
	}
}

func TestAudioPluginConvertToYMMPWithAutoLength(t *testing.T) {
	// Red: This test will fail because AudioPlugin doesn't exist yet
	plugin := NewAudioPlugin()
	
	audioItem := &AudioItem{
		FilePath:         "./assets/music.mp3",
		Volume:           0.5,
		length:           models.LengthSpec{Type: models.LengthTypeAuto},
		calculatedLength: 7.5, // Simulated calculated length
	}
	
	basePath := "/base/path"
	ymmItem, err := plugin.ConvertToYMMP(audioItem, basePath)
	if err != nil {
		t.Errorf("ConvertToYMMP failed: %v", err)
	}
	
	if ymmItem.Length != 450 { // 7.5 seconds * 60 FPS
		t.Errorf("Expected length 450, got %d", ymmItem.Length)
	}
}