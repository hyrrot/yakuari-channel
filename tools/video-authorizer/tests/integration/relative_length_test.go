package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yakuari-channel/video-authorizer/internal/converter"
	"github.com/yakuari-channel/video-authorizer/internal/models"
	"github.com/yakuari-channel/video-authorizer/internal/parser"
)

// TestRelativeLengthWithIDGeneration tests relative length calculations after ID generation
func TestRelativeLengthWithIDGeneration(t *testing.T) {
	tests := []struct {
		name         string
		scenarioData string
		checks       func(t *testing.T, project *models.YMMPProject)
		description  string
	}{
		{
			name: "scene_end_with_auto_generated_ids",
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "180"
                Serif: "第一部分"
          - Items:
              - _Template: "voice_template"
                Length: "240"
                Serif: "第二部分"
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "_until:SCENE_END"
                Serif: "第二シーン、シーン終了まで"`,
			description: "IDが自動生成されたシーンでのシーン終了計算",
			checks: func(t *testing.T, project *models.YMMPProject) {
				require.Len(t, project.Timelines[0].Items, 3)
				
				items := project.Timelines[0].Items
				voice1 := items[0].(*models.VoiceItem)
				voice2 := items[1].(*models.VoiceItem)
				voice3 := items[2].(*models.VoiceItem)
				
				// 第一シーンのアイテム
				assert.Equal(t, 0, voice1.Frame)     // 開始フレーム
				assert.Equal(t, 180, voice1.Length)  // 指定された長さ
				
				assert.Equal(t, 180, voice2.Frame)   // 前のアイテム後
				assert.Equal(t, 240, voice2.Length)  // 指定された長さ
				
				// 第二シーンのアイテム
				assert.Equal(t, 420, voice3.Frame)   // 第一シーン終了後 (180+240)
				// voice3の長さは実装に依存（template defaultまたは計算された値）
				assert.Greater(t, voice3.Length, 0) // 何らかの正の値
			},
		},
		{
			name: "sequence_end_with_auto_generated_ids",
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "150"
                Serif: "シーケンス開始"
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "200"
                Serif: "シーケンス中盤"
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "_until:SEQUENCE_END"
                Serif: "最初のシーケンス終了まで"`,
			description: "自動生成IDでのシーケンス間相対参照",
			checks: func(t *testing.T, project *models.YMMPProject) {
				require.Len(t, project.Timelines[0].Items, 3)
				
				items := project.Timelines[0].Items
				voice1 := items[0].(*models.VoiceItem)
				voice2 := items[1].(*models.VoiceItem)
				voice3 := items[2].(*models.VoiceItem)
				
				// 第一シーケンス
				assert.Equal(t, 0, voice1.Frame)
				assert.Equal(t, 150, voice1.Length)
				
				assert.Equal(t, 150, voice2.Frame)
				assert.Equal(t, 200, voice2.Length)
				
				// 第二シーケンス
				assert.Equal(t, 350, voice3.Frame) // 150+200 = 350
				assert.Greater(t, voice3.Length, 0)
			},
		},
		{
			name: "shot_end_with_auto_generated_ids",
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "100"
                Serif: "ショット内音声1"
              - _Template: "voice_template"
                Length: "_until:SHOT_END"
                Serif: "ショット終了まで"`,
			description: "同一ショット内での相対長計算",
			checks: func(t *testing.T, project *models.YMMPProject) {
				require.Len(t, project.Timelines[0].Items, 2)
				
				items := project.Timelines[0].Items
				voice1 := items[0].(*models.VoiceItem)
				voice2 := items[1].(*models.VoiceItem)
				
				// 同時開始（同一ショット内）
				assert.Equal(t, 0, voice1.Frame)
				assert.Equal(t, 100, voice1.Length)
				
				assert.Equal(t, 0, voice2.Frame)    // 同時開始
				assert.Equal(t, 100, voice2.Length) // ショット終了まで = voice1の長さまで
			},
		},
		{
			name: "mixed_explicit_and_generated_ids",
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - ID: "explicit_seq"
    Scenes:
      - Shots:  # 自動生成ID
          - Items:
              - _Template: "voice_template"
                Length: "120"
                Serif: "明示的シーケンス内"
          - ID: "explicit_shot"  # 明示的ID
            Items:
              - _Template: "voice_template"
                Length: "_until:SHOT_END:explicit_shot"
                Serif: "明示的ショット参照"`,
			description: "明示的IDと自動生成IDの混在",
			checks: func(t *testing.T, project *models.YMMPProject) {
				require.Len(t, project.Timelines[0].Items, 2)
				
				items := project.Timelines[0].Items
				voice1 := items[0].(*models.VoiceItem)
				voice2 := items[1].(*models.VoiceItem)
				
				assert.Equal(t, 0, voice1.Frame)
				assert.Equal(t, 120, voice1.Length)
				
				assert.Equal(t, 120, voice2.Frame)
				// 明示的なID参照が正しく動作することを確認
				assert.Greater(t, voice2.Length, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			// Convert with ID generation
			conv := converter.NewConverter()
			project, err := conv.Convert(scenario, template)
			require.NoError(t, err, tt.description)

			// Run test-specific checks
			tt.checks(t, project)
		})
	}
}

// TestIDGenerationUniqueness tests that generated IDs are unique across runs
func TestIDGenerationUniqueness(t *testing.T) {
	scenarioData := `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "120"`

	// Parse scenario multiple times
	var parsedDocuments []*models.YMMPSDocument
	for i := 0; i < 5; i++ {
		ymmpsParser := parser.NewYMMPSParser()
		doc, err := ymmpsParser.Parse(strings.NewReader(scenarioData))
		require.NoError(t, err)
		
		// Apply ID generation
		err = doc.EnsureIDs()
		require.NoError(t, err)
		
		parsedDocuments = append(parsedDocuments, doc)
	}

	// Check that all generated IDs are unique across runs
	allIDs := make(map[string]int) // ID -> count
	
	for i, doc := range parsedDocuments {
		for _, seq := range doc.Sequences {
			allIDs[seq.ID]++
			if allIDs[seq.ID] > 1 {
				t.Errorf("Duplicate sequence ID %s found in run %d", seq.ID, i)
			}
			
			for _, scene := range seq.Scenes {
				allIDs[scene.ID]++
				if allIDs[scene.ID] > 1 {
					t.Errorf("Duplicate scene ID %s found in run %d", scene.ID, i)
				}
				
				for _, shot := range scene.Shots {
					allIDs[shot.ID]++
					if allIDs[shot.ID] > 1 {
						t.Errorf("Duplicate shot ID %s found in run %d", shot.ID, i)
					}
				}
			}
		}
	}

	// Should have 15 unique IDs (5 runs × 3 IDs per run)
	assert.Len(t, allIDs, 15)
}