package converter

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/hyrrot/ymm4-relativizer/internal/model"
	"github.com/hyrrot/ymm4-relativizer/internal/utils"
)

type AbsolutizerConfig struct {
	InputPath   string
	OutputDir   string
	SkipMissing bool
}

func Absolutize(config AbsolutizerConfig) error {
	// 出力ディレクトリを作成
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("出力ディレクトリの作成に失敗: %w", err)
	}

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

	// 入力ファイルのディレクトリパス
	inputDir := filepath.Dir(config.InputPath)

	// 出力ファイル名を構築
	outName := filepath.Base(config.InputPath)
	outName = outName[:len(outName)-len(filepath.Ext(outName))] + ".ymmp"
	outPath := filepath.Join(config.OutputDir, outName)
	absOutPath, err := filepath.Abs(outPath)
	if err != nil {
		return fmt.Errorf("出力パスの絶対パス化に失敗: %w", err)
	}

	// ファイルパスを処理する関数
	processPath := func(path string, isRoot bool) string {
		if isRoot {
			// ルート要素のFilePathは出力ファイルの絶対パスを設定
			return absOutPath
		}

		if path == "" {
			return path
		}

		// 相対パスを絶対パスに変換
		absPath := filepath.Join(inputDir, path)
		absPath, err := filepath.Abs(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告: パスの絶対パス化に失敗 %s: %v\n", path, err)
			return path
		}
		
		// ファイルの存在確認
		if !utils.FileExists(absPath) {
			if config.SkipMissing {
				return path
			}
			fmt.Fprintf(os.Stderr, "警告: ファイルが存在しません %s\n", absPath)
			return path
		}

		return absPath
	}

	// FilePathを更新
	ymmp.UpdateFilePaths(processPath)

	// 結果を保存
	output, err := ymmp.ToJSON()
	if err != nil {
		return fmt.Errorf("JSONの生成に失敗: %w", err)
	}

	if err := os.WriteFile(outPath, output, 0644); err != nil {
		return fmt.Errorf("出力ファイルの保存に失敗: %w", err)
	}

	return nil
} 