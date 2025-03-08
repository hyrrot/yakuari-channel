package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyFile はファイルをコピーします
func CopyFile(src, dst string) error {
	// 出力先ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました %s: %w", filepath.Dir(dst), err)
	}

	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("ソースファイルのオープンに失敗しました %s: %w", src, err)
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("出力ファイルの作成に失敗しました %s: %w", dst, err)
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return fmt.Errorf("ファイルのコピーに失敗しました %s -> %s: %w", src, dst, err)
	}

	return nil
}

// FileExists はファイルが存在するかチェックします
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
} 