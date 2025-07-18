package parser

import (
	"strings"
	"testing"
	
	"github.com/user/ymmp-compiler/internal/plugins"
	"github.com/user/ymmp-compiler/internal/plugins/items"
)

func TestParseSYMMPWithDynamicItems(t *testing.T) {
	// Red: This test will fail because ParseSYMMPWithPlugins doesn't exist yet
	yamlContent := `
symmp_format_version: 0.1
project:
  name: "Dynamic Test Project"
sequences:
  - id: seq1
    scenes:
      - id: scene1
        shots:
          - image:
              file_path: "./test.png"
              length: 5.0
            voice:
              line: "こんにちは"
              length: auto
`
	
	// Create plugin registry
	registry := plugins.NewPluginRegistry()
	registry.RegisterItemPlugin(items.NewImagePlugin())
	registry.RegisterItemPlugin(items.NewVoicePlugin())
	
	reader := strings.NewReader(yamlContent)
	episode, err := ParseSYMMPWithPlugins(reader, registry)
	
	if err != nil {
		t.Errorf("ParseSYMMPWithPlugins failed: %v", err)
	}
	
	if len(episode.Sequences) != 1 {
		t.Errorf("Expected 1 sequence, got %d", len(episode.Sequences))
	}
	
	sequence := episode.Sequences[0]
	if len(sequence.Scenes) != 1 {
		t.Errorf("Expected 1 scene, got %d", len(sequence.Scenes))
	}
	
	scene := sequence.Scenes[0]
	if len(scene.Shots) != 1 {
		t.Errorf("Expected 1 shot, got %d", len(scene.Shots))
	}
	
	shot := scene.Shots[0]
	if len(shot.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(shot.Items))
	}
	
	// Check first item (image)
	imageItem, ok := shot.Items[0].(*items.ImageItem)
	if !ok {
		t.Error("Expected first item to be *ImageItem")
	}
	
	if imageItem.FilePath != "./test.png" {
		t.Errorf("Expected image file path './test.png', got %s", imageItem.FilePath)
	}
	
	if imageItem.GetLength().Value != 5.0 {
		t.Errorf("Expected image length 5.0, got %f", imageItem.GetLength().Value)
	}
	
	// Check second item (voice)
	voiceItem, ok := shot.Items[1].(*items.VoiceItem)
	if !ok {
		t.Error("Expected second item to be *VoiceItem")
	}
	
	if voiceItem.Line != "こんにちは" {
		t.Errorf("Expected voice line 'こんにちは', got %s", voiceItem.Line)
	}
}

func TestParseSYMMPWithMultipleItemTypes(t *testing.T) {
	// Red: This test will fail because ParseSYMMPWithPlugins doesn't exist yet
	yamlContent := `
symmp_format_version: 0.1
project:
  name: "Multi-Item Test"
sequences:
  - id: seq1
    scenes:
      - id: scene1
        shots:
          - image:
              file_path: "./background.jpg"
              length: 10.0
            tachie:
              item: "ずんだもん"
              length: 10.0
            voice:
              line: "複数のアイテムのテストです"
              length: auto
            audio:
              file_path: "./bgm.mp3"
              volume: 0.5
              length: auto
`
	
	// Create plugin registry with all plugins
	registry := plugins.NewPluginRegistry()
	registry.RegisterItemPlugin(items.NewImagePlugin())
	registry.RegisterItemPlugin(items.NewVoicePlugin())
	registry.RegisterItemPlugin(items.NewAudioPlugin())
	registry.RegisterItemPlugin(items.NewTachiePlugin())
	
	reader := strings.NewReader(yamlContent)
	episode, err := ParseSYMMPWithPlugins(reader, registry)
	
	if err != nil {
		t.Errorf("ParseSYMMPWithPlugins failed: %v", err)
	}
	
	shot := episode.Sequences[0].Scenes[0].Shots[0]
	if len(shot.Items) != 4 {
		t.Errorf("Expected 4 items, got %d", len(shot.Items))
	}
	
	// Check item types (order may vary due to map iteration)
	expectedTypes := map[string]bool{
		"image":  true,
		"tachie": true,
		"voice":  true,
		"audio":  true,
	}
	
	actualTypes := make(map[string]bool)
	for _, item := range shot.Items {
		actualTypes[item.GetType()] = true
	}
	
	for expectedType := range expectedTypes {
		if !actualTypes[expectedType] {
			t.Errorf("Expected to find item type %s, but it was not found", expectedType)
		}
	}
}

func TestParseSYMMPWithUnknownItemType(t *testing.T) {
	// Red: This test will fail because ParseSYMMPWithPlugins doesn't exist yet
	yamlContent := `
symmp_format_version: 0.1
project:
  name: "Unknown Item Test"
sequences:
  - id: seq1
    scenes:
      - id: scene1
        shots:
          - unknown_item_type:
              some_field: "value"
              length: 5.0
`
	
	// Create plugin registry without unknown_item_type plugin
	registry := plugins.NewPluginRegistry()
	registry.RegisterItemPlugin(items.NewImagePlugin())
	
	reader := strings.NewReader(yamlContent)
	_, err := ParseSYMMPWithPlugins(reader, registry)
	
	if err == nil {
		t.Error("Expected error for unknown item type")
	}
	
	if !strings.Contains(err.Error(), "unknown_item_type") {
		t.Errorf("Expected error to mention 'unknown_item_type', got: %v", err)
	}
}

func TestParseSYMMPWithInvalidItemData(t *testing.T) {
	// Red: This test will fail because ParseSYMMPWithPlugins doesn't exist yet
	yamlContent := `
symmp_format_version: 0.1
project:
  name: "Invalid Item Test"
sequences:
  - id: seq1
    scenes:
      - id: scene1
        shots:
          - image:
              # Missing required file_path
              length: 5.0
`
	
	registry := plugins.NewPluginRegistry()
	registry.RegisterItemPlugin(items.NewImagePlugin())
	
	reader := strings.NewReader(yamlContent)
	_, err := ParseSYMMPWithPlugins(reader, registry)
	
	if err == nil {
		t.Error("Expected error for invalid item data")
	}
	
	if !strings.Contains(err.Error(), "file_path") {
		t.Errorf("Expected error to mention 'file_path', got: %v", err)
	}
}