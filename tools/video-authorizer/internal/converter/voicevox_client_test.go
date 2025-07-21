package converter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yakuari-channel/video-authorizer/internal/converter"
)

func TestVoicevoxClient_CalculateDurationLogic(t *testing.T) {
	// Test duration calculation logic by manually implementing it
	mockQuery := &converter.AudioQueryResponse{
		PrePhonemeLength:  0.1,
		PostPhonemeLength: 0.1,
		SpeedScale:        1.0,
		AccentPhrases: []converter.AccentPhrase{
			{
				Moras: []converter.Mora{
					{
						Text:        "こ",
						Vowel:       "o",
						VowelLength: 0.2,
					},
					{
						Text:        "ん",
						Vowel:       "N",
						VowelLength: 0.15,
					},
					{
						Text:        "に",
						Vowel:       "i",
						VowelLength: 0.18,
					},
					{
						Text:        "ち",
						Vowel:       "i",
						VowelLength: 0.17,
					},
					{
						Text:        "は",
						Vowel:       "a",
						VowelLength: 0.20,
					},
				},
			},
		},
	}
	
	// Calculate duration manually to verify logic
	totalDuration := 0.0
	totalDuration += mockQuery.PrePhonemeLength
	
	for _, phrase := range mockQuery.AccentPhrases {
		for _, mora := range phrase.Moras {
			if mora.ConsonantLength != nil {
				totalDuration += *mora.ConsonantLength
			}
			totalDuration += mora.VowelLength
		}
	}
	
	totalDuration += mockQuery.PostPhonemeLength
	
	if mockQuery.SpeedScale > 0 {
		totalDuration /= mockQuery.SpeedScale
	}
	
	// Expected: 0.1 (pre) + 0.2 + 0.15 + 0.18 + 0.17 + 0.20 + 0.1 (post) = 1.1
	expected := 1.1
	tolerance := 0.0001
	if totalDuration < expected-tolerance || totalDuration > expected+tolerance {
		t.Errorf("expected duration %f, got %f", expected, totalDuration)
	}
}

func TestVoicevoxClient_WithMockServer(t *testing.T) {
	// Create a mock VOICEVOX server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/version":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"version": "0.14.0"}`))
		case "/audio_query":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"accent_phrases": [
					{
						"moras": [
							{
								"text": "こ",
								"vowel": "o",
								"vowel_length": 0.2,
								"pitch": 5.5
							},
							{
								"text": "ん",
								"vowel": "N", 
								"vowel_length": 0.15,
								"pitch": 5.2
							}
						],
						"accent": 1
					}
				],
				"speedScale": 1.0,
				"pitchScale": 0.0,
				"volumeScale": 1.0,
				"prePhonemeLength": 0.1,
				"postPhonemeLength": 0.1
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	
	client := converter.NewVoicevoxClient(server.URL)
	
	// Test availability
	if !client.IsAvailable() {
		t.Error("expected client to be available")
	}
	
	// Test audio duration calculation
	duration, err := client.GetAudioDuration("こん", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// Expected: 0.1 (pre) + 0.2 + 0.15 + 0.1 (post) = 0.55
	expected := 0.55
	if duration != expected {
		t.Errorf("expected duration %f, got %f", expected, duration)
	}
}

func TestConvertDurationToFrames(t *testing.T) {
	tests := []struct {
		name     string
		duration float64
		fps      float64
		expected int
	}{
		{
			name:     "1 second at 30 FPS",
			duration: 1.0,
			fps:      30.0,
			expected: 30,
		},
		{
			name:     "0.5 seconds at 60 FPS",
			duration: 0.5,
			fps:      60.0,
			expected: 30,
		},
		{
			name:     "2.5 seconds at 24 FPS",
			duration: 2.5,
			fps:      24.0,
			expected: 60,
		},
		{
			name:     "default FPS when fps <= 0",
			duration: 1.0,
			fps:      0,
			expected: 30, // Default FPS is 30
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

