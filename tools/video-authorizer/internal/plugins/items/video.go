package items

import (
	"fmt"
	"path/filepath"
	
	"github.com/user/ymmp-compiler/internal/models"
)

// VideoPlugin handles video items
type VideoPlugin struct{}

// NewVideoPlugin creates a new video plugin
func NewVideoPlugin() *VideoPlugin {
	return &VideoPlugin{}
}

// GetType returns the type of this plugin
func (p *VideoPlugin) GetType() string {
	return "video"
}

// ParseYAML parses YAML data into a VideoItem
func (p *VideoPlugin) ParseYAML(data interface{}) (models.Item, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("video item data must be a map")
	}
	
	// Required field: file_path
	filePathRaw, ok := dataMap["file_path"]
	if !ok {
		return nil, fmt.Errorf("video item requires 'file_path' field")
	}
	
	filePath, ok := filePathRaw.(string)
	if !ok {
		return nil, fmt.Errorf("file_path must be a string")
	}
	
	item := &VideoItem{
		FilePath: filePath,
		X:        getIntValue(dataMap, "x", 0),
		Y:        getIntValue(dataMap, "y", 0),
		Z:        getIntValue(dataMap, "z", 0),
		Zoom:     getFloatValue(dataMap, "zoom", 1.0),
		Rotation: getFloatValue(dataMap, "rotation", 0.0),
		Alpha:    getIntValue(dataMap, "alpha", 255),
	}
	
	// Parse length
	if lengthRaw, ok := dataMap["length"]; ok {
		length, err := parseLengthSpec(lengthRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid length specification: %v", err)
		}
		item.length = length
	} else {
		// Default to auto length for video items
		item.length = models.LengthSpec{Type: models.LengthTypeAuto}
	}
	
	return item, nil
}

// ConvertToYMMP converts a VideoItem to YMMP format
func (p *VideoPlugin) ConvertToYMMP(item models.Item, basePath string) (*models.YMMPItem, error) {
	videoItem, ok := item.(*VideoItem)
	if !ok {
		return nil, fmt.Errorf("expected VideoItem, got %T", item)
	}
	
	// Convert relative path to absolute path
	absolutePath := filepath.Join(basePath, videoItem.FilePath)
	
	// Convert seconds to frames (assuming 60 FPS)
	var lengthInFrames int
	if videoItem.length.Type == models.LengthTypeNumeric {
		lengthInFrames = int(videoItem.length.Value * 60)
	} else {
		// For auto length, use calculated length
		lengthInFrames = int(videoItem.calculatedLength * 60)
	}
	
	ymmItem := &models.YMMPItem{
		Type:     "YMM.Parts.VideoPart",
		Layer:    4, // Default layer for video
		Frame:    int(videoItem.startTime * 60), // Convert start time to frames
		Length:   lengthInFrames,
		FilePath: absolutePath,
		Extended: map[string]interface{}{
			"X":        videoItem.X,
			"Y":        videoItem.Y,
			"Z":        videoItem.Z,
			"Zoom":     videoItem.Zoom,
			"Rotation": videoItem.Rotation,
			"Alpha":    videoItem.Alpha,
		},
	}
	
	return ymmItem, nil
}

// VideoItem represents a video item
type VideoItem struct {
	FilePath         string
	X, Y, Z          int
	Zoom, Rotation   float64
	Alpha            int
	length           models.LengthSpec
	calculatedLength float64
	startTime        float64
}

// GetType returns the type of this item
func (v *VideoItem) GetType() string {
	return "video"
}

// GetLength returns the length specification
func (v *VideoItem) GetLength() models.LengthSpec {
	return v.length
}

// SetCalculatedLength sets the calculated length
func (v *VideoItem) SetCalculatedLength(seconds float64) {
	v.calculatedLength = seconds
}

// GetStartTime returns the start time
func (v *VideoItem) GetStartTime() float64 {
	return v.startTime
}

// SetStartTime sets the start time
func (v *VideoItem) SetStartTime(seconds float64) {
	v.startTime = seconds
}