package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/yakuari-channel/video-authorizer/internal/models"
)

// YMMPParser handles parsing of YMMP files
type YMMPParser struct{}

// NewYMMPParser creates a new YMMP parser
func NewYMMPParser() *YMMPParser {
	return &YMMPParser{}
}

// Parse parses YMMP content from a reader
func (p *YMMPParser) Parse(reader io.Reader) (*models.YMMPProject, error) {
	var project models.YMMPProject
	decoder := json.NewDecoder(reader)
	
	if err := decoder.Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate the project
	if err := project.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &project, nil
}

// ParseFile parses YMMP content from a file
func (p *YMMPParser) ParseFile(filename string) (*models.YMMPProject, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return p.Parse(file)
}