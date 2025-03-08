package utils

import (
	"encoding/hex"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Windows invalid characters",
			input:    "file<>:\"/\\|?*.txt",
			expected: "file.txt",
		},
		{
			name:     "Japanese characters",
			input:    "四国めたん立ち絵素材2.1.psd",
			expected: "四国めたん立ち絵素材2.1.psd",
		},
		{
			name:     "Path with drive letter",
			input:    "C:\\test:file.txt",
			expected: "Ctestfile.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizePath(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveDriveLetter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Windows path with drive letter",
			input:    "C:\\Users\\test\\file.txt",
			expected: filepath.Join("c", "Users", "test", "file.txt"),
		},
		{
			name:     "UNC path",
			input:    "\\\\server\\share\\file.txt",
			expected: "\\\\server\\share\\file.txt",
		},
		{
			name:     "Relative path",
			input:    "test/file.txt",
			expected: "test/file.txt",
		},
		{
			name:     "Japanese path with drive letter",
			input:    "D:\\作業\\素材\\file.txt",
			expected: filepath.Join("d", "作業", "素材", "file.txt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveDriveLetter(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveDriveLetter(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestProcessPathByMode(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		mode           string
		levels         int
		validateResult func(t *testing.T, result string)
	}{
		{
			name:   "Full mode with drive letter",
			path:   "C:\\Users\\test\\file.txt",
			mode:   "full",
			levels: 0,
			validateResult: func(t *testing.T, result string) {
				expected := filepath.Join("c", "Users", "test", "file.txt")
				if result != expected {
					t.Errorf("got %q, want %q", result, expected)
				}
			},
		},
		{
			name:   "Partial mode with 2 levels",
			path:   "C:\\Users\\test\\documents\\file.txt",
			mode:   "partial",
			levels: 2,
			validateResult: func(t *testing.T, result string) {
				expected := filepath.Join("test", "documents", "file.txt")
				if result != expected {
					t.Errorf("got %q, want %q", result, expected)
				}
			},
		},
		{
			name:   "Flat mode",
			path:   "C:\\Users\\test\\file.txt",
			mode:   "flat",
			levels: 0,
			validateResult: func(t *testing.T, result string) {
				// 1. ハッシュ部分（8文字）+ ハイフン + ファイル名
				parts := strings.Split(result, "-")
				if len(parts) != 2 {
					t.Errorf("expected format 'hash-filename', got %q", result)
					return
				}
				
				// 2. ハッシュ部分が8文字であることを確認
				if len(parts[0]) != 8 {
					t.Errorf("hash part should be 8 characters, got %d characters: %q", len(parts[0]), parts[0])
					return
				}
				
				// 3. ファイル名部分が正しいことを確認
				if parts[1] != "file.txt" {
					t.Errorf("filename part should be 'file.txt', got %q", parts[1])
					return
				}
				
				// 4. ハッシュ部分が16進数であることを確認
				if _, err := hex.DecodeString(parts[0]); err != nil {
					t.Errorf("hash part should be hexadecimal, got %q", parts[0])
					return
				}
			},
		},
		{
			name:   "Japanese path in partial mode",
			path:   "D:\\作業\\素材\\アセット\\file.txt",
			mode:   "partial",
			levels: 2,
			validateResult: func(t *testing.T, result string) {
				expected := filepath.Join("素材", "アセット", "file.txt")
				if result != expected {
					t.Errorf("got %q, want %q", result, expected)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProcessPathByMode(tt.path, tt.mode, tt.levels)
			tt.validateResult(t, result)
		})
	}
}

func TestProcessPathByModeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		mode     string
		levels   int
		expected string
	}{
		{
			name:     "Empty path",
			path:     "",
			mode:     "full",
			levels:   0,
			expected: "",
		},
		{
			name:     "Path with only filename",
			path:     "test.txt",
			mode:     "partial",
			levels:   2,
			expected: "test.txt",
		},
		{
			name:     "Path with spaces",
			path:     "C:\\Program Files\\My App\\file with spaces.txt",
			mode:     "full",
			levels:   0,
			expected: filepath.Join("c", "Program Files", "My App", "file with spaces.txt"),
		},
		{
			name:     "UNC path in flat mode",
			path:     "\\\\server\\share\\folder\\file.txt",
			mode:     "flat",
			levels:   0,
			expected: "39c857e7-file.txt",
		},
		{
			name:     "Path with dot segments",
			path:     "C:\\folder\\.\\subfolder\\..\\file.txt",
			mode:     "full",
			levels:   0,
			expected: filepath.Join("c", "folder", "file.txt"),
		},
		{
			name:     "Very deep path in partial mode",
			path:     "C:\\1\\2\\3\\4\\5\\6\\7\\8\\9\\10\\file.txt",
			mode:     "partial",
			levels:   3,
			expected: filepath.Join("8", "9", "10", "file.txt"),
		},
		{
			name:     "Path with special characters",
			path:     "C:\\フォルダー\\🎮\\test～!@#$%^&().txt",
			mode:     "full",
			levels:   0,
			expected: filepath.Join("c", "フォルダー", "🎮", "test～!@#$%^&().txt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProcessPathByMode(tt.path, tt.mode, tt.levels)
			if result != tt.expected {
				t.Errorf("ProcessPathByMode(%q, %q, %d) = %q, want %q",
					tt.path, tt.mode, tt.levels, result, tt.expected)
			}
		})
	}
}