package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yakuari-channel/video-authorizer/internal/converter"
	"github.com/yakuari-channel/video-authorizer/internal/models"
	"github.com/yakuari-channel/video-authorizer/internal/parser"
)

func TestCompleteWorkflow(t *testing.T) {
	tests := []struct {
		name         string
		ymmpsContent string
		ymmpTemplate string
		wantItems    int
		wantLength   int
		checkItems   func(t *testing.T, items []interface{})
	}{
		{
			name: "basic_conversion",
			ymmpsContent: `YMMPSVersion: "1"
Sequences:
  - ID: seq1
    Scenes:
      - ID: scene1
        Shots:
          - ID: shot1
            Items:
              - Type: VoiceItem
                Template: voice_template
                Length: "120"
                Properties:
                  Serif: "Hello, World!"
`,
			ymmpTemplate: `{
  "FilePath": "template.ymmp",
  "Timelines": [{
    "ID": "timeline1",
    "Items": [{
      "$type": "YukkuriMovieMaker.Project.VoiceItem",
      "Remark": "voice_template",
      "Serif": "Template text",
      "Frame": 0,
      "Length": 100,
      "Layer": 0
    }]
  }]
}`,
			wantItems:  1,
			wantLength: 120,
			checkItems: func(t *testing.T, items []interface{}) {
				voiceItem := items[0].(*models.VoiceItem)
				assert.Equal(t, "Hello, World!", voiceItem.Serif)
				assert.Equal(t, 0, voiceItem.Frame)
				assert.Equal(t, 120, voiceItem.Length)
			},
		},
		{
			name: "relative_length_until_shot_end",
			ymmpsContent: `YMMPSVersion: "1"
Sequences:
  - ID: seq1
    Scenes:
      - ID: scene1
        Shots:
          - ID: shot1
            Items:
              - Type: VoiceItem
                Template: voice1
                Length: "100"
              - Type: TachieItem
                Template: tachie1
                Length: "UNTIL SHOT END"
`,
			ymmpTemplate: `{
  "FilePath": "template.ymmp",
  "Timelines": [{
    "ID": "timeline1",
    "Items": [{
      "$type": "YukkuriMovieMaker.Project.VoiceItem",
      "Remark": "voice1",
      "Serif": "Voice",
      "Frame": 0,
      "Length": 50
    },{
      "$type": "YukkuriMovieMaker.Project.TachieItem",
      "Remark": "tachie1",
      "VideoItem": {
        "FilePath": "tachie.png"
      },
      "Frame": 0,
      "Length": 50
    }]
  }]
}`,
			wantItems:  2,
			wantLength: 100,
			checkItems: func(t *testing.T, items []interface{}) {
				assert.Equal(t, 100, items[0].(*models.VoiceItem).Length)
				assert.Equal(t, 100, items[1].(*models.TachieItem).Length)
			},
		},
		{
			name: "multiple_shots",
			ymmpsContent: `YMMPSVersion: "1"
Sequences:
  - ID: seq1
    Scenes:
      - ID: scene1
        Shots:
          - ID: shot1
            Items:
              - Type: VoiceItem
                Template: voice1
                Length: "60"
          - ID: shot2
            Items:
              - Type: VoiceItem
                Template: voice1
                Length: "40"
`,
			ymmpTemplate: `{
  "FilePath": "template.ymmp",
  "Timelines": [{
    "ID": "timeline1",
    "Items": [{
      "$type": "YukkuriMovieMaker.Project.VoiceItem",
      "Remark": "voice1",
      "Serif": "Voice",
      "Frame": 0,
      "Length": 50
    }]
  }]
}`,
			wantItems:  2,
			wantLength: 100,
			checkItems: func(t *testing.T, items []interface{}) {
				assert.Equal(t, 0, items[0].(*models.VoiceItem).Frame)
				assert.Equal(t, 60, items[0].(*models.VoiceItem).Length)
				assert.Equal(t, 60, items[1].(*models.VoiceItem).Frame)
				assert.Equal(t, 40, items[1].(*models.VoiceItem).Length)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary files
			tmpDir := t.TempDir()
			ymmpsPath := filepath.Join(tmpDir, "test.ymmps")
			ymmpPath := filepath.Join(tmpDir, "template.ymmp")

			// Write test files
			err := os.WriteFile(ymmpsPath, []byte(tt.ymmpsContent), 0644)
			require.NoError(t, err)

			err = os.WriteFile(ymmpPath, []byte(tt.ymmpTemplate), 0644)
			require.NoError(t, err)

			// Parse files
			ymmpsParser := parser.NewYMMPSParser()
			ymmpsDoc, err := ymmpsParser.ParseFile(ymmpsPath)
			require.NoError(t, err)

			ymmpParser := parser.NewYMMPParser()
			template, err := ymmpParser.ParseFile(ymmpPath)
			require.NoError(t, err)

			// Convert
			conv := converter.NewConverter()
			result, err := conv.Convert(ymmpsDoc, template)
			require.NoError(t, err)

			// Check results
			assert.Len(t, result.Timelines, 1)
			timeline := result.Timelines[0]
			
			assert.Len(t, timeline.Items, tt.wantItems)
			assert.Equal(t, tt.wantLength, timeline.Length)

			if tt.checkItems != nil {
				tt.checkItems(t, timeline.Items)
			}
		})
	}
}

