package converter

import (
	"path/filepath"

	"github.com/yakuari-channel/video-authorizer/internal/models"
)

// FilePathUpdater handles updating FilePath elements in YMMP projects
type FilePathUpdater struct{}

// NewFilePathUpdater creates a new FilePath updater
func NewFilePathUpdater() *FilePathUpdater {
	return &FilePathUpdater{}
}

// UpdateProjectFilePath updates the main FilePath element of the YMMP project
func (fpu *FilePathUpdater) UpdateProjectFilePath(project *models.YMMPProject, outputPath string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		return err
	}

	// Update the project's FilePath
	project.FilePath = absPath
	
	return nil
}

// UpdateItemFilePaths updates FilePath fields in timeline items to use absolute paths
func (fpu *FilePathUpdater) UpdateItemFilePaths(project *models.YMMPProject, basePath string) error {
	for timelineIdx := range project.Timelines {
		timeline := &project.Timelines[timelineIdx]
		
		for itemIdx := range timeline.Items {
			if err := fpu.updateItemFilePath(timeline.Items[itemIdx], basePath); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// updateItemFilePath updates FilePath fields in a single item
func (fpu *FilePathUpdater) updateItemFilePath(item interface{}, basePath string) error {
	switch v := item.(type) {
	case *models.VideoItem:
		if v.FilePath != "" {
			absPath, err := fpu.resolveToAbsolutePath(v.FilePath, basePath)
			if err != nil {
				return err
			}
			v.FilePath = absPath
		}
		
	case *models.TachieItem:
		// TachieItem might have file paths in TachieItemParameter
		if v.TachieItemParameter != nil {
			fpu.updateNestedFilePaths(v.TachieItemParameter, basePath)
		}
		
	case map[string]interface{}:
		// Handle generic items
		if filePath, exists := v["FilePath"]; exists {
			if filePathStr, ok := filePath.(string); ok && filePathStr != "" {
				absPath, err := fpu.resolveToAbsolutePath(filePathStr, basePath)
				if err != nil {
					return err
				}
				v["FilePath"] = absPath
			}
		}
		
		// Handle nested file paths in properties
		fpu.updateNestedFilePaths(v, basePath)
	}
	
	return nil
}

// updateNestedFilePaths updates file paths in nested properties
func (fpu *FilePathUpdater) updateNestedFilePaths(obj map[string]interface{}, basePath string) {
	filePathKeys := []string{
		"ImagePath", "VideoPath", "AudioPath", "SourcePath", 
		"TargetPath", "OutputPath", "InputPath", "Path",
	}
	
	for _, key := range filePathKeys {
		if value, exists := obj[key]; exists {
			if pathStr, ok := value.(string); ok && pathStr != "" {
				if absPath, err := fpu.resolveToAbsolutePath(pathStr, basePath); err == nil {
					obj[key] = absPath
				}
			}
		}
	}
	
	// Recursively handle nested objects
	for _, value := range obj {
		if nested, ok := value.(map[string]interface{}); ok {
			fpu.updateNestedFilePaths(nested, basePath)
		}
	}
}

// resolveToAbsolutePath resolves a path to absolute path
func (fpu *FilePathUpdater) resolveToAbsolutePath(path, basePath string) (string, error) {
	// If already absolute, return as-is
	if filepath.IsAbs(path) {
		return path, nil
	}
	
	// If no base path, use current directory
	if basePath == "" {
		return filepath.Abs(path)
	}
	
	// Resolve relative to base path
	fullPath := filepath.Join(basePath, path)
	return filepath.Abs(fullPath)
}

// UpdateAllFilePaths updates both project FilePath and item FilePaths
func (fpu *FilePathUpdater) UpdateAllFilePaths(project *models.YMMPProject, outputPath, basePath string) error {
	// Update project FilePath
	if err := fpu.UpdateProjectFilePath(project, outputPath); err != nil {
		return err
	}
	
	// Update item FilePaths
	if err := fpu.UpdateItemFilePaths(project, basePath); err != nil {
		return err
	}
	
	return nil
}

// GetOutputDirectory returns the directory where the output file will be written
func (fpu *FilePathUpdater) GetOutputDirectory(outputPath string) string {
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		// Fallback to using the path as-is
		return filepath.Dir(outputPath)
	}
	
	return filepath.Dir(absPath)
}