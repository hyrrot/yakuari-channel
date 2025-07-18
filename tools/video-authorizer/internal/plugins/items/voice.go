package items

import (
	"fmt"
	
	"github.com/user/ymmp-compiler/internal/models"
)

// VoicePlugin handles voice items
type VoicePlugin struct{}

// NewVoicePlugin creates a new voice plugin
func NewVoicePlugin() *VoicePlugin {
	return &VoicePlugin{}
}

// GetType returns the type of this plugin
func (p *VoicePlugin) GetType() string {
	return "voice"
}

// ParseYAML parses YAML data into a VoiceItem
func (p *VoicePlugin) ParseYAML(data interface{}) (models.Item, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("voice item data must be a map")
	}
	
	// Required field: line
	lineRaw, ok := dataMap["line"]
	if !ok {
		return nil, fmt.Errorf("voice item requires 'line' field")
	}
	
	line, ok := lineRaw.(string)
	if !ok {
		return nil, fmt.Errorf("line must be a string")
	}
	
	item := &VoiceItem{
		Line:          line,
		CharacterName: getStringValue(dataMap, "character_name", ""),
		VoiceID:       getIntValue(dataMap, "voice_id", 0),
		Speed:         getFloatValue(dataMap, "speed", 1.0),
		Pitch:         getFloatValue(dataMap, "pitch", 0.0),
		Volume:        getFloatValue(dataMap, "volume", 1.0),
	}
	
	// Parse length
	if lengthRaw, ok := dataMap["length"]; ok {
		length, err := parseLengthSpec(lengthRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid length specification: %v", err)
		}
		item.length = length
	} else {
		// Default to auto length for voice items
		item.length = models.LengthSpec{Type: models.LengthTypeAuto}
	}
	
	return item, nil
}

// ConvertToYMMP converts a VoiceItem to YMMP format
func (p *VoicePlugin) ConvertToYMMP(item models.Item, basePath string) (*models.YMMPItem, error) {
	voiceItem, ok := item.(*VoiceItem)
	if !ok {
		return nil, fmt.Errorf("expected VoiceItem, got %T", item)
	}
	
	// Convert seconds to frames (assuming 60 FPS)
	var lengthInFrames int
	if voiceItem.length.Type == models.LengthTypeNumeric {
		lengthInFrames = int(voiceItem.length.Value * 60)
	} else {
		// For auto length, use calculated length
		lengthInFrames = int(voiceItem.calculatedLength * 60)
	}
	
	ymmItem := &models.YMMPItem{
		Type:     "YMM.Parts.VoicePart",
		Layer:    2, // Default layer for voice
		Frame:    int(voiceItem.startTime * 60), // Convert start time to frames
		Length:   lengthInFrames,
		FilePath: "", // Voice items don't have file paths directly
		Extended: map[string]interface{}{
			"Line":          voiceItem.Line,
			"CharacterName": voiceItem.CharacterName,
			"VoiceID":       voiceItem.VoiceID,
			"Speed":         voiceItem.Speed,
			"Pitch":         voiceItem.Pitch,
			"Volume":        voiceItem.Volume,
		},
	}
	
	return ymmItem, nil
}

// VoiceItem represents a voice item
type VoiceItem struct {
	Line          string
	CharacterName string
	VoiceID       int
	Speed         float64
	Pitch         float64
	Volume        float64
	length        models.LengthSpec
	calculatedLength float64
	startTime     float64
}

// GetType returns the type of this item
func (v *VoiceItem) GetType() string {
	return "voice"
}

// GetLength returns the length specification
func (v *VoiceItem) GetLength() models.LengthSpec {
	return v.length
}

// SetCalculatedLength sets the calculated length
func (v *VoiceItem) SetCalculatedLength(seconds float64) {
	v.calculatedLength = seconds
}

// GetStartTime returns the start time
func (v *VoiceItem) GetStartTime() float64 {
	return v.startTime
}

// SetStartTime sets the start time
func (v *VoiceItem) SetStartTime(seconds float64) {
	v.startTime = seconds
}

// SetLength sets the length specification (for testing purposes)
func (v *VoiceItem) SetLength(length models.LengthSpec) {
	v.length = length
}

// GetCalculatedLength returns the calculated length
func (v *VoiceItem) GetCalculatedLength() float64 {
	return v.calculatedLength
}

// Helper function for string values
func getStringValue(data map[string]interface{}, key string, defaultValue string) string {
	if value, ok := data[key]; ok {
		if stringValue, ok := value.(string); ok {
			return stringValue
		}
	}
	return defaultValue
}