package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yakuari-channel/video-authorizer/internal/parser"
)

func TestYMMPSValidation(t *testing.T) {
	tests := []struct {
		name         string
		ymmpsContent string
		expectError  bool
		errorMessage string
	}{
		{
			name: "empty_shot_without_id",
			ymmpsContent: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - ID: shot1
            # Empty shot - no Items array
`,
			expectError:  true,
			errorMessage: "missing required field: Items",
		},
		{
			name: "empty_shot_with_empty_items",
			ymmpsContent: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - ID: shot1
            Items: []  # Empty Items array
`,
			expectError:  true,
			errorMessage: "Items array cannot be empty",
		},
		{
			name: "empty_shot_without_id_or_items",
			ymmpsContent: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - {}  # Completely empty shot
`,
			expectError:  true,
			errorMessage: "missing required field: Items",
		},
		{
			name: "valid_shot_with_items",
			ymmpsContent: `YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - ID: shot1
            Items:
              - Type: VoiceItem
                Template: voice_template
                Length: "120"
`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			ymmpsPath := filepath.Join(tmpDir, "test.ymmps")

			err := os.WriteFile(ymmpsPath, []byte(tt.ymmpsContent), 0644)
			require.NoError(t, err)

			// Parse and validate
			ymmpsParser := parser.NewYMMPSParser()
			_, err = ymmpsParser.ParseFile(ymmpsPath)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOldFormatCompatibility(t *testing.T) {
	// Test that old format gives helpful error
	oldFormatContent := `YMMPSVersion: "1"
Sequences:
- ID: sequence1
  Scenes:
  - ID: scene1
    Shots:
    - _Template: "ずんだもん立ち絵01"
      Length: "until:SEQUENCE_END"
`

	tmpDir := t.TempDir()
	ymmpsPath := filepath.Join(tmpDir, "old_format.ymmps")

	err := os.WriteFile(ymmpsPath, []byte(oldFormatContent), 0644)
	require.NoError(t, err)

	ymmpsParser := parser.NewYMMPSParser()
	_, err = ymmpsParser.ParseFile(ymmpsPath)

	assert.Error(t, err)
	// Should give helpful error about missing Items structure
	assert.Contains(t, err.Error(), "missing required field: Items")
}