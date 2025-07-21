package parser

import (
	"bytes"
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
	// Read all content to handle BOM
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	// Remove BOM if present
	content = removeBOM(content)

	var project models.YMMPProject
	decoder := json.NewDecoder(bytes.NewReader(content))
	
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

// removeBOM removes the UTF-8 BOM (Byte Order Mark) if present
func removeBOM(content []byte) []byte {
	// UTF-8 BOM is 0xEF, 0xBB, 0xBF
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		return content[3:]
	}
	return content
}