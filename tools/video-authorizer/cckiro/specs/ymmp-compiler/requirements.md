# YMMP Compiler 要件定義書

## 1. プロジェクト概要

YAMLファイルから、ゆっくりムービーメーカー（YMM）のプロジェクトデータ（YMMP）を生成するコマンドラインツールを開発する。

### 1.1 SYMMPファイルとは

SYMMP（Structured YMMP）ファイルは、YMMPファイルの構造を人間が読み書きしやすいYAML形式で表現したものである。

#### 特徴
- **階層的な構造化**: 動画コンテンツをepisode、sequence、scene、shot、itemという階層で整理
- **相対パス記述**: 素材ファイルを相対パスで指定し、プロジェクトの可搬性を向上
- **自動長さ計算**: 音声や動画の長さを自動的に計算し、手動調整を削減
- **プログラマブル**: YAMLファイルとして編集可能で、スクリプトによる自動生成も容易

#### ファイル構造例
```yaml
symmp_format_version: 0.1

project:
  name: サンプル動画プロジェクト

sequences:
  - id: opening
    scenes:
      - id: title_scene
        shots:
          - voice:
              line: こんにちは、今日は動画編集について説明します
              character_name: ずんだもん
              length: auto
          - image:
              file_path: ./assets/title.png
              length: until:SHOT_END
```

## 2. ユーザーストーリー

### 2.1 動画制作者として
- **As a** 動画制作者
- **I want to** YAMLファイルで動画の構成を定義し、YMMPファイルを自動生成したい
- **So that** プロジェクトファイルのポータビリティを確保し、編集効率を向上させることができる

### 2.2 コンテンツクリエイターとして
- **As a** ゆっくり実況動画のクリエイター
- **I want to** タイムライン上の要素をプログラム的に編集したい
- **So that** 繰り返し作業を自動化し、制作時間を短縮できる

### 2.3 チーム制作者として
- **As a** チームで動画を制作する編集者
- **I want to** 相対パスで素材を管理したい
- **So that** プロジェクトファイルを他のメンバーと簡単に共有できる

## 3. EARS形式による要件定義

### 3.1 基本機能要件

#### REQ-001: SYMMPファイル読み込み
- **When** ユーザーがSYMMPファイル（.symmp）を指定した場合
- **The system shall** YAMLフォーマットのファイルを読み込み、構造を解析する

#### REQ-002: 階層構造の解析
- **The system shall** episode > sequence > scene > shot > item の階層構造を解析する
- **Where** 各要素は一意のIDを持つ

#### REQ-003: YMMPファイル生成
- **When** 有効なSYMMPファイルが解析された場合
- **The system shall** YMM互換のJSONフォーマット（.ymmp）ファイルを生成する

### 3.2 アイテムタイプ要件

#### REQ-004: 音声アイテム（voice）
- **The system shall** voice アイテムをサポートする
- **Where** 必須プロパティ: line（テキスト）
- **Where** オプションプロパティ: character_name, voice_id, speed, pitch, volume

#### REQ-005: 動画アイテム（video）
- **The system shall** video アイテムをサポートする
- **Where** 必須プロパティ: file_path
- **Where** オプションプロパティ: x, y, z, zoom, rotation, alpha

#### REQ-006: 画像アイテム（image）
- **The system shall** image アイテムをサポートする
- **Where** 必須プロパティ: file_path
- **Where** オプションプロパティ: x, y, z, zoom, rotation, alpha

#### REQ-007: 立ち絵アイテム（tachie）
- **The system shall** tachie アイテムをサポートする
- **Where** 必須プロパティ: item（キャラクター名）
- **Where** オプションプロパティ: x, y, z, zoom, rotation, alpha, fade_in, fade_out

#### REQ-008: オーディオアイテム（audio）
- **The system shall** audio アイテムをサポートする
- **Where** 必須プロパティ: file_path
- **Where** オプションプロパティ: volume

### 3.3 長さ計算要件

#### REQ-009: 数値指定
- **When** アイテムの長さが数値で指定された場合
- **The system shall** その値を秒単位の長さとして使用する

#### REQ-010: 自動長さ計算（voice）
- **When** voice アイテムの長さが "auto" に設定された場合
- **The system shall** VOICEVOX APIを使用して音声の長さを自動計算する

#### REQ-011: 自動長さ計算（video/audio）
- **When** video/audio アイテムの長さが "auto" に設定された場合
- **The system shall** ffprobeを使用してメディアファイルの長さを取得する

#### REQ-012: 相対長さ指定
- **When** アイテムの長さが "until:SEQUENCE_END", "until:SCENE_END", "until:SHOT_END" のいずれかに設定された場合
- **The system shall** 指定された階層の終了時刻まで続くように長さを計算する

### 3.4 パス管理要件

