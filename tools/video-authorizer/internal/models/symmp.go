package models

// Episode represents the entire project
type Episode struct {
	SYMMPFormatVersion string                 `yaml:"symmp_format_version"`
	Project            ProjectConfig          `yaml:"project"`
	Defaults           map[string]interface{} `yaml:"defaults,omitempty"`
	Sequences          []Sequence             `yaml:"sequences"`
}

// ProjectConfig contains project settings
type ProjectConfig struct {
	Name      string `yaml:"name"`
	OutputDir string `yaml:"output_dir"`
}

// Sequence represents a sequence of scenes
type Sequence struct {
	ID     string  `yaml:"id"`
	Scenes []Scene `yaml:"scenes"`
}

// Scene represents a scene containing shots
type Scene struct {
	ID    string `yaml:"id"`
	Shots []Shot `yaml:"shots"`
}

// Shot represents a shot containing items
type Shot struct {
	Items []Item `yaml:"-"` // Will be unmarshaled dynamically
}

// Item is the base interface for all item types
type Item interface {
	GetType() string
	GetLength() LengthSpec
	SetCalculatedLength(seconds float64)
	GetStartTime() float64
	SetStartTime(seconds float64)
}

// LengthSpec represents different types of length specifications
type LengthSpec struct {
	Type  LengthType
	Value float64 // Used for numeric values
}

// LengthType defines the type of length specification
type LengthType string

const (
	LengthTypeNumeric           LengthType = "numeric"
	LengthTypeAuto              LengthType = "auto"
	LengthTypeUntilShotEnd      LengthType = "until:SHOT_END"
	LengthTypeUntilSceneEnd     LengthType = "until:SCENE_END"
	LengthTypeUntilSequenceEnd  LengthType = "until:SEQUENCE_END"
)