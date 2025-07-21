package models

import (
	"encoding/json"
	"fmt"
)

// YMMPProject represents the entire YMMP file
type YMMPProject struct {
	FilePath              string       `json:"FilePath"`
	SelectedTimelineIndex int          `json:"SelectedTimelineIndex"`
	Timelines             []Timeline   `json:"Timelines"`
	Characters            []Character  `json:"Characters"`
	CollapsedGroups       []string     `json:"CollapsedGroups"`
}

// Validate validates the YMMP project
func (p *YMMPProject) Validate() error {
	if len(p.Timelines) == 0 {
		return fmt.Errorf("at least one timeline is required")
	}

	if p.SelectedTimelineIndex >= len(p.Timelines) {
		return fmt.Errorf("selected timeline index out of range")
	}

	return nil
}

// Timeline represents a timeline in YMMP
type Timeline struct {
	ID            string        `json:"ID"`
	Name          string        `json:"Name"`
	VideoInfo     VideoInfo     `json:"VideoInfo"`
	VerticalLine  VerticalLine  `json:"VerticalLine"`
	Items         []interface{} `json:"Items"`
	LayerSettings LayerSettings `json:"LayerSettings"`
	CurrentFrame  int           `json:"CurrentFrame"`
	Length        int           `json:"Length"`
	MaxLayer      int           `json:"MaxLayer"`
}

// VideoInfo represents video settings
type VideoInfo struct {
	FPS    int `json:"FPS"`
	Hz     int `json:"Hz"`
	Width  int `json:"Width"`
	Height int `json:"Height"`
}

// VerticalLine represents vertical line settings
type VerticalLine struct {
	IsEnabled  bool        `json:"IsEnabled"`
	StartFrame int         `json:"StartFrame"`
	LineType   interface{} `json:"LineType"`
	Line       interface{} `json:"Line"`
	Group      int         `json:"Group"`
}

// LayerSettings represents layer settings
type LayerSettings struct {
	Items []LayerSetting `json:"Items"`
}

// LayerSetting represents individual layer setting
type LayerSetting struct {
	Layer    int     `json:"Layer"`
	Label    *string `json:"Label"`
	Color    string  `json:"Color"`
	IsHidden bool    `json:"IsHidden"`
	Volume   float64 `json:"Volume"`
}

// Character represents a character configuration
type Character struct {
	Name      string                 `json:"Name"`
	GroupName string                 `json:"GroupName"`
	Color     string                 `json:"Color"`
	Layer     int                    `json:"Layer"`
	Voice     map[string]interface{} `json:"Voice"`
	// Simplified for now - can be expanded as needed
}

// BaseItem contains common fields for all item types
type BaseItem struct {
	Frame     int    `json:"Frame"`
	Layer     int    `json:"Layer"`
	Length    int    `json:"Length"`
	Remark    string `json:"Remark"`
	IsLocked  bool   `json:"IsLocked"`
	IsHidden  bool   `json:"IsHidden"`
	Group     int    `json:"Group"`
	KeyFrames struct {
		Frames []interface{} `json:"Frames"`
		Count  int           `json:"Count"`
	} `json:"KeyFrames"`
}

// GetRemark returns the remark of the item
func (b *BaseItem) GetRemark() string {
	return b.Remark
}

// VoiceItem represents a voice item
type VoiceItem struct {
	Type          string                 `json:"$type"`
	BaseItem                             // Embedded
	CharacterName string                 `json:"CharacterName"`
	Serif         string                 `json:"Serif"`
	Hatsuon       string                 `json:"Hatsuon"`
	VoiceLength   interface{}            `json:"VoiceLength"`
	X             AnimatedValue          `json:"X"`
	Y             AnimatedValue          `json:"Y"`
	Font          string                 `json:"Font"`
	FontSize      AnimatedValue          `json:"FontSize"`
	FontColor     string                 `json:"FontColor"`
	Style         string                 `json:"Style"`
	StyleColor    string                 `json:"StyleColor"`
	// Additional fields can be added as needed
}

// GetType returns the type of the voice item
func (v *VoiceItem) GetType() string {
	return v.Type
}

// VideoItem represents a video item
type VideoItem struct {
	Type         string                 `json:"$type"`
	BaseItem                            // Embedded
	FilePath     string                 `json:"FilePath"`
	Volume       AnimatedValue          `json:"Volume"`
	PlaybackRate float64                `json:"PlaybackRate"`
	X            AnimatedValue          `json:"X"`
	Y            AnimatedValue          `json:"Y"`
	Zoom         AnimatedValue          `json:"Zoom"`
	// Additional fields can be added as needed
}

// TachieItem represents a tachie (standing picture) item
type TachieItem struct {
	Type               string                 `json:"$type"`
	BaseItem                                  // Embedded
	CharacterName      string                 `json:"CharacterName"`
	TachieItemParameter map[string]interface{} `json:"TachieItemParameter"`
	X                  AnimatedValue          `json:"X"`
	Y                  AnimatedValue          `json:"Y"`
	Zoom               AnimatedValue          `json:"Zoom"`
	// Additional fields can be added as needed
}

// AnimatedValue represents an animated property value
type AnimatedValue struct {
	Values        []ValuePoint `json:"Values"`
	Span          float64      `json:"Span"`
	AnimationType string       `json:"AnimationType"`
}

// ValuePoint represents a value at a specific point
type ValuePoint struct {
	Value float64 `json:"Value"`
}

// ItemType constants
const (
	ItemTypeVoice  = "YukkuriMovieMaker.Project.Items.VoiceItem, YukkuriMovieMaker"
	ItemTypeVideo  = "YukkuriMovieMaker.Project.Items.VideoItem, YukkuriMovieMaker"
	ItemTypeTachie = "YukkuriMovieMaker.Project.Items.TachieItem, YukkuriMovieMaker"
)

// UnmarshalJSON custom unmarshaler for Timeline to handle polymorphic Items
func (t *Timeline) UnmarshalJSON(data []byte) error {
	type Alias Timeline
	aux := &struct {
		Items []json.RawMessage `json:"Items"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse items based on their type
	t.Items = make([]interface{}, len(aux.Items))
	for i, rawItem := range aux.Items {
		var typeCheck struct {
			Type string `json:"$type"`
		}

		if err := json.Unmarshal(rawItem, &typeCheck); err != nil {
			return err
		}

		switch typeCheck.Type {
		case ItemTypeVoice, "YukkuriMovieMaker.Project.VoiceItem":
			var voiceItem VoiceItem
			if err := json.Unmarshal(rawItem, &voiceItem); err != nil {
				return err
			}
			t.Items[i] = &voiceItem
		case ItemTypeVideo, "YukkuriMovieMaker.Project.VideoItem":
			var videoItem VideoItem
			if err := json.Unmarshal(rawItem, &videoItem); err != nil {
				return err
			}
			t.Items[i] = &videoItem
		case ItemTypeTachie, "YukkuriMovieMaker.Project.TachieItem":
			var tachieItem TachieItem
			if err := json.Unmarshal(rawItem, &tachieItem); err != nil {
				return err
			}
			t.Items[i] = &tachieItem
		default:
			// Store as map for unknown types
			var unknownItem map[string]interface{}
			if err := json.Unmarshal(rawItem, &unknownItem); err != nil {
				return err
			}
			t.Items[i] = unknownItem
		}
	}

	return nil
}