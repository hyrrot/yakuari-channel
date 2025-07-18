package validator

import (
	"strings"
	"testing"
	
	"github.com/user/ymmp-compiler/internal/models"
)

func TestValidateMinimalEpisode(t *testing.T) {
	// Red: This test will fail because Validate doesn't exist yet
	episode := &models.Episode{
		SYMMPFormatVersion: "0.1",
		Project: models.ProjectConfig{
			Name:      "Test Project",
			OutputDir: "./output",
		},
		Sequences: []models.Sequence{
			{
				ID: "seq1",
				Scenes: []models.Scene{
					{
						ID:    "scene1",
						Shots: []models.Shot{},
					},
				},
			},
		},
	}
	
	err := Validate(episode)
	if err != nil {
		t.Errorf("Validate failed on valid episode: %v", err)
	}
}

func TestValidateMissingFormatVersion(t *testing.T) {
	// Red: This test will fail because Validate doesn't exist yet
	episode := &models.Episode{
		// Missing SYMMPFormatVersion
		Project: models.ProjectConfig{
			Name: "Test Project",
		},
		Sequences: []models.Sequence{
			{
				ID: "seq1",
				Scenes: []models.Scene{
					{
						ID:    "scene1",
						Shots: []models.Shot{},
					},
				},
			},
		},
	}
	
	err := Validate(episode)
	if err == nil {
		t.Error("Expected validation to fail when format version is missing")
	}
	
	if !strings.Contains(err.Error(), "symmp_format_version") {
		t.Errorf("Expected error to mention 'symmp_format_version', got: %v", err)
	}
}

func TestValidateMissingProjectName(t *testing.T) {
	// Red: This test will fail because Validate doesn't exist yet
	episode := &models.Episode{
		SYMMPFormatVersion: "0.1",
		Project: models.ProjectConfig{
			// Missing Name
			OutputDir: "./output",
		},
		Sequences: []models.Sequence{
			{
				ID: "seq1",
				Scenes: []models.Scene{
					{
						ID:    "scene1",
						Shots: []models.Shot{},
					},
				},
			},
		},
	}
	
	err := Validate(episode)
	if err == nil {
		t.Error("Expected validation to fail when project name is missing")
	}
	
	if !strings.Contains(err.Error(), "project.name") {
		t.Errorf("Expected error to mention 'project.name', got: %v", err)
	}
}

func TestValidateEmptySequenceID(t *testing.T) {
	// Red: This test will fail because Validate doesn't exist yet
	episode := &models.Episode{
		SYMMPFormatVersion: "0.1",
		Project: models.ProjectConfig{
			Name:      "Test Project",
			OutputDir: "./output",
		},
		Sequences: []models.Sequence{
			{
				ID: "", // Empty ID
				Scenes: []models.Scene{
					{
						ID:    "scene1",
						Shots: []models.Shot{},
					},
				},
			},
		},
	}
	
	err := Validate(episode)
	if err == nil {
		t.Error("Expected validation to fail when sequence ID is empty")
	}
	
	if !strings.Contains(err.Error(), "sequence ID") {
		t.Errorf("Expected error to mention 'sequence ID', got: %v", err)
	}
}

func TestValidateDuplicateSequenceIDs(t *testing.T) {
	// Red: This test will fail because Validate doesn't exist yet
	episode := &models.Episode{
		SYMMPFormatVersion: "0.1",
		Project: models.ProjectConfig{
			Name:      "Test Project",
			OutputDir: "./output",
		},
		Sequences: []models.Sequence{
			{
				ID: "seq1",
				Scenes: []models.Scene{
					{
						ID:    "scene1",
						Shots: []models.Shot{},
					},
				},
			},
			{
				ID: "seq1", // Duplicate ID
				Scenes: []models.Scene{
					{
						ID:    "scene2",
						Shots: []models.Shot{},
					},
				},
			},
		},
	}
	
	err := Validate(episode)
	if err == nil {
		t.Error("Expected validation to fail when sequence IDs are duplicated")
	}
	
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("Expected error to mention 'duplicate', got: %v", err)
	}
}

func TestValidateEmptySceneID(t *testing.T) {
	// Red: This test will fail because Validate doesn't exist yet
	episode := &models.Episode{
		SYMMPFormatVersion: "0.1",
		Project: models.ProjectConfig{
			Name:      "Test Project",
			OutputDir: "./output",
		},
		Sequences: []models.Sequence{
			{
				ID: "seq1",
				Scenes: []models.Scene{
					{
						ID:    "", // Empty ID
						Shots: []models.Shot{},
					},
				},
			},
		},
	}
	
	err := Validate(episode)
	if err == nil {
		t.Error("Expected validation to fail when scene ID is empty")
	}
	
	if !strings.Contains(err.Error(), "scene ID") {
		t.Errorf("Expected error to mention 'scene ID', got: %v", err)
	}
}