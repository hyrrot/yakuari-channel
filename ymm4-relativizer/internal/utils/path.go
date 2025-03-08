package utils

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Windowsのドライブレター部分を検出する正規表現
	driveLetterRegex = regexp.MustCompile(`^[A-Za-z]:[\\/]`)
)

// SanitizePath はWindowsのファイルパスとして使用できない文字を除去します
func SanitizePath(path string) string {
	// Windowsで使用できない文字を空文字に置換
	invalid := []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}
	result := path
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "")
	}
	return result
}

// RemoveDriveLetter はパスからドライブレター部分を除去します
func RemoveDriveLetter(path string) string {
	if driveLetterRegex.MatchString(path) {
		// ドライブレター部分を取得（例：D:\）
		driveLetter := path[:2] // D:
		// ドライブレターをディレクトリ名として使用（コロンを除去）
		dirName := strings.ToLower(driveLetter[:1])
		// パスの残りの部分を結合
		remainingPath := path[3:] // 3は "D:\" の長さ
		return filepath.Join(dirName, remainingPath)
	}
	return path
}

// ProcessPathByMode はディレクトリモードに応じてパスを処理します
func ProcessPathByMode(path string, mode string, levels int) string {
	// まずドライブレターを処理
	path = RemoveDriveLetter(path)

	switch mode {
	case "full":
		// フルパスを維持（ドライブレター除去済み）
		return path
		
	case "partial":
		// 指定レベルのディレクトリのみ保持
		parts := strings.Split(filepath.ToSlash(path), "/")
		if len(parts) <= levels {
			return path
		}
		return filepath.Join(parts[len(parts)-levels:]...)
		
	case "flat":
		// ファイル名のみ（ハッシュ付き）
		return GenerateHashedFilename(path)
		
	default:
		return path
	}
}

// GenerateHashedFilename はファイル名にハッシュを付加します
func GenerateHashedFilename(originalPath string) string {
	hash := sha256.Sum256([]byte(originalPath))
	hashStr := fmt.Sprintf("%x", hash)[:8] // 最初の8文字のみ使用
	
	base := filepath.Base(originalPath)
	return fmt.Sprintf("%s-%s", hashStr, SanitizePath(base))
} 