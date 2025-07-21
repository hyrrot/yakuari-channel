package converter

import "strings"

// MockVoicevoxClient is a mock implementation of VoicevoxClientInterface for testing
type MockVoicevoxClient struct {
	Available bool
	Durations map[string]float64 // text -> duration mapping
}

// NewMockVoicevoxClient creates a new mock VOICEVOX client
func NewMockVoicevoxClient(available bool) *MockVoicevoxClient {
	return &MockVoicevoxClient{
		Available: available,
		Durations: make(map[string]float64),
	}
}

// GetAudioDuration returns mock duration based on text length
func (m *MockVoicevoxClient) GetAudioDuration(text string, speaker int) (float64, error) {
	// If specific duration is set for this text, use it
	if duration, exists := m.Durations[text]; exists {
		return duration, nil
	}
	
	// Default fallback: calculate based on text length
	// Assume 1 character = 0.15 seconds (rough estimate for Japanese)
	runes := []rune(strings.TrimSpace(text))
	if len(runes) == 0 {
		return 1.0, nil // minimum duration for empty text
	}
	
	return float64(len(runes)) * 0.15, nil
}

// IsAvailable returns the mock availability status
func (m *MockVoicevoxClient) IsAvailable() bool {
	return m.Available
}

// IsAvailableWithLogging returns the mock availability status
func (m *MockVoicevoxClient) IsAvailableWithLogging(verbose bool) bool {
	return m.Available
}

// SetDuration sets a specific duration for a given text
func (m *MockVoicevoxClient) SetDuration(text string, duration float64) {
	m.Durations[text] = duration
}

// SetAvailable sets the availability status
func (m *MockVoicevoxClient) SetAvailable(available bool) {
	m.Available = available
}