package converter

// VoicevoxClientInterface defines the interface for VOICEVOX client operations
type VoicevoxClientInterface interface {
	GetAudioDuration(text string, speaker int) (float64, error)
	IsAvailable() bool
	IsAvailableWithLogging(verbose bool) bool
}

// Ensure VoicevoxClient implements the interface
var _ VoicevoxClientInterface = (*VoicevoxClient)(nil)