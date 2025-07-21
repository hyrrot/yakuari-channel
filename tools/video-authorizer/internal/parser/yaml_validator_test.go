package parser_test

import (
	"strings"
	"testing"

	"github.com/yakuari-channel/video-authorizer/internal/parser"
)

func TestYAMLStructureValidator_ValidateYAMLSyntax(t *testing.T) {
	validator := parser.NewYAMLStructureValidator()
	
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid YAML",
			yaml: `
YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - Template: "test"
                Length: "100"
`,
			wantErr: false,
		},
		{
			name: "invalid YAML - bad indentation",
			yaml: `
YMMPSVersion: "1"
Sequences:
- Scenes:
    - Shots:
        - Items:
            - Template: "test"
             Length: "100"  # Bad indentation
`,
			wantErr: true,
		},
		{
			name: "invalid YAML - unclosed quote",
			yaml: `
YMMPSVersion: "1
Sequences: []
`,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateYAMLSyntax([]byte(tt.yaml))
			if tt.wantErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestYAMLStructureValidator_ValidateYMMPSStructure(t *testing.T) {
	validator := parser.NewYAMLStructureValidator()
	
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid structure",
			input: map[string]interface{}{
				"YMMPSVersion": "1",
				"Sequences": []interface{}{
					map[string]interface{}{
						"ID": "seq1",
						"Scenes": []interface{}{
							map[string]interface{}{
								"ID": "scene1",
								"Shots": []interface{}{
									map[string]interface{}{
										"ID": "shot1",
										"Items": []interface{}{
											map[string]interface{}{
												"Template": "test",
												"Length":   "100",
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
			name:    "invalid root type",
			input:   []interface{}{},
			wantErr: true,
			errMsg:  "root must be a map",
		},
		{
			name: "missing YMMPSVersion",
			input: map[string]interface{}{
				"Sequences": []interface{}{},
			},
			wantErr: true,
			errMsg:  "missing required field: YMMPSVersion",
		},
		{
			name: "invalid YMMPSVersion type",
			input: map[string]interface{}{
				"YMMPSVersion": 1,
				"Sequences":    []interface{}{},
			},
			wantErr: true,
			errMsg:  "YMMPSVersion must be a string",
		},
		{
			name: "unsupported YMMPSVersion",
			input: map[string]interface{}{
				"YMMPSVersion": "2",
				"Sequences":    []interface{}{},
			},
			wantErr: true,
			errMsg:  "unsupported YMMPSVersion: 2",
		},
		{
			name: "missing Sequences",
			input: map[string]interface{}{
				"YMMPSVersion": "1",
			},
			wantErr: true,
			errMsg:  "missing required field: Sequences",
		},
		{
			name: "empty Sequences",
			input: map[string]interface{}{
				"YMMPSVersion": "1",
				"Sequences":    []interface{}{},
			},
			wantErr: true,
			errMsg:  "Sequences array cannot be empty",
		},
		{
			name: "missing Template in item",
			input: map[string]interface{}{
				"YMMPSVersion": "1",
				"Sequences": []interface{}{
					map[string]interface{}{
						"Scenes": []interface{}{
							map[string]interface{}{
								"Shots": []interface{}{
									map[string]interface{}{
										"Items": []interface{}{
											map[string]interface{}{
												"Length": "100",
												// Missing Template
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing required field: Template or _Template",
		},
		{
			name: "missing Length in item",
			input: map[string]interface{}{
				"YMMPSVersion": "1",
				"Sequences": []interface{}{
					map[string]interface{}{
						"Scenes": []interface{}{
							map[string]interface{}{
								"Shots": []interface{}{
									map[string]interface{}{
										"Items": []interface{}{
											map[string]interface{}{
												"Template": "test",
												// Missing Length
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing required field: Length",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateYMMPSStructure(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing '%s', got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestYAMLStructureValidator_DetectCommonYAMLIssues(t *testing.T) {
	validator := parser.NewYAMLStructureValidator()
	
	tests := []struct {
		name           string
		input          interface{}
		expectedIssues []string
	}{
		{
			name: "no issues",
			input: map[string]interface{}{
				"YMMPSVersion": "1",
				"Sequences": []interface{}{
					map[string]interface{}{
						"ID": "sequence_1",
						"Scenes": []interface{}{
							map[string]interface{}{
								"ID": "scene_1",
								"Shots": []interface{}{
									map[string]interface{}{
										"ID": "shot_1",
										"Items": []interface{}{
											map[string]interface{}{
												"Template": "test",
												"Length":   "100",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedIssues: []string{},
		},
		{
			name: "unexpected root field",
			input: map[string]interface{}{
				"YMMPSVersion": "1",
				"Sequences":    []interface{}{},
				"ExtraField":   "value",
			},
			expectedIssues: []string{"unexpected field at root level: ExtraField"},
		},
		{
			name: "mixed naming conventions",
			input: map[string]interface{}{
				"YMMPSVersion": "1",
				"Sequences": []interface{}{
					map[string]interface{}{
						"ID": "sequence_1", // snake_case
						"Scenes": []interface{}{
							map[string]interface{}{
								"ID": "sceneOne", // camelCase
								"Shots": []interface{}{
									map[string]interface{}{
										"ID": "shot_1",
										"Items": []interface{}{
											map[string]interface{}{
												"Template": "test",
												"Length":   "100",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedIssues: []string{"inconsistent ID naming"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := validator.DetectCommonYAMLIssues(tt.input)
			
			if len(issues) != len(tt.expectedIssues) {
				t.Errorf("expected %d issues, got %d: %v", len(tt.expectedIssues), len(issues), issues)
				return
			}
			
			for i, expectedIssue := range tt.expectedIssues {
				if i >= len(issues) {
					t.Errorf("missing expected issue: %s", expectedIssue)
					continue
				}
				
				if !strings.Contains(issues[i], expectedIssue) {
					t.Errorf("expected issue containing '%s', got: %s", expectedIssue, issues[i])
				}
			}
		})
	}
}