func TestRelativeLengthCalculations(t *testing.T) {
	tests := []struct {
		name         string
		ymmpsContent string
		expectError  bool
	}{
		{
			name: "until_scene_end",
			ymmpsContent: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - ID: scene1
        Shots:
          - Items:
              - Type: VoiceItem
                Template: voice1
                Length: "UNTIL SCENE END"
`,
			expectError: false,
		},
		{
			name: "until_sequence_end",
			ymmpsContent: `YMMPSVersion: "1"
Sequences:
  - ID: seq1
    Scenes:
      - Shots:
          - Items:
              - Type: VoiceItem
                Template: voice1
                Length: "UNTIL SEQUENCE END"
`,
			expectError: false,
		},
		{
			name: "until_specific_id",
			ymmpsContent: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - ID: shot1
            Items:
              - Type: VoiceItem
                Template: voice1
                Length: "100"
          - Items:
              - Type: VoiceItem
                Template: voice1
                Length: "UNTIL SHOT shot1 END"
`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic template
			tmplContent := `{
  "FilePath": "template.ymmp",
  "Timelines": [{
    "ID": "timeline1",
    "Items": [{
      "$type": "YukkuriMovieMaker.Project.VoiceItem",
      "Remark": "voice1"
    }]
  }]
}`

			// Create temporary files
			tmpDir := t.TempDir()
			ymmpsPath := filepath.Join(tmpDir, "test.ymmps")
			ymmpPath := filepath.Join(tmpDir, "template.ymmp")

			err := os.WriteFile(ymmpsPath, []byte(tt.ymmpsContent), 0644)
			require.NoError(t, err)

			err = os.WriteFile(ymmpPath, []byte(tmplContent), 0644)
			require.NoError(t, err)

			// Parse and convert
			ymmpsParser := parser.NewYMMPSParser()
			ymmpsDoc, err := ymmpsParser.ParseFile(ymmpsPath)
			require.NoError(t, err)

			ymmpParser := parser.NewYMMPParser()
			template, err := ymmpParser.ParseFile(ymmpPath)
			require.NoError(t, err)

			conv := converter.NewConverter()
			_, err = conv.Convert(ymmpsDoc, template)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCLIIntegration(t *testing.T) {
	// Create test files
	tmpDir := t.TempDir()
	ymmpsPath := filepath.Join(tmpDir, "scenario.ymmps")
	ymmpPath := filepath.Join(tmpDir, "template.ymmp")
	outputPath := filepath.Join(tmpDir, "output.ymmp")

	ymmpsContent := `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - Type: VoiceItem
                Template: voice_narration
                Length: "150"
                Properties:
                  Serif: "This is a test narration"
`

	ymmpContent := `{
  "FilePath": "template.ymmp",
  "SelectedTimelineIndex": 0,
  "Timelines": [{
    "ID": "timeline1",
    "Name": "Main Timeline",
    "VideoInfo": {
      "FPS": 30,
      "Width": 1920,
      "Height": 1080
    },
    "Items": [{
      "$type": "YukkuriMovieMaker.Project.VoiceItem",
      "Remark": "voice_narration",
      "Serif": "Default text",
      "Frame": 0,
      "Length": 100,
      "Layer": 0,
      "IsLocked": false,
      "IsSelected": false
    }]
  }]
}`

	err := os.WriteFile(ymmpsPath, []byte(ymmpsContent), 0644)
	require.NoError(t, err)

	err = os.WriteFile(ymmpPath, []byte(ymmpContent), 0644)
	require.NoError(t, err)

	// Test the conversion workflow
	ymmpsParser := parser.NewYMMPSParser()
	ymmpsDoc, err := ymmpsParser.ParseFile(ymmpsPath)
	require.NoError(t, err)

	ymmpParser := parser.NewYMMPParser()
	template, err := ymmpParser.ParseFile(ymmpPath)
	require.NoError(t, err)

	conv := converter.NewConverter()
	result, err := conv.Convert(ymmpsDoc, template)
	require.NoError(t, err)

	// Write output
	output, err := os.Create(outputPath)
	require.NoError(t, err)
	defer output.Close()

	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(result)
	require.NoError(t, err)

	// Verify output file exists and is valid
	_, err = os.Stat(outputPath)
	assert.NoError(t, err)

	// Parse output to verify
	outputProject, err := ymmpParser.ParseFile(outputPath)
	require.NoError(t, err)

	assert.Len(t, outputProject.Timelines, 1)
	assert.Len(t, outputProject.Timelines[0].Items, 1)
	
	voiceItem := outputProject.Timelines[0].Items[0].(*models.VoiceItem)
	assert.Equal(t, "This is a test narration", voiceItem.Serif)
	assert.Equal(t, 150, voiceItem.Length)
}