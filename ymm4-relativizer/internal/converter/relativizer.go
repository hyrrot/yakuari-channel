package converter

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/hyrrot/ymm4-relativizer/internal/model"
	"github.com/hyrrot/ymm4-relativizer/internal/utils"
)

type RelativizerConfig struct {
	InputPath       string
	OutputDir       string
	AssetsDir       string
	DirectoryMode   string
	SkipMissing     bool
	DirectoryLevels int
}

func Relativize(config RelativizerConfig) error {
	// 入力ファイルを読み込む
	data, err := os.ReadFile(config.InputPath)
	if err != nil {
		return fmt.Errorf("入力ファイルの読み込みに失敗: %w", err)
	}

	// YMMPファイルをパース
	ymmp, err := model.ParseYMMP(data)
	if err != nil {
		return fmt.Errorf("YMMPファイルのパースに失敗: %w", err)
	}

	// アセットディレクトリのパスを作成
	assetsDir := filepath.Join(config.OutputDir, config.AssetsDir)

	// ファイルパスを処理する関数
	processPath := func(path string, isRoot bool) string {
		if path == "" {
			return path
		}

		// ルート要素のFilePathはnullに設定
		if isRoot {
			return ""
		}

		// ファイルの存在確認
		if !utils.FileExists(path) {
			if config.SkipMissing {
				return path
			}
			return path // エラー処理は呼び出し側で行う
		}

		// 新しいパスを生成
		newPath := utils.ProcessPathByMode(path, config.DirectoryMode, config.DirectoryLevels)
		dstPath := filepath.Join(assetsDir, newPath)

		// ファイルをコピー
		if err := utils.CopyFile(path, dstPath); err != nil {
			fmt.Fprintf(os.Stderr, "警告: ファイルのコピーに失敗 %s: %v\n", path, err)
			return path
		}

		// 相対パスを返す
		rel, err := filepath.Rel(config.OutputDir, dstPath)
		if err != nil {
			return path
		}
		return filepath.ToSlash(rel)
	}

	// FilePathを更新
	ymmp.UpdateFilePaths(processPath)

	// 結果を保存
	output, err := ymmp.ToJSON()
	if err != nil {
		return fmt.Errorf("JSONの生成に失敗: %w", err)
	}

	// 出力ファイルを保存
	outPath := filepath.Join(config.OutputDir, filepath.Base(config.InputPath[:len(config.InputPath)-len(filepath.Ext(config.InputPath))]+".ymmpr"))
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("出力ディレクトリの作成に失敗: %w", err)
	}
	if err := os.WriteFile(outPath, output, 0644); err != nil {
		return fmt.Errorf("出力ファイルの保存に失敗: %w", err)
	}

	return nil
} 