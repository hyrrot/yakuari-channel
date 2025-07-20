# YMMP Scenario Converter 実装計画書

## 1. 概要

本ドキュメントは、YMMP Scenario Converterの実装計画を記述する。Kent BeckのTest Driven Development (TDD)手法に従い、段階的かつ反復的な開発を行う。

## 2. 開発方針

### 2.1 TDD原則
1. **Red**: 失敗するテストを書く
2. **Green**: テストを通す最小限の実装を行う
3. **Refactor**: コードをクリーンにする

### 2.2 実装の優先順位
1. コアとなるデータモデルと基本的な変換
2. 段階的に機能を追加
3. 外部依存（VOICEVOX API、ffprobe）は最後に統合

## 3. 実装フェーズ

### Phase 1: 基礎構築（基本的なデータモデルとI/O）

#### Step 1.1: プロジェクト構造の設定
```
ymmp-scenario-converter/
├── cmd/
│   └── ymmps-authorizer/
│       └── main.go
├── internal/
│   ├── models/
│   │   ├── ymmps.go
│   │   ├── ymmps_test.go
│   │   ├── ymmp.go
│   │   └── ymmp_test.go
│   ├── parser/
│   │   ├── ymmps_parser.go
│   │   ├── ymmps_parser_test.go
│   │   ├── ymmp_parser.go
│   │   └── ymmp_parser_test.go
│   ├── converter/
│   │   ├── converter.go
│   │   └── converter_test.go
│   └── cli/
│       ├── cli.go
│       └── cli_test.go
├── go.mod
├── go.sum
└── README.md
```

#### Step 1.2: 基本的なデータモデル
- [ ] YMMPSの基本構造体を定義
- [ ] YMMPの基本構造体を定義
- [ ] 単体テストで構造体の基本動作を確認

#### Step 1.3: パーサーの実装
- [ ] YMMPSパーサー（YAML読み込み）
- [ ] YMMPパーサー（JSON読み込み）
- [ ] 各パーサーの単体テスト

### Phase 2: 基本変換機能

#### Step 2.1: 最小限の変換機能
- [ ] 固定長のアイテムを配置する機能
- [ ] テンプレートからアイテムをコピーする機能
- [ ] Remarkによるテンプレート参照の解決

#### Step 2.2: 階層構造の処理
- [ ] Sequence → Scene → Shot → Item の階層処理
- [ ] 各階層でのフレーム位置計算
- [ ] 統合テストで階層処理を検証

#### Step 2.3: プロパティのマージ
- [ ] テンプレートアイテムへのプロパティ上書き
- [ ] 型安全なプロパティ処理

### Phase 3: 長さ計算機能

#### Step 3.1: 相対的な長さ指定
- [ ] `_until:SEQUENCE_END`の実装
- [ ] `_until:SCENE_END`の実装
- [ ] `_until:SHOT_END`の実装

#### Step 3.2: ID参照による長さ指定
- [ ] IDの収集と管理
- [ ] `_until:(SEQUENCE|SCENE|SHOT)_END:<ID>`の実装
- [ ] 循環参照の検出

#### Step 3.3: 自動長さ計算のスタブ
- [ ] `_auto:VOICEVOX`のインターフェース定義
- [ ] `_auto:VIDEO`のインターフェース定義
- [ ] モック実装でのテスト

### Phase 4: バリデーション機能

#### Step 4.1: スキーマバリデーション
- [ ] YMMPSバージョンチェック
- [ ] 必須フィールドの検証
- [ ] データ型の検証

#### Step 4.2: 参照バリデーション
- [ ] テンプレート要素の存在確認
- [ ] ID参照の妥当性確認
- [ ] エラーメッセージの詳細化

### Phase 5: パス処理

#### Step 5.1: 相対パスの解決
- [ ] YMMPSファイル基準の相対パス解決
- [ ] カスタムベースパスのサポート

#### Step 5.2: 出力パスの生成
- [ ] FilePath要素の更新
- [ ] 出力ファイル名の生成ロジック

### Phase 6: CLI実装

#### Step 6.1: 基本的なCLI
- [ ] コマンドライン引数のパース
- [ ] ヘルプメッセージ
- [ ] エラーハンドリング

