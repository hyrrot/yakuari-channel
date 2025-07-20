package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYMMPSDocument_EnsureIDs(t *testing.T) {
	tests := []struct {
		name     string
		ymmps    *YMMPSDocument
		validate func(t *testing.T, ymmps *YMMPSDocument)
	}{
		{
			name: "all_elements_without_ids",
			ymmps: &YMMPSDocument{
				YMMPSVersion: "1",
				Sequences: []Sequence{
					{
						Scenes: []Scene{
							{
								Shots: []Shot{
									{
										Items: []ItemSpec{
											{
												Template: "test_template",
												Length:   "120",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, ymmps *YMMPSDocument) {
				// All sequences should have IDs
				assert.NotEmpty(t, ymmps.Sequences[0].ID)
				assert.True(t, strings.HasPrefix(ymmps.Sequences[0].ID, "__auto_seq_"))
				
				// All scenes should have IDs
				assert.NotEmpty(t, ymmps.Sequences[0].Scenes[0].ID)
				assert.True(t, strings.HasPrefix(ymmps.Sequences[0].Scenes[0].ID, "__auto_scene_"))
				
				// All shots should have IDs
				assert.NotEmpty(t, ymmps.Sequences[0].Scenes[0].Shots[0].ID)
				assert.True(t, strings.HasPrefix(ymmps.Sequences[0].Scenes[0].Shots[0].ID, "__auto_shot_"))
			},
		},
		{
			name: "mixed_existing_and_missing_ids",
			ymmps: &YMMPSDocument{
				YMMPSVersion: "1",
				Sequences: []Sequence{
					{
						ID: "existing_seq",
						Scenes: []Scene{
							{
								// No ID - should get auto-generated
								Shots: []Shot{
									{
										ID: "existing_shot",
										Items: []ItemSpec{
											{
												Template: "test_template",
												Length:   "120",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, ymmps *YMMPSDocument) {
				// Existing sequence ID should be preserved
				assert.Equal(t, "existing_seq", ymmps.Sequences[0].ID)
				
				// Scene should get auto-generated ID
				assert.NotEmpty(t, ymmps.Sequences[0].Scenes[0].ID)
				assert.True(t, strings.HasPrefix(ymmps.Sequences[0].Scenes[0].ID, "__auto_scene_"))
				
				// Existing shot ID should be preserved
				assert.Equal(t, "existing_shot", ymmps.Sequences[0].Scenes[0].Shots[0].ID)
			},
		},
		{
			name: "multiple_sequences_scenes_shots",
			ymmps: &YMMPSDocument{
				YMMPSVersion: "1",
				Sequences: []Sequence{
					{
						Scenes: []Scene{
							{
								Shots: []Shot{
									{
										Items: []ItemSpec{
											{Template: "test1", Length: "120"},
										},
									},
									{
										Items: []ItemSpec{
											{Template: "test2", Length: "120"},
										},
									},
								},
							},
							{
								Shots: []Shot{
									{
										Items: []ItemSpec{
											{Template: "test3", Length: "120"},
										},
									},
								},
							},
						},
					},
					{
						Scenes: []Scene{
							{
								Shots: []Shot{
									{
										Items: []ItemSpec{
											{Template: "test4", Length: "120"},
										},
									},
								},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, ymmps *YMMPSDocument) {
				// All elements should have unique IDs
				seenIDs := make(map[string]bool)
				
				for i, sequence := range ymmps.Sequences {
					assert.NotEmpty(t, sequence.ID, "Sequence %d should have ID", i)
					assert.False(t, seenIDs[sequence.ID], "Sequence ID should be unique: %s", sequence.ID)
					seenIDs[sequence.ID] = true
					
					for j, scene := range sequence.Scenes {
						assert.NotEmpty(t, scene.ID, "Scene %d.%d should have ID", i, j)
						assert.False(t, seenIDs[scene.ID], "Scene ID should be unique: %s", scene.ID)
						seenIDs[scene.ID] = true
						
						for k, shot := range scene.Shots {
							assert.NotEmpty(t, shot.ID, "Shot %d.%d.%d should have ID", i, j, k)
							assert.False(t, seenIDs[shot.ID], "Shot ID should be unique: %s", shot.ID)
							seenIDs[shot.ID] = true
						}
					}
				}
			},
		},
		{
			name: "all_elements_already_have_ids",
			ymmps: &YMMPSDocument{
				YMMPSVersion: "1",
				Sequences: []Sequence{
					{
						ID: "seq1",
						Scenes: []Scene{
							{
								ID: "scene1",
								Shots: []Shot{
									{
										ID: "shot1",
										Items: []ItemSpec{
											{Template: "test", Length: "120"},
										},
									},
								},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, ymmps *YMMPSDocument) {
				// All original IDs should be preserved
				assert.Equal(t, "seq1", ymmps.Sequences[0].ID)
				assert.Equal(t, "scene1", ymmps.Sequences[0].Scenes[0].ID)
				assert.Equal(t, "shot1", ymmps.Sequences[0].Scenes[0].Shots[0].ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ymmps.EnsureIDs()
			require.NoError(t, err)
			
			// Run specific validations
			tt.validate(t, tt.ymmps)
		})
	}
}

func TestGenerateUniqueID(t *testing.T) {
	existingIDs := make(map[string]bool)
	
	// Test basic ID generation
	id1, err := generateUniqueID("test", existingIDs)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(id1, "__auto_test_"))
	assert.Len(t, id1, len("__auto_test_")+12) // prefix + 12 hex chars
	
	// Test uniqueness
	id2, err := generateUniqueID("test", existingIDs)
	require.NoError(t, err)
	assert.NotEqual(t, id1, id2)
	
	// Test different prefixes
	id3, err := generateUniqueID("scene", existingIDs)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(id3, "__auto_scene_"))
}

func TestEnsureIDs_Integration(t *testing.T) {
	// Test with a realistic YMMPS document structure
	ymmps := &YMMPSDocument{
		YMMPSVersion: "1",
		Sequences: []Sequence{
			{
				Scenes: []Scene{
					{
						Shots: []Shot{
							{
								Items: []ItemSpec{
									{
										Template: "voice_template",
										Length:   "120",
									},
								},
							},
							{
								Items: []ItemSpec{
									{
										Template: "voice_template",
										Length:   "_until:SCENE_END",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	// Before EnsureIDs - no IDs
	assert.Empty(t, ymmps.Sequences[0].ID)
	assert.Empty(t, ymmps.Sequences[0].Scenes[0].ID)
	assert.Empty(t, ymmps.Sequences[0].Scenes[0].Shots[0].ID)
	assert.Empty(t, ymmps.Sequences[0].Scenes[0].Shots[1].ID)
	
	// Apply EnsureIDs
	err := ymmps.EnsureIDs()
	require.NoError(t, err)
	
	// After EnsureIDs - all should have IDs
	assert.NotEmpty(t, ymmps.Sequences[0].ID)
	assert.NotEmpty(t, ymmps.Sequences[0].Scenes[0].ID)
	assert.NotEmpty(t, ymmps.Sequences[0].Scenes[0].Shots[0].ID)
	assert.NotEmpty(t, ymmps.Sequences[0].Scenes[0].Shots[1].ID)
	
	// IDs should be different
	assert.NotEqual(t, ymmps.Sequences[0].Scenes[0].Shots[0].ID, 
		ymmps.Sequences[0].Scenes[0].Shots[1].ID)
	
	// The document should still validate correctly
	err = ymmps.Validate()
	assert.NoError(t, err)
}