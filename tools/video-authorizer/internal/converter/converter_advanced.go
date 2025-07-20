package converter

import (
	"fmt"

	"github.com/yakuari-channel/video-authorizer/internal/models"
)

// ConvertWithRelativeLengths performs conversion with full relative length support
func (c *Converter) ConvertWithRelativeLengths(ymmps *models.YMMPSDocument, template *models.YMMPProject) (*models.YMMPProject, error) {
	// Deep copy the template
	result, err := c.deepCopyProject(template)
	if err != nil {
		return nil, fmt.Errorf("failed to copy template: %w", err)
	}

	// Clear the items in the timeline
	if len(result.Timelines) == 0 {
		return nil, fmt.Errorf("template must have at least one timeline")
	}

	timeline := &result.Timelines[0]
	timeline.Items = []interface{}{}
	timeline.CurrentFrame = 0
	timeline.Length = 0

	// First pass: calculate all end frames
	ctx, err := c.lengthCalculator.CalculateWithPrepass(ymmps)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate end frames: %w", err)
	}

	// Second pass: convert with full context
	currentFrame := 0
	for _, sequence := range ymmps.Sequences {
		ctx.CurrentSequence = &sequence
		endFrame, err := c.processSequenceWithContext(template, timeline, sequence, currentFrame, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to process sequence %s: %w", sequence.ID, err)
		}
		currentFrame = endFrame
	}

	// Update timeline length
	timeline.Length = currentFrame

	return result, nil
}

// processSequenceWithContext processes a sequence with full compile context
func (c *Converter) processSequenceWithContext(template *models.YMMPProject, timeline *models.Timeline, sequence models.Sequence, startFrame int, ctx *CompileContext) (int, error) {
	currentFrame := startFrame
	ctx.CurrentFrame = currentFrame

	for _, scene := range sequence.Scenes {
		ctx.CurrentScene = &scene
		endFrame, err := c.processSceneWithContext(template, timeline, scene, currentFrame, ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to process scene %s: %w", scene.ID, err)
		}
		currentFrame = endFrame
	}

	return currentFrame, nil
}

// processSceneWithContext processes a scene with full compile context
func (c *Converter) processSceneWithContext(template *models.YMMPProject, timeline *models.Timeline, scene models.Scene, startFrame int, ctx *CompileContext) (int, error) {
	currentFrame := startFrame
	ctx.CurrentFrame = currentFrame

	for _, shot := range scene.Shots {
		ctx.CurrentShot = &shot
		endFrame, err := c.processShotWithContext(template, timeline, shot, currentFrame, ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to process shot %s: %w", shot.ID, err)
		}
		currentFrame = endFrame
	}

	return currentFrame, nil
}

// processShotWithContext processes a shot with full compile context
func (c *Converter) processShotWithContext(template *models.YMMPProject, timeline *models.Timeline, shot models.Shot, startFrame int, ctx *CompileContext) (int, error) {
	maxEndFrame := startFrame

	for _, itemSpec := range shot.Items {
		ctx.CurrentFrame = startFrame
		
		// Calculate actual length using context
		length, err := c.lengthCalculator.basicCalculator.CalculateLength(itemSpec.Length, ctx)
		if err != nil {
			// If relative length calculation fails, try to get from prepass
			if shot.ID != "" {
				if endFrame, found := ctx.ShotEndFrames[shot.ID]; found {
					length = endFrame - startFrame
				} else {
					return 0, fmt.Errorf("failed to calculate length for item: %w", err)
				}
			} else {
				return 0, fmt.Errorf("failed to calculate length for item: %w", err)
			}
		}

		// Process item with calculated length
		item, err := c.processItemWithLength(template, itemSpec, startFrame, length)
		if err != nil {
			return 0, fmt.Errorf("failed to process item with template %s: %w", itemSpec.Template, err)
		}

		timeline.Items = append(timeline.Items, item)

		// Update max end frame
		endFrame := startFrame + length
		if endFrame > maxEndFrame {
			maxEndFrame = endFrame
		}
	}

	return maxEndFrame, nil
}