#### REQ-013: 相対パス解決
- **When** SYMMPファイル内でファイルパスが相対パスで指定された場合
- **The system shall** 基準ディレクトリを使用して絶対パスに変換する
- **Where** 基準ディレクトリは `--base-dir` オプションで指定可能
- **Where** デフォルトの基準ディレクトリはSYMMPファイルの配置ディレクトリ

### 3.5 デフォルト値要件

#### REQ-014: デフォルト値の優先順位
- **The system shall** 以下の優先順位でデフォルト値を適用する:
  1. SYMMPファイル内の個別指定
  2. SYMMPファイル内のdefaultsセクション
  3. 外部設定ファイル（defaults/ディレクトリ）

### 3.6 バリデーション要件

#### REQ-015: YAMLスキーマ検証
- **When** SYMMPファイルを読み込む場合
- **The system shall** YAMLスキーマの妥当性を検証する
- **If** 無効な構造が検出された場合
- **Then** 詳細なエラーメッセージと行番号を表示する

#### REQ-016: 必須フィールド検証
- **The system shall** 各アイテムタイプの必須フィールドの存在を検証する
- **If** 必須フィールドが欠如している場合
- **Then** フィールド名と場所を明示してエラーを表示する

#### REQ-017: ID参照検証
- **When** ID参照が使用されている場合
- **The system shall** 参照先のIDが存在することを検証する
- **If** 参照先が見つからない場合
- **Then** 参照元と参照先を明示してエラーを表示する

### 3.7 外部連携要件

#### REQ-018: 音声合成サービス連携
- **When** voice アイテムの長さを自動計算する場合
- **The system shall** プラガブルな音声合成サービスインターフェースを通じて音声長を取得する
- **Where** デフォルトの音声合成サービスはVOICEVOX
- **Where** 音声合成サービスはプラグインとして追加可能（例: AquesTalk, Amazon Polly）
- **Where** サービス固有の設定:
  - VOICEVOX: エンドポイント（デフォルト: http://localhost:50021）、タイムアウト: 30秒、リトライ: 最大3回（間隔5秒）

#### REQ-019: キャッシュ機能
- **The system shall** 音声合成サービスの結果をキャッシュする
- **So that** 同じテキストの再計算を回避できる
- **Where** キャッシュはJSON形式で保存される
- **Where** キャッシュキーにはサービス名とパラメータを含む

### 3.8 コマンドライン要件

#### REQ-020: 必須引数
- **The system shall** 以下の必須引数を受け付ける:
  - `-i, --input <path>`: 入力YAMLファイルのパス
  - `-o, --output <path>`: 出力YMMPファイルのパス

#### REQ-021: オプション引数
- **The system shall** 以下のオプション引数を受け付ける:
  - `--defaults-dir <path>`: デフォルト値ディレクトリ
  - `--validate-only`: バリデーションのみ実行
  - `--verbose`: 詳細ログ出力
  - `--cache-dir <path>`: キャッシュディレクトリ
  - `--voicevox-url <url>`: VOICEVOX APIエンドポイント（VOICEVOXサービス使用時）
  - `--voice-service <name>`: 使用する音声合成サービス（デフォルト: voicevox）
  - `--dry-run`: 実際の変換を行わない
  - `--base-dir <path>`: 相対パスを絶対パスに変換する際の基準ディレクトリ（デフォルト: SYMMPファイルのディレクトリ）

### 3.9 エラーハンドリング要件

#### REQ-022: ファイルエラー
- **If** ファイルが見つからない場合
- **Then the system shall** ファイルパスと種類を明示してエラーを表示する

#### REQ-023: 循環参照検出
- **When** 長さ計算時に循環参照が検出された場合
- **The system shall** 依存関係の詳細を表示してエラーを報告する

### 3.10 制約事項

#### REQ-024: タイムライン制限
- **The system shall** 1つのタイムラインのみをサポートする

#### REQ-025: 最小長さ制限
- **The system shall** 各アイテムの最小長さを0.1秒とする

#### REQ-026: オーバーラップ制限
- **The system shall not** アイテムのオーバーラップをサポートする（将来の拡張として検討）

## 4. 非機能要件

### NFR-001: 開発言語
- **The system shall** Go言語で実装される

### NFR-002: 開発手法
- **The system shall** Test Driven Development（TDD）に従って開発される
- **Where** Kent Beckの手法に準拠する

### NFR-003: 拡張性
- **The system shall** 以下の拡張を容易にするプラガブルアーキテクチャを採用する:
  - 新しいアイテムタイプの追加
  - 新しい音声合成サービスの追加（REQ-018に関連）
  - 新しい長さ計算プロバイダーの追加

## 5. 成果物

- `ymmp-compiler` 実行可能ファイル
- 単体テスト・統合テスト一式
- デフォルト設定ファイルサンプル（defaults/ディレクトリ）