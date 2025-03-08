package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyrrot/ymm4-relativizer/internal/converter"
)

type Config struct {
	Mode            string
	InputPath       string
	OutputDir       string
	AssetsDir       string
	DirectoryMode   string
	SkipMissing     bool
	DirectoryLevels int
}

func main() {
	config := parseFlags()
	
	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() *Config {
	config := &Config{}
	
	flag.StringVar(&config.Mode, "mode", "relativize", "変換モード (relativize/absolutize)")
	flag.StringVar(&config.InputPath, "input", "", "入力ファイルまたはディレクトリのパス")
	flag.StringVar(&config.OutputDir, "output", "", "出力ディレクトリ")
	flag.StringVar(&config.AssetsDir, "assets", "assets", "アセットディレクトリ名")
	flag.StringVar(&config.DirectoryMode, "dirmode", "full", "ディレクトリモード (full/partial/flat)")
	flag.IntVar(&config.DirectoryLevels, "levels", 2, "保持するディレクトリレベル数")
	flag.BoolVar(&config.SkipMissing, "skip-missing", false, "存在しないファイルをスキップ")
	
	flag.Parse()
	
	if config.InputPath == "" {
		fmt.Fprintln(os.Stderr, "エラー: 入力パスを指定してください")
		flag.Usage()
		os.Exit(1)
	}

	if config.OutputDir == "" {
		fmt.Fprintln(os.Stderr, "エラー: 出力ディレクトリを指定してください")
		flag.Usage()
		os.Exit(1)
	}

	if config.DirectoryMode != "full" && config.DirectoryMode != "partial" && config.DirectoryMode != "flat" {
		fmt.Fprintln(os.Stderr, "エラー: 不正なディレクトリモードです。full, partial, または flat を指定してください")
		flag.Usage()
		os.Exit(1)
	}

	if config.DirectoryMode == "partial" && config.DirectoryLevels < 1 {
		fmt.Fprintln(os.Stderr, "エラー: ディレクトリレベルは1以上を指定してください")
		flag.Usage()
		os.Exit(1)
	}
	
	return config
}

func run(config *Config) error {
	// 出力ディレクトリを作成
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("出力ディレクトリの作成に失敗しました %s: %w", config.OutputDir, err)
	}

	switch config.Mode {
	case "relativize":
		if err := converter.Relativize(converter.RelativizerConfig{
			InputPath:       config.InputPath,
			OutputDir:       config.OutputDir,
			AssetsDir:       config.AssetsDir,
			DirectoryMode:   config.DirectoryMode,
			SkipMissing:     config.SkipMissing,
			DirectoryLevels: config.DirectoryLevels,
		}); err != nil {
			return fmt.Errorf("相対化処理に失敗しました: %w", err)
		}

	case "absolutize":
		if err := converter.Absolutize(converter.AbsolutizerConfig{
			InputPath:   config.InputPath,
			OutputDir:   config.OutputDir,
			SkipMissing: config.SkipMissing,
		}); err != nil {
			return fmt.Errorf("絶対化処理に失敗しました: %w", err)
		}

	default:
		return fmt.Errorf("不正なモード: %s", config.Mode)
	}

	return nil
} 