package converter

import (
	"fmt"
	"strconv"

	"github.com/yakuari-channel/video-authorizer/internal/models"
)

// CompileContext holds context information during compilation
type CompileContext struct {
	CurrentSequence *models.Sequence
	CurrentScene    *models.Scene
	CurrentShot     *models.Shot
	
	// Maps to track end frames for different elements
	SequenceEndFrames map[string]int // ID -> end frame
	SceneEndFrames    map[string]int // ID -> end frame
	ShotEndFrames     map[string]int // ID -> end frame
	
	// Current frame tracking
	CurrentFrame int
}

// NewCompileContext creates a new compile context
func NewCompileContext() *CompileContext {
	return &CompileContext{
		SequenceEndFrames: make(map[string]int),
		SceneEndFrames:    make(map[string]int),
		ShotEndFrames:     make(map[string]int),
	}
}

// RegisterSequenceEnd registers the end frame for a sequence
func (ctx *CompileContext) RegisterSequenceEnd(id string, endFrame int) {
	if id != "" {
		ctx.SequenceEndFrames[id] = endFrame
	}
}

// RegisterSceneEnd registers the end frame for a scene
func (ctx *CompileContext) RegisterSceneEnd(id string, endFrame int) {
	if id != "" {
		ctx.SceneEndFrames[id] = endFrame
	}
}

// RegisterShotEnd registers the end frame for a shot
func (ctx *CompileContext) RegisterShotEnd(id string, endFrame int) {
	if id != "" {
		ctx.ShotEndFrames[id] = endFrame
	}
}

// LengthCalculator handles calculation of item lengths
type LengthCalculator struct {
	voicevoxClient VoicevoxClientInterface
	ffprobeClient  FFProbeClientInterface
}

// NewLengthCalculator creates a new length calculator
func NewLengthCalculator() *LengthCalculator {
	return &LengthCalculator{
		voicevoxClient: NewVoicevoxClient(""),
		ffprobeClient:  NewFFProbeClient(""),
	}
}

// NewLengthCalculatorWithVoicevox creates a new length calculator with a custom VOICEVOX client
func NewLengthCalculatorWithVoicevox(voicevoxClient VoicevoxClientInterface) *LengthCalculator {
	return &LengthCalculator{
		voicevoxClient: voicevoxClient,
		ffprobeClient:  NewFFProbeClient(""),
	}
}

// NewLengthCalculatorWithFFProbe creates a new length calculator with a custom FFProbe client
func NewLengthCalculatorWithFFProbe(ffprobeClient FFProbeClientInterface) *LengthCalculator {
	return &LengthCalculator{
		voicevoxClient: NewVoicevoxClient(""),
		ffprobeClient:  ffprobeClient,
	}
}

// NewLengthCalculatorWithClients creates a new length calculator with custom VOICEVOX and FFProbe clients
func NewLengthCalculatorWithClients(voicevoxClient VoicevoxClientInterface, ffprobeClient FFProbeClientInterface) *LengthCalculator {
	return &LengthCalculator{
		voicevoxClient: voicevoxClient,
		ffprobeClient:  ffprobeClient,
	}
}

// CalculateLength calculates the length for an item based on the length specification
func (lc *LengthCalculator) CalculateLength(lengthStr string, ctx *CompileContext) (int, error) {
	parsed, err := models.ParseLength(lengthStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse length: %w", err)
	}

	switch parsed.Type {
	case models.LengthTypeNumeric:
		return parsed.Value, nil

	case models.LengthTypeUntilSeqEnd:
		return lc.calculateUntilSequenceEnd(ctx)

	case models.LengthTypeUntilSceneEnd:
		return lc.calculateUntilSceneEnd(ctx)

	case models.LengthTypeUntilShotEnd:
		return lc.calculateUntilShotEnd(ctx)

	case models.LengthTypeUntilSeqEndID:
		return lc.calculateUntilIDEnd("SEQUENCE", parsed.TargetID, ctx)
		
	case models.LengthTypeUntilSceneEndID:
		return lc.calculateUntilIDEnd("SCENE", parsed.TargetID, ctx)
		
	case models.LengthTypeUntilShotEndID:
		return lc.calculateUntilIDEnd("SHOT", parsed.TargetID, ctx)
		
	case models.LengthTypeUntilIDEnd:
		return lc.calculateUntilIDEnd(parsed.TargetType, parsed.TargetID, ctx)

	case models.LengthTypeAutoVoice:
		return lc.calculateAutoVoiceLength(ctx)

	case models.LengthTypeAutoVideo:
		return lc.calculateAutoVideoLength(ctx)

	default:
		return 0, fmt.Errorf("unsupported length type: %v", parsed.Type)
	}
}

