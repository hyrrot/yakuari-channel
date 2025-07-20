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
type LengthCalculator struct{}

// NewLengthCalculator creates a new length calculator
func NewLengthCalculator() *LengthCalculator {
	return &LengthCalculator{}
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

	case models.LengthTypeUntilIDEnd:
		return lc.calculateUntilIDEnd(parsed.TargetType, parsed.TargetID, ctx)

	case models.LengthTypeAutoVoice:
		// TODO: Implement in Phase 7
		return 0, fmt.Errorf("auto voice length calculation not yet implemented")

	case models.LengthTypeAutoVideo:
		// TODO: Implement in Phase 7
		return 0, fmt.Errorf("auto video length calculation not yet implemented")

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
			// For non-numeric lengths, assume a default length for pre-pass
			// This will be refined in the actual implementation
			defaultLength := 100 // frames
			endFrame := startFrame + defaultLength
			if endFrame > maxEndFrame {
				maxEndFrame = endFrame
			}
		}
	}
	
	return maxEndFrame, nil
}