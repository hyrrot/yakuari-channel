package converter

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/yakuari-channel/video-authorizer/internal/models"
)

// Converter handles the conversion from YMMPS to YMMP
type Converter struct {
	lengthCalculator  *AdvancedLengthCalculator
	pathResolver      *PathResolver
	templateValidator *TemplateValidator
	filePathUpdater   *FilePathUpdater
}

// NewConverter creates a new converter
func NewConverter() *Converter {
	return &Converter{
		lengthCalculator:  NewAdvancedLengthCalculator(),
		pathResolver:      NewPathResolver(""),
		templateValidator: NewTemplateValidator(),
		filePathUpdater:   NewFilePathUpdater(),
	}
}

// NewConverterWithBasePath creates a new converter with a base path for relative path resolution
func NewConverterWithBasePath(basePath string) *Converter {
	return &Converter{
		lengthCalculator:  NewAdvancedLengthCalculator(),
		pathResolver:      NewPathResolver(basePath),
		templateValidator: NewTemplateValidator(),
		filePathUpdater:   NewFilePathUpdater(),
	}
}

// Convert converts YMMPS document to YMMP project using the template
func (c *Converter) Convert(ymmps *models.YMMPSDocument, template *models.YMMPProject) (*models.YMMPProject, error) {
	// Validate template references before conversion
	if err := c.templateValidator.ValidateTemplateReferences(ymmps, template); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}
	
	// Check template compatibility (warnings only)
	if err := c.templateValidator.ValidateTemplateCompatibility(ymmps, template); err != nil {
		// Template compatibility issues are warnings, not errors
		fmt.Printf("Template compatibility warnings: %v\n", err)
	}
	
	// Use advanced conversion with relative length support
	return c.ConvertWithRelativeLengths(ymmps, template)
}

// ConvertWithOutput converts YMMPS document to YMMP project and updates FilePath elements
func (c *Converter) ConvertWithOutput(ymmps *models.YMMPSDocument, template *models.YMMPProject, outputPath string) (*models.YMMPProject, error) {
	// Perform normal conversion
	project, err := c.Convert(ymmps, template)
	if err != nil {
		return nil, err
	}
	
	// Update FilePath elements
	basePath := c.pathResolver.GetBasePath()
	if basePath == "" {
		// Use output directory as base path if no explicit base path is set
		basePath = c.filePathUpdater.GetOutputDirectory(outputPath)
	}
	
	if err := c.filePathUpdater.UpdateAllFilePaths(project, outputPath, basePath); err != nil {
		return nil, fmt.Errorf("failed to update file paths: %w", err)
	}
	
	return project, nil
}

// processSequence processes a sequence and adds items to the timeline
func (c *Converter) processSequence(template *models.YMMPProject, timeline *models.Timeline, sequence models.Sequence, startFrame int) (int, error) {
	currentFrame := startFrame

	for _, scene := range sequence.Scenes {
		endFrame, err := c.processScene(template, timeline, scene, currentFrame)
		if err != nil {
			return 0, fmt.Errorf("failed to process scene %s: %w", scene.ID, err)
		}
		currentFrame = endFrame
	}

	return currentFrame, nil
}

// processScene processes a scene and adds items to the timeline
func (c *Converter) processScene(template *models.YMMPProject, timeline *models.Timeline, scene models.Scene, startFrame int) (int, error) {
	currentFrame := startFrame

	for _, shot := range scene.Shots {
		endFrame, err := c.processShot(template, timeline, shot, currentFrame)
		if err != nil {
			return 0, fmt.Errorf("failed to process shot %s: %w", shot.ID, err)
		}
		currentFrame = endFrame
	}

	return currentFrame, nil
}

// processShot processes a shot and adds items to the timeline
func (c *Converter) processShot(template *models.YMMPProject, timeline *models.Timeline, shot models.Shot, startFrame int) (int, error) {
	maxEndFrame := startFrame

	for _, itemSpec := range shot.Items {
		item, err := c.processItem(template, itemSpec, startFrame)
		if err != nil {
			return 0, fmt.Errorf("failed to process item with template %s: %w", itemSpec.Template, err)
		}

		timeline.Items = append(timeline.Items, item)

		// Calculate end frame for this item
		length, err := c.parseLength(itemSpec.Length)
		if err != nil {
			return 0, fmt.Errorf("failed to parse length %s: %w", itemSpec.Length, err)
		}

		endFrame := startFrame + length
		if endFrame > maxEndFrame {
			maxEndFrame = endFrame
		}
	}

	return maxEndFrame, nil
}

