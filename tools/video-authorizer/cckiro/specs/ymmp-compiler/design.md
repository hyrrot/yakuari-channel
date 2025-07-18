# YMMP Compiler 設計書

## 1. システムアーキテクチャ

### 1.1 全体構成

```
┌─────────────────────────────────────────────────────────────────┐
│                          CLI Interface                           │
├─────────────────────────────────────────────────────────────────┤
│                        Core Components                           │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────────┐ │
│  │   Parser    │  │  Validator   │  │    Converter           │ │
│  │  (YAML)     │  │              │  │  (SYMMP → YMMP)        │ │
│  └─────────────┘  └──────────────┘  └────────────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                         Plugin System                            │
│  ┌────────────────┐  ┌────────────────┐  ┌─────────────────┐   │
│  │  Item Types   │  │ Voice Services │  │ Length Providers│   │
│  │  - voice      │  │  - VOICEVOX    │  │  - ffprobe      │   │
│  │  - video      │  │  - AquesTalk   │  │  - auto         │   │
│  │  - image      │  │  - Amazon Polly│  │                 │   │
│  │  - tachie     │  │                │  │                 │   │
│  │  - audio      │  │                │  │                 │   │
│  └────────────────┘  └────────────────┘  └─────────────────┘   │
├─────────────────────────────────────────────────────────────────┤
│                      Support Components                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │ File Manager │  │ Cache Manager│  │  Config Manager      │  │
│  │              │  │              │  │                      │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 ディレクトリ構造

```
ymmp-compiler/
├── cmd/
│   └── ymmp-compiler/
│       └── main.go              # エントリーポイント
├── internal/
│   ├── cli/                     # CLIインターフェース
│   │   ├── command.go
│   │   └── flags.go
│   ├── core/                    # コアコンポーネント
│   │   ├── parser/
│   │   │   └── symmp_parser.go
│   │   ├── validator/
│   │   │   └── validator.go
│   │   └── converter/
│   │       ├── converter.go
│   │       └── timeline_calculator.go
│   ├── models/                  # データモデル
│   │   ├── symmp.go
│   │   └── ymmp.go
│   ├── plugins/                 # プラグインシステム
│   │   ├── interfaces.go
│   │   ├── items/
│   │   │   ├── voice.go
│   │   │   ├── video.go
│   │   │   ├── image.go
│   │   │   ├── tachie.go
│   │   │   └── audio.go
│   │   ├── voice_services/
│   │   │   ├── interface.go
│   │   │   └── voicevox.go
│   │   └── length_providers/
│   │       ├── interface.go
│   │       └── ffprobe.go
│   └── support/                 # サポートコンポーネント
│       ├── cache/
│       │   └── cache_manager.go
│       ├── config/
│       │   └── config_manager.go
│       └── file/
│           └── file_manager.go
├── pkg/                         # 外部パッケージ
│   └── errors/
│       └── errors.go
├── defaults/                    # デフォルト設定
│   ├── voice.yaml
│   ├── video.yaml
│   ├── image.yaml
│   ├── tachie.yaml
│   └── audio.yaml
└── tests/                       # テストコード
    ├── unit/
    └── integration/
```

## 2. データモデル設計

### 2.1 SYMMP データモデル

```go
// Episode represents the entire project
type Episode struct {
    SYMMPFormatVersion string                 `yaml:"symmp_format_version"`
    Project           ProjectConfig           `yaml:"project"`
    Defaults          map[string]interface{}  `yaml:"defaults,omitempty"`
    Sequences         []Sequence              `yaml:"sequences"`
}

// ProjectConfig contains project settings
type ProjectConfig struct {
    Name      string `yaml:"name"`
    OutputDir string `yaml:"output_dir"`
}

// Sequence represents a sequence of scenes
type Sequence struct {
    ID     string  `yaml:"id"`
    Scenes []Scene `yaml:"scenes"`
}

// Scene represents a scene containing shots
type Scene struct {
    ID    string `yaml:"id"`
    Shots []Shot `yaml:"shots"`
}

// Shot represents a shot containing items
type Shot struct {
    Items []Item `yaml:"-"` // Will be unmarshaled dynamically
}

// Item is the base interface for all item types
type Item interface {
    GetType() string
    GetLength() LengthSpec
    SetCalculatedLength(seconds float64)
    GetStartTime() float64
    SetStartTime(seconds float64)
}

// LengthSpec represents different types of length specifications
type LengthSpec struct {
    Type  LengthType
    Value float64 // Used for numeric values
}

type LengthType string

const (
    LengthTypeNumeric      LengthType = "numeric"
    LengthTypeAuto         LengthType = "auto"
    LengthTypeUntilShotEnd LengthType = "until:SHOT_END"
    LengthTypeUntilSceneEnd LengthType = "until:SCENE_END"
    LengthTypeUntilSequenceEnd LengthType = "until:SEQUENCE_END"
)
```

### 2.2 YMMP データモデル

```go
// YMMPProject represents the YMMP file structure
type YMMPProject struct {
    Timeline Timeline `json:"Timeline"`
    // Other YMMP fields...
}

// Timeline contains all items in the timeline
type Timeline struct {
    Items []YMMPItem `json:"Items"`
}

