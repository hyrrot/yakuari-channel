package models

import (
	"fmt"
	"strconv"
	"strings"
)

// YMMPSDocument represents the entire YMMPS file
type YMMPSDocument struct {
	YMMPSVersion string     `yaml:"YMMPSVersion"`
	Sequences    []Sequence `yaml:"Sequences"`
}

// Validate validates the YMMPS document
func (d *YMMPSDocument) Validate() error {
	if d.YMMPSVersion != "1" {
		return fmt.Errorf("unsupported YMMPS version: %s", d.YMMPSVersion)
	}

	if len(d.Sequences) == 0 {
		return fmt.Errorf("at least one sequence is required")
	}

	// Validate each sequence
	for i, seq := range d.Sequences {
		if err := seq.Validate(); err != nil {
			return fmt.Errorf("sequence[%d]: %w", i, err)
		}
	}

	return nil
}

// Sequence represents a sequence containing multiple scenes
type Sequence struct {
	ID     string  `yaml:"ID,omitempty"`
	Scenes []Scene `yaml:"Scenes"`
}

// Validate validates the sequence
func (s *Sequence) Validate() error {
	if len(s.Scenes) == 0 {
		return fmt.Errorf("at least one scene is required")
	}

	for i, scene := range s.Scenes {
		if err := scene.Validate(); err != nil {
			return fmt.Errorf("scene[%d]: %w", i, err)
		}
	}

	return nil
}

// Scene represents a scene containing multiple shots
type Scene struct {
	ID    string `yaml:"ID,omitempty"`
	Shots []Shot `yaml:"Shots"`
}

// Validate validates the scene
func (s *Scene) Validate() error {
	if len(s.Shots) == 0 {
		return fmt.Errorf("at least one shot is required")
	}

	for i, shot := range s.Shots {
		if err := shot.Validate(); err != nil {
			return fmt.Errorf("shot[%d]: %w", i, err)
		}
	}

	return nil
}

// Shot represents a shot containing multiple items that start simultaneously
type Shot struct {
	ID    string     `yaml:"ID,omitempty"`
	Items []ItemSpec `yaml:"Items,omitempty"`
}

// Validate validates the shot
func (s *Shot) Validate() error {
	if len(s.Items) == 0 {
		shotDesc := "shot"
		if s.ID != "" {
			shotDesc = fmt.Sprintf("shot '%s'", s.ID)
		}
		return fmt.Errorf("%s must contain at least one item in the 'Items' array", shotDesc)
	}

	for i, item := range s.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("item[%d]: %w", i, err)
		}
	}

	return nil
}

// ItemSpec represents an item specification in YMMPS
type ItemSpec struct {
	Template   string                 `yaml:"_Template"`
	Length     string                 `yaml:"Length"`
	Properties map[string]interface{} `yaml:",inline"`
}

// Validate validates the item spec
func (i *ItemSpec) Validate() error {
	if i.Template == "" {
		return fmt.Errorf("_Template is required")
	}

	if i.Length == "" {
		return fmt.Errorf("Length is required")
	}

	// Validate length format
	_, err := ParseLength(i.Length)
	if err != nil {
		return fmt.Errorf("invalid Length format: %w", err)
	}

	return nil
}

// LengthType represents the type of length specification
type LengthType string

const (
	LengthTypeNumeric       LengthType = "numeric"
	LengthTypeUntilSeqEnd   LengthType = "until_seq"
	LengthTypeUntilSceneEnd LengthType = "until_scene"
	LengthTypeUntilShotEnd  LengthType = "until_shot"
	LengthTypeUntilIDEnd    LengthType = "until_id"
	LengthTypeAutoVoice     LengthType = "auto_voice"
	LengthTypeAutoVideo     LengthType = "auto_video"
)

// ParsedLength represents parsed length information
type ParsedLength struct {
	Type       LengthType
	Value      int    // for numeric type
	TargetID   string // for until_id type
	TargetType string // for until_id type (SEQUENCE/SCENE/SHOT)
}

// ParseLength parses a length string into ParsedLength
func ParseLength(length string) (ParsedLength, error) {
	// Check if it's a numeric value
	if num, err := strconv.Atoi(length); err == nil {
		return ParsedLength{
			Type:  LengthTypeNumeric,
			Value: num,
		}, nil
	}

	// Handle natural language format (e.g., "UNTIL SHOT END")
	upperLength := strings.ToUpper(length)
	if strings.HasPrefix(upperLength, "UNTIL ") {
		if strings.HasSuffix(upperLength, " SEQUENCE END") {
			return ParsedLength{Type: LengthTypeUntilSeqEnd}, nil
		}
		if strings.HasSuffix(upperLength, " SCENE END") {
			return ParsedLength{Type: LengthTypeUntilSceneEnd}, nil
		}
		if strings.HasSuffix(upperLength, " SHOT END") {
			return ParsedLength{Type: LengthTypeUntilShotEnd}, nil
		}
		
		// Handle "UNTIL <TYPE> <ID> END" format
		parts := strings.Fields(upperLength)
		if len(parts) == 4 && parts[0] == "UNTIL" && parts[3] == "END" {
			targetType := parts[1]
			targetID := strings.Fields(length)[2] // Get original case ID
			
			switch targetType {
			case "SEQUENCE":
				return ParsedLength{
					Type:       LengthTypeUntilIDEnd,
					TargetType: "SEQUENCE",
					TargetID:   targetID,
				}, nil
			case "SCENE":
				return ParsedLength{
					Type:       LengthTypeUntilIDEnd,
					TargetType: "SCENE",
					TargetID:   targetID,
				}, nil
			case "SHOT":
				return ParsedLength{
					Type:       LengthTypeUntilIDEnd,
					TargetType: "SHOT",
					TargetID:   targetID,
				}, nil
			}
		}
	}

	// Check for special formats (legacy support)
	if strings.HasPrefix(length, "_until:") {
		parts := strings.Split(length, ":")
		if len(parts) < 2 {
			return ParsedLength{}, fmt.Errorf("invalid _until format: %s", length)
		}

		switch parts[1] {
		case "SEQUENCE_END":
			if len(parts) == 2 {
				return ParsedLength{Type: LengthTypeUntilSeqEnd}, nil
			} else if len(parts) == 3 {
				return ParsedLength{
					Type:       LengthTypeUntilIDEnd,
					TargetType: "SEQUENCE",
					TargetID:   parts[2],
				}, nil
			}
		case "SCENE_END":
			if len(parts) == 2 {
				return ParsedLength{Type: LengthTypeUntilSceneEnd}, nil
			} else if len(parts) == 3 {
				return ParsedLength{
					Type:       LengthTypeUntilIDEnd,
					TargetType: "SCENE",
					TargetID:   parts[2],
				}, nil
			}
		case "SHOT_END":
			if len(parts) == 2 {
				return ParsedLength{Type: LengthTypeUntilShotEnd}, nil
			} else if len(parts) == 3 {
				return ParsedLength{
					Type:       LengthTypeUntilIDEnd,
					TargetType: "SHOT",
					TargetID:   parts[2],
				}, nil
			}
		}
	}

	if strings.HasPrefix(length, "_auto:") {
		parts := strings.Split(length, ":")
		if len(parts) != 2 {
			return ParsedLength{}, fmt.Errorf("invalid _auto format: %s", length)
		}

		switch parts[1] {
		case "VOICEVOX":
			return ParsedLength{Type: LengthTypeAutoVoice}, nil
		case "VIDEO":
			return ParsedLength{Type: LengthTypeAutoVideo}, nil
		}
	}

	return ParsedLength{}, fmt.Errorf("unknown length format: %s", length)
}