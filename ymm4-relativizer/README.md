# ymm4-relativizer

YMM4のプロジェクトファイル（.ymmp）内のファイルパスを相対パス化/絶対パス化するツールです。

## 機能

### 相対化 (relativize)
YMMPファイルのFilePath要素を相対パス化し、参照されているファイルを指定されたディレクトリにコピーします。

### 絶対化 (absolutize)
相対パス化されたYMMPファイル（.ymmpr）を元のYMMPファイル形式に戻します。

## 使用方法

### 相対化
```bash
ymm4-relativizer -mode relativize -input input.ymmp -output ./output -dirmode [mode]
```

オプション：
- `-mode`: 変換モード（relativize/absolutize）
- `-input`: 入力ファイルパス
- `-output`: 出力ディレクトリ
- `-assets`: アセットディレクトリ名（デフォルト: assets）
- `-dirmode`: ディレクトリモード（full/partial/flat）
- `-levels`: partialモード時の保持するディレクトリレベル数
- `-skip-missing`: 存在しないファイルをスキップ

### 絶対化
```bash
ymm4-relativizer -mode absolutize -input input.ymmpr -output ./output
```

## ディレクトリモード

### full
元のディレクトリ構造を維持します（ドライブレター除去）

### partial
指定されたレベル数のディレクトリ階層のみを保持します

### flat
すべてのファイルを1つのディレクトリに配置し、ハッシュ値を付加して名前の衝突を防ぎます 