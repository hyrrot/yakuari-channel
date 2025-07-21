package converter

// FFProbeClientInterface defines the interface for FFProbe client operations
// This interface allows for dependency injection and mocking in tests
type FFProbeClientInterface interface {
	// GetVideoDuration gets the duration of a video file in seconds
	GetVideoDuration(filePath string) (float64, error)
	
	// GetVideoInfo gets comprehensive video information
	GetVideoInfo(filePath string) (*VideoInfo, error)
	
	// IsAvailable checks if ffprobe is available
	IsAvailable() bool
}

// Ensure FFProbeClient implements FFProbeClientInterface
var _ FFProbeClientInterface = (*FFProbeClient)(nil)