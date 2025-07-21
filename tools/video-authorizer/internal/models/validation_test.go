package models_test

import (
	"strings"
	"testing"

	"github.com/yakuari-channel/video-authorizer/internal/models"
)

func TestValidateReferences(t *testing.T) {
	tests := []struct {
		name    string
		doc     *models.YMMPSDocument
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid ID references",
			doc: &models.YMMPSDocument{
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
												Template: "item1",
												Length:   "_until:SEQUENCE_END:seq1",
											},
											{
												Template: "item2",
												Length:   "_until:SCENE_END:scene1",
											},
											{
												Template: "item3",
												Length:   "_until:SHOT_END:shot1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "undefined sequence ID reference",
			doc: &models.YMMPSDocument{
				YMMPSVersion: "1",
				Sequences: []models.Sequence{
					{
						ID: "seq1",
						Scenes: []models.Scene{
							{
								Shots: []models.Shot{
									{
										Items: []models.ItemSpec{
											{
												Template: "item1",
												Length:   "_until:SEQUENCE_END:nonexistent",
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
			errMsg:  "undefined SEQUENCE ID 'nonexistent'",
		},
		{
			name: "undefined scene ID reference",
			doc: &models.YMMPSDocument{
				YMMPSVersion: "1",
				Sequences: []models.Sequence{
					{
						Scenes: []models.Scene{
							{
								ID: "scene1",
								Shots: []models.Shot{
									{
										Items: []models.ItemSpec{
											{
												Template: "item1",
												Length:   "_until:SCENE_END:nonexistent",
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
			errMsg:  "undefined SCENE ID 'nonexistent'",
		},
		{
			name: "undefined shot ID reference",
			doc: &models.YMMPSDocument{
				YMMPSVersion: "1",
				Sequences: []models.Sequence{
					{
						Scenes: []models.Scene{
							{
								Shots: []models.Shot{
									{
										ID: "shot1",
										Items: []models.ItemSpec{
											{
												Template: "item1",
												Length:   "_until:SHOT_END:nonexistent",
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
			errMsg:  "undefined SHOT ID 'nonexistent'",
		},
		{
			name: "wrong type reference",
			doc: &models.YMMPSDocument{
				YMMPSVersion: "1",
				Sequences: []models.Sequence{
					{
						ID: "seq1",
						Scenes: []models.Scene{
							{
								Shots: []models.Shot{
									{
										Items: []models.ItemSpec{
											{
												Template: "item1",
												Length:   "_until:SCENE_END:seq1", // seq1 is a sequence, not a scene
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
			errMsg:  "ID 'seq1' is a SEQUENCE but referenced as SCENE",
		},
		{
			name: "empty ID reference",
			doc: &models.YMMPSDocument{
				YMMPSVersion: "1",
				Sequences: []models.Sequence{
					{
						Scenes: []models.Scene{
							{
								Shots: []models.Shot{
									{
										Items: []models.ItemSpec{
											{
												Template: "item1",
												Length:   "_until:SEQUENCE_END:", // Empty ID
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
			errMsg:  "empty ID reference",
		},
		{
			name: "natural language format with ID",
			doc: &models.YMMPSDocument{
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
												Template: "item1",
												Length:   "UNTIL SEQUENCE seq1 END",
											},
											{
												Template: "item2",
												Length:   "UNTIL SCENE scene1 END",
											},
											{
												Template: "item3",
												Length:   "UNTIL SHOT shot1 END",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.doc.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', but got none", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing '%s', but got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}