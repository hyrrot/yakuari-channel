package converter

import (
	"fmt"
	"path/filepath"
	"strings"
)

// MockFFProbeClient is a mock implementation of FFProbeClientInterface for testing
type MockFFProbeClient struct {
	Available bool
	// Durations maps file paths to their durations in seconds
	Durations map[string]float64
	// VideoInfos maps file paths to their complete video information
	VideoInfos map[string]*VideoInfo
}

// NewMockFFProbeClient creates a new mock FFProbe client
func NewMockFFProbeClient(available bool) *MockFFProbeClient {
	return &MockFFProbeClient{
		Available:  available,
		Durations:  make(map[string]float64),
		VideoInfos: make(map[string]*VideoInfo),
	}
}

// SetDuration sets the mock duration for a specific file path
func (m *MockFFProbeClient) SetDuration(filePath string, duration float64) {
	m.Durations[filePath] = duration
}

// SetVideoInfo sets the mock video information for a specific file path
func (m *MockFFProbeClient) SetVideoInfo(filePath string, info *VideoInfo) {
	m.VideoInfos[filePath] = info
}

// GetVideoDuration returns the mock duration for the specified file path
func (m *MockFFProbeClient) GetVideoDuration(filePath string) (float64, error) {
	if !m.Available {
		return 0, fmt.Errorf("ffprobe is not available")
	}

	// Check for exact path match first
	if duration, exists := m.Durations[filePath]; exists {
		return duration, nil
	}

	// Check for filename-only match (for relative paths)
	fileName := filepath.Base(filePath)
	for path, duration := range m.Durations {
		if filepath.Base(path) == fileName {
			return duration, nil
		}
	}

	// Fallback: generate duration based on file extension and name
	return m.generateFallbackDuration(filePath), nil
}

// GetVideoInfo returns mock video information for the specified file path
func (m *MockFFProbeClient) GetVideoInfo(filePath string) (*VideoInfo, error) {
	if !m.Available {
		return nil, fmt.Errorf("ffprobe is not available")
	}

	// Check for exact path match first
	if info, exists := m.VideoInfos[filePath]; exists {
		return info, nil
	}

	// Check for filename-only match
	fileName := filepath.Base(filePath)
	for path, info := range m.VideoInfos {
		if filepath.Base(path) == fileName {
			return info, nil
		}
	}

	// Generate fallback video info
	duration := m.generateFallbackDuration(filePath)
	return &VideoInfo{
		FilePath:  filePath,
		Duration:  duration,
		Width:     1920, // Standard HD width
		Height:    1080, // Standard HD height
		FrameRate: 30.0, // Standard frame rate
	}, nil
}

// IsAvailable returns whether the mock FFProbe client is available
func (m *MockFFProbeClient) IsAvailable() bool {
	return m.Available
}

// generateFallbackDuration generates a fallback duration based on file characteristics
func (m *MockFFProbeClient) generateFallbackDuration(filePath string) float64 {
	fileName := strings.ToLower(filepath.Base(filePath))
	
	// Generate duration based on filename patterns
	if strings.Contains(fileName, "short") {
		return 5.0 // 5 seconds for files with "short" in name
	} else if strings.Contains(fileName, "long") {
		return 30.0 // 30 seconds for files with "long" in name
	} else if strings.Contains(fileName, "intro") {
		return 10.0 // 10 seconds for intro videos
	} else if strings.Contains(fileName, "outro") {
		return 8.0 // 8 seconds for outro videos
	} else if strings.Contains(fileName, "bgm") || strings.Contains(fileName, "music") {
		return 180.0 // 3 minutes for background music videos
	} else {
		// Default fallback: 15 seconds
		return 15.0
	}
}

// Ensure MockFFProbeClient implements FFProbeClientInterface
var _ FFProbeClientInterface = (*MockFFProbeClient)(nil)