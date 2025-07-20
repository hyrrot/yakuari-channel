package converter

import (
	"path/filepath"
	"strings"
)

// PathResolver handles path resolution for relative paths
type PathResolver struct {
	basePath string
}

// GetBasePath returns the current base path
func (pr *PathResolver) GetBasePath() string {
	return pr.basePath
}

// NewPathResolver creates a new path resolver
func NewPathResolver(basePath string) *PathResolver {
	return &PathResolver{
		basePath: basePath,
	}
}

// ResolvePath resolves a path that may be relative
func (pr *PathResolver) ResolvePath(path string) string {
	// If path is already absolute, return as-is
	if filepath.IsAbs(path) {
		return path
	}
	
	// If no base path is set, return the path as-is
	if pr.basePath == "" {
		return path
	}
	
	// Resolve relative path against base path
	return filepath.Join(pr.basePath, path)
}

// ResolveFilePath resolves file paths in item properties
func (pr *PathResolver) ResolveFilePath(value interface{}) interface{} {
	// Handle string values
	if str, ok := value.(string); ok {
		// Check if this looks like a file path
		if isFilePath(str) {
			return pr.ResolvePath(str)
		}
	}
	
	// Handle map values (nested properties)
	if m, ok := value.(map[string]interface{}); ok {
		resolved := make(map[string]interface{})
		for k, v := range m {
			resolved[k] = pr.ResolveFilePath(v)
		}
		return resolved
	}
	
	// Handle slice values
	if slice, ok := value.([]interface{}); ok {
		resolved := make([]interface{}, len(slice))
		for i, v := range slice {
			resolved[i] = pr.ResolveFilePath(v)
		}
		return resolved
	}
	
	// Return value as-is for other types
	return value
}

// isFilePath checks if a string looks like a file path
func isFilePath(s string) bool {
	// Check for common file extensions
	extensions := []string{
		".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp", // Images
		".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv",   // Videos
		".mp3", ".wav", ".ogg", ".m4a", ".aac", ".flac",  // Audio
		".txt", ".json", ".yaml", ".yml", ".xml",         // Text files
	}
	
	lower := strings.ToLower(s)
	for _, ext := range extensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	
	// Check for path separators
	if strings.Contains(s, "/") || strings.Contains(s, "\\") {
		return true
	}
	
	return false
}

// ResolveItemProperties resolves paths in item properties
func (pr *PathResolver) ResolveItemProperties(properties map[string]interface{}) map[string]interface{} {
	resolved := make(map[string]interface{})
	
	for key, value := range properties {
		// Special handling for known file path properties
		if isFilePathProperty(key) {
			resolved[key] = pr.ResolveFilePath(value)
		} else {
			// Recursively resolve nested properties
			resolved[key] = pr.ResolveFilePath(value)
		}
	}
	
	return resolved
}

// isFilePathProperty checks if a property name is known to contain file paths
func isFilePathProperty(key string) bool {
	filePathProperties := []string{
		"FilePath",
		"Path",
		"ImagePath",
		"VideoPath",
		"AudioPath",
		"SourcePath",
		"TargetPath",
		"OutputPath",
		"InputPath",
	}
	
	for _, prop := range filePathProperties {
		if strings.EqualFold(key, prop) {
			return true
		}
	}
	
	return false
}