// CalculateLengthForItem calculates the length for a specific item with context
func (lc *LengthCalculator) CalculateLengthForItem(lengthStr string, item interface{}, ctx *CompileContext) (int, error) {
	parsed, err := models.ParseLength(lengthStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse length: %w", err)
	}

	switch parsed.Type {
	case models.LengthTypeNumeric:
		return parsed.Value, nil

	case models.LengthTypeUntilSeqEnd:
		return lc.calculateUntilSequenceEnd(ctx)

	case models.LengthTypeUntilSceneEnd:
		return lc.calculateUntilSceneEnd(ctx)

	case models.LengthTypeUntilShotEnd:
		return lc.calculateUntilShotEnd(ctx)

	case models.LengthTypeUntilSeqEndID:
		return lc.calculateUntilIDEnd("SEQUENCE", parsed.TargetID, ctx)
		
	case models.LengthTypeUntilSceneEndID:
		return lc.calculateUntilIDEnd("SCENE", parsed.TargetID, ctx)
		
	case models.LengthTypeUntilShotEndID:
		return lc.calculateUntilIDEnd("SHOT", parsed.TargetID, ctx)
		
	case models.LengthTypeUntilIDEnd:
		return lc.calculateUntilIDEnd(parsed.TargetType, parsed.TargetID, ctx)

	case models.LengthTypeAutoVoice:
		return lc.calculateAutoVoiceLengthForItem(item)

	case models.LengthTypeAutoVideo:
		return lc.calculateAutoVideoLengthForItem(item)

	default:
		return 0, fmt.Errorf("unsupported length type: %v", parsed.Type)
	}
}

// calculateUntilSequenceEnd calculates length until current sequence ends
func (lc *LengthCalculator) calculateUntilSequenceEnd(ctx *CompileContext) (int, error) {
	return lc.calculateUntilSequenceEndWithContext(ctx)
}

// calculateUntilSceneEnd calculates length until current scene ends
func (lc *LengthCalculator) calculateUntilSceneEnd(ctx *CompileContext) (int, error) {
	return lc.calculateUntilSceneEndWithContext(ctx)
}

// calculateUntilShotEnd calculates length until current shot ends
func (lc *LengthCalculator) calculateUntilShotEnd(ctx *CompileContext) (int, error) {
	return lc.calculateUntilShotEndWithContext(ctx)
}

// calculateUntilIDEnd calculates length until specified ID element ends
func (lc *LengthCalculator) calculateUntilIDEnd(targetType, targetID string, ctx *CompileContext) (int, error) {
	var endFrame int
	var found bool

	switch targetType {
	case "SEQUENCE":
		endFrame, found = ctx.SequenceEndFrames[targetID]
	case "SCENE":
		endFrame, found = ctx.SceneEndFrames[targetID]
	case "SHOT":
		endFrame, found = ctx.ShotEndFrames[targetID]
	default:
		return 0, fmt.Errorf("unsupported target type: %s", targetType)
	}

	if !found {
		return 0, fmt.Errorf("target %s with ID %s not found", targetType, targetID)
	}

	length := endFrame - ctx.CurrentFrame
	if length < 0 {
		return 0, fmt.Errorf("calculated length is negative: %d", length)
	}

	return length, nil
}

// AdvancedLengthCalculator handles multi-pass length calculation
// This is needed for relative length calculations that depend on future elements
type AdvancedLengthCalculator struct {
	basicCalculator *LengthCalculator
}

// NewAdvancedLengthCalculator creates a new advanced length calculator
func NewAdvancedLengthCalculator() *AdvancedLengthCalculator {
	return &AdvancedLengthCalculator{
		basicCalculator: NewLengthCalculator(),
	}
}

// NewAdvancedLengthCalculatorWithVoicevox creates a new advanced length calculator with a custom VOICEVOX client
func NewAdvancedLengthCalculatorWithVoicevox(voicevoxClient VoicevoxClientInterface) *AdvancedLengthCalculator {
	return &AdvancedLengthCalculator{
		basicCalculator: NewLengthCalculatorWithVoicevox(voicevoxClient),
	}
}

