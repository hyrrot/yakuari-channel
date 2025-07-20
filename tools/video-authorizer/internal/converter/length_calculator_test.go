package converter

import (
	"testing"

	"github.com/yakuari-channel/video-authorizer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLengthCalculator_CalculateLength(t *testing.T) {
	calculator := NewLengthCalculator()
	ctx := NewCompileContext()
	
	// Set up some test context data
	ctx.RegisterSequenceEnd("seq1", 300)
	ctx.RegisterSceneEnd("scene1", 200)
	ctx.RegisterShotEnd("shot1", 100)
	ctx.CurrentFrame = 50

	tests := []struct {
		name      string
		lengthStr string
		want      int
		wantErr   bool
	}{
		{
			name:      "numeric length",
			lengthStr: "100",
			want:      100,
			wantErr:   false,
		},
		{
			name:      "until sequence end with ID",
			lengthStr: "_until:SEQUENCE_END:seq1",
			want:      250, // 300 - 50
			wantErr:   false,
		},
		{
			name:      "until scene end with ID",
			lengthStr: "_until:SCENE_END:scene1",
			want:      150, // 200 - 50
			wantErr:   false,
		},
		{
			name:      "until shot end with ID",
			lengthStr: "_until:SHOT_END:shot1",
			want:      50, // 100 - 50
			wantErr:   false,
		},
		{
			name:      "until nonexistent ID",
			lengthStr: "_until:SEQUENCE_END:nonexistent",
			wantErr:   true,
		},
		{
			name:      "auto voice (not implemented)",
			lengthStr: "_auto:VOICEVOX",
			wantErr:   true,
		},
		{
			name:      "auto video (not implemented)",
			lengthStr: "_auto:VIDEO",
			wantErr:   true,
		},
		{
			name:      "until sequence end (no prepass)",
			lengthStr: "_until:SEQUENCE_END",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculator.CalculateLength(tt.lengthStr, ctx)
			
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestAdvancedLengthCalculator_CalculateWithPrepass(t *testing.T) {
	calculator := NewAdvancedLengthCalculator()

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
										Template: "item1",
										Length:   "100",
									},
									{
										Template: "item2",
										Length:   "50",
									},
								},
							},
						},
					},
					{
						ID: "scene2",
						Shots: []models.Shot{
							{
								Items: []models.ItemSpec{
									{
										Template: "item3",
										Length:   "75",
									},
								},
							},
						},
					},
				},
			},
			{
				ID: "seq2",
				Scenes: []models.Scene{
					{
						Shots: []models.Shot{
							{
								Items: []models.ItemSpec{
									{
										Template: "item4",
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

	ctx, err := calculator.CalculateWithPrepass(ymmps)
	require.NoError(t, err)

	// Check sequence end frames
	assert.Equal(t, 175, ctx.SequenceEndFrames["seq1"]) // 100 (scene1) + 75 (scene2)
	assert.Equal(t, 375, ctx.SequenceEndFrames["seq2"]) // 175 + 200

	// Check scene end frames
	assert.Equal(t, 100, ctx.SceneEndFrames["scene1"]) // max(100, 50) = 100
	assert.Equal(t, 175, ctx.SceneEndFrames["scene2"]) // 100 + 75

	// Check shot end frames
	assert.Equal(t, 100, ctx.ShotEndFrames["shot1"]) // max(100, 50) = 100
}

func TestCompileContext_RegisterAndLookup(t *testing.T) {
	ctx := NewCompileContext()

	// Test sequence registration
	ctx.RegisterSequenceEnd("seq1", 100)
	ctx.RegisterSequenceEnd("", 200) // Should be ignored
	
	endFrame, found := ctx.SequenceEndFrames["seq1"]
	assert.True(t, found)
	assert.Equal(t, 100, endFrame)
	
	_, found = ctx.SequenceEndFrames[""]
	assert.False(t, found)

	// Test scene registration
	ctx.RegisterSceneEnd("scene1", 50)
	endFrame, found = ctx.SceneEndFrames["scene1"]
	assert.True(t, found)
	assert.Equal(t, 50, endFrame)

	// Test shot registration
	ctx.RegisterShotEnd("shot1", 25)
	endFrame, found = ctx.ShotEndFrames["shot1"]
	assert.True(t, found)
	assert.Equal(t, 25, endFrame)
}