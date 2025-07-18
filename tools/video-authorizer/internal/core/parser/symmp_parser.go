package parser

import (
	"io"
	
	"gopkg.in/yaml.v3"
	"github.com/user/ymmp-compiler/internal/models"
)

// ParseSYMMP parses a SYMMP file from a reader
func ParseSYMMP(reader io.Reader) (*models.Episode, error) {
	var episode models.Episode
	
	decoder := yaml.NewDecoder(reader)
	err := decoder.Decode(&episode)
	if err != nil {
		return nil, err
	}
	
	return &episode, nil
}