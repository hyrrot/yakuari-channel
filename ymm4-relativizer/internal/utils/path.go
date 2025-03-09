package utils

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Regular expression to detect Windows drive letter
	driveLetterRegex = regexp.MustCompile(`^[A-Za-z]:[\\/]`)
)

// SanitizePath removes characters that are invalid in Windows file paths
func SanitizePath(path string) string {
	// Replace characters that are invalid in Windows
	invalid := []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}
	result := path
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "")
	}
	return result
}

// RemoveDriveLetter removes the drive letter portion from a path
func RemoveDriveLetter(path string) string {
	if driveLetterRegex.MatchString(path) {
		// Get drive letter part (e.g., D:\)
		driveLetter := path[:2] // D:
		// Use drive letter as directory name (remove colon)
		dirName := strings.ToLower(driveLetter[:1])
		// Join with remaining path
		remainingPath := path[3:] // 3 is length of "D:\"
		return filepath.Join(dirName, remainingPath)
	}
	return path
}

// ProcessPathByMode processes the path according to the directory mode
func ProcessPathByMode(path string, mode string, levels int) string {
	// Process drive letter first
	path = RemoveDriveLetter(path)

	switch mode {
	case "full":
		return path
		
	case "partial":
		parts := strings.Split(filepath.ToSlash(path), "/")
		if len(parts) <= levels+1 {
			return path
		}
		return filepath.FromSlash(strings.Join(parts[len(parts)-levels-1:], "/"))
		
	case "flat":
		// パスを正規化してからハッシュを生成
		normalizedPath := filepath.ToSlash(path)
		hash := sha256.Sum256([]byte(normalizedPath))
		hashStr := fmt.Sprintf("%x", hash)[:8]
		return fmt.Sprintf("%s-%s", hashStr, filepath.Base(path))
		
	default:
		return path
	}
}

// GenerateHashedFilename adds a hash to the filename
func GenerateHashedFilename(originalPath string) string {
	hash := sha256.Sum256([]byte(originalPath))
	hashStr := fmt.Sprintf("%x", hash)[:8] // Use only first 8 characters
	
	base := filepath.Base(originalPath)
	return fmt.Sprintf("%s-%s", hashStr, SanitizePath(base))
} 