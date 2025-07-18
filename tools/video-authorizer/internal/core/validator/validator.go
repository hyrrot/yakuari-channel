package validator

import (
	"fmt"
	
	"github.com/user/ymmp-compiler/internal/models"
)

// Validate validates the structure and content of a SYMMP episode
func Validate(episode *models.Episode) error {
	if err := validateRequiredFields(episode); err != nil {
		return err
	}
	
	if err := validateIDs(episode); err != nil {
		return err
	}
	
	return nil
}

// validateRequiredFields checks that all required fields are present
func validateRequiredFields(episode *models.Episode) error {
	if episode.SYMMPFormatVersion == "" {
		return fmt.Errorf("required field 'symmp_format_version' is missing")
	}
	
	if episode.Project.Name == "" {
		return fmt.Errorf("required field 'project.name' is missing")
	}
	
	return nil
}

// validateIDs checks for empty and duplicate IDs
func validateIDs(episode *models.Episode) error {
	sequenceIDs := make(map[string]bool)
	
	for seqIdx, sequence := range episode.Sequences {
		// Check for empty sequence ID
		if sequence.ID == "" {
			return fmt.Errorf("sequence ID cannot be empty (sequence at index %d)", seqIdx)
		}
		
		// Check for duplicate sequence IDs
		if sequenceIDs[sequence.ID] {
			return fmt.Errorf("duplicate sequence ID '%s' found", sequence.ID)
		}
		sequenceIDs[sequence.ID] = true
		
		// Validate scene IDs within this sequence
		sceneIDs := make(map[string]bool)
		for sceneIdx, scene := range sequence.Scenes {
			// Check for empty scene ID
			if scene.ID == "" {
				return fmt.Errorf("scene ID cannot be empty (scene at index %d in sequence '%s')", sceneIdx, sequence.ID)
			}
			
			// Check for duplicate scene IDs within the same sequence
			if sceneIDs[scene.ID] {
				return fmt.Errorf("duplicate scene ID '%s' found in sequence '%s'", scene.ID, sequence.ID)
			}
			sceneIDs[scene.ID] = true
		}
	}
	
	return nil
}