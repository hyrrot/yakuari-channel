package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// VoicevoxClient handles communication with VOICEVOX API
type VoicevoxClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewVoicevoxClient creates a new VOICEVOX client
func NewVoicevoxClient(baseURL string) *VoicevoxClient {
	if baseURL == "" {
		baseURL = "http://localhost:50021" // Default VOICEVOX API URL
	}
	
	return &VoicevoxClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AudioQueryRequest represents the request for audio query
type AudioQueryRequest struct {
	Text     string `json:"text"`
	Speaker  int    `json:"speaker"`
}

// AudioQueryResponse represents the response from audio query
type AudioQueryResponse struct {
	AccentPhrases []AccentPhrase `json:"accent_phrases"`
	SpeedScale    float64        `json:"speedScale"`
	PitchScale    float64        `json:"pitchScale"`
	VolumeScale   float64        `json:"volumeScale"`
	PrePhonemeLength float64     `json:"prePhonemeLength"`
	PostPhonemeLength float64    `json:"postPhonemeLength"`
}

// AccentPhrase represents an accent phrase
type AccentPhrase struct {
	Moras  []Mora `json:"moras"`
	Accent int    `json:"accent"`
}

// Mora represents a mora (sound unit)
type Mora struct {
	Text            string  `json:"text"`
	Consonant       *string `json:"consonant"`
	ConsonantLength *float64 `json:"consonant_length"`
	Vowel           string  `json:"vowel"`
	VowelLength     float64 `json:"vowel_length"`
	Pitch           float64 `json:"pitch"`
}

// GetAudioDuration calculates the duration of text when spoken by VOICEVOX
func (vc *VoicevoxClient) GetAudioDuration(text string, speaker int) (float64, error) {
	// Step 1: Get audio query
	audioQuery, err := vc.getAudioQuery(text, speaker)
	if err != nil {
		return 0, fmt.Errorf("failed to get audio query: %w", err)
	}
	
	// Step 2: Calculate duration from audio query
	duration := vc.calculateDurationFromQuery(audioQuery)
	
	return duration, nil
}

// getAudioQuery gets audio query from VOICEVOX API
func (vc *VoicevoxClient) getAudioQuery(text string, speaker int) (*AudioQueryResponse, error) {
	// Prepare URL
	apiURL := fmt.Sprintf("%s/audio_query", vc.baseURL)
	params := url.Values{}
	params.Add("text", text)
	params.Add("speaker", fmt.Sprintf("%d", speaker))
	
	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())
	
	// Make POST request
	resp, err := vc.httpClient.Post(fullURL, "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var audioQuery AudioQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&audioQuery); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return &audioQuery, nil
}

// calculateDurationFromQuery calculates total duration from audio query
func (vc *VoicevoxClient) calculateDurationFromQuery(query *AudioQueryResponse) float64 {
	totalDuration := 0.0
	
	// Add pre-phoneme length
	totalDuration += query.PrePhonemeLength
	
	// Calculate duration from accent phrases
	for _, phrase := range query.AccentPhrases {
		for _, mora := range phrase.Moras {
			// Add consonant length if present
			if mora.ConsonantLength != nil {
				totalDuration += *mora.ConsonantLength
			}
			// Add vowel length
			totalDuration += mora.VowelLength
		}
	}
	
	// Add post-phoneme length
	totalDuration += query.PostPhonemeLength
	
	// Apply speed scale
	if query.SpeedScale > 0 {
		totalDuration /= query.SpeedScale
	}
	
	return totalDuration
}

// IsAvailable checks if VOICEVOX API is available
func (vc *VoicevoxClient) IsAvailable() bool {
	resp, err := vc.httpClient.Get(fmt.Sprintf("%s/version", vc.baseURL))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

// ConvertDurationToFrames converts duration in seconds to frame count
func ConvertDurationToFrames(durationSeconds float64, fps float64) int {
	if fps <= 0 {
		fps = 30.0 // Default FPS
	}
	return int(durationSeconds * fps)
}