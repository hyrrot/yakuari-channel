package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYMMPSDocument_Validate(t *testing.T) {
	tests := []struct {
		name    string
		doc     YMMPSDocument
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid document",
			doc: YMMPSDocument{
				YMMPSVersion: "1",
				Sequences: []Sequence{
					{
						ID: "seq1",
						Scenes: []Scene{
							{
								ID: "scene1",
								Shots: []Shot{
									{
										Items: []ItemSpec{
											{
												Template: "ずんだもんvoice 01",
												Length:   "100",
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
			name: "invalid version",
			doc: YMMPSDocument{
				YMMPSVersion: "2",
			},
			wantErr: true,
			errMsg:  "unsupported YMMPS version: 2",
		},
		{
			name: "empty sequences",
			doc: YMMPSDocument{
				YMMPSVersion: "1",
				Sequences:    []Sequence{},
			},
			wantErr: true,
			errMsg:  "at least one sequence is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.doc.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParsedLength_Parse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      ParsedLength
		wantErr   bool
	}{
		{
			name:  "numeric length",
			input: "100",
			want: ParsedLength{
				Type:  LengthTypeNumeric,
				Value: 100,
			},
		},
		{
			name:  "until sequence end",
			input: "_until:SEQUENCE_END",
			want: ParsedLength{
				Type: LengthTypeUntilSeqEnd,
			},
		},
		{
			name:  "until scene end",
			input: "_until:SCENE_END",
			want: ParsedLength{
				Type: LengthTypeUntilSceneEnd,
			},
		},
		{
			name:  "until shot end",
			input: "_until:SHOT_END",
			want: ParsedLength{
				Type: LengthTypeUntilShotEnd,
			},
		},
		{
			name:  "until sequence end with ID",
			input: "_until:SEQUENCE_END:seq1",
			want: ParsedLength{
				Type:       LengthTypeUntilSeqEndID,
				TargetType: "SEQUENCE",
				TargetID:   "seq1",
			},
		},
		{
			name:  "until scene end with ID",
			input: "_until:SCENE_END:scene1",
			want: ParsedLength{
				Type:       LengthTypeUntilSceneEndID,
				TargetType: "SCENE",
				TargetID:   "scene1",
			},
		},
		{
			name:  "auto voice",
			input: "_auto:VOICEVOX",
			want: ParsedLength{
				Type: LengthTypeAutoVoice,
			},
		},
		{
			name:  "auto video",
			input: "_auto:VIDEO",
			want: ParsedLength{
				Type: LengthTypeAutoVideo,
			},
		},
		{
			name:    "invalid format",
			input:   "_invalid:format",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLength(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}