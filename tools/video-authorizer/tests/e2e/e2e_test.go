package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	
	"github.com/user/ymmp-compiler/internal/core/converter"
	"github.com/user/ymmp-compiler/internal/core/parser"
	"github.com/user/ymmp-compiler/internal/core/validator"
	"github.com/user/ymmp-compiler/internal/models"
	"github.com/user/ymmp-compiler/internal/plugins"
	"github.com/user/ymmp-compiler/internal/plugins/items"
)

func TestE2ESimpleConversion(t *testing.T) {
	// Red: This test will fail because the integration doesn't exist yet
	
	// Create a temporary SYMMP file
	symmContent := `symmp_format_version: 0.1
project:
  name: "E2E Test Project"
  output_dir: "./output"
sequences:
  - id: opening
    scenes:
      - id: intro
        shots:
          - image:
              file_path: "./assets/background.png"
              x: 0
              y: 0
              length: 5.0
            voice:
              line: "こんにちは、テストです"
              character_name: "テスト"
              length: 3.0
`
	
	// Write temp file
	tmpFile, err := os.CreateTemp("", "test*.symmp")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(symmContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()
	
	// Test the full conversion pipeline
	result, err := ConvertSYMMPToYMMP(tmpFile.Name(), "/test/base/path")
	if err != nil {
		t.Fatalf("ConvertSYMMPToYMMP failed: %v", err)
	}
	
	// Verify the result
	if result.Timeline.Items == nil {
		t.Error("Expected timeline items to be non-nil")
	}
	
	if len(result.Timeline.Items) != 2 {
		t.Errorf("Expected 2 timeline items, got %d", len(result.Timeline.Items))
	}
	
	// Check first item (image)
	imageItem := result.Timeline.Items[0]
	if imageItem.Type != "YMM.Parts.ImagePart" {
		t.Errorf("Expected first item type 'YMM.Parts.ImagePart', got %s", imageItem.Type)
	}
	
	if imageItem.Length != 300 { // 5.0 seconds * 60 FPS
		t.Errorf("Expected first item length 300, got %d", imageItem.Length)
	}
	
	// Check second item (voice)
	voiceItem := result.Timeline.Items[1]
	if voiceItem.Type != "YMM.Parts.VoicePart" {
		t.Errorf("Expected second item type 'YMM.Parts.VoicePart', got %s", voiceItem.Type)
	}
	
	if voiceItem.Length != 180 { // 3.0 seconds * 60 FPS
		t.Errorf("Expected second item length 180, got %d", voiceItem.Length)
	}
	
	// Check that both items start at the same time (same shot)
	if imageItem.Frame != voiceItem.Frame {
		t.Errorf("Expected both items to start at the same frame, got %d and %d", imageItem.Frame, voiceItem.Frame)
	}
}

func TestE2ESequentialScenes(t *testing.T) {
	// Red: This test will fail because the integration doesn't exist yet
	
	symmContent := `symmp_format_version: 0.1
project:
  name: "Sequential Test"
sequences:
  - id: main
    scenes:
      - id: scene1
        shots:
          - image:
              file_path: "./scene1.png"
              length: 2.0
      - id: scene2
        shots:
          - image:
              file_path: "./scene2.png"
              length: 3.0
`
	
	tmpFile, err := os.CreateTemp("", "test*.symmp")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(symmContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()
	
	result, err := ConvertSYMMPToYMMP(tmpFile.Name(), "/test/base")
	if err != nil {
		t.Fatalf("ConvertSYMMPToYMMP failed: %v", err)
	}
	
	if len(result.Timeline.Items) != 2 {
		t.Errorf("Expected 2 timeline items, got %d", len(result.Timeline.Items))
	}
	
	// First scene should start at frame 0
	if result.Timeline.Items[0].Frame != 0 {
		t.Errorf("Expected first item to start at frame 0, got %d", result.Timeline.Items[0].Frame)
	}
	
	// Second scene should start after first scene ends (2.0 seconds = 120 frames)
	if result.Timeline.Items[1].Frame != 120 {
		t.Errorf("Expected second item to start at frame 120, got %d", result.Timeline.Items[1].Frame)
	}
}

func TestE2EWriteYMMPFile(t *testing.T) {
	// Red: This test will fail because WriteYMMPFile doesn't exist yet
	
	symmContent := `symmp_format_version: 0.1
project:
  name: "Write Test"
sequences:
  - id: test
    scenes:
      - id: scene1
        shots:
          - image:
              file_path: "./test.png"
              length: 1.0
`
	
	tmpInputFile, err := os.CreateTemp("", "input*.symmp")
	if err != nil {
		t.Fatalf("Failed to create temp input file: %v", err)
	}
	defer os.Remove(tmpInputFile.Name())
	
	if _, err := tmpInputFile.WriteString(symmContent); err != nil {
		t.Fatalf("Failed to write temp input file: %v", err)
	}
	tmpInputFile.Close()
	
	tmpOutputFile, err := os.CreateTemp("", "output*.ymmp")
	if err != nil {
		t.Fatalf("Failed to create temp output file: %v", err)
	}
	defer os.Remove(tmpOutputFile.Name())
	tmpOutputFile.Close()
	
	// Test writing the YMMP file
	err = ConvertAndWriteYMMPFile(tmpInputFile.Name(), tmpOutputFile.Name(), "/test/base")
	if err != nil {
		t.Fatalf("ConvertAndWriteYMMPFile failed: %v", err)
	}
	
	// Verify the output file exists and contains valid JSON
	data, err := os.ReadFile(tmpOutputFile.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	
	var ymmProject models.YMMPProject
	if err := json.Unmarshal(data, &ymmProject); err != nil {
		t.Fatalf("Output file is not valid JSON: %v", err)
	}
	
	if len(ymmProject.Timeline.Items) != 1 {
		t.Errorf("Expected 1 timeline item in output file, got %d", len(ymmProject.Timeline.Items))
	}
}

// ConvertSYMMPToYMMP converts a SYMMP file to YMMP format
func ConvertSYMMPToYMMP(inputPath, basePath string) (*models.YMMPProject, error) {
	// Open the input file
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open input file: %v", err)
	}
	defer file.Close()
	
	// Create plugin registry
	registry := plugins.NewPluginRegistry()
	registry.RegisterItemPlugin(items.NewImagePlugin())
	registry.RegisterItemPlugin(items.NewVoicePlugin())
	registry.RegisterItemPlugin(items.NewAudioPlugin())
	registry.RegisterItemPlugin(items.NewVideoPlugin())
	registry.RegisterItemPlugin(items.NewTachiePlugin())
	
	// Parse the SYMMP file
	episode, err := parser.ParseSYMMPWithPlugins(file, registry)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SYMMP file: %v", err)
	}
	
	// Validate the episode
	if err := validator.Validate(episode); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}
	
	// Calculate timeline
	if err := converter.CalculateTimeline(episode); err != nil {
		return nil, fmt.Errorf("failed to calculate timeline: %v", err)
	}
	
	// Convert to YMMP format
	ymmProject, err := converter.ConvertToYMMP(episode, basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to YMMP: %v", err)
	}
	
	return ymmProject, nil
}

// ConvertAndWriteYMMPFile converts a SYMMP file and writes it to a YMMP file
func ConvertAndWriteYMMPFile(inputPath, outputPath, basePath string) error {
	// Convert SYMMP to YMMP
	ymmProject, err := ConvertSYMMPToYMMP(inputPath, basePath)
	if err != nil {
		return fmt.Errorf("failed to convert SYMMP to YMMP: %v", err)
	}
	
	// Marshal to JSON
	jsonData, err := json.MarshalIndent(ymmProject, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal YMMP data: %v", err)
	}
	
	// Write to output file
	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}
	
	return nil
}