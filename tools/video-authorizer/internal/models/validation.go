package models

import (
	"fmt"
	"strings"
)

// ValidateReferences validates all ID references in the YMMPS document
func (d *YMMPSDocument) ValidateReferences() error {
	// Collect all defined IDs
	definedIDs := make(map[string]string) // ID -> type (SEQUENCE/SCENE/SHOT)
	
	for _, seq := range d.Sequences {
		if seq.ID != "" {
			definedIDs[seq.ID] = "SEQUENCE"
		}
		for _, scene := range seq.Scenes {
			if scene.ID != "" {
				definedIDs[scene.ID] = "SCENE"
			}
			for _, shot := range scene.Shots {
				if shot.ID != "" {
					definedIDs[shot.ID] = "SHOT"
				}
			}
		}
	}
	
	// Validate all references
	var errors []string
	
	for seqIdx, seq := range d.Sequences {
		for sceneIdx, scene := range seq.Scenes {
			for shotIdx, shot := range scene.Shots {
				for itemIdx, item := range shot.Items {
					if err := validateItemReference(item, definedIDs, seqIdx, sceneIdx, shotIdx, itemIdx); err != nil {
						errors = append(errors, err.Error())
					}
				}
			}
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("ID reference validation failed:\n%s", strings.Join(errors, "\n"))
	}
	
	return nil
}

func validateItemReference(item ItemSpec, definedIDs map[string]string, seqIdx, sceneIdx, shotIdx, itemIdx int) error {
	// Skip if no length specified
	if item.Length == "" {
		return nil
	}
	
	// Parse length to check for ID references
	parsed, err := ParseLength(item.Length)
	if err != nil {
		return nil // ParseLength errors are handled elsewhere
	}
	
	// Check if this is an ID-based reference
	var targetID string
	var expectedType string
	
	switch parsed.Type {
	case LengthTypeUntilSeqEndID:
		targetID = parsed.TargetID
		expectedType = "SEQUENCE"
	case LengthTypeUntilSceneEndID:
		targetID = parsed.TargetID
		expectedType = "SCENE"
	case LengthTypeUntilShotEndID:
		targetID = parsed.TargetID
		expectedType = "SHOT"
	default:
		return nil // Not an ID-based reference
	}
	
	// Validate the reference
	if targetID == "" {
		return fmt.Errorf("empty ID reference at Sequence[%d].Scene[%d].Shot[%d].Item[%d]", 
			seqIdx, sceneIdx, shotIdx, itemIdx)
	}
	
	actualType, exists := definedIDs[targetID]
	if !exists {
		return fmt.Errorf("undefined %s ID '%s' referenced at Sequence[%d].Scene[%d].Shot[%d].Item[%d]", 
			expectedType, targetID, seqIdx, sceneIdx, shotIdx, itemIdx)
	}
	
	if actualType != expectedType {
		return fmt.Errorf("ID '%s' is a %s but referenced as %s at Sequence[%d].Scene[%d].Shot[%d].Item[%d]", 
			targetID, actualType, expectedType, seqIdx, sceneIdx, shotIdx, itemIdx)
	}
	
	return nil
}