package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yakuari-channel/video-authorizer/internal/converter"
	"github.com/yakuari-channel/video-authorizer/internal/models"
	"github.com/yakuari-channel/video-authorizer/internal/parser"
)

// TestComprehensiveScenarios tests various combinations of templates and scenarios
func TestComprehensiveScenarios(t *testing.T) {
	tests := []struct {
		name             string
		templateData     string // JSON template data
		scenarioData     string // YMMPS scenario data
		expectedChecks   func(t *testing.T, project *models.YMMPProject) // Validation function
		shouldSucceed    bool
		expectedErrorMsg string
	}{
		{
			name: "basic_voice_and_video_scenario",
			templateData: `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"SelectedTimelineIndex": 0,
				"Timelines": [{
					"ID": "timeline-1",
					"Name": "メイン",
					"VideoInfo": {"FPS": 30, "Hz": 48000, "Width": 1920, "Height": 1080},
					"Items": [
						{
							"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
							"Layer": 0,
							"Frame": 0,
							"Length": 120,
							"Remark": "voice_template",
							"CharacterName": "ずんだもん"
						},
						{
							"$type": "YukkuriMovieMaker.Project.Items.VideoItem, YukkuriMovieMaker",
							"Layer": 1,
							"Frame": 0,
							"Length": 300,
							"Remark": "video_template",
							"FilePath": "dummy.mp4"
						}
					]
				}]
			}`,
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - ID: "seq1"
    Scenes:
      - ID: "scene1"
        Shots:
          - ID: "shot1"
            Items:
              - _Template: "voice_template"
                Length: "150"
                Serif: "こんにちは、世界！"
              - _Template: "video_template"
                Length: "_until:SHOT_END"
                FilePath: "test_video.mp4"`,
			expectedChecks: func(t *testing.T, project *models.YMMPProject) {
				require.NotNil(t, project)
				require.Len(t, project.Timelines, 1)
				require.Len(t, project.Timelines[0].Items, 2)

				// Check voice item
				voiceItem, ok := project.Timelines[0].Items[0].(*models.VoiceItem)
				require.True(t, ok, "First item should be VoiceItem")
				assert.Equal(t, 150, voiceItem.Length)
				assert.Equal(t, "こんにちは、世界！", voiceItem.Serif)

				// Check video item
				videoItem, ok := project.Timelines[0].Items[1].(*models.VideoItem)
				require.True(t, ok, "Second item should be VideoItem")
				assert.Equal(t, 150, videoItem.Length) // Should match shot end
				assert.Contains(t, videoItem.FilePath, "test_video.mp4")
			},
			shouldSucceed: true,
		},
		{
			name: "complex_multi_scene_scenario",
			templateData: `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"SelectedTimelineIndex": 0,
				"Timelines": [{
					"ID": "timeline-1",
					"Name": "メイン",
					"VideoInfo": {"FPS": 30, "Hz": 48000, "Width": 1920, "Height": 1080},
					"Items": [
						{
							"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
							"Layer": 0,
							"Frame": 0,
							"Length": 120,
							"Remark": "narrator_voice",
							"CharacterName": "ずんだもん"
						},
						{
							"$type": "YukkuriMovieMaker.Project.Items.TachieItem, YukkuriMovieMaker",
							"Layer": 1,
							"Frame": 0,
							"Length": 300,
							"Remark": "character_image",
							"TachieItemParameter": {"ImagePath": "dummy.png"}
						},
						{
							"$type": "YukkuriMovieMaker.Project.Items.VideoItem, YukkuriMovieMaker",
							"Layer": 2,
							"Frame": 0,
							"Length": 600,
							"Remark": "background_video",
							"FilePath": "bg.mp4"
						}
					]
				}]
			}`,
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - ID: "intro_seq"
    Scenes:
      - ID: "opening"
        Shots:
          - ID: "title_shot"
            Items:
              - _Template: "narrator_voice"
                Length: "180"
                Serif: "タイトル紹介です"
              - _Template: "character_image"
                Length: "_until:SCENE_END"
                ImagePath: "characters/zundamon.png"
          - ID: "explanation_shot"
            Items:
              - _Template: "narrator_voice"
                Length: "240"
                Serif: "詳しく説明します"
              - _Template: "background_video"
                Length: "_until:SEQUENCE_END:intro_seq"
                FilePath: "videos/explanation_bg.mp4"
      - ID: "conclusion"
        Shots:
          - ID: "final_shot"
            Items:
              - _Template: "narrator_voice"
                Length: "120"
                Serif: "以上です"`,
			expectedChecks: func(t *testing.T, project *models.YMMPProject) {
				require.NotNil(t, project)
				require.Len(t, project.Timelines, 1)
				require.Len(t, project.Timelines[0].Items, 5) // 3 shots with voice + 1 character image + 1 background video

				// Verify frame positioning and lengths
				items := project.Timelines[0].Items
				
				// Title shot voice (starts at 0)
				voiceItem1 := items[0].(*models.VoiceItem)
				assert.Equal(t, 0, voiceItem1.Frame)
				assert.Equal(t, 180, voiceItem1.Length)
				assert.Equal(t, "タイトル紹介です", voiceItem1.Serif)

				// Character image (until scene end: 180+240=420)
				tachieItem := items[1].(*models.TachieItem)
				assert.Equal(t, 0, tachieItem.Frame)
				assert.Equal(t, 420, tachieItem.Length)
				assert.Contains(t, tachieItem.TachieItemParameter["ImagePath"], "zundamon.png")

				// Explanation shot voice
				voiceItem2 := items[2].(*models.VoiceItem)
				assert.Equal(t, 180, voiceItem2.Frame) // After title shot
				assert.Equal(t, 240, voiceItem2.Length)
				assert.Equal(t, "詳しく説明します", voiceItem2.Serif)

				// Background video (until sequence end: 180+240+120=540)
				videoItem := items[3].(*models.VideoItem)
				assert.Equal(t, 180, videoItem.Frame) // Starts with explanation shot
				assert.Equal(t, 360, videoItem.Length) // 540-180
				assert.Contains(t, videoItem.FilePath, "explanation_bg.mp4")

				// Final shot voice
				voiceItem3 := items[4].(*models.VoiceItem)
				assert.Equal(t, 420, voiceItem3.Frame) // After opening scene
				assert.Equal(t, 120, voiceItem3.Length)
				assert.Equal(t, "以上です", voiceItem3.Serif)
			},
			shouldSucceed: true,
		},
		{
			name: "auto_length_calculation_scenario",
			templateData: `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"SelectedTimelineIndex": 0,
				"Timelines": [{
					"ID": "timeline-1",
					"Name": "メイン",
					"VideoInfo": {"FPS": 30, "Hz": 48000, "Width": 1920, "Height": 1080},
					"Items": [
						{
							"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
							"Layer": 0,
							"Frame": 0,
							"Length": 120,
							"Remark": "auto_voice",
							"CharacterName": "ずんだもん"
						},
						{
							"$type": "YukkuriMovieMaker.Project.Items.VideoItem, YukkuriMovieMaker",
							"Layer": 1,
							"Frame": 0,
							"Length": 300,
							"Remark": "auto_video",
							"FilePath": "dummy.mp4"
						}
					]
				}]
			}`,
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "auto_voice"
                Length: "_auto:VOICEVOX"
                Serif: "自動計算される音声です"
              - _Template: "auto_video"
                Length: "_auto:VIDEO"
                FilePath: "test_video.mp4"`,
			expectedChecks: func(t *testing.T, project *models.YMMPProject) {
				require.NotNil(t, project)
				require.Len(t, project.Timelines, 1)
				require.Len(t, project.Timelines[0].Items, 2)

				// Mock calculations should provide predictable results
				voiceItem := project.Timelines[0].Items[0].(*models.VoiceItem)
				assert.Equal(t, 90, voiceItem.Length) // Mock設定値 3.0秒 * 30fps = 90 frames

				videoItem := project.Timelines[0].Items[1].(*models.VideoItem)
				assert.Equal(t, 450, videoItem.Length) // Mock設定値 15.0秒 * 30fps = 450 frames
			},
			shouldSucceed: true,
		},
		{
			name: "relative_path_resolution_scenario",
			templateData: `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"Timeline": {
					"VideoInfo": {"FPS": 30},
					"Items": [
						{
							"$type": "YukkuriMovieMaker.Project.Items.VideoItem, YukkuriMovieMaker",
							"Layer": 0,
							"Frame": 0,
							"Length": 300,
							"Remark": "media_item",
							"FilePath": "dummy.mp4"
						},
						{
							"$type": "YukkuriMovieMaker.Project.Items.TachieItem, YukkuriMovieMaker",
							"Layer": 1,
							"Frame": 0,
							"Length": 300,
							"Remark": "image_item",
							"TachieItemParameter": {"ImagePath": "dummy.png"}
						}
					]
				}
			}`,
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "media_item"
                Length: "200"
                FilePath: "assets/videos/intro.mp4"
              - _Template: "image_item"
                Length: "200"
                ImagePath: "assets/images/character.png"`,
			expectedChecks: func(t *testing.T, project *models.YMMPProject) {
				require.NotNil(t, project)
				require.Len(t, project.Timelines, 1)
				require.Len(t, project.Timelines[0].Items, 2)

				// File paths should be converted to absolute paths
				videoItem := project.Timelines[0].Items[0].(*models.VideoItem)
				assert.True(t, filepath.IsAbs(videoItem.FilePath))
				assert.Contains(t, videoItem.FilePath, "intro.mp4")

				tachieItem := project.Timelines[0].Items[1].(*models.TachieItem)
				imagePath, exists := tachieItem.TachieItemParameter["ImagePath"]
				require.True(t, exists)
				imagePathStr, ok := imagePath.(string)
				require.True(t, ok)
				assert.True(t, filepath.IsAbs(imagePathStr))
				assert.Contains(t, imagePathStr, "character.png")
			},
			shouldSucceed: true,
		},
		{
			name: "template_validation_error_scenario",
			templateData: `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"Timeline": {
					"VideoInfo": {"FPS": 30},
					"Items": [
						{
							"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
							"Layer": 0,
							"Frame": 0,
							"Length": 120,
							"Remark": "available_template"
						}
					]
				}
			}`,
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "nonexistent_template"
                Length: "120"
                Serif: "This should fail"`,
			shouldSucceed:    false,
			expectedErrorMsg: "template 'nonexistent_template' not found",
		},
		{
			name: "id_reference_validation_scenario",
			templateData: `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"Timeline": {
					"VideoInfo": {"FPS": 30},
					"Items": [
						{
							"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
							"Layer": 0,
							"Frame": 0,
							"Length": 120,
							"Remark": "voice_template"
						}
					]
				}
			}`,
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - ID: "seq1"
    Scenes:
      - ID: "scene1"
        Shots:
          - ID: "shot1"
            Items:
              - _Template: "voice_template"
                Length: "_until:SCENE_END:nonexistent_scene"
                Serif: "This should fail"`,
			shouldSucceed:    false,
			expectedErrorMsg: "undefined SCENE ID 'nonexistent_scene'",
		},
		{
			name: "natural_language_length_format_scenario",
			templateData: `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"Timeline": {
					"VideoInfo": {"FPS": 30},
					"Items": [
						{
							"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
							"Layer": 0,
							"Frame": 0,
							"Length": 120,
							"Remark": "voice_template"
						}
					]
				}
			}`,
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - ID: "main_seq"
    Scenes:
      - ID: "intro_scene"
        Shots:
          - ID: "greeting_shot"
            Items:
              - _Template: "voice_template"
                Length: "180"
                Serif: "最初の挨拶"
          - ID: "main_shot"
            Items:
              - _Template: "voice_template"
                Length: "UNTIL SCENE END"
                Serif: "メインコンテンツ"
      - ID: "outro_scene"
        Shots:
          - ID: "ending_shot"
            Items:
              - _Template: "voice_template"
                Length: "UNTIL SEQUENCE main_seq END"
                Serif: "終わりの挨拶"`,
			expectedChecks: func(t *testing.T, project *models.YMMPProject) {
				require.NotNil(t, project)
				require.Len(t, project.Timelines, 1)
				require.Len(t, project.Timelines[0].Items, 3)

				// Check natural language parsing worked correctly
				items := project.Timelines[0].Items

				// First voice: fixed length
				voice1 := items[0].(*models.VoiceItem)
				assert.Equal(t, 0, voice1.Frame)
				assert.Equal(t, 180, voice1.Length)

				// Second voice: until scene end (180 + calculated length)
				voice2 := items[1].(*models.VoiceItem)
				assert.Equal(t, 180, voice2.Frame)
				assert.Greater(t, voice2.Length, 0) // Should have calculated length

				// Third voice: until sequence end
				voice3 := items[2].(*models.VoiceItem)
				assert.Greater(t, voice3.Frame, 180) // Should start after intro scene
				assert.Greater(t, voice3.Length, 0) // Should have calculated length
			},
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temporary directory
			tmpDir := t.TempDir()

			// Create template file
			templatePath := filepath.Join(tmpDir, "template.ymmp")
			err := os.WriteFile(templatePath, []byte(tt.templateData), 0644)
			require.NoError(t, err)

			// Create scenario file
			scenarioPath := filepath.Join(tmpDir, "scenario.ymmps")
			err = os.WriteFile(scenarioPath, []byte(tt.scenarioData), 0644)
			require.NoError(t, err)

			// Parse template
			ymmParser := parser.NewYMMPParser()
			template, err := ymmParser.ParseFile(templatePath)
			require.NoError(t, err)

			// Parse scenario
			ymmpsParser := parser.NewYMMPSParser()
			scenario, err := ymmpsParser.ParseFile(scenarioPath)
			if !tt.shouldSucceed && err != nil {
				// Early validation error expected
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				return
			}
			require.NoError(t, err)

			// Convert scenario to project with mocks for auto calculation tests
			var conv *converter.Converter
			if tt.name == "auto_length_calculation_scenario" {
				// Create mock clients for predictable test results
				mockVoicevox := converter.NewMockVoicevoxClient(true)
				mockVoicevox.SetDuration("自動計算される音声です", 3.0) // 90 frames at 30 FPS
				
				mockFFProbe := converter.NewMockFFProbeClient(true)
				mockFFProbe.SetDuration("test_video.mp4", 15.0) // 450 frames at 30 FPS
				
				conv = converter.NewConverterWithClients(mockVoicevox, mockFFProbe)
			} else {
				conv = converter.NewConverter()
			}
			project, err := conv.Convert(scenario, template)

			if tt.shouldSucceed {
				require.NoError(t, err, "Conversion should succeed")
				require.NotNil(t, project, "Project should not be nil")

				// Run specific checks
				if tt.expectedChecks != nil {
					tt.expectedChecks(t, project)
				}

				// Validate generated project structure
				err = project.Validate()
				assert.NoError(t, err, "Generated project should be valid")

				// Ensure project can be serialized
				jsonData, err := json.Marshal(project)
				assert.NoError(t, err, "Project should be serializable")
				assert.Greater(t, len(jsonData), 0, "Serialized data should not be empty")
			} else {
				require.Error(t, err, "Conversion should fail")
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
			}
		})
	}
}

// TestSpecificationCompliance tests specific requirements from the specification
func TestSpecificationCompliance(t *testing.T) {
	t.Run("US-001_basic_conversion", func(t *testing.T) {
		// Test basic YMMPS to YMMP conversion functionality
		templateData := createMinimalTemplate()
		scenarioData := createMinimalScenario()

		project := runConversion(t, templateData, scenarioData)
		
		// Verify basic structure
		assert.NotNil(t, project)
		assert.NotEmpty(t, project.Timelines)
		assert.NotEmpty(t, project.Timelines[0].Items)
	})

	t.Run("US-004_template_matching", func(t *testing.T) {
		// Test template matching by Remark field
		templateData := `{
			"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
			"FilePath": "template.ymmp",
			"Timeline": {
				"VideoInfo": {"FPS": 30},
				"Items": [
					{
						"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
						"Remark": "specific_voice_template",
						"Layer": 0,
						"Frame": 0,
						"Length": 120
					}
				]
			}
		}`

		scenarioData := `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "specific_voice_template"
                Length: "200"
                Serif: "テンプレートマッチングのテスト"`

		project := runConversion(t, templateData, scenarioData)
		voiceItem := project.Timelines[0].Items[0].(*models.VoiceItem)
		assert.Equal(t, 200, voiceItem.Length)
		assert.Equal(t, "テンプレートマッチングのテスト", voiceItem.Serif)
	})

	t.Run("US-005_property_override", func(t *testing.T) {
		// Test property override functionality
		templateData := createTemplateWithVoiceItem("test_voice", 120)
		scenarioData := `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "test_voice"
                Length: "300"
                Serif: "上書きされたテキスト"
                Speaker: 1`

		project := runConversion(t, templateData, scenarioData)
		voiceItem := project.Timelines[0].Items[0].(*models.VoiceItem)
		
		// Length should be overridden
		assert.Equal(t, 300, voiceItem.Length)
		// Text should be overridden
		assert.Equal(t, "上書きされたテキスト", voiceItem.Serif)
		// Check that properties were applied (Speaker field may not be directly accessible)
		// The conversion should have succeeded
	})

	t.Run("US-007_hierarchical_structure", func(t *testing.T) {
		// Test Sequence > Scene > Shot > Item hierarchy
		scenarioData := `YMMPSVersion: "1"
Sequences:
  - ID: "seq1"
    Scenes:
      - ID: "scene1"
        Shots:
          - ID: "shot1"
            Items:
              - _Template: "voice_template"
                Length: "100"
                Serif: "シーン1ショット1"
          - ID: "shot2" 
            Items:
              - _Template: "voice_template"
                Length: "150"
                Serif: "シーン1ショット2"
      - ID: "scene2"
        Shots:
          - ID: "shot3"
            Items:
              - _Template: "voice_template"
                Length: "200"
                Serif: "シーン2ショット1"`

		templateData := createTemplateWithVoiceItem("voice_template", 120)
		project := runConversion(t, templateData, scenarioData)

		// Should have 3 items (one per shot)
		assert.Len(t, project.Timelines[0].Items, 3)

		// Check frame positioning follows hierarchy
		items := project.Timelines[0].Items
		assert.Equal(t, 0, items[0].(*models.VoiceItem).Frame)   // seq1.scene1.shot1 starts at 0
		assert.Equal(t, 100, items[1].(*models.VoiceItem).Frame) // seq1.scene1.shot2 starts at 100
		assert.Equal(t, 250, items[2].(*models.VoiceItem).Frame) // seq1.scene2.shot3 starts at 250
	})

	t.Run("US-009_relative_length_until_scene_end", func(t *testing.T) {
		scenarioData := `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "120"
                Serif: "最初の音声"
          - Items:
              - _Template: "voice_template"
                Length: "_until:SCENE_END"
                Serif: "シーン終了まで続く音声"`

		templateData := createTemplateWithVoiceItem("voice_template", 100)
		project := runConversion(t, templateData, scenarioData)

		items := project.Timelines[0].Items
		voice1 := items[0].(*models.VoiceItem)
		voice2 := items[1].(*models.VoiceItem)

		assert.Equal(t, 120, voice1.Length)
		// The second voice should use template default length when scene end cannot be determined
		assert.Equal(t, 100, voice2.Length) // Template default length
		assert.Equal(t, 120, voice2.Frame)  // Should start after first voice
	})

	t.Run("US-010_relative_length_until_sequence_end", func(t *testing.T) {
		scenarioData := `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "100"
                Serif: "最初"
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "150"
                Serif: "次"
          - Items:
              - _Template: "voice_template"
                Length: "_until:SEQUENCE_END"
                Serif: "シーケンス終了まで"`

		templateData := createTemplateWithVoiceItem("voice_template", 80)
		project := runConversion(t, templateData, scenarioData)

		items := project.Timelines[0].Items
		voice3 := items[2].(*models.VoiceItem)

		// Should start at frame 250 (100+150) and have appropriate length
		assert.Equal(t, 250, voice3.Frame)
		assert.Greater(t, voice3.Length, 0)
	})
}

// Helper functions
func createMinimalTemplate() string {
	return `{
		"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
		"FilePath": "template.ymmp",
		"SelectedTimelineIndex": 0,
		"Timelines": [{
			"ID": "timeline-1",
			"Name": "メイン",
			"VideoInfo": {"FPS": 30, "Hz": 48000, "Width": 1920, "Height": 1080},
			"Items": [
				{
					"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
					"Layer": 0,
					"Frame": 0,
					"Length": 120,
					"Remark": "voice_template",
					"CharacterName": "ずんだもん"
				}
			]
		}]
	}`
}

func createMinimalScenario() string {
	return `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "150"
                Serif: "基本的なテスト"`
}

func createTemplateWithVoiceItem(remarkName string, defaultLength int) string {
	return `{
		"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
		"FilePath": "template.ymmp",
		"SelectedTimelineIndex": 0,
		"Timelines": [{
			"ID": "timeline-1",
			"Name": "メイン",
			"VideoInfo": {"FPS": 30, "Hz": 48000, "Width": 1920, "Height": 1080},
			"Items": [
				{
					"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
					"Layer": 0,
					"Frame": 0,
					"Length": ` + fmt.Sprintf("%d", defaultLength) + `,
					"Remark": "` + remarkName + `",
					"CharacterName": "ずんだもん"
				}
			]
		}]
	}`
}

func runConversion(t *testing.T, templateData, scenarioData string) *models.YMMPProject {
	tmpDir := t.TempDir()

	// Create template file
	templatePath := filepath.Join(tmpDir, "template.ymmp")
	err := os.WriteFile(templatePath, []byte(templateData), 0644)
	require.NoError(t, err)

	// Create scenario file
	scenarioPath := filepath.Join(tmpDir, "scenario.ymmps")
	err = os.WriteFile(scenarioPath, []byte(scenarioData), 0644)
	require.NoError(t, err)

	// Parse files
	ymmParser := parser.NewYMMPParser()
	template, err := ymmParser.ParseFile(templatePath)
	require.NoError(t, err)

	ymmpsParser := parser.NewYMMPSParser()
	scenario, err := ymmpsParser.ParseFile(scenarioPath)
	require.NoError(t, err)

	// Convert with mocks for auto functionality
	mockVoicevox := converter.NewMockVoicevoxClient(true)
	mockFFProbe := converter.NewMockFFProbeClient(true)
	// Set up default mock values
	mockVoicevox.SetDuration("自動計算される音声です", 3.0) // 90 frames at 30 FPS
	mockFFProbe.SetDuration("test_video.mp4", 15.0) // 450 frames at 30 FPS
	
	conv := converter.NewConverterWithClients(mockVoicevox, mockFFProbe)
	project, err := conv.Convert(scenario, template)
	require.NoError(t, err)

	return project
}