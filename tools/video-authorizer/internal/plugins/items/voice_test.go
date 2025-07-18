package items

import (
	"testing"
	
	"github.com/user/ymmp-compiler/internal/models"
)

func TestVoicePluginGetType(t *testing.T) {
	// Red: This test will fail because VoicePlugin doesn't exist yet
	plugin := NewVoicePlugin()
	
	if plugin.GetType() != "voice" {
		t.Errorf("Expected type 'voice', got %s", plugin.GetType())
	}
}

func TestVoicePluginParseYAML(t *testing.T) {
	// Red: This test will fail because VoicePlugin doesn't exist yet
	plugin := NewVoicePlugin()
	
	yamlData := map[string]interface{}{
		"line":           "こんにちは、今日はいい天気ですね。",
		"character_name": "ずんだもん",
		"voice_id":       1,
		"speed":          1.2,
		"pitch":          0.1,
		"volume":         0.8,
		"length":         "auto",
	}
	
	item, err := plugin.ParseYAML(yamlData)
	if err != nil {
		t.Errorf("ParseYAML failed: %v", err)
	}
	
	voiceItem, ok := item.(*VoiceItem)
	if !ok {
		t.Error("Expected item to be *VoiceItem")
	}
	
	if voiceItem.Line != "こんにちは、今日はいい天気ですね。" {
		t.Errorf("Expected line 'こんにちは、今日はいい天気ですね。', got %s", voiceItem.Line)
	}
	
	if voiceItem.CharacterName != "ずんだもん" {
		t.Errorf("Expected character_name 'ずんだもん', got %s", voiceItem.CharacterName)
	}
	
	if voiceItem.VoiceID != 1 {
		t.Errorf("Expected voice_id 1, got %d", voiceItem.VoiceID)
	}
	
	if voiceItem.Speed != 1.2 {
		t.Errorf("Expected speed 1.2, got %f", voiceItem.Speed)
	}
	
	if voiceItem.Pitch != 0.1 {
		t.Errorf("Expected pitch 0.1, got %f", voiceItem.Pitch)
	}
	
	if voiceItem.Volume != 0.8 {
		t.Errorf("Expected volume 0.8, got %f", voiceItem.Volume)
	}
	
	expectedLength := models.LengthSpec{Type: models.LengthTypeAuto}
	if voiceItem.GetLength().Type != expectedLength.Type {
		t.Errorf("Expected length type %s, got %s", expectedLength.Type, voiceItem.GetLength().Type)
	}
}

func TestVoicePluginParseYAMLMissingLine(t *testing.T) {
	// Red: This test will fail because VoicePlugin doesn't exist yet
	plugin := NewVoicePlugin()
	
	yamlData := map[string]interface{}{
		"character_name": "ずんだもん",
		"length":         "auto",
		// Missing line
	}
	
	_, err := plugin.ParseYAML(yamlData)
	if err == nil {
		t.Error("Expected error when line is missing")
	}
}

func TestVoicePluginParseYAMLWithDefaults(t *testing.T) {
	// Red: This test will fail because VoicePlugin doesn't exist yet
	plugin := NewVoicePlugin()
	
	yamlData := map[string]interface{}{
		"line":   "テストメッセージ",
		"length": 3.5,
	}
	
	item, err := plugin.ParseYAML(yamlData)
	if err != nil {
		t.Errorf("ParseYAML failed: %v", err)
	}
	
	voiceItem, ok := item.(*VoiceItem)
	if !ok {
		t.Error("Expected item to be *VoiceItem")
	}
	
	// Check default values
	if voiceItem.CharacterName != "" {
		t.Errorf("Expected empty character_name (default), got %s", voiceItem.CharacterName)
	}
	
	if voiceItem.VoiceID != 0 {
		t.Errorf("Expected voice_id 0 (default), got %d", voiceItem.VoiceID)
	}
	
	if voiceItem.Speed != 1.0 {
		t.Errorf("Expected speed 1.0 (default), got %f", voiceItem.Speed)
	}
	
	if voiceItem.Pitch != 0.0 {
		t.Errorf("Expected pitch 0.0 (default), got %f", voiceItem.Pitch)
	}
	
	if voiceItem.Volume != 1.0 {
		t.Errorf("Expected volume 1.0 (default), got %f", voiceItem.Volume)
	}
	
	expectedLength := models.LengthSpec{Type: models.LengthTypeNumeric, Value: 3.5}
	if voiceItem.GetLength().Type != expectedLength.Type || voiceItem.GetLength().Value != expectedLength.Value {
		t.Errorf("Expected length %+v, got %+v", expectedLength, voiceItem.GetLength())
	}
}

func TestVoicePluginConvertToYMMP(t *testing.T) {
	// Red: This test will fail because VoicePlugin doesn't exist yet
	plugin := NewVoicePlugin()
	
	voiceItem := &VoiceItem{
		Line:          "テスト音声",
		CharacterName: "ずんだもん",
		VoiceID:       1,
		Speed:         1.2,
		Pitch:         0.1,
		Volume:        0.8,
		length:        models.LengthSpec{Type: models.LengthTypeNumeric, Value: 3.0},
	}
	
	basePath := "/base/path"
	ymmItem, err := plugin.ConvertToYMMP(voiceItem, basePath)
	if err != nil {
		t.Errorf("ConvertToYMMP failed: %v", err)
	}
	
	if ymmItem.Type != "YMM.Parts.VoicePart" {
		t.Errorf("Expected type 'YMM.Parts.VoicePart', got %s", ymmItem.Type)
	}
	
	if ymmItem.Length != 180 { // 3.0 seconds * 60 FPS
		t.Errorf("Expected length 180, got %d", ymmItem.Length)
	}
	
	// Check extended properties
	if ymmItem.Extended["Line"] != "テスト音声" {
		t.Errorf("Expected Line 'テスト音声', got %v", ymmItem.Extended["Line"])
	}
	
	if ymmItem.Extended["CharacterName"] != "ずんだもん" {
		t.Errorf("Expected CharacterName 'ずんだもん', got %v", ymmItem.Extended["CharacterName"])
	}
	
	if ymmItem.Extended["VoiceID"] != 1 {
		t.Errorf("Expected VoiceID 1, got %v", ymmItem.Extended["VoiceID"])
	}
	
	if ymmItem.Extended["Speed"] != 1.2 {
		t.Errorf("Expected Speed 1.2, got %v", ymmItem.Extended["Speed"])
	}
}