// processItemWithLength processes an item with a pre-calculated length
func (c *Converter) processItemWithLength(template *models.YMMPProject, itemSpec models.ItemSpec, frame int, length int) (interface{}, error) {
	// Find the template item
	templateItem, found := c.findTemplateItem(template, itemSpec.Template)
	if !found {
		return nil, fmt.Errorf("template item not found: %s", itemSpec.Template)
	}

	// Deep copy the template item
	item, err := c.deepCopyItem(templateItem)
	if err != nil {
		return nil, fmt.Errorf("failed to copy template item: %w", err)
	}

	// Set frame and length
	c.setItemFrame(item, frame)
	c.setItemLength(item, length)

	// Apply property overrides
	if err := c.applyPropertyOverrides(item, itemSpec.Properties); err != nil {
		return nil, fmt.Errorf("failed to apply property overrides: %w", err)
	}

	return item, nil
}

// UpdateLengthCalculation updates the length calculation implementation
func (lc *LengthCalculator) calculateUntilSequenceEndWithContext(ctx *CompileContext) (int, error) {
	if ctx.CurrentSequence == nil {
		return 0, fmt.Errorf("no current sequence in context")
	}

	endFrame, found := ctx.SequenceEndFrames[ctx.CurrentSequence.ID]
	if !found && ctx.CurrentSequence.ID != "" {
		return 0, fmt.Errorf("sequence end frame not found for ID: %s", ctx.CurrentSequence.ID)
	}

	if !found {
		// Try to find by reference
		for i, seq := range ctx.CurrentSequence.Scenes {
			if i == len(ctx.CurrentSequence.Scenes)-1 {
				// This is the last scene in sequence
				if seq.ID != "" {
					if sceneEnd, ok := ctx.SceneEndFrames[seq.ID]; ok {
						endFrame = sceneEnd
						break
					}
				}
			}
		}
		
		if endFrame == 0 {
			return 0, fmt.Errorf("unable to determine sequence end")
		}
	}

	length := endFrame - ctx.CurrentFrame
	if length < 0 {
		return 0, fmt.Errorf("calculated length is negative: %d", length)
	}

	return length, nil
}

// calculateUntilSceneEndWithContext calculates length until current scene ends
func (lc *LengthCalculator) calculateUntilSceneEndWithContext(ctx *CompileContext) (int, error) {
	if ctx.CurrentScene == nil {
		return 0, fmt.Errorf("no current scene in context")
	}

	endFrame, found := ctx.SceneEndFrames[ctx.CurrentScene.ID]
	if !found && ctx.CurrentScene.ID != "" {
		return 0, fmt.Errorf("scene end frame not found for ID: %s", ctx.CurrentScene.ID)
	}

	if !found {
		// Try to find by reference
		for i, shot := range ctx.CurrentScene.Shots {
			if i == len(ctx.CurrentScene.Shots)-1 {
				// This is the last shot in scene
				if shot.ID != "" {
					if shotEnd, ok := ctx.ShotEndFrames[shot.ID]; ok {
						endFrame = shotEnd
						break
					}
				}
			}
		}
		
		if endFrame == 0 {
			return 0, fmt.Errorf("unable to determine scene end")
		}
	}

	length := endFrame - ctx.CurrentFrame
	if length < 0 {
		return 0, fmt.Errorf("calculated length is negative: %d", length)
	}

	return length, nil
}

// calculateUntilShotEndWithContext calculates length until current shot ends
func (lc *LengthCalculator) calculateUntilShotEndWithContext(ctx *CompileContext) (int, error) {
	if ctx.CurrentShot == nil {
		return 0, fmt.Errorf("no current shot in context")
	}

	endFrame, found := ctx.ShotEndFrames[ctx.CurrentShot.ID]
	if !found && ctx.CurrentShot.ID != "" {
		return 0, fmt.Errorf("shot end frame not found for ID: %s", ctx.CurrentShot.ID)
	}

	if !found {
		return 0, fmt.Errorf("unable to determine shot end")
	}

	length := endFrame - ctx.CurrentFrame
	if length < 0 {
		return 0, fmt.Errorf("calculated length is negative: %d", length)
	}

	return length, nil
}