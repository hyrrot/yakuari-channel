package utils_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hyrrot/ymm4-relativizer/internal/utils"
)

func TestFileExists(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() string
		cleanup  func(string)
		expected bool
	}{
		{
			name: "存在するファイル",
			setup: func() string {
				f, err := os.CreateTemp("", "test")
				if err != nil {
					t.Fatal(err)
				}
				f.Close()
				return f.Name()
			},
			cleanup: func(path string) {
				os.Remove(path)
			},
			expected: true,
		},
		{
			name: "存在しないファイル",
			setup: func() string {
				return "non_existent_file.txt"
			},
			cleanup: func(string) {},
			expected: false,
		},
		{
			name: "アクセス権限のないファイル",
			setup: func() string {
				f, err := os.CreateTemp("", "test")
				if err != nil {
					t.Fatal(err)
				}
				path := f.Name()
				f.Close()
				os.Chmod(path, 0000)
				return path
			},
			cleanup: func(path string) {
				os.Chmod(path, 0666)
				os.Remove(path)
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			defer tt.cleanup(path)

			if got := utils.FileExists(path); got != tt.expected {
				t.Errorf("FileExists() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (string, string)
		cleanup     func(string, string)
		wantErr     bool
		errContains string
	}{
		{
			name: "正常なコピー",
			setup: func() (string, string) {
				src, err := os.CreateTemp("", "src")
				if err != nil {
					t.Fatal(err)
				}
				src.WriteString("test content")
				src.Close()
				dst := src.Name() + ".copy"
				return src.Name(), dst
			},
			cleanup: func(src, dst string) {
				os.Remove(src)
				os.Remove(dst)
			},
			wantErr: false,
		},
		{
			name: "存在しないソースファイル",
			setup: func() (string, string) {
				return "non_existent_source.txt", "dest.txt"
			},
			cleanup: func(_, dst string) {
				os.Remove(dst)
			},
			wantErr:     true,
			errContains: "no such file",
		},
		{
			name: "書き込み権限のない出力先",
			setup: func() (string, string) {
				src, err := os.CreateTemp("", "src")
				if err != nil {
					t.Fatal(err)
				}
				src.WriteString("test content")
				src.Close()

				dstDir, err := os.MkdirTemp("", "readonly")
				if err != nil {
					t.Fatal(err)
				}
				os.Chmod(dstDir, 0500)
				return src.Name(), filepath.Join(dstDir, "dest.txt")
			},
			cleanup: func(src, dst string) {
				os.Remove(src)
				dir := filepath.Dir(dst)
				os.Chmod(dir, 0700)
				os.RemoveAll(dir)
			},
			wantErr:     true,
			errContains: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := tt.setup()
			defer tt.cleanup(src, dst)

			err := utils.CopyFile(src, dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("CopyFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("CopyFile() error = %v, want error containing %v", err, tt.errContains)
				}
			}

			// 正常なコピーの場合、内容を検証
			if !tt.wantErr {
				srcContent, err := os.ReadFile(src)
				if err != nil {
					t.Fatalf("ソースファイルの読み込みに失敗: %v", err)
				}
				dstContent, err := os.ReadFile(dst)
				if err != nil {
					t.Fatalf("コピーされたファイルの読み込みに失敗: %v", err)
				}
				if string(srcContent) != string(dstContent) {
					t.Errorf("コピーされたファイルの内容が一致しません")
				}
			}
		})
	}
}

func TestGenerateHashedFilename(t *testing.T) {
	tests := []struct {
		name         string
		originalPath string
		wantPattern string
	}{
		{
			name:         "通常のファイルパス",
			originalPath: "test.wav",
			wantPattern: `^[a-f0-9]{32}\.wav$`,
		},
		{
			name:         "パスを含むファイル名",
			originalPath: "path/to/test.wav",
			wantPattern: `^[a-f0-9]{32}\.wav$`,
		},
		{
			name:         "拡張子なしのファイル",
			originalPath: "test",
			wantPattern: `^[a-f0-9]{32}$`,
		},
		{
			name:         "複数の拡張子",
			originalPath: "test.tar.gz",
			wantPattern: `^[a-f0-9]{32}\.tar\.gz$`,
		},
		{
			name:         "日本語のファイル名",
			originalPath: "テスト.wav",
			wantPattern: `^[a-f0-9]{32}\.wav$`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.GenerateHashedFilename(tt.originalPath)
			matched, err := regexp.MatchString(tt.wantPattern, got)
			if err != nil {
				t.Errorf("正規表現のマッチングに失敗: %v", err)
			}
			if !matched {
				t.Errorf("GenerateHashedFilename() = %v, want pattern %v", got, tt.wantPattern)
			}

			// 同じ入力で2回呼び出した場合、同じハッシュが生成されることを確認
			got2 := utils.GenerateHashedFilename(tt.originalPath)
			if got != got2 {
				t.Errorf("同じ入力に対して異なるハッシュが生成されました: %v != %v", got, got2)
			}
		})
	}
} 