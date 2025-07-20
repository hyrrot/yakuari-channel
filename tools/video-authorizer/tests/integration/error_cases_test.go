package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yakuari-channel/video-authorizer/internal/converter"
	"github.com/yakuari-channel/video-authorizer/internal/parser"
)

// TestErrorCases tests various error conditions and edge cases
func TestErrorCases(t *testing.T) {
	tests := []struct {
		name            string
		templateData    string
		scenarioData    string
		expectedError   string
		errorType      string // "parse", "validate", "convert"
	}{
		{
			name: "invalid_ymmps_version",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "999"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "120"`,
			expectedError: "unsupported YMMPS version",
			errorType:     "parse",
		},
		{
			name: "missing_ymmps_version",
			templateData: createBasicTemplate(),
			scenarioData: `Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "120"`,
			expectedError: "missing required field: YMMPSVersion",
			errorType:     "parse",
		},
		{
			name: "empty_sequences",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences: []`,
			expectedError: "Sequences array cannot be empty",
			errorType:     "parse",
		},
		{
			name: "missing_sequences",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"`,
			expectedError: "missing required field: Sequences",
			errorType:     "parse",
		},
		{
			name: "empty_scenes_in_sequence",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - ID: "seq1"
    Scenes: []`,
			expectedError: "Scenes array cannot be empty",
			errorType:     "parse",
		},
		{
			name: "empty_shots_in_scene",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - ID: "scene1"
        Shots: []`,
			expectedError: "Shots array cannot be empty",
			errorType:     "parse",
		},
		{
			name: "empty_items_in_shot",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - ID: "shot1"
            Items: []`,
			expectedError: "Items array cannot be empty",
			errorType:     "parse",
		},
		{
			name: "missing_template_field",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - Length: "120"
                Serif: "Missing template"`,
			expectedError: "missing required field: _Template",
			errorType:     "parse",
		},
		{
			name: "missing_length_field",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Serif: "Missing length"`,
			expectedError: "missing required field: Length",
			errorType:     "parse",
		},
		{
			name: "nonexistent_template_reference",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "nonexistent_template"
                Length: "120"`,
			expectedError: "template 'nonexistent_template' not found",
			errorType:     "validate",
		},
		{
			name: "invalid_id_reference_sequence",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - ID: "seq1"
    Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "_until:SEQUENCE_END:nonexistent_seq"`,
			expectedError: "undefined SEQUENCE ID 'nonexistent_seq'",
			errorType:     "validate",
		},
		{
			name: "invalid_id_reference_scene",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - ID: "scene1"
        Shots:
          - Items:
              - _Template: "voice_template"
                Length: "_until:SCENE_END:nonexistent_scene"`,
			expectedError: "undefined SCENE ID 'nonexistent_scene'",
			errorType:     "validate",
		},
		{
			name: "invalid_id_reference_shot",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - ID: "shot1"
            Items:
              - _Template: "voice_template"
                Length: "_until:SHOT_END:nonexistent_shot"`,
			expectedError: "undefined SHOT ID 'nonexistent_shot'",
			errorType:     "validate",
		},
		{
			name: "wrong_type_id_reference",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - ID: "seq1"
    Scenes:
      - ID: "scene1"
        Shots:
          - Items:
              - _Template: "voice_template"
                Length: "_until:SHOT_END:seq1"`,
			expectedError: "ID 'seq1' is of type SEQUENCE, but expected SHOT",
			errorType:     "validate",
		},
		{
			name: "empty_id_reference",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "_until:SCENE_END:"`,
			expectedError: "empty ID in reference",
			errorType:     "validate",
		},
		{
			name: "invalid_length_format",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "invalid_length_format"`,
			expectedError: "invalid length format",
			errorType:     "parse",
		},
		{
			name: "invalid_numeric_length",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "-50"`,
			expectedError: "length must be positive",
			errorType:     "parse",
		},
		{
			name: "malformed_yaml_structure",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
    - Shots:  # Invalid indentation
        - Items:
            - _Template: "voice_template"
              Length: "120"`,
			expectedError: "YAML syntax",
			errorType:     "parse",
		},
		{
			name: "unclosed_yaml_quote",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "120"`,
			expectedError: "YAML syntax",
			errorType:     "parse",
		},
		{
			name: "invalid_json_template",
			templateData: `{
				"$type": "YukkuriMovieMaker.Project.YmmProject, YukkuriMovieMaker",
				"FilePath": "template.ymmp",
				"Timeline": {
					"VideoInfo": {"FPS": 30},
					"Items": [
						{
							"$type": "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker",
							"Remark": "voice_template"
							// Missing comma - invalid JSON
						}
					]
				}
			}`,
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "120"`,
			expectedError: "invalid character",
			errorType:     "parse",
		},
		{
			name: "circular_reference_detection",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - ID: "seq1"
    Scenes:
      - ID: "scene1"
        Shots:
          - ID: "shot1"
            Items:
              - _Template: "voice_template"
                Length: "_until:SHOT_END:shot1"`,
			expectedError: "self-reference not allowed",
			errorType:     "validate",
		},
		{
			name: "duplicate_sequence_ids",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - ID: "duplicate_id"
    Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "120"
  - ID: "duplicate_id"
    Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "120"`,
			expectedError: "duplicate SEQUENCE ID 'duplicate_id'",
			errorType:     "validate",
		},
		{
			name: "duplicate_scene_ids",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - ID: "duplicate_scene"
        Shots:
          - Items:
              - _Template: "voice_template"
                Length: "120"
      - ID: "duplicate_scene"
        Shots:
          - Items:
              - _Template: "voice_template"
                Length: "120"`,
			expectedError: "duplicate SCENE ID 'duplicate_scene'",
			errorType:     "validate",
		},
		{
			name: "duplicate_shot_ids",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - ID: "duplicate_shot"
            Items:
              - _Template: "voice_template"
                Length: "120"
          - ID: "duplicate_shot"
            Items:
              - _Template: "voice_template"
                Length: "120"`,
			expectedError: "duplicate SHOT ID 'duplicate_shot'",
			errorType:     "validate",
		},
		{
			name: "unsupported_auto_format",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "_auto:UNSUPPORTED"`,
			expectedError: "invalid length format",
			errorType:     "parse",
		},
		{
			name: "natural_language_invalid_format",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "INVALID NATURAL LANGUAGE FORMAT"`,
			expectedError: "invalid length format",
			errorType:     "parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create template file
			templatePath := filepath.Join(tmpDir, "template.ymmp")
			err := os.WriteFile(templatePath, []byte(tt.templateData), 0644)
			require.NoError(t, err)

			// Create scenario file
			scenarioPath := filepath.Join(tmpDir, "scenario.ymmps")
			err = os.WriteFile(scenarioPath, []byte(tt.scenarioData), 0644)
			require.NoError(t, err)

			var finalError error

			switch tt.errorType {
			case "parse":
				// Test parsing stage
				ymmParser := parser.NewYMMPParser()
				_, err = ymmParser.ParseFile(templatePath)
				if err != nil {
					finalError = err
					break
				}

				ymmpsParser := parser.NewYMMPSParser()
				_, err = ymmpsParser.ParseFile(scenarioPath)
				finalError = err

			case "validate":
				// Test validation stage (after successful parsing)
				ymmParser := parser.NewYMMPParser()
				template, err := ymmParser.ParseFile(templatePath)
				require.NoError(t, err, "Template parsing should succeed")

				ymmpsParser := parser.NewYMMPSParser()
				scenario, err := ymmpsParser.ParseFile(scenarioPath)
				require.NoError(t, err, "Scenario parsing should succeed")

				// Validation should fail during conversion
				conv := converter.NewConverter()
				_, err = conv.Convert(scenario, template)
				finalError = err

			case "convert":
				// Test conversion stage (after successful parsing and validation)
				ymmParser := parser.NewYMMPParser()
				template, err := ymmParser.ParseFile(templatePath)
				require.NoError(t, err, "Template parsing should succeed")

				ymmpsParser := parser.NewYMMPSParser()
				scenario, err := ymmpsParser.ParseFile(scenarioPath)
				require.NoError(t, err, "Scenario parsing should succeed")

				conv := converter.NewConverter()
				_, err = conv.Convert(scenario, template)
				finalError = err
			}

			// Verify that the expected error occurred
			require.Error(t, finalError, "Expected error should occur")
			assert.Contains(t, strings.ToLower(finalError.Error()), 
				strings.ToLower(tt.expectedError),
				"Error message should contain expected text")
		})
	}
}

