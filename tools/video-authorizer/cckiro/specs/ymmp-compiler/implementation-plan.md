# YMMP Compiler 実装計画書

## 1. 実装方針

### 1.1 Test Driven Development (TDD) アプローチ

Kent BeckのTDD手法に従い、以下のサイクルで開発を進める：

1. **Red**: 失敗するテストを書く
2. **Green**: テストを通す最小限のコードを実装
3. **Refactor**: コードの品質を改善

### 1.2 実装の優先順位

1. 基本的なデータモデルとパーサー
2. バリデーション機能
3. 基本的な変換機能（長さ計算なし）
4. 長さ計算機能
5. プラグインシステム
6. 外部連携（VOICEVOX、ffprobe）
7. キャッシュ機能
8. エラーハンドリングの改善

## 2. 実装ステップ

### Phase 1: プロジェクトセットアップ（Day 1）

#### Step 1.1: Go モジュールの初期化
```bash
go mod init github.com/user/ymmp-compiler
```

#### Step 1.2: 基本的なディレクトリ構造の作成
```bash
mkdir -p cmd/ymmp-compiler
mkdir -p internal/{cli,core,models,plugins,support}
mkdir -p pkg/errors
mkdir -p tests/{unit,integration,e2e}
mkdir -p defaults
```

#### Step 1.3: 最初のテストとmain関数
- `cmd/ymmp-compiler/main_test.go`: CLIの基本的な動作テスト
- `cmd/ymmp-compiler/main.go`: 最小限のmain関数

### Phase 2: データモデルとパーサー（Day 2-3）

#### Step 2.1: SYMMPデータモデル
**テストファイル**: `internal/models/symmp_test.go`
```go
// Test: Episode構造体の基本的なフィールド
// Test: Sequence、Scene、Shotの階層構造
// Test: LengthSpecの各種タイプ
```

**実装ファイル**: `internal/models/symmp.go`
- Episode, Sequence, Scene, Shot構造体
- Item インターフェース
- LengthSpec構造体

#### Step 2.2: YAMLパーサー
**テストファイル**: `internal/core/parser/symmp_parser_test.go`
```go
// Test: 最小限のSYMMPファイルのパース
// Test: 複数のsequenceを持つファイルのパース
// Test: 不正なYAMLのエラーハンドリング
```

**実装ファイル**: `internal/core/parser/symmp_parser.go`
- ParseSYMMP関数
- 基本的なYAML構造の読み込み

### Phase 3: バリデーション（Day 4-5）

#### Step 3.1: スキーマバリデーション
**テストファイル**: `internal/core/validator/validator_test.go`
```go
// Test: 必須フィールドの検証
// Test: ID参照の検証
// Test: 無効なYAML構造の検証
```

**実装ファイル**: `internal/core/validator/validator.go`
- Validate関数
- エラーメッセージの生成

### Phase 4: 基本的な変換機能（Day 6-7）

#### Step 4.1: YMMPデータモデル
**テストファイル**: `internal/models/ymmp_test.go`
```go
// Test: YMMPProject構造体
// Test: Timeline構造体
// Test: YMMPItem構造体
```

**実装ファイル**: `internal/models/ymmp.go`

#### Step 4.2: 基本的な変換
**テストファイル**: `internal/core/converter/converter_test.go`
```go
// Test: 単一のitemの変換
// Test: 複数itemの変換
// Test: 相対パスから絶対パスへの変換
```

**実装ファイル**: `internal/core/converter/converter.go`
- Convert関数（長さ計算なしバージョン）

### Phase 5: プラグインシステム（Day 8-10）

#### Step 5.1: プラグインインターフェース
**テストファイル**: `internal/plugins/interfaces_test.go`
```go
// Test: ItemPluginインターフェースのモック実装
// Test: プラグインレジストリの動作
```

**実装ファイル**: `internal/plugins/interfaces.go`
- ItemPlugin, VoiceService, LengthProviderインターフェース
- PluginRegistry構造体

#### Step 5.2: 基本的なアイテムプラグイン
**各アイテムタイプごとに**:
- テストファイル: `internal/plugins/items/[type]_test.go`
- 実装ファイル: `internal/plugins/items/[type].go`

実装順序:
1. image（最もシンプル）
2. tachie
3. audio
4. video
5. voice（最も複雑）

### Phase 6: 長さ計算機能（Day 11-13）

