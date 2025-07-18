package items

import (
	"fmt"
	"path/filepath"
	
	"github.com/user/ymmp-compiler/internal/models"
)

// ImagePlugin handles image items
type ImagePlugin struct{}

// NewImagePlugin creates a new image plugin
func NewImagePlugin() *ImagePlugin {
	return &ImagePlugin{}
}

// GetType returns the type of this plugin
func (p *ImagePlugin) GetType() string {
	return "image"
}

// ParseYAML parses YAML data into an ImageItem
func (p *ImagePlugin) ParseYAML(data interface{}) (models.Item, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("image item data must be a map")
	}
	
	// Required field: file_path
	filePathRaw, ok := dataMap["file_path"]
	if !ok {
		return nil, fmt.Errorf("image item requires 'file_path' field")
	}
	
	filePath, ok := filePathRaw.(string)
	if !ok {
		return nil, fmt.Errorf("file_path must be a string")
	}
	
	item := &ImageItem{
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
		// Default to numeric length (must be specified for images)
		return nil, fmt.Errorf("image item requires 'length' field")
	}
	
	return item, nil
}

// ConvertToYMMP converts an ImageItem to YMMP format
func (p *ImagePlugin) ConvertToYMMP(item models.Item, basePath string) (*models.YMMPItem, error) {
	imageItem, ok := item.(*ImageItem)
	if !ok {
		return nil, fmt.Errorf("expected ImageItem, got %T", item)
	}
	
	// Convert relative path to absolute path
	absolutePath := filepath.Join(basePath, imageItem.FilePath)
	
	// Convert seconds to frames (assuming 60 FPS)
	lengthInFrames := int(imageItem.length.Value * 60)
	
	ymmItem := &models.YMMPItem{
		Type:     "YMM.Parts.ImagePart",
		Layer:    1, // Default layer
		Frame:    int(imageItem.startTime * 60), // Convert start time to frames
		Length:   lengthInFrames,
		FilePath: absolutePath,
		Extended: map[string]interface{}{
			"X":        imageItem.X,
			"Y":        imageItem.Y,
			"Z":        imageItem.Z,
			"Zoom":     imageItem.Zoom,
			"Rotation": imageItem.Rotation,
			"Alpha":    imageItem.Alpha,
		},
	}
	
	return ymmItem, nil
}

// ImageItem represents an image item
type ImageItem struct {
	FilePath         string
	X, Y, Z          int
	Zoom, Rotation   float64
	Alpha            int
	length           models.LengthSpec
	calculatedLength float64
	startTime        float64
}

// GetType returns the type of this item
func (i *ImageItem) GetType() string {
	return "image"
}

// GetLength returns the length specification
func (i *ImageItem) GetLength() models.LengthSpec {
	return i.length
}

// SetCalculatedLength sets the calculated length
func (i *ImageItem) SetCalculatedLength(seconds float64) {
	i.calculatedLength = seconds
}

// GetStartTime returns the start time
func (i *ImageItem) GetStartTime() float64 {
	return i.startTime
}

// SetStartTime sets the start time
func (i *ImageItem) SetStartTime(seconds float64) {
	i.startTime = seconds
}

// SetLength sets the length specification (for testing purposes)
func (i *ImageItem) SetLength(length models.LengthSpec) {
	i.length = length
}

// Helper functions
func getIntValue(data map[string]interface{}, key string, defaultValue int) int {
	if value, ok := data[key]; ok {
		if intValue, ok := value.(int); ok {
			return intValue
		}
		if floatValue, ok := value.(float64); ok {
			return int(floatValue)
		}
	}
	return defaultValue
}

func getFloatValue(data map[string]interface{}, key string, defaultValue float64) float64 {
	if value, ok := data[key]; ok {
		if floatValue, ok := value.(float64); ok {
			return floatValue
		}
		if intValue, ok := value.(int); ok {
			return float64(intValue)
		}
	}
	return defaultValue
}

func parseLengthSpec(lengthRaw interface{}) (models.LengthSpec, error) {
	switch v := lengthRaw.(type) {
	case float64:
		return models.LengthSpec{Type: models.LengthTypeNumeric, Value: v}, nil
	case int:
		return models.LengthSpec{Type: models.LengthTypeNumeric, Value: float64(v)}, nil
	case string:
		switch v {
		case "auto":
			return models.LengthSpec{Type: models.LengthTypeAuto}, nil
		case "until:SHOT_END":
			return models.LengthSpec{Type: models.LengthTypeUntilShotEnd}, nil
		case "until:SCENE_END":
			return models.LengthSpec{Type: models.LengthTypeUntilSceneEnd}, nil
		case "until:SEQUENCE_END":
			return models.LengthSpec{Type: models.LengthTypeUntilSequenceEnd}, nil
		default:
			return models.LengthSpec{}, fmt.Errorf("unknown length specification: %s", v)
		}
	default:
		return models.LengthSpec{}, fmt.Errorf("length must be a number or string, got %T", v)
	}
}