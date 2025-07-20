package parser

import (
	"fmt"
	"strings"
	
	"gopkg.in/yaml.v3"
)

// YAMLStructureValidator validates YAML structure before parsing into models
type YAMLStructureValidator struct{}

// NewYAMLStructureValidator creates a new YAML structure validator
func NewYAMLStructureValidator() *YAMLStructureValidator {
	return &YAMLStructureValidator{}
}

// ValidateYMMPSStructure validates the raw YAML structure of a YMMPS document
func (v *YAMLStructureValidator) ValidateYMMPSStructure(raw interface{}) error {
	// Ensure top level is a map
	rootMap, ok := raw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid YAML structure: root must be a map, got %T", raw)
	}

	// Validate YMMPSVersion
	if err := v.validateYMMPSVersion(rootMap); err != nil {
		return err
	}

	// Validate Sequences
	if err := v.validateSequences(rootMap); err != nil {
		return err
	}

	return nil
}

// validateYMMPSVersion validates the YMMPSVersion field
func (v *YAMLStructureValidator) validateYMMPSVersion(root map[string]interface{}) error {
	version, exists := root["YMMPSVersion"]
	if !exists {
		return fmt.Errorf("missing required field: YMMPSVersion")
	}

	versionStr, ok := version.(string)
	if !ok {
		return fmt.Errorf("YMMPSVersion must be a string, got %T", version)
	}

	if versionStr == "" {
		return fmt.Errorf("YMMPSVersion cannot be empty")
	}

	// Currently only version "1" is supported
	if versionStr != "1" {
		return fmt.Errorf("unsupported YMMPSVersion: %s (only \"1\" is supported)", versionStr)
	}

	return nil
}

// validateSequences validates the Sequences field
func (v *YAMLStructureValidator) validateSequences(root map[string]interface{}) error {
	sequences, exists := root["Sequences"]
	if !exists {
		return fmt.Errorf("missing required field: Sequences")
	}

	sequenceSlice, ok := sequences.([]interface{})
	if !ok {
		return fmt.Errorf("Sequences must be an array, got %T", sequences)
	}

	if len(sequenceSlice) == 0 {
		return fmt.Errorf("Sequences array cannot be empty")
	}

	// Validate each sequence
	for i, seq := range sequenceSlice {
		if err := v.validateSequence(seq, i); err != nil {
			return fmt.Errorf("sequence[%d]: %w", i, err)
		}
	}

	return nil
}

// validateSequence validates a single sequence
func (v *YAMLStructureValidator) validateSequence(seq interface{}, index int) error {
	seqMap, ok := seq.(map[string]interface{})
	if !ok {
		return fmt.Errorf("sequence must be a map, got %T", seq)
	}

	// ID is optional, but if present must be a string
	if id, exists := seqMap["ID"]; exists {
		if _, ok := id.(string); !ok {
			return fmt.Errorf("ID must be a string, got %T", id)
		}
	}

	// Validate Scenes
	scenes, exists := seqMap["Scenes"]
	if !exists {
		return fmt.Errorf("missing required field: Scenes")
	}

	sceneSlice, ok := scenes.([]interface{})
	if !ok {
		return fmt.Errorf("Scenes must be an array, got %T", scenes)
	}

	if len(sceneSlice) == 0 {
		return fmt.Errorf("Scenes array cannot be empty")
	}

	// Validate each scene
	for i, scene := range sceneSlice {
		if err := v.validateScene(scene, i); err != nil {
			return fmt.Errorf("scene[%d]: %w", i, err)
		}
	}

	return nil
}

// validateScene validates a single scene
func (v *YAMLStructureValidator) validateScene(scene interface{}, index int) error {
	sceneMap, ok := scene.(map[string]interface{})
	if !ok {
		return fmt.Errorf("scene must be a map, got %T", scene)
	}

	// ID is optional, but if present must be a string
	if id, exists := sceneMap["ID"]; exists {
		if _, ok := id.(string); !ok {
			return fmt.Errorf("ID must be a string, got %T", id)
		}
	}

	// Validate Shots
	shots, exists := sceneMap["Shots"]
	if !exists {
		return fmt.Errorf("missing required field: Shots")
	}

	shotSlice, ok := shots.([]interface{})
	if !ok {
		return fmt.Errorf("Shots must be an array, got %T", shots)
	}

	if len(shotSlice) == 0 {
		return fmt.Errorf("Shots array cannot be empty")
	}

	// Validate each shot
	for i, shot := range shotSlice {
		if err := v.validateShot(shot, i); err != nil {
			return fmt.Errorf("shot[%d]: %w", i, err)
		}
	}

	return nil
}

