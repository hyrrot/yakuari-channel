# Video Authorizer

YukkuriMovieMaker向けのシナリオベースのビデオプロジェクト生成ツール

## 概要

Video Authorizerは、YMMPS（YukkuriMovieMaker Project Scenario）形式のシナリオファイルから、YMMP（YukkuriMovieMaker Project）形式のプロジェクトファイルを生成するツールです。シナリオをYAML形式で記述することで、動画プロジェクトを効率的に作成できます。

## システム構成

```
tools/video-authorizer/
├── internal/
│   ├── converter/      # YMMPS → YMMP変換ロジック
│   ├── parser/         # YMMP/YMMPSファイルパーサー
│   └── models/         # データモデル定義
├── cmd/
│   └── ymmps-authorizer/  # CLIエントリポイント
└── work/               # 作業ディレクトリ
```

## 主要機能

### 1. シナリオベースの動画構成
- **階層構造**: シーケンス → シーン → ショット → アイテム
- **テンプレート参照**: 事前定義されたアイテムテンプレートを使用
- **プロパティオーバーライド**: テンプレートのプロパティを個別に上書き可能

### 2. 柔軟な長さ指定
- **数値指定**: フレーム数での直接指定（例: `"120"`）
- **相対指定**: 自然言語での長さ指定
  - `"UNTIL SHOT END"` - 現在のショット終了まで
  - `"UNTIL SCENE END"` - 現在のシーン終了まで
  - `"UNTIL SEQUENCE END"` - 現在のシーケンス終了まで
  - `"UNTIL SHOT shot1 END"` - 特定IDのショット終了まで
- **自動計算**: 音声やビデオの長さに基づく自動設定（今後実装予定）

### 3. サポートするアイテムタイプ
- **VoiceItem**: 音声・セリフアイテム
- **VideoItem**: ビデオクリップ
- **TachieItem**: 立ち絵・キャラクター画像

## ファイル形式

### YMMPS（シナリオファイル）
```yaml
YMMPSVersion: "1"
Sequences:
  - ID: sequence1
    Scenes:
      - ID: scene1
        Shots:
          - ID: shot1
            Items:
              - Type: VoiceItem
                Template: voice_template_1
                Length: "120"  # フレーム数
                Properties:
                  Serif: "こんにちは"
```

### YMMP（プロジェクトファイル）
```json
{
  "FilePath": "project.ymmp",
  "Timelines": [
    {
      "ID": "timeline1",
      "Items": [
        {
          "$type": "YukkuriMovieMaker.Project.VoiceItem",
          "Serif": "こんにちは",
          "Frame": 0,
          "Length": 120,
          "Layer": 0
        }
      ]
    }
  ]
}
```

## 使用方法

### インストール

#### Linux/macOS
```bash
cd tools/video-authorizer
go build -o ymmps-authorizer ./cmd/ymmps-authorizer
```

#### Windows (PowerShell)
```powershell
cd tools/video-authorizer
go build -o ymmps-authorizer.exe ./cmd/ymmps-authorizer
```

### 基本的な使用方法

#### Linux/macOS
```bash
# YMMPSファイルからYMMPを生成
./ymmps-authorizer convert scenario.ymmps template.ymmp -o output.ymmp

# ヘルプの表示
./ymmps-authorizer --help

# バージョン情報
./ymmps-authorizer version

# 変換コマンドの詳細ヘルプ
./ymmps-authorizer convert --help
```

#### Windows (PowerShell)
```powershell
# YMMPSファイルからYMMPを生成
.\ymmps-authorizer.exe convert scenario.ymmps template.ymmp -o output.ymmp

# ヘルプの表示
.\ymmps-authorizer.exe --help

# バージョン情報
.\ymmps-authorizer.exe version

# 変換コマンドの詳細ヘルプ
.\ymmps-authorizer.exe convert --help
```

### 使用例

#### Linux/macOS
```bash
# サンプルファイルでの変換
./ymmps-authorizer convert tests/fixtures/simple_scenario.ymmps tests/fixtures/simple_template.ymmp -o output.ymmp
```

#### Windows (PowerShell)
```powershell
# サンプルファイルでの変換
.\ymmps-authorizer.exe convert tests/fixtures/simple_scenario.ymmps tests/fixtures/simple_template.ymmp -o output.ymmp
```

## 実装状況

- ✅ コアコンバーター実装
- ✅ YMMP/YMMPSパーサー実装
- ✅ 基本的な長さ計算
- ✅ 相対長さ計算（完全実装）
- ✅ CLIコマンド実装
- ✅ 包括的なテストスイート
- ✅ プロパティオーバーライド機能
- ✅ エラーハンドリングと検証

## テスト

全ての機能は包括的にテストされています：

#### Linux/macOS/Windows共通
```bash
# 全テストの実行
go test ./... -v

# 統合テストのみ実行
go test ./tests/integration/... -v

# 特定パッケージのテスト
go test ./internal/converter -v
```

## 今後の予定

1. 音声・動画の自動長さ計算機能
2. より多くのアイテムタイプのサポート
3. パフォーマンス最適化
4. GUIツールの開発検討

## ライセンス

[ライセンス情報を追加してください]