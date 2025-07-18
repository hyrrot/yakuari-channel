package models

// YMMPProject represents the YMMP file structure
type YMMPProject struct {
	Timeline Timeline `json:"Timeline"`
}

// Timeline contains all items in the timeline
type Timeline struct {
	Items []YMMPItem `json:"Items"`
}

// YMMPItem represents an item in YMMP format
type YMMPItem struct {
	Type     string                 `json:"$type"`
	Layer    int                    `json:"Layer"`
	Frame    int                    `json:"Frame"`
	Length   int                    `json:"Length"`
	FilePath string                 `json:"FilePath,omitempty"`
	Extended map[string]interface{} `json:",inline"`
}