// NewAdvancedLengthCalculatorWithFFProbe creates a new advanced length calculator with a custom FFProbe client
func NewAdvancedLengthCalculatorWithFFProbe(ffprobeClient FFProbeClientInterface) *AdvancedLengthCalculator {
	return &AdvancedLengthCalculator{
		basicCalculator: NewLengthCalculatorWithFFProbe(ffprobeClient),
	}
}

// NewAdvancedLengthCalculatorWithClients creates a new advanced length calculator with custom VOICEVOX and FFProbe clients
func NewAdvancedLengthCalculatorWithClients(voicevoxClient VoicevoxClientInterface, ffprobeClient FFProbeClientInterface) *AdvancedLengthCalculator {
	return &AdvancedLengthCalculator{
		basicCalculator: NewLengthCalculatorWithClients(voicevoxClient, ffprobeClient),
	}
}

// CalculateWithPrepass performs a two-pass calculation:
// 1. First pass: Calculate all end frames for sequences, scenes, and shots
// 2. Second pass: Calculate actual lengths with full context
func (alc *AdvancedLengthCalculator) CalculateWithPrepass(ymmps *models.YMMPSDocument) (*CompileContext, error) {
	ctx := NewCompileContext()

	// First pass: calculate all end frames
	currentFrame := 0
	for _, sequence := range ymmps.Sequences {
		seqEndFrame, err := alc.calculateSequenceEndFrame(sequence, currentFrame, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate sequence end frame: %w", err)
		}
		
		ctx.RegisterSequenceEnd(sequence.ID, seqEndFrame)
		currentFrame = seqEndFrame
	}

	return ctx, nil
}

// calculateSequenceEndFrame calculates when a sequence ends
func (alc *AdvancedLengthCalculator) calculateSequenceEndFrame(sequence models.Sequence, startFrame int, ctx *CompileContext) (int, error) {
	currentFrame := startFrame
	
	// Scenes are executed sequentially
	for _, scene := range sequence.Scenes {
		sceneEndFrame, err := alc.calculateSceneEndFrame(scene, currentFrame, ctx)
		if err != nil {
			return 0, err
		}
		
		ctx.RegisterSceneEnd(scene.ID, sceneEndFrame)
		currentFrame = sceneEndFrame
	}
	
	return currentFrame, nil
}

// calculateSceneEndFrame calculates when a scene ends
func (alc *AdvancedLengthCalculator) calculateSceneEndFrame(scene models.Scene, startFrame int, ctx *CompileContext) (int, error) {
	currentFrame := startFrame
	
	for _, shot := range scene.Shots {
		shotEndFrame, err := alc.calculateShotEndFrame(shot, currentFrame, ctx)
		if err != nil {
			return 0, err
		}
		
		ctx.RegisterShotEnd(shot.ID, shotEndFrame)
		currentFrame = shotEndFrame
	}
	
	return currentFrame, nil
}

// calculateShotEndFrame calculates when a shot ends
func (alc *AdvancedLengthCalculator) calculateShotEndFrame(shot models.Shot, startFrame int, ctx *CompileContext) (int, error) {
	maxEndFrame := startFrame
	
	for _, item := range shot.Items {
		// For now, only handle numeric lengths in pre-pass
		// Relative lengths will be resolved in second pass
		if length, err := strconv.Atoi(item.Length); err == nil {
			endFrame := startFrame + length
			if endFrame > maxEndFrame {
				maxEndFrame = endFrame
			}
		} else {
			// For non-numeric lengths, use smarter defaults based on length type
			parsed, parseErr := models.ParseLength(item.Length)
			var defaultLength int
			
			if parseErr == nil {
				switch parsed.Type {
				case models.LengthTypeUntilIDEnd:
					// For _until: types, don't assume a default length in prepass
					// These will be resolved in the second pass
					continue
				case models.LengthTypeAutoVoice:
					// Voice items typically 3-10 seconds, assume 5 seconds at 30 FPS
					defaultLength = 150
				case models.LengthTypeAutoVideo:
					// Video items vary widely, use moderate default
					defaultLength = 300
				default:
					defaultLength = 100
				}
			} else {
				// Unknown format, use conservative default
				defaultLength = 100
			}
			
			endFrame := startFrame + defaultLength
			if endFrame > maxEndFrame {
				maxEndFrame = endFrame
			}
		}
	}
	
	return maxEndFrame, nil
}

