package converter

import (
	"fmt"
	"strings"

	"github.com/yakuari-channel/video-authorizer/internal/models"
)

// TemplateValidator validates template references and availability
type TemplateValidator struct{}

// NewTemplateValidator creates a new template validator
func NewTemplateValidator() *TemplateValidator {
	return &TemplateValidator{}
}

// ValidateTemplateReferences validates that all template references in YMMPS can be found in the template YMMP
func (tv *TemplateValidator) ValidateTemplateReferences(ymmps *models.YMMPSDocument, template *models.YMMPProject) error {
	// Collect all template references from YMMPS
	templateRefs := tv.collectTemplateReferences(ymmps)
	
	// Collect all available templates from YMMP
	availableTemplates := tv.collectAvailableTemplates(template)
	
	// Validate each reference
	var errors []string
	
	for _, ref := range templateRefs {
		if !tv.isTemplateAvailable(ref.TemplateName, availableTemplates) {
			errors = append(errors, fmt.Sprintf("template '%s' not found (referenced at %s)", 
				ref.TemplateName, ref.Location))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("template validation failed:\n%s", strings.Join(errors, "\n"))
	}
	
	return nil
}

// TemplateReference represents a template reference in YMMPS
type TemplateReference struct {
	TemplateName string
	Location     string // Description of where it's referenced
}

// AvailableTemplate represents an available template in YMMP
type AvailableTemplate struct {
	Remark   string
	ItemType string
}

// collectTemplateReferences collects all template references from YMMPS document
func (tv *TemplateValidator) collectTemplateReferences(ymmps *models.YMMPSDocument) []TemplateReference {
	var refs []TemplateReference
	
	for seqIdx, seq := range ymmps.Sequences {
		seqLocation := fmt.Sprintf("Sequence[%d]", seqIdx)
		if seq.ID != "" {
			seqLocation = fmt.Sprintf("Sequence[%d](%s)", seqIdx, seq.ID)
		}
		
		for sceneIdx, scene := range seq.Scenes {
			sceneLocation := fmt.Sprintf("%s.Scene[%d]", seqLocation, sceneIdx)
			if scene.ID != "" {
				sceneLocation = fmt.Sprintf("%s.Scene[%d](%s)", seqLocation, sceneIdx, scene.ID)
			}
			
			for shotIdx, shot := range scene.Shots {
				shotLocation := fmt.Sprintf("%s.Shot[%d]", sceneLocation, shotIdx)
				if shot.ID != "" {
					shotLocation = fmt.Sprintf("%s.Shot[%d](%s)", sceneLocation, shotIdx, shot.ID)
				}
				
				for itemIdx, item := range shot.Items {
					itemLocation := fmt.Sprintf("%s.Item[%d]", shotLocation, itemIdx)
					
					refs = append(refs, TemplateReference{
						TemplateName: item.Template,
						Location:     itemLocation,
					})
				}
			}
		}
	}
	
	return refs
}

// collectAvailableTemplates collects all available templates from YMMP project
func (tv *TemplateValidator) collectAvailableTemplates(template *models.YMMPProject) []AvailableTemplate {
	var templates []AvailableTemplate
	
	// Check if template has timelines
	if len(template.Timelines) == 0 {
		return templates
	}
	
	// Only check the first timeline (as per specification)
	timeline := template.Timelines[0]
	
	for _, item := range timeline.Items {
		remark := tv.getItemRemark(item)
		itemType := tv.getItemType(item)
		
		if remark != "" {
			templates = append(templates, AvailableTemplate{
				Remark:   remark,
				ItemType: itemType,
			})
		}
	}
	
	return templates
}

// getItemRemark extracts the remark from any item type
func (tv *TemplateValidator) getItemRemark(item interface{}) string {
	switch v := item.(type) {
	case *models.VoiceItem:
		return v.Remark
	case *models.VideoItem:
		return v.Remark
	case *models.TachieItem:
		return v.Remark
	case map[string]interface{}:
		// Handle generic item as map
		if remark, ok := v["Remark"].(string); ok {
			return remark
		}
	}
	return ""
}

// getItemType extracts the type from any item type
func (tv *TemplateValidator) getItemType(item interface{}) string {
	switch v := item.(type) {
	case *models.VoiceItem:
		return "VoiceItem"
	case *models.VideoItem:
		return "VideoItem"
	case *models.TachieItem:
		return "TachieItem"
	case map[string]interface{}:
		// Handle generic item as map
		if itemType, ok := v["$type"].(string); ok {
			return itemType
		}
	}
	return "Unknown"
}

// isTemplateAvailable checks if a template name is available in the template list
func (tv *TemplateValidator) isTemplateAvailable(templateName string, available []AvailableTemplate) bool {
	for _, template := range available {
		if template.Remark == templateName {
			return true
		}
	}
	return false
}

// ValidateTemplateCompatibility validates that template items are compatible with their usage
func (tv *TemplateValidator) ValidateTemplateCompatibility(ymmps *models.YMMPSDocument, template *models.YMMPProject) error {
	// Collect template references with more context
	templateUsage := tv.collectTemplateUsageWithContext(ymmps)
	
	// Collect available templates with types
	availableTemplates := tv.collectAvailableTemplates(template)
	
	// Validate compatibility
	var warnings []string
	
	for _, usage := range templateUsage {
		availableTemplate := tv.findTemplateByRemark(usage.TemplateName, availableTemplates)
		if availableTemplate == nil {
			continue // This will be caught by ValidateTemplateReferences
		}
		
		// Check for potential compatibility issues
		if warning := tv.checkCompatibilityWarning(usage, *availableTemplate); warning != "" {
			warnings = append(warnings, warning)
		}
	}
	
	// For now, we only log warnings. In a production system, these could be returned as warnings
	if len(warnings) > 0 {
		fmt.Printf("Template compatibility warnings:\n")
		for _, warning := range warnings {
			fmt.Printf("  - %s\n", warning)
		}
	}
	
	return nil
}

// TemplateUsage represents how a template is being used
type TemplateUsage struct {
	TemplateName   string
	Location       string
	HasVoiceProps  bool // Has voice-related properties
	HasVideoProps  bool // Has video-related properties
	HasImageProps  bool // Has image-related properties
}

// collectTemplateUsageWithContext collects template usage with context about properties
func (tv *TemplateValidator) collectTemplateUsageWithContext(ymmps *models.YMMPSDocument) []TemplateUsage {
	var usage []TemplateUsage
	
	for seqIdx, seq := range ymmps.Sequences {
		seqLocation := fmt.Sprintf("Sequence[%d]", seqIdx)
		if seq.ID != "" {
			seqLocation = fmt.Sprintf("Sequence[%d](%s)", seqIdx, seq.ID)
		}
		
		for sceneIdx, scene := range seq.Scenes {
			sceneLocation := fmt.Sprintf("%s.Scene[%d]", seqLocation, sceneIdx)
			if scene.ID != "" {
				sceneLocation = fmt.Sprintf("%s.Scene[%d](%s)", seqLocation, sceneIdx, scene.ID)
			}
			
			for shotIdx, shot := range scene.Shots {
				shotLocation := fmt.Sprintf("%s.Shot[%d]", sceneLocation, shotIdx)
				if shot.ID != "" {
					shotLocation = fmt.Sprintf("%s.Shot[%d](%s)", sceneLocation, shotIdx, shot.ID)
				}
				
				for itemIdx, item := range shot.Items {
					itemLocation := fmt.Sprintf("%s.Item[%d]", shotLocation, itemIdx)
					
					usage = append(usage, TemplateUsage{
						TemplateName:  item.Template,
						Location:      itemLocation,
						HasVoiceProps: tv.hasVoiceProperties(item.Properties),
						HasVideoProps: tv.hasVideoProperties(item.Properties),
						HasImageProps: tv.hasImageProperties(item.Properties),
					})
				}
			}
		}
	}
	
	return usage
}

// findTemplateByRemark finds a template by its remark
func (tv *TemplateValidator) findTemplateByRemark(remark string, templates []AvailableTemplate) *AvailableTemplate {
	for _, template := range templates {
		if template.Remark == remark {
			return &template
		}
	}
	return nil
}

// checkCompatibilityWarning checks for potential compatibility issues
func (tv *TemplateValidator) checkCompatibilityWarning(usage TemplateUsage, template AvailableTemplate) string {
	// Check if voice properties are being used with non-voice templates
	if usage.HasVoiceProps && !strings.Contains(template.ItemType, "Voice") {
		return fmt.Sprintf("voice properties used with non-voice template '%s' at %s", 
			usage.TemplateName, usage.Location)
	}
	
	// Check if video properties are being used with non-video templates
	if usage.HasVideoProps && !strings.Contains(template.ItemType, "Video") {
		return fmt.Sprintf("video properties used with non-video template '%s' at %s", 
			usage.TemplateName, usage.Location)
	}
	
	return ""
}

// hasVoiceProperties checks if properties contain voice-related fields
func (tv *TemplateValidator) hasVoiceProperties(props map[string]interface{}) bool {
	voiceProps := []string{"Serif", "Hatsuon", "CharacterName", "VoiceLength"}
	
	for _, prop := range voiceProps {
		if _, exists := props[prop]; exists {
			return true
		}
	}
	
	return false
}

// hasVideoProperties checks if properties contain video-related fields
func (tv *TemplateValidator) hasVideoProperties(props map[string]interface{}) bool {
	videoProps := []string{"FilePath", "Volume", "PlaybackRate"}
	
	for _, prop := range videoProps {
		if _, exists := props[prop]; exists {
			return true
		}
	}
	
	return false
}

// hasImageProperties checks if properties contain image-related fields
func (tv *TemplateValidator) hasImageProperties(props map[string]interface{}) bool {
	imageProps := []string{"ImagePath", "Width", "Height", "X", "Y", "Zoom"}
	
	for _, prop := range imageProps {
		if _, exists := props[prop]; exists {
			return true
		}
	}
	
	return false
}