// validateShot validates a single shot
func (v *YAMLStructureValidator) validateShot(shot interface{}, index int) error {
	shotMap, ok := shot.(map[string]interface{})
	if !ok {
		return fmt.Errorf("shot must be a map, got %T", shot)
	}

	// ID is optional, but if present must be a string
	if id, exists := shotMap["ID"]; exists {
		if _, ok := id.(string); !ok {
			return fmt.Errorf("ID must be a string, got %T", id)
		}
	}

	// Validate Items
	items, exists := shotMap["Items"]
	if !exists {
		return fmt.Errorf("missing required field: Items")
	}

	itemSlice, ok := items.([]interface{})
	if !ok {
		return fmt.Errorf("Items must be an array, got %T", items)
	}

	if len(itemSlice) == 0 {
		return fmt.Errorf("Items array cannot be empty")
	}

	// Validate each item
	for i, item := range itemSlice {
		if err := v.validateItem(item, i); err != nil {
			return fmt.Errorf("item[%d]: %w", i, err)
		}
	}

	return nil
}

// validateItem validates a single item
func (v *YAMLStructureValidator) validateItem(item interface{}, index int) error {
	itemMap, ok := item.(map[string]interface{})
	if !ok {
		return fmt.Errorf("item must be a map, got %T", item)
	}

	// Validate Template or _Template
	hasTemplate := false
	if template, exists := itemMap["Template"]; exists {
		if _, ok := template.(string); !ok {
			return fmt.Errorf("Template must be a string, got %T", template)
		}
		hasTemplate = true
	}
	if template, exists := itemMap["_Template"]; exists {
		if _, ok := template.(string); !ok {
			return fmt.Errorf("_Template must be a string, got %T", template)
		}
		hasTemplate = true
	}

	if !hasTemplate {
		return fmt.Errorf("missing required field: Template or _Template")
	}

	// Validate Length
	length, exists := itemMap["Length"]
	if !exists {
		return fmt.Errorf("missing required field: Length")
	}

	// Length can be string or number
	switch length.(type) {
	case string, int, float64:
		// Valid types
	default:
		return fmt.Errorf("Length must be a string or number, got %T", length)
	}

	// Validate Properties if present
	if props, exists := itemMap["Properties"]; exists {
		if _, ok := props.(map[string]interface{}); !ok {
			return fmt.Errorf("Properties must be a map, got %T", props)
		}
	}

	return nil
}

// ValidateYAMLSyntax performs basic YAML syntax validation
func (v *YAMLStructureValidator) ValidateYAMLSyntax(data []byte) error {
	var temp interface{}
	
	// Try to parse the YAML to check for syntax errors
	if err := yaml.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	return nil
}

// DetectCommonYAMLIssues detects common YAML structure issues
func (v *YAMLStructureValidator) DetectCommonYAMLIssues(raw interface{}) []string {
	var issues []string

	// Check for common structural issues
	if rootMap, ok := raw.(map[string]interface{}); ok {
		// Check for extra fields at root level
		allowedRootFields := map[string]bool{
			"YMMPSVersion": true,
			"Sequences":    true,
		}

		for field := range rootMap {
			if !allowedRootFields[field] {
				issues = append(issues, fmt.Sprintf("unexpected field at root level: %s", field))
			}
		}

		// Check for inconsistent ID naming
		ids := v.collectAllIDs(rootMap)
		if inconsistentIDs := v.findInconsistentIDs(ids); len(inconsistentIDs) > 0 {
			issues = append(issues, fmt.Sprintf("inconsistent ID naming: %s", strings.Join(inconsistentIDs, ", ")))
		}
	}

	return issues
}

// collectAllIDs collects all IDs from the document
func (v *YAMLStructureValidator) collectAllIDs(root map[string]interface{}) []string {
	var ids []string

	if sequences, ok := root["Sequences"].([]interface{}); ok {
		for _, seq := range sequences {
			if seqMap, ok := seq.(map[string]interface{}); ok {
				if id, exists := seqMap["ID"]; exists {
					if idStr, ok := id.(string); ok {
						ids = append(ids, idStr)
					}
				}

				if scenes, ok := seqMap["Scenes"].([]interface{}); ok {
					for _, scene := range scenes {
						if sceneMap, ok := scene.(map[string]interface{}); ok {
							if id, exists := sceneMap["ID"]; exists {
								if idStr, ok := id.(string); ok {
									ids = append(ids, idStr)
								}
							}

							if shots, ok := sceneMap["Shots"].([]interface{}); ok {
								for _, shot := range shots {
									if shotMap, ok := shot.(map[string]interface{}); ok {
										if id, exists := shotMap["ID"]; exists {
											if idStr, ok := id.(string); ok {
												ids = append(ids, idStr)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return ids
}

// findInconsistentIDs finds IDs that don't follow consistent naming patterns
func (v *YAMLStructureValidator) findInconsistentIDs(ids []string) []string {
	var inconsistent []string

	// Simple heuristic: check for mixed naming conventions
	hasUnderscores := false
	hasCamelCase := false

	for _, id := range ids {
		if strings.Contains(id, "_") {
			hasUnderscores = true
		}
		if strings.ToLower(id) != id && strings.ToUpper(id) != id {
			hasCamelCase = true
		}
	}

	// If we have both underscores and camelCase, suggest consistency
	if hasUnderscores && hasCamelCase {
		inconsistent = append(inconsistent, "mixed snake_case and camelCase naming")
	}

	return inconsistent
}