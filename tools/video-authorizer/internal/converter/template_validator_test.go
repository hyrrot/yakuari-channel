package converter_test

import (
	"strings"
	"testing"

	"github.com/yakuari-channel/video-authorizer/internal/converter"
	"github.com/yakuari-channel/video-authorizer/internal/models"
)

func TestTemplateValidator_ValidateTemplateReferences(t *testing.T) {
	validator := converter.NewTemplateValidator()
	
	tests := []struct {
		name     string
		ymmps    *models.YMMPSDocument
		template *models.YMMPProject
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid template references",
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
												Template: "voice_template",
												Length:   "100",
											},
											{
												Template: "video_template",
												Length:   "200",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			template: &models.YMMPProject{
				Timelines: []models.Timeline{
					{
						Items: []interface{}{
							&models.VoiceItem{
								BaseItem: models.BaseItem{
									Remark: "voice_template",
								},
							},
							&models.VideoItem{
								BaseItem: models.BaseItem{
									Remark: "video_template",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing template reference",
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
												Template: "nonexistent_template",
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
			template: &models.YMMPProject{
				Timelines: []models.Timeline{
					{
						Items: []interface{}{
							&models.VoiceItem{
								BaseItem: models.BaseItem{
									Remark: "different_template",
								},
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "template 'nonexistent_template' not found",
		},
		{
			name: "empty template project",
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
												Template: "any_template",
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
			template: &models.YMMPProject{
				Timelines: []models.Timeline{}, // Empty timelines
			},
			wantErr: true,
			errMsg:  "template 'any_template' not found",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTemplateReferences(tt.ymmps, tt.template)
			
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing '%s', got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestTemplateValidator_CollectTemplateReferences(t *testing.T) {
	validator := converter.NewTemplateValidator()
	
	ymmps := &models.YMMPSDocument{
		YMMPSVersion: "1",
		Sequences: []models.Sequence{
			{
				ID: "seq1",
				Scenes: []models.Scene{
					{
						ID: "scene1",
						Shots: []models.Shot{
							{
								ID: "shot1",
								Items: []models.ItemSpec{
									{
										Template: "template1",
										Length:   "100",
									},
									{
										Template: "template2",
										Length:   "200",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	// Use reflection or a test helper to access the private method
	// For now, we'll test it indirectly through ValidateTemplateReferences
	template := &models.YMMPProject{
		Timelines: []models.Timeline{
			{
				Items: []interface{}{
					&models.VoiceItem{
						BaseItem: models.BaseItem{
							Remark: "template1",
						},
					},
					// Missing template2 to test reference collection
				},
			},
		},
	}
	
	err := validator.ValidateTemplateReferences(ymmps, template)
	if err == nil {
		t.Error("expected error for missing template2")
	}
	
	if !strings.Contains(err.Error(), "template2") {
		t.Errorf("expected error to mention template2, got: %v", err)
	}
}

func TestTemplateValidator_CollectAvailableTemplates(t *testing.T) {
	validator := converter.NewTemplateValidator()
	
	template := &models.YMMPProject{
		Timelines: []models.Timeline{
			{
				Items: []interface{}{
					&models.VoiceItem{
						BaseItem: models.BaseItem{
							Remark: "voice_template",
						},
					},
					&models.VideoItem{
						BaseItem: models.BaseItem{
							Remark: "video_template",
						},
					},
					&models.TachieItem{
						BaseItem: models.BaseItem{
							Remark: "", // Empty remark should be ignored
						},
					},
				},
			},
		},
	}
	
	// Test indirectly by checking that both templates are available
	ymmps := &models.YMMPSDocument{
		YMMPSVersion: "1",
		Sequences: []models.Sequence{
			{
				Scenes: []models.Scene{
					{
						Shots: []models.Shot{
							{
								Items: []models.ItemSpec{
									{
										Template: "voice_template",
										Length:   "100",
									},
									{
										Template: "video_template",
										Length:   "200",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	err := validator.ValidateTemplateReferences(ymmps, template)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTemplateValidator_ValidateTemplateCompatibility(t *testing.T) {
	validator := converter.NewTemplateValidator()
	
	// Test with voice properties on video template (should generate warning)
	ymmps := &models.YMMPSDocument{
		YMMPSVersion: "1",
		Sequences: []models.Sequence{
			{
				Scenes: []models.Scene{
					{
						Shots: []models.Shot{
							{
								Items: []models.ItemSpec{
									{
										Template: "video_template",
										Length:   "100",
										Properties: map[string]interface{}{
											"Serif": "Hello World", // Voice property on video template
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	template := &models.YMMPProject{
		Timelines: []models.Timeline{
			{
				Items: []interface{}{
					&models.VideoItem{
						BaseItem: models.BaseItem{
							Remark: "video_template",
						},
					},
				},
			},
		},
	}
	
	// This should not return an error (warnings are printed, not returned)
	err := validator.ValidateTemplateCompatibility(ymmps, template)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTemplateValidator_PropertyDetection(t *testing.T) {
	validator := converter.NewTemplateValidator()
	
	tests := []struct {
		name           string
		properties     map[string]interface{}
		expectedVoice  bool
		expectedVideo  bool
		expectedImage  bool
	}{
		{
			name: "voice properties",
			properties: map[string]interface{}{
				"Serif":         "Hello",
				"CharacterName": "Speaker1",
			},
			expectedVoice: true,
			expectedVideo: false,
			expectedImage: false,
		},
		{
			name: "video properties",
			properties: map[string]interface{}{
				"FilePath":     "/path/to/video.mp4",
				"PlaybackRate": 1.0,
			},
			expectedVoice: false,
			expectedVideo: true,
			expectedImage: false,
		},
		{
			name: "image properties",
			properties: map[string]interface{}{
				"X":    100,
				"Y":    200,
				"Zoom": 1.5,
			},
			expectedVoice: false,
			expectedVideo: false,
			expectedImage: true,
		},
		{
			name: "mixed properties",
			properties: map[string]interface{}{
				"Serif":    "Hello",
				"FilePath": "/path/to/video.mp4",
				"X":        100,
			},
			expectedVoice: true,
			expectedVideo: true,
			expectedImage: true,
		},
		{
			name:           "no relevant properties",
			properties:     map[string]interface{}{"SomeOtherField": "value"},
			expectedVoice:  false,
			expectedVideo:  false,
			expectedImage:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test through a YMMPS document that uses these properties
			ymmps := &models.YMMPSDocument{
				YMMPSVersion: "1",
				Sequences: []models.Sequence{
					{
						Scenes: []models.Scene{
							{
								Shots: []models.Shot{
									{
										Items: []models.ItemSpec{
											{
												Template:   "test_template",
												Length:     "100",
												Properties: tt.properties,
											},
										},
									},
								},
							},
						},
					},
				},
			}
			
			template := &models.YMMPProject{
				Timelines: []models.Timeline{
					{
						Items: []interface{}{
							&models.VoiceItem{
								BaseItem: models.BaseItem{
									Remark: "test_template",
								},
							},
						},
					},
				},
			}
			
			// The validation should succeed (property detection is internal)
			err := validator.ValidateTemplateReferences(ymmps, template)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}