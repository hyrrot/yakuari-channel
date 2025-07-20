package converter

import (
	"testing"

	"github.com/yakuari-channel/video-authorizer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_Convert(t *testing.T) {
	tests := []struct {
		name     string
		template *models.YMMPProject
		ymmps    *models.YMMPSDocument
		check    func(t *testing.T, result *models.YMMPProject)
		wantErr  bool
	}{
		{
			name: "simple conversion with fixed length",
			template: &models.YMMPProject{
				FilePath:              "/template.ymmp",
				SelectedTimelineIndex: 0,
				Timelines: []models.Timeline{
					{
						ID:   "timeline-1",
						Name: "メイン",
						VideoInfo: models.VideoInfo{
							FPS:    60,
							Hz:     48000,
							Width:  1920,
							Height: 1080,
						},
						Items: []interface{}{
							&models.VoiceItem{
								Type: models.ItemTypeVoice,
								BaseItem: models.BaseItem{
									Remark: "ずんだもんvoice 01",
								},
								CharacterName: "ずんだもん",
								Serif:         "template serif",
							},
						},
						CurrentFrame: 0,
						Length:       0,
						MaxLayer:     0,
					},
				},
				Characters:      []models.Character{},
				CollapsedGroups: []string{},
			},
			ymmps: &models.YMMPSDocument{
				YMMPSVersion: "1",
				Sequences: []models.Sequence{
					{
						ID: "seq1",
						Scenes: []models.Scene{
							{
								ID: "scene1",
								Shots: []models.Shot{
									{
										Items: []models.ItemSpec{
											{
												Template: "ずんだもんvoice 01",
												Length:   "100",
												Properties: map[string]interface{}{
													"Serif": "こんにちは",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			check: func(t *testing.T, result *models.YMMPProject) {
				require.Equal(t, 1, len(result.Timelines))
				timeline := result.Timelines[0]
				
				require.Equal(t, 1, len(timeline.Items))
				voiceItem, ok := timeline.Items[0].(*models.VoiceItem)
				require.True(t, ok)
				
				assert.Equal(t, 0, voiceItem.Frame)   // First item starts at frame 0
				assert.Equal(t, 100, voiceItem.Length) // Fixed length
				assert.Equal(t, "こんにちは", voiceItem.Serif) // Overridden property
				assert.Equal(t, "ずんだもん", voiceItem.CharacterName) // Preserved from template
			},
			wantErr: false,
		},
		{
			name: "template item not found",
			template: &models.YMMPProject{
				FilePath:              "/template.ymmp",
				SelectedTimelineIndex: 0,
				Timelines: []models.Timeline{
					{
						Items: []interface{}{
							&models.VoiceItem{
								BaseItem: models.BaseItem{
									Remark: "different remark",
								},
							},
						},
					},
				},
			},
			ymmps: &models.YMMPSDocument{
				YMMPSVersion: "1",
				Sequences: []models.Sequence{
					{
						Scenes: []models.Scene{
							{
								Shots: []models.Shot{
									{
										Items: []models.ItemSpec{
											{
												Template: "nonexistent template",
												Length:   "100",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewConverter()
			result, err := converter.Convert(tt.ymmps, tt.template)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, result)
				}
			}
		})
	}
}

func TestConverter_findTemplateItem(t *testing.T) {
	template := &models.YMMPProject{
		Timelines: []models.Timeline{
			{
				Items: []interface{}{
					&models.VoiceItem{
						BaseItem: models.BaseItem{
							Remark: "voice template",
						},
					},
					&models.VideoItem{
						BaseItem: models.BaseItem{
							Remark: "video template",
						},
					},
				},
			},
		},
	}

	converter := NewConverter()

	tests := []struct {
		name         string
		templateName string
		expectFound  bool
		expectType   string
	}{
		{
			name:         "find voice item",
			templateName: "voice template",
			expectFound:  true,
			expectType:   "*models.VoiceItem",
		},
		{
			name:         "find video item",
			templateName: "video template",
			expectFound:  true,
			expectType:   "*models.VideoItem",
		},
		{
			name:         "item not found",
			templateName: "nonexistent",
			expectFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, found := converter.findTemplateItem(template, tt.templateName)
			
			assert.Equal(t, tt.expectFound, found)
			if tt.expectFound {
				assert.Contains(t, tt.expectType, "*models.")
			} else {
				assert.Nil(t, item)
			}
		})
	}
}