# YMMPS Frame/Length Combination Test Cases Summary

このドキュメントは、YMMPSファイルのSequence情報と各アイテムのFrame, Lengthの値の組み合わせについて生成したテストケースの概要です。

## 更新履歴
- 2025-07-21: `_auto:VOICEVOX`の実装をモック化し、テストの一貫性と信頼性を向上
- 2025-07-21: `_auto:VIDEO`の実装もモック化し、FFProbe依存を排除

## 生成したテストファイル

### 1. `comprehensive_frame_length_test.go`
**基本的なFrame/Lengthの組み合わせテスト**

#### テストケース一覧：
- **single_sequence_fixed_lengths**: 単一シーケンス、固定長のみ
- **sequential_scenes_with_fixed_lengths**: 順次実行されるシーン、固定長
- **sequential_sequences**: 複数シーケンスの順次実行
- **until_scene_end_calculation**: `_until:SCENE_END`の正確な計算
- **until_sequence_end_three_pass**: `_until:SEQUENCE_END`の3パス処理テスト
- **complex_multi_sequence_scenario**: 複数シーケンス、複数シーン、異なるLength指定の組み合わせ
- **parallel_items_in_shot**: 同一ショット内の並列アイテム
- **nested_relative_references**: 入れ子になった相対参照

#### カバーする機能：
- 固定長の計算
- シーケンス間の順次実行
- 相対長参照（`_until:SCENE_END`, `_until:SEQUENCE_END`）
- 3パス処理システム
- 同一ショット内のアイテム配置

### 2. `voicevox_auto_length_test.go`
**VOICEVOX自動長さ計算のテスト**

#### テストケース一覧：
- **simple_auto_voicevox**: シンプルなVOICEVOX自動長さ計算
- **auto_voicevox_with_fallback**: VOICEVOX使用不可時のフォールバック
- **auto_voicevox_with_sequence_end**: `_auto:VOICEVOX`と`_until:SEQUENCE_END`の組み合わせ
- **multiple_auto_voicevox_sequential**: 複数の`_auto:VOICEVOX`音声の順次実行
- **complex_multi_type_scenario**: 複数アイテムタイプと`_auto:VOICEVOX`の複雑な組み合わせ
- **auto_voicevox_with_mixed_lengths**: `_auto:VOICEVOX`と固定長、相対長の混在

#### カバーする機能：
- VOICEVOX API連携（モック使用）
- フォールバック計算（文字数ベース）
- `_auto:VOICEVOX`と相対長の組み合わせ
- 複数音声の順次配置

#### モック実装：
- `MockVoicevoxClient`を使用し、テキストごとに予測可能な音声長を設定
- 実際のVOICEVOX APIへの依存を排除し、CI/CD環境での安定性を確保

### 3. `frame_length_error_cases_test.go`
**エラーケースとエッジケース**

#### エラーケース：
- **invalid_length_reference**: 存在しないIDへの参照
- **circular_reference**: 循環参照の検出
- **negative_length_handling**: 負の長さの処理
- **invalid_template_reference**: 存在しないテンプレートの参照
- **malformed_length_syntax**: 不正なLength構文
- **auto_voicevox_without_serif**: `_auto:VOICEVOX`でSerifが空（モック使用）

#### エッジケース：
- **zero_length_handling**: 長さ0のアイテムの処理
- **empty_sequence/scene/shot**: 空の要素の処理
- **very_large_length**: 非常に大きな長さ値
- **unicode_in_serif**: Unicode文字を含むSerif
- **deeply_nested_structure**: 深くネストした構造
- **single_frame_items**: 1フレームのアイテム
- **max_frame_precision**: フレーム精度の限界テスト

#### モック実装の利点：
- 空のSerifに対する一貫したフォールバック動作のテスト
- ネットワーク環境に依存しないエラーケーステスト

### 4. `real_world_scenarios_test.go`
**実際のユースケースベーステスト**

#### 実世界シナリオ：
- **gyakuten_saiban_style_video**: 逆転裁判風動画のシナリオ
  - オープニング、メイン部分の構造
  - BGM、背景、立ち絵、音声の組み合わせ
  - `_until:SEQUENCE_END`を使った背景要素の管理

- **tutorial_video_structure**: チュートリアル動画の典型的な構造
  - イントロ、コンテンツ、アウトロの3部構成
  - ステップごとのシーン分割
  - スクリーンショット画像と音声の同期

- **news_report_style**: ニュース報道風の動画構造
  - オープニングジングル
  - ヘッドライン、ストーリー部分の構造
  - 複数話者（ずんだもん、四国めたん）の使い分け

#### モック実装により実現：
- 長い日本語テキストに対する一貫した音声長計算
- 複雑なシナリオでの予測可能なフレーム位置
- 実際のVOICEVOX環境に依存しない安定したテスト実行

### 5. `auto_video_length_test.go`
**VIDEO自動長さ計算のテスト**

#### テストケース一覧：
- **simple_auto_video**: シンプルな`_auto:VIDEO`自動長さ計算
- **multiple_auto_video_sequential**: 複数の`_auto:VIDEO`動画の順次実行
- **mixed_auto_video_and_voice**: `_auto:VIDEO`と`_auto:VOICEVOX`の混在
- **auto_video_with_sequence_end**: `_auto:VIDEO`と`_until:SEQUENCE_END`の組み合わせ
- **different_video_types**: 異なる動画ファイルタイプの処理
- **auto_video_fallback_calculation**: 不明なファイルに対するフォールバック計算

