package converter

import (
	"testing"
	
	"github.com/user/ymmp-compiler/internal/models"
)

func TestConvertMinimalEpisode(t *testing.T) {
	// Red: This test will fail because Convert doesn't exist yet
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
	
	project, err := Convert(episode, "/base/path")
	if err != nil {
		t.Errorf("Convert failed: %v", err)
	}
	
	if project == nil {
		t.Error("Expected project to be non-nil")
	}
	
	if len(project.Timeline.Items) != 0 {
		t.Errorf("Expected 0 items for empty shots, got %d", len(project.Timeline.Items))
	}
}

func TestConvertWithBasePath(t *testing.T) {
	// Red: This test will fail because Convert doesn't exist yet
	episode := &models.Episode{
		SYMMPFormatVersion: "0.1",
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
	
	basePath := "/test/base/path"
	project, err := Convert(episode, basePath)
	if err != nil {
		t.Errorf("Convert failed: %v", err)
	}
	
	if project == nil {
		t.Error("Expected project to be non-nil")
	}
}