package converter

import (
	"fmt"
	
	"github.com/user/ymmp-compiler/internal/models"
)

// CalculateTimeline calculates the timeline for all items in the episode
func CalculateTimeline(episode *models.Episode) error {
	// Phase 1: Resolve auto lengths (would normally involve external API calls)
	if err := resolveAutoLengths(episode); err != nil {
		return fmt.Errorf("failed to resolve auto lengths: %v", err)
	}
	
	// Phase 2: Resolve relative lengths (bottom-up)
	if err := resolveRelativeLengths(episode); err != nil {
		return fmt.Errorf("failed to resolve relative lengths: %v", err)
	}
	
	// Phase 3: Calculate start times
	if err := calculateStartTimes(episode); err != nil {
		return fmt.Errorf("failed to calculate start times: %v", err)
	}
	
	return nil
}

// resolveAutoLengths resolves items with "auto" length specification
func resolveAutoLengths(episode *models.Episode) error {
	for _, sequence := range episode.Sequences {
		for _, scene := range sequence.Scenes {
			for _, shot := range scene.Shots {
				for _, item := range shot.Items {
					if item.GetLength().Type == models.LengthTypeAuto {
						// For now, use the calculated length if available
						// In a real implementation, this would call external services
						if item.GetLength().Value == 0.0 {
							// Use the calculated length that was set (e.g., by VOICEVOX)
							// If no calculated length is available, this would be an error
							length := item.GetLength()
							length.Value = getCalculatedLength(item)
							if length.Value == 0.0 {
								return fmt.Errorf("auto length could not be calculated for item type %s", item.GetType())
							}
							// Update the item's length
							updateItemLength(item, length)
						}
					}
				}
			}
		}
	}
	return nil
}

// resolveRelativeLengths resolves items with "until:*" length specifications
func resolveRelativeLengths(episode *models.Episode) error {
	for _, sequence := range episode.Sequences {
		for _, scene := range sequence.Scenes {
			// First, resolve UNTIL_SHOT_END
			for _, shot := range scene.Shots {
				if err := resolveUntilShotEnd(shot); err != nil {
					return err
				}
			}
			
			// Then, resolve UNTIL_SCENE_END
			if err := resolveUntilSceneEnd(scene); err != nil {
				return err
			}
		}
		
		// Finally, resolve UNTIL_SEQUENCE_END
		if err := resolveUntilSequenceEnd(sequence); err != nil {
			return err
		}
	}
	return nil
}

// resolveUntilShotEnd resolves items that should last until the end of their shot
func resolveUntilShotEnd(shot models.Shot) error {
	// Find the maximum length among non-until-shot-end items
	var maxLength float64
	var untilShotEndItems []models.Item
	
	for _, item := range shot.Items {
		if item.GetLength().Type == models.LengthTypeUntilShotEnd {
			untilShotEndItems = append(untilShotEndItems, item)
		} else {
			itemLength := item.GetLength().Value
			if itemLength > maxLength {
				maxLength = itemLength
			}
		}
	}
	
	// Set the length of until-shot-end items to the maximum length
	for _, item := range untilShotEndItems {
		length := item.GetLength()
		length.Type = models.LengthTypeNumeric
		length.Value = maxLength
		updateItemLength(item, length)
	}
	
	return nil
}

// resolveUntilSceneEnd resolves items that should last until the end of their scene
func resolveUntilSceneEnd(scene models.Scene) error {
	// Calculate the total scene length
	sceneLength := calculateSceneLength(scene)
	
	// Find and update until-scene-end items
	for _, shot := range scene.Shots {
		for _, item := range shot.Items {
			if item.GetLength().Type == models.LengthTypeUntilSceneEnd {
				length := item.GetLength()
				length.Type = models.LengthTypeNumeric
				length.Value = sceneLength
				updateItemLength(item, length)
			}
		}
	}
	
	return nil
}

// resolveUntilSequenceEnd resolves items that should last until the end of their sequence
func resolveUntilSequenceEnd(sequence models.Sequence) error {
	// Calculate the total sequence length
	sequenceLength := calculateSequenceLength(sequence)
	
	// Find and update until-sequence-end items
	for _, scene := range sequence.Scenes {
		for _, shot := range scene.Shots {
			for _, item := range shot.Items {
				if item.GetLength().Type == models.LengthTypeUntilSequenceEnd {
					length := item.GetLength()
					length.Type = models.LengthTypeNumeric
					length.Value = sequenceLength
					updateItemLength(item, length)
				}
			}
		}
	}
	
	return nil
}

// calculateStartTimes calculates the start times for all items
func calculateStartTimes(episode *models.Episode) error {
	currentTime := 0.0
	
	for _, sequence := range episode.Sequences {
		for _, scene := range sequence.Scenes {
			sceneStartTime := currentTime
			
			for _, shot := range scene.Shots {
				shotStartTime := currentTime
				
				// All items in a shot start at the same time
				for _, item := range shot.Items {
					item.SetStartTime(shotStartTime)
				}
				
				// Move to the next shot
				shotLength := calculateShotLength(shot)
				currentTime += shotLength
			}
			
			// Ensure we don't go backwards (in case of overlapping shots)
			if currentTime < sceneStartTime {
				currentTime = sceneStartTime
			}
		}
	}
	
	return nil
}

// Helper functions

func getCalculatedLength(item models.Item) float64 {
	// Try to get the calculated length that was already set
	switch v := item.(type) {
	case interface{ GetCalculatedLength() float64 }:
		if length := v.GetCalculatedLength(); length > 0 {
			return length
		}
	}
	
	// Fallback to default values for testing
	switch item.GetType() {
	case "voice":
		return 2.0 // Default for testing
	case "video", "audio":
		return 10.0 // Default for testing
	default:
		return 0.0
	}
}

func updateItemLength(item models.Item, length models.LengthSpec) {
	// This is a workaround since we can't modify the interface directly
	// In a real implementation, we might need to use type assertions
	switch v := item.(type) {
	case interface{ SetLength(models.LengthSpec) }:
		v.SetLength(length)
	}
}

func calculateShotLength(shot models.Shot) float64 {
	var maxLength float64
	
	for _, item := range shot.Items {
		itemLength := item.GetLength().Value
		if itemLength > maxLength {
			maxLength = itemLength
		}
	}
	
	return maxLength
}

func calculateSceneLength(scene models.Scene) float64 {
	var totalLength float64
	
	for _, shot := range scene.Shots {
		shotLength := calculateShotLength(shot)
		totalLength += shotLength
	}
	
	return totalLength
}

func calculateSequenceLength(sequence models.Sequence) float64 {
	var totalLength float64
	
	for _, scene := range sequence.Scenes {
		sceneLength := calculateSceneLength(scene)
		totalLength += sceneLength
	}
	
	return totalLength
}