// calculateAutoVoiceLength calculates voice length using VOICEVOX (fallback method)
func (lc *LengthCalculator) calculateAutoVoiceLength(ctx *CompileContext) (int, error) {
	return 0, fmt.Errorf("auto voice length calculation requires item context; use CalculateLengthForItem instead")
}

// calculateAutoVoiceLengthForItem calculates voice length for a specific voice item
func (lc *LengthCalculator) calculateAutoVoiceLengthForItem(item interface{}) (int, error) {
	// Extract text and speaker from voice item first
	text, speaker, err := lc.extractVoiceItemData(item)
	if err != nil {
		return 0, fmt.Errorf("failed to extract voice data: %w", err)
	}

	// Try VOICEVOX if available
	if lc.voicevoxClient.IsAvailableWithLogging(true) {
		duration, err := lc.voicevoxClient.GetAudioDuration(text, speaker)
		if err == nil {
			// Convert to frames (assuming 30 FPS)
			frames := ConvertDurationToFrames(duration, 30.0)
			return frames, nil
		}
		// Log the VOICEVOX error but continue with fallback
		fmt.Printf("VOICEVOX calculation failed: %v, using fallback estimation\n", err)
	}

	// Fallback: estimate length based on text length
	// Typical speech rate: ~5-6 characters per second in Japanese
	// Use 5 chars/sec = 150 frames (5 seconds at 30 FPS) for default estimation
	textLength := len([]rune(text)) // Count Unicode characters properly
	if textLength == 0 {
		return 150, nil // Default 5 seconds if no text
	}
	
	// Estimate: 5 characters per second = 150 frames per 5 characters
	estimatedSeconds := float64(textLength) / 5.0
	if estimatedSeconds < 1.0 {
		estimatedSeconds = 1.0 // Minimum 1 second
	}
	if estimatedSeconds > 30.0 {
		estimatedSeconds = 30.0 // Maximum 30 seconds
	}
	
	estimatedFrames := int(estimatedSeconds * 30.0) // 30 FPS
	return estimatedFrames, nil
}

// calculateAutoVideoLength calculates video length using ffprobe (fallback method)
func (lc *LengthCalculator) calculateAutoVideoLength(ctx *CompileContext) (int, error) {
	return 0, fmt.Errorf("auto video length calculation requires item context; use CalculateLengthForItem instead")
}

// calculateAutoVideoLengthForItem calculates video length for a specific video item
func (lc *LengthCalculator) calculateAutoVideoLengthForItem(item interface{}) (int, error) {
	// Check if ffprobe is available
	if !lc.ffprobeClient.IsAvailable() {
		return 0, fmt.Errorf("ffprobe is not available")
	}

	// Extract file path from video item
	filePath, err := lc.extractVideoItemFilePath(item)
	if err != nil {
		return 0, fmt.Errorf("failed to extract video file path: %w", err)
	}

	// Get duration from ffprobe
	duration, err := lc.ffprobeClient.GetVideoDuration(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get video duration from ffprobe: %w", err)
	}

	// Convert to frames (assuming 30 FPS by default, could be improved to use actual video FPS)
	frames := ConvertDurationToFrames(duration, 30.0)
	
	return frames, nil
}

// extractVoiceItemData extracts text and speaker from voice item
func (lc *LengthCalculator) extractVoiceItemData(item interface{}) (string, int, error) {
	switch v := item.(type) {
	case *models.VoiceItem:
		// Extract text from Serif field
		text := v.Serif
		if text == "" {
			return "", 0, fmt.Errorf("voice item has no Serif text")
		}
		
		// Extract speaker from CharacterName or default to 0
		// In a real implementation, you would look up the character configuration
		// and get the speaker ID from the Voice settings
		speaker := 0
		
		// TODO: Implement character-to-speaker mapping
		// For now, use a simple default
		
		return text, speaker, nil
		
	default:
		return "", 0, fmt.Errorf("item is not a voice item")
	}
}

// extractVideoItemFilePath extracts file path from video item
func (lc *LengthCalculator) extractVideoItemFilePath(item interface{}) (string, error) {
	switch v := item.(type) {
	case *models.VideoItem:
		// Extract file path from FilePath field
		if v.FilePath == "" {
			return "", fmt.Errorf("video item has no FilePath")
		}
		return v.FilePath, nil
		
	default:
		return "", fmt.Errorf("item is not a video item")
	}
}