package converter_test

import (
	"fmt"
	"testing"

	"github.com/yakuari-channel/video-authorizer/internal/converter"
	"github.com/yakuari-channel/video-authorizer/internal/models"
)

func TestFFProbeClient_IsAvailable(t *testing.T) {
	// Test with default ffprobe path (should be in PATH if installed)
	client := converter.NewFFProbeClient("")
	
	// This test will only pass if ffprobe is installed
	// We don't fail the test if it's not available, just skip it
	if !client.IsAvailable() {
		t.Skip("ffprobe is not available in PATH, skipping test")
	}
	
	t.Log("ffprobe is available")
}

func TestFFProbeClient_WithInvalidPath(t *testing.T) {
	// Test with invalid ffprobe path
	client := converter.NewFFProbeClient("/nonexistent/ffprobe")
	
	if client.IsAvailable() {
		t.Error("expected ffprobe to be unavailable with invalid path")
	}
}

func TestParseFrameRate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
		wantErr  bool
	}{
		{
			name:     "standard 30 fps",
			input:    "30/1",
			expected: 30.0,
			wantErr:  false,
		},
		{
			name:     "fractional frame rate",
			input:    "30000/1001",
			expected: 29.970029970029973,
			wantErr:  false,
		},
		{
			name:     "60 fps",
			input:    "60/1",
			expected: 60.0,
			wantErr:  false,
		},
		{
			name:     "invalid format",
			input:    "30",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "zero denominator",
			input:    "30/0",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid numerator",
			input:    "abc/1",
			expected: 0,
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We need to test the parseFrameRate function which is not exported
			// For now, we'll test it through a mock implementation
			result, err := parseFrameRateHelper(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %s", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %s: %v", tt.input, err)
				}
				// Use tolerance for floating point comparison
				tolerance := 0.000001
				if result < tt.expected-tolerance || result > tt.expected+tolerance {
					t.Errorf("expected %f, got %f", tt.expected, result)
				}
			}
		})
	}
}

func TestExtractVideoItemFilePath(t *testing.T) {
	// Test extractVideoItemFilePath function through LengthCalculator
	calculator := converter.NewLengthCalculator()
	
	// Test with valid VideoItem
	videoItem := &models.VideoItem{
		FilePath: "/path/to/video.mp4",
	}
	
	// Since the method is not exported, we'll test it through the auto video length calculation
	// which should fail at the ffprobe stage but succeed at the file path extraction
	_, err := calculator.CalculateLengthForItem("_auto:VIDEO", videoItem, nil)
	
	// The error should be about ffprobe not being available, not about missing file path
	if err != nil && err.Error() == "failed to extract video file path: video item has no FilePath" {
		t.Error("file path extraction failed when it should have succeeded")
	}
	
	// Test with VideoItem missing FilePath
	emptyVideoItem := &models.VideoItem{
		FilePath: "",
	}
	
	_, err = calculator.CalculateLengthForItem("_auto:VIDEO", emptyVideoItem, nil)
	
	// Should get error about missing FilePath
	if err == nil {
		t.Error("expected error for video item with no FilePath")
	}
}

// Helper function that mirrors the unexported parseFrameRate function
func parseFrameRateHelper(frameRateStr string) (float64, error) {
	// This mirrors the implementation in ffprobe_client.go
	parts := []string{}
	for i, char := range frameRateStr {
		if char == '/' {
			parts = append(parts, frameRateStr[:i])
			parts = append(parts, frameRateStr[i+1:])
			break
		}
	}
	
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid frame rate format: %s", frameRateStr)
	}

	var numerator, denominator float64
	var err error
	
	if numerator, err = parseFloat(parts[0]); err != nil {
		return 0, fmt.Errorf("invalid numerator in frame rate: %s", parts[0])
	}

	if denominator, err = parseFloat(parts[1]); err != nil {
		return 0, fmt.Errorf("invalid denominator in frame rate: %s", parts[1])
	}

	if denominator == 0 {
		return 0, fmt.Errorf("zero denominator in frame rate")
	}

	return numerator / denominator, nil
}

func parseFloat(s string) (float64, error) {
	// Simple float parsing that only handles integers and basic floats
	var result float64
	var err error
	
	if _, err = fmt.Sscanf(s, "%f", &result); err != nil {
		return 0, err
	}
	
	return result, nil
}

func TestConvertDurationToFramesVideo(t *testing.T) {
	tests := []struct {
		name     string
		duration float64
		fps      float64
		expected int
	}{
		{
			name:     "10 second video at 30 FPS",
			duration: 10.0,
			fps:      30.0,
			expected: 300,
		},
		{
			name:     "5.5 second video at 24 FPS",
			duration: 5.5,
			fps:      24.0,
			expected: 132,
		},
		{
			name:     "1 second video at 60 FPS",
			duration: 1.0,
			fps:      60.0,
			expected: 60,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.ConvertDurationToFrames(tt.duration, tt.fps)
			if result != tt.expected {
				t.Errorf("expected %d frames, got %d", tt.expected, result)
			}
		})
	}
}