#### Step 6.1: タイムライン計算
**テストファイル**: `internal/core/converter/timeline_calculator_test.go`
```go
// Test: 数値指定の長さ
// Test: until:SHOT_ENDの解決
// Test: until:SCENE_ENDの解決
// Test: until:SEQUENCE_ENDの解決
// Test: 開始時刻の計算
```

**実装ファイル**: `internal/core/converter/timeline_calculator.go`

#### Step 6.2: 自動長さ計算（モック）
**テストファイル**: `internal/plugins/length_providers/auto_test.go`
```go
// Test: voiceアイテムの長さ計算（モック）
// Test: video/audioアイテムの長さ計算（モック）
```

### Phase 7: 外部連携（Day 14-16）

#### Step 7.1: VOICEVOX連携
**テストファイル**: `internal/plugins/voice_services/voicevox_test.go`
```go
// Test: API呼び出しのモック
// Test: タイムアウト処理
// Test: リトライ処理
```

**実装ファイル**: `internal/plugins/voice_services/voicevox.go`

#### Step 7.2: ffprobe連携
**テストファイル**: `internal/plugins/length_providers/ffprobe_test.go`
```go
// Test: ffprobeコマンドの実行（モック）
// Test: 動画/音声ファイルの長さ取得
```

**実装ファイル**: `internal/plugins/length_providers/ffprobe.go`

### Phase 8: サポート機能（Day 17-18）

#### Step 8.1: キャッシュマネージャー
**テストファイル**: `internal/support/cache/cache_manager_test.go`
```go
// Test: キャッシュの保存と取得
// Test: キャッシュキーの生成
```

**実装ファイル**: `internal/support/cache/cache_manager.go`

#### Step 8.2: 設定マネージャー
**テストファイル**: `internal/support/config/config_manager_test.go`
```go
// Test: デフォルト値の読み込み
// Test: 設定の優先順位
```

**実装ファイル**: `internal/support/config/config_manager.go`

### Phase 9: CLIインターフェース（Day 19-20）

#### Step 9.1: コマンドライン引数処理
**テストファイル**: `internal/cli/command_test.go`
```go
// Test: 必須引数の検証
// Test: オプション引数の処理
// Test: ヘルプメッセージ
```

**実装ファイル**: `internal/cli/command.go`

### Phase 10: 統合とE2Eテスト（Day 21-22）

#### Step 10.1: 統合テスト
- 各コンポーネント間の連携テスト
- エンドツーエンドのワークフロー

#### Step 10.2: E2Eテスト
- 実際のSYMMPファイルを使った変換テスト
- エラーケースのテスト

## 3. テストデータの準備

### 3.1 テスト用SYMMPファイル

```yaml
# tests/fixtures/minimal.symmp
symmp_format_version: 0.1
project:
  name: "Minimal Test"
sequences:
  - id: seq1
    scenes:
      - id: scene1
        shots:
          - image:
              file_path: "./test.png"
              length: 5.0
```

### 3.2 モックデータ

- VOICEVOX APIレスポンス
- ffprobe出力
- 各種エラーケース

## 4. CI/CD設定

### 4.1 GitHub Actions

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test ./...
      - run: go build ./cmd/ymmp-compiler
```

## 5. 品質基準

### 5.1 テストカバレッジ
- 単体テスト: 80%以上
- 統合テスト: 主要なワークフローをカバー

### 5.2 コード品質
- `go fmt`でフォーマット
- `go vet`でエラーチェック
- `golangci-lint`で静的解析

### 5.3 ドキュメント
- 各パッケージにGoDoc
- README.mdの作成
- 使用例の提供

## 6. リスクと対策

### 6.1 技術的リスク
- **リスク**: YMMPフォーマットの仕様が不明確
- **対策**: 実際のYMMPファイルを解析し、リバースエンジニアリング

### 6.2 スケジュールリスク
- **リスク**: 外部連携の実装に時間がかかる
- **対策**: モックを使った開発を先行し、実装を後回しにする

## 7. マイルストーン

1. **Week 1**: 基本機能の実装（Phase 1-4）
2. **Week 2**: プラグインシステムと長さ計算（Phase 5-6）
3. **Week 3**: 外部連携とサポート機能（Phase 7-8）
4. **Week 4**: CLI完成と品質向上（Phase 9-10）

## 8. 次のアクション

1. Goプロジェクトの初期化
2. 最初のテスト（main関数の存在確認）を書く
3. テストを通す最小限のコードを実装
4. 次のテストに進む