package converter

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// FFProbeClient handles video/audio file analysis using ffprobe
type FFProbeClient struct {
	ffprobePath string
}

// NewFFProbeClient creates a new ffprobe client
func NewFFProbeClient(ffprobePath string) *FFProbeClient {
	if ffprobePath == "" {
		ffprobePath = "ffprobe" // Assume ffprobe is in PATH
	}
	
	return &FFProbeClient{
		ffprobePath: ffprobePath,
	}
}

// FFProbeFormat represents format information from ffprobe
type FFProbeFormat struct {
	Duration string `json:"duration"`
	Size     string `json:"size"`
	BitRate  string `json:"bit_rate"`
}

// FFProbeStream represents stream information from ffprobe
type FFProbeStream struct {
	CodecType string `json:"codec_type"`
	Duration  string `json:"duration"`
	Width     *int   `json:"width,omitempty"`
	Height    *int   `json:"height,omitempty"`
	RFrameRate string `json:"r_frame_rate"`
}

// FFProbeOutput represents the complete ffprobe output
type FFProbeOutput struct {
	Format  FFProbeFormat   `json:"format"`
	Streams []FFProbeStream `json:"streams"`
}

// GetVideoDuration gets the duration of a video file in seconds
func (fpc *FFProbeClient) GetVideoDuration(filePath string) (float64, error) {
	// Check if ffprobe is available
	if !fpc.IsAvailable() {
		return 0, fmt.Errorf("ffprobe is not available at path: %s", fpc.ffprobePath)
	}

	// Run ffprobe command to get duration
	cmd := exec.Command(fpc.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe command failed: %w", err)
	}

	// Parse JSON output
	var probeOutput FFProbeOutput
	if err := json.Unmarshal(output, &probeOutput); err != nil {
		return 0, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Try to get duration from format first
	if probeOutput.Format.Duration != "" {
		duration, err := strconv.ParseFloat(probeOutput.Format.Duration, 64)
		if err == nil {
			return duration, nil
		}
	}

	// If format duration is not available, try to get from video stream
	for _, stream := range probeOutput.Streams {
		if stream.CodecType == "video" && stream.Duration != "" {
			duration, err := strconv.ParseFloat(stream.Duration, 64)
			if err == nil {
				return duration, nil
			}
		}
	}

	return 0, fmt.Errorf("could not determine video duration from ffprobe output")
}

// GetVideoInfo gets comprehensive video information
func (fpc *FFProbeClient) GetVideoInfo(filePath string) (*VideoInfo, error) {
	if !fpc.IsAvailable() {
		return nil, fmt.Errorf("ffprobe is not available at path: %s", fpc.ffprobePath)
	}

	cmd := exec.Command(fpc.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe command failed: %w", err)
	}

	var probeOutput FFProbeOutput
	if err := json.Unmarshal(output, &probeOutput); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	info := &VideoInfo{
		FilePath: filePath,
	}

	// Extract format information
	if probeOutput.Format.Duration != "" {
		if duration, err := strconv.ParseFloat(probeOutput.Format.Duration, 64); err == nil {
			info.Duration = duration
		}
	}

	// Extract video stream information
	for _, stream := range probeOutput.Streams {
		if stream.CodecType == "video" {
			if stream.Width != nil {
				info.Width = *stream.Width
			}
			if stream.Height != nil {
				info.Height = *stream.Height
			}
			if stream.RFrameRate != "" {
				if fps, err := parseFrameRate(stream.RFrameRate); err == nil {
					info.FrameRate = fps
				}
			}
			if info.Duration == 0 && stream.Duration != "" {
				if duration, err := strconv.ParseFloat(stream.Duration, 64); err == nil {
					info.Duration = duration
				}
			}
			break
		}
	}

	return info, nil
}

// VideoInfo represents video file information
type VideoInfo struct {
	FilePath  string
	Duration  float64 // in seconds
	Width     int
	Height    int
	FrameRate float64
}

// IsAvailable checks if ffprobe is available
func (fpc *FFProbeClient) IsAvailable() bool {
	cmd := exec.Command(fpc.ffprobePath, "-version")
	err := cmd.Run()
	return err == nil
}

// parseFrameRate parses frame rate from ffprobe format (e.g., "30/1" -> 30.0)
func parseFrameRate(frameRateStr string) (float64, error) {
	parts := strings.Split(frameRateStr, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid frame rate format: %s", frameRateStr)
	}

	numerator, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numerator in frame rate: %s", parts[0])
	}

	denominator, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid denominator in frame rate: %s", parts[1])
	}

	if denominator == 0 {
		return 0, fmt.Errorf("zero denominator in frame rate")
	}

	return numerator / denominator, nil
}