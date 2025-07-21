package converter

import (
	"strings"
	"testing"

	"github.com/yakuari-channel/video-authorizer/internal/models"
	"github.com/yakuari-channel/video-authorizer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ConvertSimpleProject(t *testing.T) {
	// Template YMMP
	templateJSON := `{
		"FilePath": "/template.ymmp",
		"SelectedTimelineIndex": 0,
		"Timelines": [
			{
				"ID": "timeline-1",
				"Name": "メイン",
				"VideoInfo": {
					"FPS": 60,
					"Hz": 48000,
					"Width": 1920,
					"Height": 1080
				},
				"Items": [
					{
						"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
						"Frame": 0,
						"Layer": 0,
						"Length": 100,
						"Remark": "ずんだもんvoice 01",
						"CharacterName": "ずんだもん",
						"Serif": "template text",
						"IsLocked": false,
						"IsHidden": false,
						"Group": 0,
						"KeyFrames": {"Frames": [], "Count": 0}
					},
					{
						"$type": "YukkuriMovieMaker.Project.Items.TachieItem, YukkuriMovieMaker",
						"Frame": 0,
						"Layer": 0,
						"Length": 100,
						"Remark": "ずんだもん立ち絵01",
						"CharacterName": "ずんだもん",
						"IsLocked": false,
						"IsHidden": false,
						"Group": 0,
						"KeyFrames": {"Frames": [], "Count": 0}
					}
				],
				"CurrentFrame": 0,
				"Length": 100,
				"MaxLayer": 0
			}
		],
		"Characters": [],
		"CollapsedGroups": []
	}`

	// YMMPS scenario
	ymmpsYAML := `YMMPSVersion: "1"
Sequences:
  - ID: sequence1
    Scenes:
      - ID: scene1
        Shots:
          - Items:
              - _Template: "ずんだもん立ち絵01"
                Length: "150"
              - _Template: "ずんだもんvoice 01"
                Length: "74"
                Serif: "こんにちは"
      - ID: scene2
        Shots:
          - Items:
              - _Template: "ずんだもんvoice 01"
                Length: "80"
                Serif: "今日はこの沼について語るよ"
`

	// Parse template
	ymmParser := parser.NewYMMPParser()
	template, err := ymmParser.Parse(strings.NewReader(templateJSON))
	require.NoError(t, err)

	// Parse YMMPS
	ymmpsParser := parser.NewYMMPSParser()
	ymmps, err := ymmpsParser.Parse(strings.NewReader(ymmpsYAML))
	require.NoError(t, err)

	// Convert
	converter := NewConverter()
	result, err := converter.Convert(ymmps, template)
	require.NoError(t, err)

	// Verify results
	timeline := result.Timelines[0]
	
	// Should have 3 items total
	assert.Equal(t, 3, len(timeline.Items))

	// Check first item (tachie)
	tachieItem, ok := timeline.Items[0].(*models.TachieItem)
	require.True(t, ok)
	assert.Equal(t, 0, tachieItem.Frame)
	assert.Equal(t, 150, tachieItem.Length)
	assert.Equal(t, "ずんだもん立ち絵01", tachieItem.Remark)

	// Check second item (voice 1)
	voiceItem1, ok := timeline.Items[1].(*models.VoiceItem)
	require.True(t, ok)
	assert.Equal(t, 0, voiceItem1.Frame)   // Same shot, so same frame
	assert.Equal(t, 74, voiceItem1.Length)
	assert.Equal(t, "こんにちは", voiceItem1.Serif)
	assert.Equal(t, "ずんだもんvoice 01", voiceItem1.Remark)

	// Check third item (voice 2)
	voiceItem2, ok := timeline.Items[2].(*models.VoiceItem)
	require.True(t, ok)
	assert.Equal(t, 150, voiceItem2.Frame) // Starts after first scene ends (max of first shot: 150)
	assert.Equal(t, 80, voiceItem2.Length)
	assert.Equal(t, "今日はこの沼について語るよ", voiceItem2.Serif)

	// Check timeline length
	assert.Equal(t, 230, timeline.Length) // 150 + 80
}