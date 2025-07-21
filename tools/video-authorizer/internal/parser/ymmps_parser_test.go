package parser

import (
	"strings"
	"testing"

	"github.com/yakuari-channel/video-authorizer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYMMPSParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *models.YMMPSDocument
		wantErr bool
	}{
		{
			name: "valid YMMPS",
			input: `YMMPSVersion: "1"
Sequences:
  - ID: seq1
    Scenes:
      - ID: scene1
        Shots:
          - Items:
              - _Template: "ずんだもんvoice 01"
                Length: "100"
                Serif: "こんにちは"
`,
			want: &models.YMMPSDocument{
				YMMPSVersion: "1",
				Sequences: []models.Sequence{
					{
						ID: "seq1",
						Scenes: []models.Scene{
							{
								ID: "scene1",
								Shots: []models.Shot{
									{
										Items: []models.ItemSpec{
											{
												Template: "ずんだもんvoice 01",
												Length:   "100",
												Properties: map[string]interface{}{
													"Serif": "こんにちは",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid YAML",
			input: `YMMPSVersion: "1"
Sequences
  invalid yaml`,
			wantErr: true,
		},
		{
			name: "invalid version",
			input: `YMMPSVersion: "2"
Sequences: []`,
			wantErr: true,
		},
	}

	parser := NewYMMPSParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			got, err := parser.Parse(reader)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.YMMPSVersion, got.YMMPSVersion)
				assert.Equal(t, len(tt.want.Sequences), len(got.Sequences))
				
				if len(tt.want.Sequences) > 0 {
					assert.Equal(t, tt.want.Sequences[0].ID, got.Sequences[0].ID)
				}
			}
		})
	}
}

func TestYMMPSParser_ParseFile(t *testing.T) {
	// Create a temporary file for testing
	content := `YMMPSVersion: "1"
Sequences:
  - ID: test
    Scenes:
      - Shots:
          - Items:
              - _Template: "test"
                Length: "10"
`
	
	// Create temp file
	tmpfile := createTempFile(t, "test.ymmps", content)
	defer tmpfile.Close()

	parser := NewYMMPSParser()
	doc, err := parser.ParseFile(tmpfile.Name())

	require.NoError(t, err)
	assert.Equal(t, "1", doc.YMMPSVersion)
	assert.Equal(t, 1, len(doc.Sequences))
}