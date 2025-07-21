package parser

import (
	"fmt"
	"io"
	"os"

	"github.com/yakuari-channel/video-authorizer/internal/models"
	"gopkg.in/yaml.v3"
)

// YMMPSParser handles parsing of YMMPS files
type YMMPSParser struct{}

// NewYMMPSParser creates a new YMMPS parser
func NewYMMPSParser() *YMMPSParser {
	return &YMMPSParser{}
}

// Parse parses YMMPS content from a reader
func (p *YMMPSParser) Parse(reader io.Reader) (*models.YMMPSDocument, error) {
	// Read all data first for validation
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	// Validate YAML syntax
	validator := NewYAMLStructureValidator()
	if err := validator.ValidateYAMLSyntax(data); err != nil {
		return nil, fmt.Errorf("YAML syntax validation failed: %w", err)
	}

	// Parse into generic structure for detailed validation
	var rawInterface interface{}
	if err := yaml.Unmarshal(data, &rawInterface); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate YMMPS structure
	if err := validator.ValidateYMMPSStructure(rawInterface); err != nil {
		return nil, fmt.Errorf("YMMPS structure validation failed: %w", err)
	}

	// Check for common issues and warn
	if issues := validator.DetectCommonYAMLIssues(rawInterface); len(issues) > 0 {
		// For now, we just log warnings. In production, these could be returned as warnings
		fmt.Printf("Warning: Detected potential issues in YMMPS file:\n")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
	}

	// Now decode into typed structure
	var raw rawYMMPSDocument
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse YAML into typed structure: %w", err)
	}

	// Convert raw document to models.YMMPSDocument
	doc := &models.YMMPSDocument{
		YMMPSVersion: raw.YMMPSVersion,
		Sequences:    make([]models.Sequence, len(raw.Sequences)),
	}

	// Process sequences
	for i, rawSeq := range raw.Sequences {
		seq := models.Sequence{
			ID:     rawSeq.ID,
			Scenes: make([]models.Scene, len(rawSeq.Scenes)),
		}

		// Process scenes
		for j, rawScene := range rawSeq.Scenes {
			scene := models.Scene{
				ID:    rawScene.ID,
				Shots: make([]models.Shot, len(rawScene.Shots)),
			}

			// Process shots
			for k, rawShot := range rawScene.Shots {
				shot := models.Shot{
					ID: rawShot.ID,
				}

				// Process items - they might be inline with the shot
				if len(rawShot.Items) > 0 {
					shot.Items = make([]models.ItemSpec, len(rawShot.Items))
					for l, rawItem := range rawShot.Items {
						// Try both "Template" and "_Template" for backward compatibility
						template, ok := rawItem["Template"].(string)
						if !ok {
							template, ok = rawItem["_Template"].(string)
							if !ok {
								return nil, fmt.Errorf("Template is required and must be a string")
							}
						}

						length := fmt.Sprintf("%v", rawItem["Length"])

						item := models.ItemSpec{
							Template:   template,
							Length:     length,
							Properties: make(map[string]interface{}),
						}

						// Handle Properties field specifically if it exists
						if props, ok := rawItem["Properties"].(map[string]interface{}); ok {
							for key, value := range props {
								item.Properties[key] = value
							}
						}
						
						// Copy other fields as properties except Template/_Template, Length, Type, and Properties
						for key, value := range rawItem {
							if key != "Template" && key != "_Template" && key != "Length" && 
							   key != "Type" && key != "Properties" {
								item.Properties[key] = value
							}
						}

						shot.Items[l] = item
					}
				} else {
					// Handle case where items might be defined directly in the shot
					shot.Items = []models.ItemSpec{}
				}

				scene.Shots[k] = shot
			}

			seq.Scenes[j] = scene
		}

		doc.Sequences[i] = seq
	}

	// Validate the document
	if err := doc.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return doc, nil
}

// ParseFile parses YMMPS content from a file
func (p *YMMPSParser) ParseFile(filename string) (*models.YMMPSDocument, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return p.Parse(file)
}

// Raw structures for YAML parsing
type rawYMMPSDocument struct {
	YMMPSVersion string        `yaml:"YMMPSVersion"`
	Sequences    []rawSequence `yaml:"Sequences"`
}

type rawSequence struct {
	ID     string     `yaml:"ID,omitempty"`
	Scenes []rawScene `yaml:"Scenes"`
}

type rawScene struct {
	ID    string    `yaml:"ID,omitempty"`
	Shots []rawShot `yaml:"Shots"`
}

type rawShot struct {
	ID    string                   `yaml:"ID,omitempty"`
	Items []map[string]interface{} `yaml:"Items,omitempty"`
}