#### Step 6.2: オプション機能
- [ ] `--validate-only`の実装
- [ ] `--dry-run`の実装
- [ ] `--verbose`の実装

### Phase 7: 外部依存の統合

#### Step 7.1: VOICEVOX統合
- [ ] VOICEVOX APIクライアントの実装
- [ ] 音声長の計算
- [ ] エラーハンドリングとリトライ

#### Step 7.2: ffprobe統合
- [ ] ffprobeコマンドの実行
- [ ] 動画情報の解析
- [ ] エラーハンドリング

### Phase 8: 最適化と品質向上

#### Step 8.1: パフォーマンス最適化
- [ ] 大規模プロジェクトでのプロファイリング
- [ ] ボトルネックの特定と改善

#### Step 8.2: エラーメッセージの改善
- [ ] ユーザーフレンドリーなメッセージ
- [ ] デバッグ情報の追加

#### Step 8.3: ドキュメンテーション
- [ ] README.mdの作成
- [ ] 使用例の追加
- [ ] APIドキュメントの生成

## 4. 各フェーズのテスト戦略

### Phase 1-2: 単体テスト中心
```go
// 例: models/ymmps_test.go
func TestSequence_Validate(t *testing.T) {
    tests := []struct {
        name    string
        seq     Sequence
        wantErr bool
    }{
        {
            name: "valid sequence with ID",
            seq: Sequence{
                ID: "seq1",
                Scenes: []Scene{{ID: "scene1"}},
            },
            wantErr: false,
        },
        {
            name: "empty sequence",
            seq: Sequence{},
            wantErr: true,
        },
    }
    // ...
}
```

### Phase 3-5: 統合テスト追加
```go
// 例: converter/converter_test.go
func TestConverter_ConvertSimpleProject(t *testing.T) {
    // テンプレートとYMMPSファイルを用意
    template := loadTestTemplate(t, "testdata/template.ymmp")
    ymmps := loadTestYMMPS(t, "testdata/simple.ymmps")
    
    converter := NewConverter()
    result, err := converter.Convert(template, ymmps)
    
    require.NoError(t, err)
    assert.Equal(t, 1, len(result.Timelines[0].Items))
    // ...
}
```

### Phase 6-7: E2Eテスト
```go
// 例: e2e/e2e_test.go
func TestE2E_BasicConversion(t *testing.T) {
    // 実際のファイルを使用したエンドツーエンドテスト
    cmd := exec.Command("go", "run", "./cmd/ymmps-authorizer",
        "--template-file", "testdata/template.ymmp",
        "testdata/scenario.ymmps")
    
    output, err := cmd.CombinedOutput()
    require.NoError(t, err)
    
    // 出力ファイルの検証
    // ...
}
```

## 5. 実装順序の根拠

1. **基礎から積み上げる**: データモデル → パーサー → コンバーター
2. **外部依存を遅らせる**: モックで開発を進め、最後に統合
3. **ユーザー価値の早期提供**: 基本機能を先に実装
4. **リスクの早期発見**: 複雑な機能（ID参照、自動長さ）を中盤で実装

## 6. マイルストーン

### M1: MVP（Phase 1-2完了）
- 基本的な変換が動作
- 固定長のアイテム配置が可能

### M2: 実用版（Phase 3-5完了）
- 相対長さ指定が動作
- バリデーション機能完備
- CLI利用可能

### M3: 完全版（Phase 6-8完了）
- VOICEVOX/ffprobe統合
- パフォーマンス最適化済み
- ドキュメント完備

## 7. リスクと対策

### リスク1: YMMP形式の仕様変更
- 対策: インターフェースを使用した疎結合設計

### リスク2: 外部API（VOICEVOX）の不安定性
- 対策: リトライ機構、キャッシュ、オフラインモード

### リスク3: 大規模プロジェクトでのパフォーマンス
- 対策: 早期のプロファイリング、必要に応じた並列化

## 8. 開発環境とツール

### 必須ツール
- Go 1.21以上
- golangci-lint（静的解析）
- go test（テスト実行）

### 推奨ツール
- mockgen（モック生成）
- go-task（タスクランナー）
- delve（デバッガー）

### 外部依存
- gopkg.in/yaml.v3（YAML解析）
- github.com/stretchr/testify（テストアサーション）
- github.com/spf13/cobra（CLIフレームワーク）