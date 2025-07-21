package parser

import (
	"strings"
	"testing"

	"github.com/yakuari-channel/video-authorizer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYMMPParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		check   func(t *testing.T, project *models.YMMPProject)
		wantErr bool
	}{
		{
			name: "valid minimal YMMP",
			input: `{
				"FilePath": "/path/to/project.ymmp",
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
						"Items": [],
						"CurrentFrame": 0,
						"Length": 100,
						"MaxLayer": 0
					}
				],
				"Characters": [],
				"CollapsedGroups": []
			}`,
			check: func(t *testing.T, project *models.YMMPProject) {
				assert.Equal(t, "/path/to/project.ymmp", project.FilePath)
				assert.Equal(t, 0, project.SelectedTimelineIndex)
				assert.Equal(t, 1, len(project.Timelines))
				assert.Equal(t, 60, project.Timelines[0].VideoInfo.FPS)
			},
			wantErr: false,
		},
		{
			name: "YMMP with voice item",
			input: `{
				"FilePath": "/path/to/project.ymmp",
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
								"Serif": "こんにちは",
								"IsLocked": false,
								"IsHidden": false
							}
						],
						"CurrentFrame": 0,
						"Length": 100,
						"MaxLayer": 0
					}
				],
				"Characters": [],
				"CollapsedGroups": []
			}`,
			check: func(t *testing.T, project *models.YMMPProject) {
				assert.Equal(t, 1, len(project.Timelines[0].Items))
				voiceItem, ok := project.Timelines[0].Items[0].(*models.VoiceItem)
				require.True(t, ok)
				assert.Equal(t, "ずんだもんvoice 01", voiceItem.Remark)
				assert.Equal(t, "ずんだもん", voiceItem.CharacterName)
				assert.Equal(t, "こんにちは", voiceItem.Serif)
			},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{"invalid": json}`,
			wantErr: true,
		},
	}

	parser := NewYMMPParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			got, err := parser.Parse(reader)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.check != nil {
					tt.check(t, got)
				}
			}
		})
	}
}

func TestYMMPParser_ParseFile(t *testing.T) {
	content := `{
		"FilePath": "/test/project.ymmp",
		"SelectedTimelineIndex": 0,
		"Timelines": [{
			"ID": "test-timeline",
			"Name": "Test",
			"VideoInfo": {"FPS": 60, "Hz": 48000, "Width": 1920, "Height": 1080},
			"Items": [],
			"CurrentFrame": 0,
			"Length": 0,
			"MaxLayer": 0
		}],
		"Characters": [],
		"CollapsedGroups": []
	}`

	tmpfile := createTempFile(t, "test.ymmp", content)
	defer tmpfile.Close()

	parser := NewYMMPParser()
	project, err := parser.ParseFile(tmpfile.Name())

	require.NoError(t, err)
	assert.Equal(t, "/test/project.ymmp", project.FilePath)
	assert.Equal(t, 1, len(project.Timelines))
}

func TestYMMPParser_ParseWithBOM(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		check func(t *testing.T, project *models.YMMPProject)
	}{
		{
			name: "JSON with UTF-8 BOM",
			input: append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{
				"FilePath": "/test/project.ymmp",
				"SelectedTimelineIndex": 0,
				"Timelines": [{
					"ID": "test-timeline",
					"Name": "Test",
					"VideoInfo": {"FPS": 60, "Hz": 48000, "Width": 1920, "Height": 1080},
					"Items": [],
					"CurrentFrame": 0,
					"Length": 0,
					"MaxLayer": 0
				}],
				"Characters": [],
				"CollapsedGroups": []
			}`)...),
			check: func(t *testing.T, project *models.YMMPProject) {
				assert.Equal(t, "/test/project.ymmp", project.FilePath)
				assert.Equal(t, 1, len(project.Timelines))
			},
		},
		{
			name: "JSON without BOM",
			input: []byte(`{
				"FilePath": "/test/project2.ymmp",
				"SelectedTimelineIndex": 0,
				"Timelines": [{
					"ID": "test-timeline-2",
					"Name": "Test2",
					"VideoInfo": {"FPS": 30, "Hz": 48000, "Width": 1920, "Height": 1080},
					"Items": [],
					"CurrentFrame": 0,
					"Length": 0,
					"MaxLayer": 0
				}],
				"Characters": [],
				"CollapsedGroups": []
			}`),
			check: func(t *testing.T, project *models.YMMPProject) {
				assert.Equal(t, "/test/project2.ymmp", project.FilePath)
				assert.Equal(t, 30, project.Timelines[0].VideoInfo.FPS)
			},
		},
	}

	parser := NewYMMPParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(string(tt.input))
			got, err := parser.Parse(reader)

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestRemoveBOM(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "content with UTF-8 BOM",
			input:    []byte{0xEF, 0xBB, 0xBF, 'h', 'e', 'l', 'l', 'o'},
			expected: []byte("hello"),
		},
		{
			name:     "content without BOM",
			input:    []byte("hello"),
			expected: []byte("hello"),
		},
		{
			name:     "empty content",
			input:    []byte{},
			expected: []byte{},
		},
		{
			name:     "only BOM",
			input:    []byte{0xEF, 0xBB, 0xBF},
			expected: []byte{},
		},
		{
			name:     "partial BOM (should not remove)",
			input:    []byte{0xEF, 0xBB},
			expected: []byte{0xEF, 0xBB},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeBOM(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}