#### カバーする機能：
- FFProbe連携（モック使用）
- ファイル名パターンベースの動画長推定
- `_auto:VIDEO`と相対長の組み合わせ
- 複数動画の順次配置

#### モック実装：
- `MockFFProbeClient`を使用し、ファイルパスごとに予測可能な動画長を設定
- パターンベースのフォールバック計算（short→5秒、long→30秒、intro→10秒など）
- 実際のFFProbeへの依存を排除し、CI/CD環境での安定性を確保

## テストで検証される要素

### Frame/Length計算の組み合わせ
1. **固定値**: `Length: "120"`
2. **VOICEVOX自動**: `Length: "_auto:VOICEVOX"`
3. **VIDEO自動**: `Length: "_auto:VIDEO"`
4. **相対参照**: 
   - `_until:SCENE_END`
   - `_until:SEQUENCE_END`
   - `_until:SHOT_END`

### Sequence構造のパターン
1. **単一Sequence**: 基本的な構造
2. **複数Sequence**: 順次実行
3. **複数Scene**: Sequence内での順次実行
4. **複数Shot**: Scene内での並列実行

### アイテムタイプの組み合わせ
- **VoiceItem**: 音声アイテム（`_auto:VOICEVOX`対応）
- **VideoItem**: 動画アイテム（`_auto:VIDEO`対応）
- **AudioItem**: BGM/効果音
- **ImageItem**: 画像/背景
- **TachieItem**: 立ち絵

### エラーハンドリング
- 無効な参照
- 循環参照
- 負の値
- 空の要素
- 制限値の境界

## 実行方法

```bash
# 全テストの実行
go test -v ./tests/integration/

# 個別テストファイルの実行
go test -v ./tests/integration/comprehensive_frame_length_test.go
go test -v ./tests/integration/voicevox_auto_length_test.go
go test -v ./tests/integration/auto_video_length_test.go
go test -v ./tests/integration/frame_length_error_cases_test.go
go test -v ./tests/integration/real_world_scenarios_test.go

# 特定のテストケースの実行
go test -v -run TestComprehensiveFrameLengthCombinations/until_sequence_end_three_pass
go test -v -run TestAutoVideoLengthCombinations/mixed_auto_video_and_voice
```

## 実装との対応

これらのテストケースは以下の実装機能を検証します：

1. **3パス処理システム**（`converter_advanced.go`）
   - 1パス目: 基本的な長さ計算
   - 2パス目: コンテキスト付き変換
   - 3パス目: `_until:SEQUENCE_END`の再計算

2. **VOICEVOX連携**（`voicevox_client.go`, `mock_voicevox_client.go`）
   - API呼び出し（本番環境）
   - モック実装（テスト環境）
   - フォールバック計算

3. **FFProbe連携**（`ffprobe_client.go`, `mock_ffprobe_client.go`）
   - 動画解析呼び出し（本番環境）
   - モック実装（テスト環境）
   - パターンベースフォールバック計算

4. **相対長さ計算**（`length_calculator.go`）
   - コンテキストベースの計算
   - 前処理パス

5. **BOM対応**（`ymmp_parser.go`）
   - UTF-8 BOM除去

## 注意事項

- 一部のテストケースは実装の仕様により期待値の調整が必要な場合があります
- **モック実装により、VOICEVOX・FFProbe環境への依存を排除済み**
- テストケースは継続的に更新し、新機能追加時には対応するテストを追加してください
- モックのテキスト→音声長マッピングは、実際のVOICEVOX API結果との整合性を定期的に確認することを推奨
- モックのファイルパス→動画長マッピングは、実際のFFProbe結果との整合性を定期的に確認することを推奨

## モック実装の技術詳細

### VoicevoxClientInterface
```go
type VoicevoxClientInterface interface {
    GetAudioDuration(text string, speaker int) (float64, error)
    IsAvailable() bool
    IsAvailableWithLogging(verbose bool) bool
}
```

### MockVoicevoxClient
- テキストごとの予測可能な音声長を提供
- SetDurationメソッドでテスト用の音声長を設定可能
- フォールバック計算（1文字=0.15秒）を実装

### FFProbeClientInterface
```go
type FFProbeClientInterface interface {
    GetVideoDuration(filePath string) (float64, error)
    GetVideoInfo(filePath string) (*VideoInfo, error)
    IsAvailable() bool
}
```

### MockFFProbeClient
- ファイルパスごとの予測可能な動画長を提供
- SetDurationメソッドでテスト用の動画長を設定可能
- パターンベースフォールバック計算を実装：
  - `short*` → 5.0秒
  - `long*` → 30.0秒  
  - `intro*` → 10.0秒
  - `outro*` → 8.0秒
  - `bgm*`/`music*` → 180.0秒
  - デフォルト → 15.0秒

## 今後の拡張

1. **パフォーマンステスト**: 大規模なYMMPSファイルの処理
2. **並行処理テスト**: 複数ファイルの同時変換
3. **メモリ使用量テスト**: リソース消費の監視
4. **統合テスト**: 実際のYMMファイルとの互換性確認
5. **VOICEVOX モック精度向上**: 実際のVOICEVOX APIとの一致率測定
6. **動的VOICEVOXモック**: 話者ID、感情パラメータに応じた音声長計算
7. **FFProbe モック精度向上**: 実際の動画ファイル解析結果との一致率測定
8. **動的FFProbeモック**: フレームレート、解像度に応じた処理時間算出
9. **総合自動長さ計算テスト**: `_auto:VOICEVOX`と`_auto:VIDEO`の複合シナリオ