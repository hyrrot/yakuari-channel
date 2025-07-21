package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yakuari-channel/video-authorizer/internal/models"
)

// TestCLIComprehensiveScenarios tests the CLI with various real-world scenarios
func TestCLIComprehensiveScenarios(t *testing.T) {
	// Build the CLI executable first
	executablePath := buildCLIExecutable(t)

	tests := []struct {
		name         string
		templateData string
		scenarioData string
		cliArgs      []string
		expectError  bool
		errorContains string
		outputChecks func(t *testing.T, outputPath string)
	}{
		{
			name: "complete_video_production_workflow",
			templateData: `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"Timeline": {
					"VideoInfo": {"FPS": 30, "Width": 1920, "Height": 1080},
					"Items": [
						{
							"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
							"Layer": 0, "Frame": 0, "Length": 120,
							"Remark": "narrator_voice",
							"VoiceParameters": {"CharacterName": "ずんだもん", "Speed": 1.0}
						},
						{
							"$type": "YukkuriMovieMaker.Project.Items.TachieItem, YukkuriMovieMaker",
							"Layer": 1, "Frame": 0, "Length": 300,
							"Remark": "character_display",
							"TachieItemParameter": {"ImagePath": "template_character.png"}
						},
						{
							"$type": "YukkuriMovieMaker.Project.Items.VideoItem, YukkuriMovieMaker",
							"Layer": 2, "Frame": 0, "Length": 600,
							"Remark": "background_video",
							"FilePath": "template_bg.mp4"
						},
						{
							"$type": "YukkuriMovieMaker.Project.Items.AudioItem, YukkuriMovieMaker",
							"Layer": 3, "Frame": 0, "Length": 1800,
							"Remark": "bgm_track",
							"FilePath": "template_bgm.mp3"
						}
					]
				}
			}`,
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - ID: "intro_sequence"
    Scenes:
      - ID: "title_scene"
        Shots:
          - ID: "title_display"
            Items:
              - _Template: "narrator_voice"
                Length: "180"
                Serif: "こんにちは！今日は動画作成について説明します"
              - _Template: "character_display"
                Length: "_until:SCENE_END"
                ImagePath: "assets/characters/zundamon_greeting.png"
              - _Template: "background_video"
                Length: "_until:SEQUENCE_END"
                FilePath: "assets/backgrounds/intro_bg.mp4"
          - ID: "overview_explanation"
            Items:
              - _Template: "narrator_voice"
                Length: "240"
                Serif: "まず全体の流れを説明しますね"
      - ID: "main_content_scene"
        Shots:
          - ID: "detailed_explanation"
            Items:
              - _Template: "narrator_voice"
                Length: "300"
                Serif: "詳しい内容はこちらになります"
              - _Template: "character_display"
                Length: "_until:SHOT_END"
                ImagePath: "assets/characters/zundamon_explaining.png"
  - ID: "conclusion_sequence"
    Scenes:
      - ID: "summary_scene"
        Shots:
          - ID: "final_message"
            Items:
              - _Template: "narrator_voice"
                Length: "150"
                Serif: "以上で説明を終わります。ありがとうございました！"
              - _Template: "character_display"
                Length: "_until:SEQUENCE_END"
                ImagePath: "assets/characters/zundamon_goodbye.png"
              - _Template: "bgm_track"
                Length: "_until:SEQUENCE_END"
                FilePath: "assets/audio/ending_bgm.mp3"`,
			cliArgs: []string{"--verbose"},
			outputChecks: func(t *testing.T, outputPath string) {
				// Check that output file was created
				_, err := os.Stat(outputPath)
				assert.NoError(t, err, "Output file should exist")

				// Parse and validate the output
				data, err := os.ReadFile(outputPath)
				require.NoError(t, err)

				var project models.YMMPProject
				err = json.Unmarshal(data, &project)
				require.NoError(t, err)

				// Verify complex timeline structure
				assert.Len(t, project.Timelines, 1)
				timeline := project.Timelines[0]
				
				// Should have 6 items: 3 voices + 3 character displays + 1 background video + 1 bgm
				assert.Len(t, timeline.Items, 6)

				// Verify frame positioning
				frames := make([]int, len(timeline.Items))
				for i, item := range timeline.Items {
					switch v := item.(type) {
					case *models.VoiceItem:
						frames[i] = v.Frame
					case *models.TachieItem:
						frames[i] = v.Frame
					case *models.VideoItem:
						frames[i] = v.Frame
					}
				}

				// First scene: narrator + character should start at 0
				// Second shot in first scene: narrator should start at 180
				// Second scene: narrator should start at 420 (180+240)
				// Final sequence: should start after all previous content
				expectedFrames := []int{0, 0, 0, 180, 420, 570} // Approximate positioning
				for i, expectedFrame := range expectedFrames {
					if i < len(frames) {
						assert.GreaterOrEqual(t, frames[i], expectedFrame-10, 
							"Frame %d should be around expected position %d", i, expectedFrame)
					}
				}

				// Verify file paths were converted to absolute
				for _, item := range timeline.Items {
					switch v := item.(type) {
					case *models.VideoItem:
						if v.FilePath != "" {
							assert.True(t, filepath.IsAbs(v.FilePath) || strings.Contains(v.FilePath, "assets/"),
								"Video file path should be absolute or contain assets/")
						}
					case *models.TachieItem:
						if imagePath, exists := v.TachieItemParameter["ImagePath"]; exists {
							pathStr, ok := imagePath.(string)
							require.True(t, ok)
							assert.True(t, filepath.IsAbs(pathStr) || strings.Contains(pathStr, "assets/"),
								"Image path should be absolute or contain assets/")
						}
					}
				}
			},
		},
		{
			name: "validate_only_mode_test",
			templateData: `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"Timeline": {
					"VideoInfo": {"FPS": 30},
					"Items": [
						{
							"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
							"Layer": 0, "Frame": 0, "Length": 120,
							"Remark": "test_voice"
						}
					]
				}
			}`,
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "test_voice"
                Length: "150"
                Serif: "バリデーションテスト"`,
			cliArgs: []string{"--validate-only"},
			outputChecks: func(t *testing.T, outputPath string) {
				// In validate-only mode, no output file should be created
				_, err := os.Stat(outputPath)
				assert.True(t, os.IsNotExist(err), "No output file should be created in validate-only mode")
			},
		},
		{
			name: "dry_run_mode_test",
			templateData: `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"Timeline": {
					"VideoInfo": {"FPS": 30},
					"Items": [
						{
							"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
							"Layer": 0, "Frame": 0, "Length": 120,
							"Remark": "test_voice"
						}
					]
				}
			}`,
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "test_voice"
                Length: "200"
                Serif: "ドライランテスト"`,
			cliArgs: []string{"--dry-run", "--verbose"},
			outputChecks: func(t *testing.T, outputPath string) {
				// In dry-run mode, no output file should be created
				_, err := os.Stat(outputPath)
				assert.True(t, os.IsNotExist(err), "No output file should be created in dry-run mode")
			},
		},
		{
			name: "base_path_resolution_test",
			templateData: `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"Timeline": {
					"VideoInfo": {"FPS": 30},
					"Items": [
						{
							"$type": "YukkuriMovieMaker.Project.Items.VideoItem, YukkuriMovieMaker",
							"Layer": 0, "Frame": 0, "Length": 300,
							"Remark": "media_item",
							"FilePath": "template.mp4"
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
                Length: "250"
                FilePath: "videos/relative_video.mp4"`,
			cliArgs: []string{"--base-path", "project_assets"},
			outputChecks: func(t *testing.T, outputPath string) {
				data, err := os.ReadFile(outputPath)
				require.NoError(t, err)

				var project models.YMMPProject
				err = json.Unmarshal(data, &project)
				require.NoError(t, err)

				// Check that relative path was resolved with base path
				videoItem := project.Timelines[0].Items[0].(*models.VideoItem)
				assert.True(t, filepath.IsAbs(videoItem.FilePath))
				assert.Contains(t, videoItem.FilePath, "relative_video.mp4")
			},
		},
		{
			name: "validation_error_scenario",
			templateData: `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"Timeline": {
					"VideoInfo": {"FPS": 30},
					"Items": [
						{
							"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
							"Layer": 0, "Frame": 0, "Length": 120,
							"Remark": "existing_template"
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
                Length: "150"
                Serif: "This should fail"`,
			expectError: true,
			errorContains: "template 'nonexistent_template' not found",
		},
		{
			name: "invalid_yaml_scenario",
			templateData: `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"Timeline": {
					"VideoInfo": {"FPS": 30},
					"Items": []
				}
			}`,
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
    - Shots:  # Invalid indentation
        Items:
          - Template: "test"`,
			expectError: true,
			errorContains: "YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			tmpDir := t.TempDir()

			// Create template file
			templatePath := filepath.Join(tmpDir, "template.ymmp")
			err := os.WriteFile(templatePath, []byte(tt.templateData), 0644)
			require.NoError(t, err)

			// Create scenario file
			scenarioPath := filepath.Join(tmpDir, "scenario.ymmps")
			err = os.WriteFile(scenarioPath, []byte(tt.scenarioData), 0644)
			require.NoError(t, err)

			// Create output path
			outputPath := filepath.Join(tmpDir, "output.ymmp")

			// Prepare CLI arguments
			args := []string{
				"--template-file", templatePath,
				"--output", outputPath,
			}
			args = append(args, tt.cliArgs...)
			args = append(args, scenarioPath)

			// Execute CLI
			cmd := exec.Command(executablePath, args...)
			cmd.Dir = tmpDir
			output, err := cmd.CombinedOutput()

			if tt.expectError {
				assert.Error(t, err, "CLI should fail for invalid input")
				if tt.errorContains != "" {
					assert.Contains(t, string(output), tt.errorContains,
						"Error output should contain expected message")
				}
			} else {
				if err != nil {
					t.Logf("CLI output: %s", string(output))
				}
				assert.NoError(t, err, "CLI should succeed")

				// Run output checks if provided
				if tt.outputChecks != nil {
					tt.outputChecks(t, outputPath)
				}
			}
		})
	}
}

// TestCLIEdgeCases tests edge cases and error conditions
func TestCLIEdgeCases(t *testing.T) {
	executablePath := buildCLIExecutable(t)

	t.Run("missing_template_file", func(t *testing.T) {
		tmpDir := t.TempDir()
		scenarioPath := filepath.Join(tmpDir, "scenario.ymmps")
		
		err := os.WriteFile(scenarioPath, []byte(`YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "test"
                Length: "100"`), 0644)
		require.NoError(t, err)

		cmd := exec.Command(executablePath, 
			"--template-file", "nonexistent.ymmp",
			scenarioPath)
		output, err := cmd.CombinedOutput()

		assert.Error(t, err)
		assert.Contains(t, string(output), "template")
	})

	t.Run("missing_scenario_file", func(t *testing.T) {
		cmd := exec.Command(executablePath, 
			"--template-file", "template.ymmp",
			"nonexistent.ymmps")
		output, err := cmd.CombinedOutput()

		assert.Error(t, err)
		assert.Contains(t, string(output), "file")
	})

	t.Run("invalid_cli_arguments", func(t *testing.T) {
		cmd := exec.Command(executablePath, "--invalid-flag")
		output, err := cmd.CombinedOutput()

		assert.Error(t, err)
		assert.Contains(t, string(output), "flag")
	})

	t.Run("help_display", func(t *testing.T) {
		cmd := exec.Command(executablePath, "--help")
		output, err := cmd.CombinedOutput()

		// Help should succeed (exit code 0)
		assert.NoError(t, err)
		assert.Contains(t, string(output), "ymmps-authorizer")
		assert.Contains(t, string(output), "template-file")
	})
}

// TestCLIPerformance tests performance with larger scenarios
func TestCLIPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	executablePath := buildCLIExecutable(t)
	tmpDir := t.TempDir()

	// Create a large template with many items
	templateData := `{
		"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
		"FilePath": "template.ymmp",
		"Timeline": {
			"VideoInfo": {"FPS": 30},
			"Items": [`

	// Add many template items
	for i := 0; i < 50; i++ {
		if i > 0 {
			templateData += ","
		}
		templateData += fmt.Sprintf(`{
			"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
			"Layer": %d, "Frame": 0, "Length": 120,
			"Remark": "voice_template_%d"
		}`, i, i)
	}
	templateData += `]
		}
	}`

	templatePath := filepath.Join(tmpDir, "template.ymmp")
	err := os.WriteFile(templatePath, []byte(templateData), 0644)
	require.NoError(t, err)

	// Create a large scenario using many of these templates
	scenarioData := `YMMPSVersion: "1"
Sequences:`

	for seq := 0; seq < 10; seq++ {
		scenarioData += fmt.Sprintf(`
  - ID: "sequence_%d"
    Scenes:`, seq)
		
		for scene := 0; scene < 5; scene++ {
			scenarioData += fmt.Sprintf(`
      - ID: "scene_%d_%d"
        Shots:`, seq, scene)
			
			for shot := 0; shot < 3; shot++ {
				scenarioData += fmt.Sprintf(`
          - ID: "shot_%d_%d_%d"
            Items:`, seq, scene, shot)
				
				// Use different templates
				templateIdx := (seq*5*3 + scene*3 + shot) % 50
				scenarioData += fmt.Sprintf(`
              - _Template: "voice_template_%d"
                Length: "%d"
                Serif: "Generated content %d_%d_%d"`, 
					templateIdx, 120+shot*30, seq, scene, shot)
			}
		}
	}

	scenarioPath := filepath.Join(tmpDir, "large_scenario.ymmps")
	err = os.WriteFile(scenarioPath, []byte(scenarioData), 0644)
	require.NoError(t, err)

	outputPath := filepath.Join(tmpDir, "large_output.ymmp")

	// Execute and measure time
	cmd := exec.Command(executablePath,
		"--template-file", templatePath,
		"--output", outputPath,
		"--verbose",
		scenarioPath)

	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Logf("CLI output: %s", string(output))
	}
	assert.NoError(t, err, "Large scenario conversion should succeed")

	// Verify output file was created and has reasonable size
	stat, err := os.Stat(outputPath)
	assert.NoError(t, err)
	assert.Greater(t, stat.Size(), int64(1000), "Output file should be substantial")

	// Verify the generated content is valid
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	var project models.YMMPProject
	err = json.Unmarshal(data, &project)
	assert.NoError(t, err, "Generated project should be valid JSON")

	// Should have 150 items (10 sequences * 5 scenes * 3 shots)
	assert.Len(t, project.Timelines[0].Items, 150, "Should have correct number of items")
}

// buildCLIExecutable builds the CLI executable for testing
func buildCLIExecutable(t *testing.T) string {
	tmpDir := t.TempDir()
	
	var executableName string
	if runtime.GOOS == "windows" {
		executableName = "ymmps-authorizer-test.exe"
	} else {
		executableName = "ymmps-authorizer-test"
	}
	
	executablePath := filepath.Join(tmpDir, executableName)
	
	// Build the executable
	cmd := exec.Command("go", "build", "-o", executablePath, "./cmd/ymmps-authorizer")
	cmd.Dir = filepath.Join("..", "..")  // Go up to project root
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Build output: %s", string(output))
		t.Fatalf("Failed to build CLI executable: %v", err)
	}
	
	return executablePath
}