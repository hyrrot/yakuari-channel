package items

import (
	"fmt"
	
	"github.com/user/ymmp-compiler/internal/models"
)

// TachiePlugin handles tachie (standing picture) items
type TachiePlugin struct{}

// NewTachiePlugin creates a new tachie plugin
func NewTachiePlugin() *TachiePlugin {
	return &TachiePlugin{}
}

// GetType returns the type of this plugin
func (p *TachiePlugin) GetType() string {
	return "tachie"
}

// ParseYAML parses YAML data into a TachieItem
func (p *TachiePlugin) ParseYAML(data interface{}) (models.Item, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("tachie item data must be a map")
	}
	
	// Required field: item (character name)
	itemRaw, ok := dataMap["item"]
	if !ok {
		return nil, fmt.Errorf("tachie item requires 'item' field")
	}
	
	item, ok := itemRaw.(string)
	if !ok {
		return nil, fmt.Errorf("item must be a string")
	}
	
	tachieItem := &TachieItem{
		Item:     item,
		X:        getIntValue(dataMap, "x", 0),
		Y:        getIntValue(dataMap, "y", 0),
		Z:        getIntValue(dataMap, "z", 0),
		Zoom:     getFloatValue(dataMap, "zoom", 1.0),
		Rotation: getFloatValue(dataMap, "rotation", 0.0),
		Alpha:    getIntValue(dataMap, "alpha", 255),
		FadeIn:   getFloatValue(dataMap, "fade_in", 0.0),
		FadeOut:  getFloatValue(dataMap, "fade_out", 0.0),
	}
	
	// Parse length
	if lengthRaw, ok := dataMap["length"]; ok {
		length, err := parseLengthSpec(lengthRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid length specification: %v", err)
		}
		tachieItem.length = length
	} else {
		// Tachie items require explicit length specification
		return nil, fmt.Errorf("tachie item requires 'length' field")
	}
	
	return tachieItem, nil
}

// ConvertToYMMP converts a TachieItem to YMMP format
func (p *TachiePlugin) ConvertToYMMP(item models.Item, basePath string) (*models.YMMPItem, error) {
	tachieItem, ok := item.(*TachieItem)
	if !ok {
		return nil, fmt.Errorf("expected TachieItem, got %T", item)
	}
	
	// Convert seconds to frames (assuming 60 FPS)
	lengthInFrames := int(tachieItem.length.Value * 60)
	
	ymmItem := &models.YMMPItem{
		Type:     "YMM.Parts.TachiePart",
		Layer:    5, // Default layer for tachie
		Frame:    int(tachieItem.startTime * 60), // Convert start time to frames
		Length:   lengthInFrames,
		FilePath: "", // Tachie items don't have file paths
		Extended: map[string]interface{}{
			"Item":     tachieItem.Item,
			"X":        tachieItem.X,
			"Y":        tachieItem.Y,
			"Z":        tachieItem.Z,
			"Zoom":     tachieItem.Zoom,
			"Rotation": tachieItem.Rotation,
			"Alpha":    tachieItem.Alpha,
			"FadeIn":   tachieItem.FadeIn,
			"FadeOut":  tachieItem.FadeOut,
		},
	}
	
	return ymmItem, nil
}

// TachieItem represents a tachie (standing picture) item
type TachieItem struct {
	Item                 string
	X, Y, Z              int
	Zoom, Rotation       float64
	Alpha                int
	FadeIn, FadeOut      float64
	length               models.LengthSpec
	calculatedLength     float64
	startTime            float64
}

// GetType returns the type of this item
func (t *TachieItem) GetType() string {
	return "tachie"
}

// GetLength returns the length specification
func (t *TachieItem) GetLength() models.LengthSpec {
	return t.length
}

// SetCalculatedLength sets the calculated length
func (t *TachieItem) SetCalculatedLength(seconds float64) {
	t.calculatedLength = seconds
}

// GetStartTime returns the start time
func (t *TachieItem) GetStartTime() float64 {
	return t.startTime
}

// SetStartTime sets the start time
func (t *TachieItem) SetStartTime(seconds float64) {
	t.startTime = seconds
}