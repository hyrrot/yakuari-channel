package parser

import (
	"strings"
	"testing"
)

func TestParseSYMMPMinimal(t *testing.T) {
	// Red: This test will fail because ParseSYMMP doesn't exist yet
	yamlContent := `
symmp_format_version: 0.1
project:
  name: "Test Project"
  output_dir: "./output"
sequences:
  - id: seq1
    scenes:
      - id: scene1
        shots: []
`
	
	reader := strings.NewReader(yamlContent)
	episode, err := ParseSYMMP(reader)
	
	if err != nil {
		t.Errorf("ParseSYMMP failed: %v", err)
	}
	
	if episode.SYMMPFormatVersion != "0.1" {
		t.Errorf("Expected version 0.1, got %s", episode.SYMMPFormatVersion)
	}
	
	if episode.Project.Name != "Test Project" {
		t.Errorf("Expected project name 'Test Project', got %s", episode.Project.Name)
	}
	
	if len(episode.Sequences) != 1 {
		t.Errorf("Expected 1 sequence, got %d", len(episode.Sequences))
	}
	
	if episode.Sequences[0].ID != "seq1" {
		t.Errorf("Expected sequence ID 'seq1', got %s", episode.Sequences[0].ID)
	}
}

func TestParseSYMMPWithDefaults(t *testing.T) {
	// Red: This test will fail because ParseSYMMP doesn't exist yet
	yamlContent := `
symmp_format_version: 0.1
project:
  name: "Test Project"
defaults:
  voice:
    character_name: "ずんだもん"
    speed: 1.0
sequences:
  - id: seq1
    scenes:
      - id: scene1
        shots: []
`
	
	reader := strings.NewReader(yamlContent)
	episode, err := ParseSYMMP(reader)
	
	if err != nil {
		t.Errorf("ParseSYMMP failed: %v", err)
	}
	
	if episode.Defaults == nil {
		t.Error("Expected defaults to be parsed")
	}
	
	voiceDefaults, ok := episode.Defaults["voice"].(map[string]interface{})
	if !ok {
		t.Error("Expected voice defaults to be a map")
	}
	
	if voiceDefaults["character_name"] != "ずんだもん" {
		t.Errorf("Expected character_name 'ずんだもん', got %v", voiceDefaults["character_name"])
	}
}

func TestParseSYMMPInvalidYAML(t *testing.T) {
	// Red: This test will fail because ParseSYMMP doesn't exist yet
	yamlContent := `
invalid: yaml: content:
  - malformed
    - structure
`
	
	reader := strings.NewReader(yamlContent)
	_, err := ParseSYMMP(reader)
	
	if err == nil {
		t.Error("Expected ParseSYMMP to fail with invalid YAML")
	}
}