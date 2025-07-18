package models

import (
	"testing"
)

func TestYMMPProjectStructure(t *testing.T) {
	// Red: This test will fail because YMMPProject doesn't exist yet
	project := YMMPProject{
		Timeline: Timeline{
			Items: []YMMPItem{
				{
					Type:     "YMM.Parts.ImagePart",
					Layer:    1,
					Frame:    0,
					Length:   300, // 5 seconds at 60fps
					FilePath: "/path/to/image.png",
				},
				{
					Type:     "YMM.Parts.VoicePart",
					Layer:    2,
					Frame:    0,
					Length:   180, // 3 seconds at 60fps
					FilePath: "/path/to/voice.wav",
				},
			},
		},
	}
	
	if len(project.Timeline.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(project.Timeline.Items))
	}
	
	imageItem := project.Timeline.Items[0]
	if imageItem.Type != "YMM.Parts.ImagePart" {
		t.Errorf("Expected type 'YMM.Parts.ImagePart', got %s", imageItem.Type)
	}
	
	if imageItem.Layer != 1 {
		t.Errorf("Expected layer 1, got %d", imageItem.Layer)
	}
	
	if imageItem.Length != 300 {
		t.Errorf("Expected length 300, got %d", imageItem.Length)
	}
}

func TestYMMPItemExtendedProperties(t *testing.T) {
	// Red: This test will fail because YMMPItem doesn't exist yet
	item := YMMPItem{
		Type:     "YMM.Parts.ImagePart",
		Layer:    1,
		Frame:    0,
		Length:   300,
		FilePath: "/path/to/image.png",
		Extended: map[string]interface{}{
			"X":      160,
			"Y":      90,
			"Zoom":   1.0,
			"Alpha":  255,
		},
	}
	
	if item.Extended["X"] != 160 {
		t.Errorf("Expected X=160, got %v", item.Extended["X"])
	}
	
	if item.Extended["Zoom"] != 1.0 {
		t.Errorf("Expected Zoom=1.0, got %v", item.Extended["Zoom"])
	}
}