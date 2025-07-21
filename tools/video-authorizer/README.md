# YMMP Scenario Converter (ymmps-authorizer)

YMMP Scenario Converter は、YukkuriMovieMaker (YMM) のプロジェクトファイル（.ymmp）をより効率的に編集するためのツールです。YMMPS（YMMP Scenario）フォーマットという中間表現を使用して、YAMLファイルベースの宣言的な動画構成を可能にします。

## 目次

- [特徴](#特徴)
- [インストール](#インストール)
- [基本的な使い方](#基本的な使い方)
- [オプション](#オプション)
- [YMMPS フォーマット](#ymmps-フォーマット)
- [長さ指定](#長さ指定)
- [使用例](#使用例)
- [トラブルシューティング](#トラブルシューティング)

## 特徴

### 🎯 主要機能
- **宣言的な動画構成**: YAMLベースのYMMPSフォーマットで動画構造を定義
- **テンプレートベース**: 既存のYMMPファイルをテンプレートとして使用
- **相対パスサポート**: ファイルの移動が容易
- **自動長さ計算**: VOICEVOX API、ffprobe を使用した自動長さ計算
- **包括的な検証**: YAML構造、ID参照、テンプレート要素の検証

### 🔧 CLI機能
- **バリデーション専用モード**: `--validate-only`
- **ドライランモード**: `--dry-run` 
- **詳細ログ出力**: `--verbose`
- **相対パス基準指定**: `--base-path`

## インストール

### ビルドから（推奨）

```bash
# リポジトリをクローン
git clone <repository-url>
cd tools/video-authorizer

# ビルド
go build -o ymmps-authorizer ./cmd/ymmps-authorizer

# Windows の場合
go build -o ymmps-authorizer.exe ./cmd/ymmps-authorizer
```

### 要件

- Go 1.19 以上
- （オプション）VOICEVOX API（自動音声長計算用）
- （オプション）ffprobe（自動動画長計算用）

## 基本的な使い方

### 1. 基本的な変換

```bash
ymmps-authorizer --template-file template.ymmp scenario.ymmps
```

### 2. 出力先を指定

```bash
ymmps-authorizer --template-file template.ymmp --output output.ymmp scenario.ymmps
```

### 3. バリデーションのみ実行

```bash
ymmps-authorizer --validate-only scenario.ymmps
```

### 4. ドライラン（変更せずに実行内容を確認）

```bash
ymmps-authorizer --dry-run --template-file template.ymmp scenario.ymmps
```

## オプション

| オプション | 説明 | 例 |
|-----------|------|-----|
| `--template-file` | テンプレートYMMPファイルのパス（必須） | `--template-file template.ymmp` |
| `--output`, `-o` | 出力YMMPファイルのパス | `--output result.ymmp` |
| `--validate-only` | バリデーションのみ実行（変換は行わない） | `--validate-only` |
| `--dry-run` | ドライラン（実際の変換は行わない） | `--dry-run` |
| `--verbose` | 詳細なログを出力 | `--verbose` |
| `--base-path` | 相対パス解決の基準パス | `--base-path /project/base` |

## YMMPS フォーマット

YMMPS ファイルは YAML 形式で記述します。

### 基本構造

```yaml
YMMPSVersion: "1"
Sequences:
  - ID: "sequence1"  # オプション
    Scenes:
      - ID: "scene1"  # オプション
        Shots:
          - ID: "shot1"  # オプション
            Items:
              - _Template: "voice_template"
                Length: "300"
                Serif: "こんにちは"
              - _Template: "video_template"
                Length: "_until:SCENE_END"
                FilePath: "videos/intro.mp4"
```

### 階層構造

- **Sequence**: 一連のシーンをまとめる最上位の単位
- **Scene**: 複数のショットから構成される場面
- **Shot**: 同時に開始する複数のアイテムをまとめる単位
- **Item**: 実際のタイムライン要素（音声、動画、立ち絵など）

## 長さ指定

### 固定長

```yaml
Length: "300"  # 300フレーム
```

### 相対的な長さ指定

```yaml
# 現在のシーケンス終了まで
Length: "_until:SEQUENCE_END"

# 現在のシーン終了まで
Length: "_until:SCENE_END"

# 現在のショット終了まで
Length: "_until:SHOT_END"
```

### ID指定による相対的な長さ

```yaml
# 特定のシーケンス終了まで
Length: "_until:SEQUENCE_END:sequence_id"

# 特定のシーン終了まで
Length: "_until:SCENE_END:scene_id"

# 特定のショット終了まで
Length: "_until:SHOT_END:shot_id"
```

### 自動長さ計算

```yaml
# VOICEVOX APIを使用した音声長計算
Length: "_auto:VOICEVOX"

# ffprobeを使用した動画長計算
Length: "_auto:VIDEO"
```

### 自然言語形式

```yaml
# 自然言語形式もサポート
Length: "UNTIL SCENE END"
Length: "UNTIL SHOT shot_id END"
```

## 使用例

### 例1: 基本的なシナリオ

```yaml
YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          # 導入音声
          - Items:
              - _Template: "voice_main"
                Length: "_auto:VOICEVOX"
                Serif: "今日は動画作成について説明します"
                
          # 説明動画
          - Items:
              - _Template: "video_main"
                Length: "_auto:VIDEO"
                FilePath: "videos/explanation.mp4"
              - _Template: "voice_main"
                Length: "_until:SHOT_END"
                Serif: "こちらの動画をご覧ください"
```

### 例2: 複雑なシーン構成

```yaml
YMMPSVersion: "1"
Sequences:
  - ID: "intro_sequence"
    Scenes:
      - ID: "opening"
        Shots:
          - ID: "title_shot"
            Items:
              - _Template: "title_template"
                Length: "180"
                Text: "チュートリアル動画"
                
      - ID: "main_content"
        Shots:
          - Items:
              - _Template: "background_video"
                Length: "_until:SCENE_END"
                FilePath: "backgrounds/tech_bg.mp4"
              - _Template: "voice_narrator"
                Length: "_auto:VOICEVOX"
                Serif: "このチュートリアルでは..."
```

### 例3: 相対パスの使用

ディレクトリ構造:
```
project/
├── scenario.ymmps
├── template.ymmp
├── assets/
│   ├── videos/
│   │   └── intro.mp4
│   └── audio/
│       └── bgm.mp3
```

```bash
# プロジェクトディレクトリを基準に相対パスを解決
ymmps-authorizer --template-file template.ymmp --base-path ./assets scenario.ymmps
```

```yaml
# scenario.ymmps内で相対パスを使用
YMMPSVersion: "1"
Sequences:
  - Scenes:
      - Shots:
          - Items:
              - _Template: "video_template"
                Length: "_auto:VIDEO"
                FilePath: "videos/intro.mp4"  # assets/videos/intro.mp4 として解決される
```

## トラブルシューティング

### よくあるエラー

#### 1. テンプレート要素が見つからない

```
Error: template validation failed:
template 'voice_main' not found (referenced at Sequence[0].Scene[0].Shot[0].Item[0])
```

**解決方法**: テンプレートYMMPファイル内のアイテムの`Remark`フィールドを確認してください。

#### 2. YAML構文エラー

```
Error: YAML syntax validation failed: yaml: line 5: found character that cannot start any token
```

**解決方法**: YAMLファイルのインデントや構文を確認してください。

#### 3. VOICEVOX APIが利用できない

```
Error: VOICEVOX API is not available at http://localhost:50021
```

**解決方法**: 
- VOICEVOXが起動していることを確認
- APIサーバーが正しいポート（50021）で動作していることを確認

#### 4. ffprobeが見つからない

```
Error: ffprobe is not available
```

**解決方法**:
- ffprobeがインストールされていることを確認
- PATHに ffprobe が含まれていることを確認

### デバッグ方法

#### 1. バリデーションで問題を特定

```bash
ymmps-authorizer --validate-only --verbose scenario.ymmps
```

#### 2. ドライランで変換内容を確認

```bash
ymmps-authorizer --dry-run --verbose --template-file template.ymmp scenario.ymmps
```

#### 3. 段階的な確認

```bash
# 1. まずバリデーション
ymmps-authorizer --validate-only scenario.ymmps

# 2. ドライランで確認
ymmps-authorizer --dry-run --template-file template.ymmp scenario.ymmps

# 3. 実際の変換
ymmps-authorizer --template-file template.ymmp scenario.ymmps
```

### パフォーマンス注意事項

- 大量のアイテム（数千個）がある場合、変換に時間がかかる場合があります
- VOICEVOX API呼び出しは音声アイテムの数に比例して時間がかかります
- ffprobe による動画長取得は動画ファイルのサイズに依存します

### サポートされるファイル形式

#### 動画ファイル
- MP4, AVI, MOV, WMV, FLV, MKV

#### 音声ファイル  
- MP3, WAV, OGG, M4A, AAC, FLAC

#### 画像ファイル
- PNG, JPG, JPEG, GIF, BMP, WEBP

---

## ライセンス

このプロジェクトのライセンスについては、LICENSEファイルを参照してください。

## 貢献

バグ報告や機能要求は、GitHubのIssuesページでお願いします。

## 更新履歴

### v1.0.0
- 初回リリース
- 全23のユーザーストーリーを実装
- VOICEVOX自動長さ計算対応
- ffprobe自動長さ計算対応
- 包括的な検証機能
- 相対パスサポート