// TestEdgeCases tests edge cases that should succeed but are unusual
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		templateData string
		scenarioData string
		description  string
		checks       func(t *testing.T, project interface{}, err error)
	}{
		{
			name:         "single_item_minimal_scenario",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "60"`,
			description: "Minimal valid scenario with just one item",
			checks: func(t *testing.T, project interface{}, err error) {
				require.NoError(t, err)
				// Should work fine with minimal structure
			},
		},
		{
			name:         "very_long_ids",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - ID: "this_is_a_very_long_sequence_identifier_that_contains_many_characters_and_underscores_and_numbers_123456789"
    Scenes:
      - ID: "this_is_also_a_very_long_scene_identifier_with_many_characters"
        Shots:
          - ID: "very_long_shot_identifier_as_well"
            Items:
              - _Template: "voice_template"
                Length: "120"`,
			description: "Very long ID names should be handled correctly",
			checks: func(t *testing.T, project interface{}, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:         "special_characters_in_text",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "120"
                Serif: "特殊文字テスト: こんにちは！🎉 \"引用符\" 'アポストロフィ' & < > % @ # $ ^ * ( ) [ ] { } | \\ / ? + = ~ ` + "`" + `"`,
			description: "Special characters and unicode should be preserved",
			checks: func(t *testing.T, project interface{}, err error) {
				require.NoError(t, err)
				// Text with special characters should be handled properly
			},
		},
		{
			name:         "zero_length_items",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "0"`,
			description: "Zero-length items should be allowed",
			checks: func(t *testing.T, project interface{}, err error) {
				require.NoError(t, err)
				// Zero length should be valid
			},
		},
		{
			name:         "very_large_numbers",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "voice_template"
                Length: "999999999"`,
			description: "Very large frame numbers should be handled",
			checks: func(t *testing.T, project interface{}, err error) {
				require.NoError(t, err)
				// Large numbers should be handled correctly
			},
		},
		{
			name:         "mixed_id_and_non_id_elements",
			templateData: createBasicTemplate(),
			scenarioData: `YMMPSVersion: "1"
Sequences:
  - ID: "named_sequence"
    Scenes:
      - Shots:  # Scene without ID
          - ID: "named_shot"
            Items:
              - _Template: "voice_template"
                Length: "120"
      - ID: "named_scene"
        Shots:
          - Items:  # Shot without ID
              - _Template: "voice_template"
                Length: "120"`,
			description: "Mix of elements with and without IDs should work",
			checks: func(t *testing.T, project interface{}, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "empty_template_project",
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
      - Shots:
          - Items:
              - _Template: "nonexistent"
                Length: "120"`,
			description: "Empty template should produce meaningful error",
			checks: func(t *testing.T, project interface{}, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "template")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create template file
			templatePath := filepath.Join(tmpDir, "template.ymmp")
			err := os.WriteFile(templatePath, []byte(tt.templateData), 0644)
			require.NoError(t, err)

			// Create scenario file
			scenarioPath := filepath.Join(tmpDir, "scenario.ymmps")
			err = os.WriteFile(scenarioPath, []byte(tt.scenarioData), 0644)
			require.NoError(t, err)

			// Parse and convert
			ymmParser := parser.NewYMMPParser()
			template, err := ymmParser.ParseFile(templatePath)
			require.NoError(t, err)

			ymmpsParser := parser.NewYMMPSParser()
			scenario, err := ymmpsParser.ParseFile(scenarioPath)

			if err != nil {
				// Parsing failed - run checks on parse error
				tt.checks(t, nil, err)
				return
			}

			// Convert
			conv := converter.NewConverter()
			project, err := conv.Convert(scenario, template)

			// Run test-specific checks
			tt.checks(t, project, err)
		})
	}
}

// createBasicTemplate creates a basic template for testing
func createBasicTemplate() string {
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