// YMMPItem represents an item in YMMP format
type YMMPItem struct {
    Type      string                 `json:"$type"`
    Layer     int                   `json:"Layer"`
    Frame     int                   `json:"Frame"`
    Length    int                   `json:"Length"`
    FilePath  string               `json:"FilePath,omitempty"`
    // Additional fields based on item type
    Extended  map[string]interface{} `json:",inline"`
}
```

## 3. インターフェース設計

### 3.1 プラグインインターフェース

```go
// ItemPlugin defines the interface for item type plugins
type ItemPlugin interface {
    GetType() string
    ParseYAML(data interface{}) (Item, error)
    ConvertToYMMP(item Item, basePath string) (*YMMPItem, error)
}

// VoiceService defines the interface for voice synthesis services
type VoiceService interface {
    GetName() string
    CalculateLength(text string, params map[string]interface{}) (float64, error)
    Configure(config map[string]interface{}) error
}

// LengthProvider defines the interface for length calculation
type LengthProvider interface {
    CanCalculate(item Item) bool
    CalculateLength(item Item, basePath string) (float64, error)
}

// PluginRegistry manages all plugins
type PluginRegistry struct {
    itemPlugins     map[string]ItemPlugin
    voiceServices   map[string]VoiceService
    lengthProviders []LengthProvider
}
```

### 3.2 キャッシュインターフェース

```go
// CacheManager handles caching of calculated values
type CacheManager interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}) error
    GenerateKey(service, text string, params map[string]interface{}) string
}
```

## 4. 処理フロー設計

### 4.1 メイン処理フロー

```
1. コマンドライン引数の解析
2. 設定の読み込み（defaults-dir, base-dir等）
3. SYMMPファイルの読み込みとパース
4. バリデーション
   - YAMLスキーマ検証
   - 必須フィールド検証
   - ID参照検証
5. 長さの計算
   - "auto"タイプの解決（キャッシュ利用）
   - 相対長さの解決（until:*）
6. タイムライン計算
   - 開始時刻の計算
   - 依存関係の解決
7. YMMP形式への変換
   - 相対パスから絶対パスへの変換
   - YMMP構造の生成
8. ファイル出力
```

### 4.2 長さ計算アルゴリズム

```go
func CalculateTimeline(episode *Episode) error {
    // Phase 1: Calculate auto lengths
    for _, seq := range episode.Sequences {
        for _, scene := range seq.Scenes {
            for _, shot := range scene.Shots {
                for _, item := range shot.Items {
                    if item.GetLength().Type == LengthTypeAuto {
                        length := calculateAutoLength(item)
                        item.SetCalculatedLength(length)
                    }
                }
            }
        }
    }
    
    // Phase 2: Resolve relative lengths (bottom-up)
    // Shot level
    resolveUntilShotEnd(episode)
    // Scene level
    resolveUntilSceneEnd(episode)
    // Sequence level
    resolveUntilSequenceEnd(episode)
    
    // Phase 3: Calculate start times
    calculateStartTimes(episode)
    
    return nil
}
```

## 5. エラーハンドリング設計

### 5.1 エラー型定義

```go
type ErrorType string

const (
    ErrorTypeValidation   ErrorType = "VALIDATION"
    ErrorTypeFileNotFound ErrorType = "FILE_NOT_FOUND"
    ErrorTypeAPIError     ErrorType = "API_ERROR"
    ErrorTypeCyclicRef    ErrorType = "CYCLIC_REFERENCE"
)

type CompilerError struct {
    Type     ErrorType
    Message  string
    Location string // File path and line number
    Cause    error
}
```

### 5.2 エラーメッセージフォーマット

```
Error: [TYPE] message
Location: file.yaml:25
Details: specific error details
Suggestion: how to fix the error
```

## 6. テスト戦略

### 6.1 Test Driven Development アプローチ

1. **Red Phase**: 失敗するテストを書く
2. **Green Phase**: テストを通す最小限のコードを書く
3. **Refactor Phase**: コードを改善する

### 6.2 テストカテゴリ

- **単体テスト**: 各コンポーネントの独立したテスト
- **統合テスト**: コンポーネント間の連携テスト
- **E2Eテスト**: 実際のSYMMPファイルを使った変換テスト

### 6.3 テストファイル構造

```
tests/
├── unit/
│   ├── parser_test.go
│   ├── validator_test.go
│   ├── converter_test.go
│   └── plugins/
│       ├── voice_test.go
│       └── ...
├── integration/
│   ├── timeline_test.go
│   └── cache_test.go
└── e2e/
    ├── fixtures/
    │   ├── simple.symmp
    │   └── complex.symmp
    └── e2e_test.go
```

## 7. 設定管理設計

### 7.1 設定の優先順位

1. コマンドライン引数
2. 環境変数
3. SYMMPファイル内の指定
4. defaults/ディレクトリの設定ファイル
5. ハードコードされたデフォルト値

### 7.2 設定ファイル例

```yaml
# defaults/voice.yaml
voice:
  character_name: "ずんだもん"
  voice_id: 1
  speed: 1.0
  pitch: 0.0
  volume: 1.0
  service: "voicevox"
  
# Service-specific settings
voicevox:
  endpoint: "http://localhost:50021"
  timeout: 30
  retry_count: 3
  retry_interval: 5
```

## 8. 拡張性の考慮

### 8.1 新しいアイテムタイプの追加

1. `ItemPlugin`インターフェースを実装
2. プラグインレジストリに登録
3. デフォルト設定ファイルを追加

### 8.2 新しい音声合成サービスの追加

1. `VoiceService`インターフェースを実装
2. サービス固有の設定を定義
3. プラグインレジストリに登録

### 8.3 新しい長さ計算プロバイダーの追加

1. `LengthProvider`インターフェースを実装
2. 対応するアイテムタイプを定義
3. プラグインレジストリに登録