package items

import (
	"fmt"
	"path/filepath"
	
	"github.com/user/ymmp-compiler/internal/models"
)

// AudioPlugin handles audio items
type AudioPlugin struct{}

// NewAudioPlugin creates a new audio plugin
func NewAudioPlugin() *AudioPlugin {
	return &AudioPlugin{}
}

// GetType returns the type of this plugin
func (p *AudioPlugin) GetType() string {
	return "audio"
}

// ParseYAML parses YAML data into an AudioItem
func (p *AudioPlugin) ParseYAML(data interface{}) (models.Item, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("audio item data must be a map")
	}
	
	// Required field: file_path
	filePathRaw, ok := dataMap["file_path"]
	if !ok {
		return nil, fmt.Errorf("audio item requires 'file_path' field")
	}
	
	filePath, ok := filePathRaw.(string)
	if !ok {
		return nil, fmt.Errorf("file_path must be a string")
	}
	
	item := &AudioItem{
		FilePath: filePath,
		Volume:   getFloatValue(dataMap, "volume", 1.0),
	}
	
	// Parse length
	if lengthRaw, ok := dataMap["length"]; ok {
		length, err := parseLengthSpec(lengthRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid length specification: %v", err)
		}
		item.length = length
	} else {
		// Default to auto length for audio items
		item.length = models.LengthSpec{Type: models.LengthTypeAuto}
	}
	
	return item, nil
}

// ConvertToYMMP converts an AudioItem to YMMP format
func (p *AudioPlugin) ConvertToYMMP(item models.Item, basePath string) (*models.YMMPItem, error) {
	audioItem, ok := item.(*AudioItem)
	if !ok {
		return nil, fmt.Errorf("expected AudioItem, got %T", item)
	}
	
	// Convert relative path to absolute path
	absolutePath := filepath.Join(basePath, audioItem.FilePath)
	
	// Convert seconds to frames (assuming 60 FPS)
	var lengthInFrames int
	if audioItem.length.Type == models.LengthTypeNumeric {
		lengthInFrames = int(audioItem.length.Value * 60)
	} else {
		// For auto length, use calculated length
		lengthInFrames = int(audioItem.calculatedLength * 60)
	}
	
	ymmItem := &models.YMMPItem{
		Type:     "YMM.Parts.AudioPart",
		Layer:    3, // Default layer for audio
		Frame:    int(audioItem.startTime * 60), // Convert start time to frames
		Length:   lengthInFrames,
		FilePath: absolutePath,
		Extended: map[string]interface{}{
			"Volume": audioItem.Volume,
		},
	}
	
	return ymmItem, nil
}

// AudioItem represents an audio item
type AudioItem struct {
	FilePath         string
	Volume           float64
	length           models.LengthSpec
	calculatedLength float64
	startTime        float64
}

// GetType returns the type of this item
func (a *AudioItem) GetType() string {
	return "audio"
}

// GetLength returns the length specification
func (a *AudioItem) GetLength() models.LengthSpec {
	return a.length
}

// SetCalculatedLength sets the calculated length
func (a *AudioItem) SetCalculatedLength(seconds float64) {
	a.calculatedLength = seconds
}

// GetStartTime returns the start time
func (a *AudioItem) GetStartTime() float64 {
	return a.startTime
}

// SetStartTime sets the start time
func (a *AudioItem) SetStartTime(seconds float64) {
	a.startTime = seconds
}