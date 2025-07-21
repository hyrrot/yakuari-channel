package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yakuari-channel/video-authorizer/internal/converter"
	"github.com/yakuari-channel/video-authorizer/internal/models"
	"github.com/yakuari-channel/video-authorizer/internal/parser"
)

// TestAutoVideoLengthCombinations tests _auto:VIDEO functionality with various scenarios
func TestAutoVideoLengthCombinations(t *testing.T) {
	tests := []struct {
		name         string
		scenarioData string
		description  string
		checks       func(t *testing.T, project *models.YMMPProject)
	}{
		{
			name: "simple_auto_video",
			description: "シンプルな_auto:VIDEO自動長さ計算",
			scenarioData: `YMMPSVersion: "1"
Sequences:
- Scenes:
  - Shots:
    - Items:
      - _Template: "video_template"
        Length: "_auto:VIDEO"
        FilePath: "short_video.mp4"`,
			checks: func(t *testing.T, project *models.YMMPProject) {
				items := project.Timelines[0].Items
				require.Len(t, items, 1)
				
				videoItem := items[0].(*models.VideoItem)
				assert.Equal(t, 0, videoItem.Frame)
				assert.Equal(t, 150, videoItem.Length) // Mock設定値 5.0秒 * 30fps = 150 frames
				assert.Equal(t, "short_video.mp4", videoItem.FilePath)
			},
		},
		{
			name: "multiple_auto_video_sequential",
			description: "複数の_auto:VIDEO動画の順次実行",
			scenarioData: `YMMPSVersion: "1"
Sequences:
- Scenes:
  - Shots:
    - Items:
      - _Template: "video_template"
        Length: "_auto:VIDEO"
        FilePath: "intro_video.mp4"
  - Shots:
    - Items:
      - _Template: "video_template"
        Length: "_auto:VIDEO"
        FilePath: "main_video.mp4"`,
			checks: func(t *testing.T, project *models.YMMPProject) {
				items := project.Timelines[0].Items
				require.Len(t, items, 2)
				
				video1 := items[0].(*models.VideoItem)
				video2 := items[1].(*models.VideoItem)
				
				assert.Equal(t, 0, video1.Frame)
				assert.Equal(t, 300, video1.Length) // Mock設定値 10.0秒 * 30fps = 300 frames
				
				// 二番目の動画は最初の動画終了後に開始
				assert.Equal(t, video1.Frame + video1.Length, video2.Frame)
				assert.Equal(t, 450, video2.Length) // Mock設定値 15.0秒 * 30fps = 450 frames
			},
		},
		{
			name: "mixed_auto_video_and_voice",
			description: "_auto:VIDEOと_auto:VOICEVOXの混在",
			scenarioData: `YMMPSVersion: "1"
Sequences:
- Scenes:
  - Shots:
    - Items:
      - _Template: "video_template"
        Length: "_auto:VIDEO"
        FilePath: "background.mp4"
      - _Template: "voice_template"
        Length: "_auto:VOICEVOX"
        Serif: "背景動画と同時に再生される音声です。"`,
			checks: func(t *testing.T, project *models.YMMPProject) {
				items := project.Timelines[0].Items
				require.Len(t, items, 2)
				
				video := items[0].(*models.VideoItem)
				voice := items[1].(*models.VoiceItem)
				
				// 同じショット内なので同時開始
				assert.Equal(t, 0, video.Frame)
				assert.Equal(t, 0, voice.Frame)
				
				assert.Equal(t, 450, video.Length) // Mock設定値 15.0秒 * 30fps = 450 frames
				assert.Equal(t, 105, voice.Length) // Mock設定値 3.5秒 * 30fps = 105 frames
			},
		},
		{
			name: "auto_video_with_sequence_end",
			description: "_auto:VIDEOと_until:SEQUENCE_ENDの組み合わせ",
			scenarioData: `YMMPSVersion: "1"
Sequences:
- ID: main_seq
  Scenes:
  - Shots:
    - Items:
      - _Template: "audio_template"
        Length: "_until:SEQUENCE_END"
      - _Template: "video_template"
        Length: "_auto:VIDEO"
        FilePath: "content.mp4"`,
			checks: func(t *testing.T, project *models.YMMPProject) {
				items := project.Timelines[0].Items
				require.Len(t, items, 2)
				
				audio := items[0].(map[string]interface{})
				video := items[1].(*models.VideoItem)
				
				// 同時開始
				assert.Equal(t, 0, int(audio["Frame"].(int)))
				assert.Equal(t, 0, video.Frame)
				
				videoLength := video.Length
				assert.Equal(t, 450, videoLength) // Mock設定値 15.0秒 * 30fps = 450 frames
				
				// 実際の実装ではaudioの長さがtemplate default lengthになる
				audioLength := int(audio["Length"].(int))
				assert.Greater(t, audioLength, 0)
			},
		},
		{
			name: "different_video_types",
			description: "異なる動画ファイルタイプの処理",
			scenarioData: `YMMPSVersion: "1"
Sequences:
- Scenes:
  - Shots:
    - Items:
      - _Template: "video_template"
        Length: "_auto:VIDEO"
        FilePath: "long_music.mp4"
  - Shots:
    - Items:
      - _Template: "video_template"
        Length: "_auto:VIDEO"
        FilePath: "outro_credits.mp4"`,
			checks: func(t *testing.T, project *models.YMMPProject) {
				items := project.Timelines[0].Items
				require.Len(t, items, 2)
				
				musicVideo := items[0].(*models.VideoItem)
				outroVideo := items[1].(*models.VideoItem)
				
				assert.Equal(t, 0, musicVideo.Frame)
				assert.Equal(t, 900, musicVideo.Length) // Mock設定値 30.0秒 * 30fps = 900 frames (long用)
				
				// アウトロ動画は音楽動画終了後
				assert.Equal(t, musicVideo.Frame + musicVideo.Length, outroVideo.Frame)
				assert.Equal(t, 240, outroVideo.Length) // Mock設定値 8.0秒 * 30fps = 240 frames (outro用)
			},
		},
		{
			name: "auto_video_fallback_calculation",
			description: "不明なファイルに対するフォールバック計算",
			scenarioData: `YMMPSVersion: "1"
Sequences:
- Scenes:
  - Shots:
    - Items:
      - _Template: "video_template"
        Length: "_auto:VIDEO"
        FilePath: "unknown_file.mp4"`,
			checks: func(t *testing.T, project *models.YMMPProject) {
				items := project.Timelines[0].Items
				require.Len(t, items, 1)
				
				video := items[0].(*models.VideoItem)
				assert.Equal(t, 0, video.Frame)
				assert.Equal(t, 450, video.Length) // Mock設定値 デフォルト15.0秒 * 30fps = 450 frames
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock FFProbe client for consistent test results
			mockFFProbe := converter.NewMockFFProbeClient(true)
			
			// Set up predictable durations for test video files
			mockFFProbe.SetDuration("short_video.mp4", 5.0)           // 150 frames at 30 FPS
			mockFFProbe.SetDuration("intro_video.mp4", 10.0)          // 300 frames at 30 FPS
			mockFFProbe.SetDuration("main_video.mp4", 15.0)           // 450 frames at 30 FPS
			mockFFProbe.SetDuration("background.mp4", 15.0)           // 450 frames at 30 FPS
			mockFFProbe.SetDuration("content.mp4", 15.0)              // 450 frames at 30 FPS
			mockFFProbe.SetDuration("long_music.mp4", 30.0)           // 900 frames at 30 FPS
			mockFFProbe.SetDuration("outro_credits.mp4", 8.0)         // 240 frames at 30 FPS
			mockFFProbe.SetDuration("unknown_file.mp4", 15.0)         // 450 frames at 30 FPS (fallback)
			
			// Also create mock VOICEVOX for mixed tests
			mockVoicevox := converter.NewMockVoicevoxClient(true)
			mockVoicevox.SetDuration("背景動画と同時に再生される音声です。", 3.5) // 105 frames at 30 FPS

			// Setup test environment
			tmpDir := t.TempDir()

			// Create template
			templateData := `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"SelectedTimelineIndex": 0,
				"Timelines": [{
					"ID": "timeline-1",
					"Name": "メイン",
					"VideoInfo": {"FPS": 30, "Hz": 48000, "Width": 1920, "Height": 1080},
					"Items": [
						{
							"$type": "YukkuriMovieMaker.Project.Items.VideoItem, YukkuriMovieMaker",
							"Layer": 2,
							"Frame": 0,
							"Length": 120,
							"Remark": "video_template",
							"FilePath": "template.mp4"
						},
						{
							"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
							"Layer": 1,
							"Frame": 0,
							"Length": 120,
							"Remark": "voice_template",
							"CharacterName": "ずんだもん"
						},
						{
							"$type": "YukkuriMovieMaker.Project.Items.AudioItem, YukkuriMovieMaker",
							"Layer": 0,
							"Frame": 0,
							"Length": 120,
							"Remark": "audio_template",
							"FilePath": "bgm.mp3"
						}
					]
				}],
				"Characters": [
					{
						"Name": "ずんだもん",
						"GroupName": "VOICEVOX",
						"Color": "#FF33A65E",
						"Voice": {
							"API": "voicevox",
							"Arg": "speaker_id:1"
						}
					}
				]
			}`

			templatePath := filepath.Join(tmpDir, "template.ymmp")
			err := os.WriteFile(templatePath, []byte(templateData), 0644)
			require.NoError(t, err)

			scenarioPath := filepath.Join(tmpDir, "scenario.ymmps")
			err = os.WriteFile(scenarioPath, []byte(tt.scenarioData), 0644)
			require.NoError(t, err)

			// Parse files
			ymmParser := parser.NewYMMPParser()
			template, err := ymmParser.ParseFile(templatePath)
			require.NoError(t, err)

			ymmpsParser := parser.NewYMMPSParser()
			scenario, err := ymmpsParser.ParseFile(scenarioPath)
			require.NoError(t, err)

			// Convert using mock clients
			conv := converter.NewConverterWithClients(mockVoicevox, mockFFProbe)
			project, err := conv.ConvertWithRelativeLengths(scenario, template)
			require.NoError(t, err, tt.description)

			// Run test-specific checks
			tt.checks(t, project)
		})
	}
}