// processItem processes an individual item spec and creates a timeline item
func (c *Converter) processItem(template *models.YMMPProject, itemSpec models.ItemSpec, frame int) (interface{}, error) {
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
	length, err := c.parseLength(itemSpec.Length)
	if err != nil {
		return nil, fmt.Errorf("failed to parse length: %w", err)
	}

	c.setItemFrame(item, frame)
	c.setItemLength(item, length)

	// Apply property overrides
	if err := c.applyPropertyOverrides(item, itemSpec.Properties); err != nil {
		return nil, fmt.Errorf("failed to apply property overrides: %w", err)
	}

	return item, nil
}

// findTemplateItem finds a template item by its remark
func (c *Converter) findTemplateItem(template *models.YMMPProject, remark string) (interface{}, bool) {
	if len(template.Timelines) == 0 {
		return nil, false
	}

	for _, item := range template.Timelines[0].Items {
		if c.getItemRemark(item) == remark {
			return item, true
		}
	}

	return nil, false
}

// getItemRemark gets the remark from any item type
func (c *Converter) getItemRemark(item interface{}) string {
	switch v := item.(type) {
	case *models.VoiceItem:
		return v.Remark
	case *models.VideoItem:
		return v.Remark
	case *models.TachieItem:
		return v.Remark
	default:
		return ""
	}
}

// setItemFrame sets the frame for any item type
func (c *Converter) setItemFrame(item interface{}, frame int) {
	switch v := item.(type) {
	case *models.VoiceItem:
		v.Frame = frame
	case *models.VideoItem:
		v.Frame = frame
	case *models.TachieItem:
		v.Frame = frame
	}
}

// setItemLength sets the length for any item type
func (c *Converter) setItemLength(item interface{}, length int) {
	switch v := item.(type) {
	case *models.VoiceItem:
		v.Length = length
	case *models.VideoItem:
		v.Length = length
	case *models.TachieItem:
		v.Length = length
	}
}

// parseLength parses a length string to frame count (simplified for now)
func (c *Converter) parseLength(lengthStr string) (int, error) {
	// For now, only handle numeric lengths
	// TODO: Implement relative lengths in Phase 3
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric length: %s", lengthStr)
	}
	return length, nil
}

// applyPropertyOverrides applies property overrides to an item
func (c *Converter) applyPropertyOverrides(item interface{}, properties map[string]interface{}) error {
	if len(properties) == 0 {
		return nil
	}

	// Resolve paths in properties
	resolvedProperties := c.pathResolver.ResolveItemProperties(properties)

	// Convert item to JSON, apply overrides, and convert back
	// This is a simple approach that works for basic property overrides
	itemJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	var itemMap map[string]interface{}
	if err := json.Unmarshal(itemJSON, &itemMap); err != nil {
		return fmt.Errorf("failed to unmarshal item: %w", err)
	}

	// Apply overrides
	for key, value := range resolvedProperties {
		itemMap[key] = value
	}

	// Convert back to specific item type
	updatedJSON, err := json.Marshal(itemMap)
	if err != nil {
		return fmt.Errorf("failed to marshal updated item: %w", err)
	}

	if err := json.Unmarshal(updatedJSON, item); err != nil {
		return fmt.Errorf("failed to unmarshal updated item: %w", err)
	}

	return nil
}

// deepCopyProject creates a deep copy of the project
func (c *Converter) deepCopyProject(project *models.YMMPProject) (*models.YMMPProject, error) {
	data, err := json.Marshal(project)
	if err != nil {
		return nil, err
	}

	var copy models.YMMPProject
	if err := json.Unmarshal(data, &copy); err != nil {
		return nil, err
	}

	return &copy, nil
}

// deepCopyItem creates a deep copy of an item
func (c *Converter) deepCopyItem(item interface{}) (interface{}, error) {
	data, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	// Create new instance of the same type
	switch item.(type) {
	case *models.VoiceItem:
		var copy models.VoiceItem
		if err := json.Unmarshal(data, &copy); err != nil {
			return nil, err
		}
		return &copy, nil
	case *models.VideoItem:
		var copy models.VideoItem
		if err := json.Unmarshal(data, &copy); err != nil {
			return nil, err
		}
		return &copy, nil
	case *models.TachieItem:
		var copy models.TachieItem
		if err := json.Unmarshal(data, &copy); err != nil {
			return nil, err
		}
		return &copy, nil
	default:
		return nil, fmt.Errorf("unsupported item type")
	}
}