package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYMMPProject_Validate(t *testing.T) {
	tests := []struct {
		name    string
		project YMMPProject
		wantErr bool
	}{
		{
			name: "valid project",
			project: YMMPProject{
				FilePath:              "/path/to/project.ymmp",
				SelectedTimelineIndex: 0,
				Timelines: []Timeline{
					{
						ID:   "timeline-1",
						Name: "メイン",
						VideoInfo: VideoInfo{
							FPS:    60,
							Hz:     48000,
							Width:  1920,
							Height: 1080,
						},
						Items: []interface{}{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty timelines",
			project: YMMPProject{
				FilePath:  "/path/to/project.ymmp",
				Timelines: []Timeline{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.project.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBaseItem_GetRemark(t *testing.T) {
	item := BaseItem{
		Frame:  0,
		Layer:  0,
		Length: 100,
		Remark: "test item",
	}

	assert.Equal(t, "test item", item.GetRemark())
}

func TestVoiceItem_GetType(t *testing.T) {
	item := VoiceItem{
		Type: "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
	}

	assert.Equal(t, "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker", item.GetType())
}