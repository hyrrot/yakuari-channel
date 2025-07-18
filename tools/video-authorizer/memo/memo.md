以下のような仕組みをGoを使って、Kent BeckによるTest Driven Developmentに忠実な開発方法で実装したい。

yamlファイルから、ゆっくりムービーメーカー Yukkuri Movie Maker (YMM) のプロジェクトデータ(YMMP)を生成するCLI.

# Motivation
YMM は素晴らしい動画オーサリングツールであるが、そのプロジェクトファイルの編集効率の向上という点について以下のような課題を抱えている。

- ポータビリティ: プロジェクトファイルに保存される素材ファイルなどのパスが絶対パスになっており、プロジェクトファイルを別の場所に移動した場合、それらを全て書き換えなければならない。
- タイムライン途中の編集が困難: タイムライン上の要素に対して、長さが変わるような変更を行ったとき、動画の後のシーンをずらす作業を手作業で行う必要があるのが大変である。
- プログラム的な編集が困難: タイムライン上の要素をプログラム的に扱う方法が用意されていない。プロジェクトファイルがJSONファイルなので、それを直接扱うスクリプトなどを経由することで可能であるものの、複雑な作業が要求される。

# Solution

YMMPファイルに含まれる要素の構造をyamlフォーマットで定義したものを、「構造化YMMP (Structured YMMP)ファイル」と呼称する。拡張子は.symmpとする。
SYMMPファイルをYMMP形式に変換するGoプログラムを作成する。

## SYMMP フォーマット

- ymmpファイルのタイムラインの構造を episode, sequence, scene, shot, item の単位で構造化して定義し、それらを適切に並べた上でymmpファイルに変換する。それぞれの単位は id 要素を持ち、他の要素から必要なときに参照することができる。
  episode: ymmpファイルで扱われるプロジェクト全体。順序のある複数のsequenceを含み、それらが指定された順番に並べられる。
  sequence: 順序のある複数の scene を含み、それらが指定された順番に並べられる。
  scene: 順序のある複数の shot を含み、それらが指定された順番に並べられる。
  shot: 順序のある複数のitemを含み、それらは同じ開始時刻を持つ（同時に開始される）。
  item: ymmpのItemに対応。

- YMMP ファイルは複数のTimelineを保持することができるが、SYMMPではタイムラインが1つであるという仮定をおく。

- item には length を指定することができる。この値は、いくつか種類がある。
  - 数値を指定: 指定された数値（秒単位）の長さのitemになる。
  - "until:SEQUENCE_END" そのitemが所属する sequence の最後まで続く。
  - "until:SCENE_END" そのitemが所属する scene の最後まで続く。
  - "until:SHOT_END" そのitemが所属する shot の最後まで続く。
  - "auto" そのitemが、itemの内容によって自動的に長さを決めることができるとき、その長さにする。
    - 長さの決め方は以下の通り。
      - item が voice である場合、音声の長さ。VOICEVOXを利用している場合、VOICEVOXのAPIを使って音声を生成し、音声の長さを取得する。
      - item が video である場合、動画の長さ。ffprobeなどを利用し取得する。
      - item に対してその長さを検出する方法はプラガブルにし、本体に対する変更なしに新たなitemのサポートを行うことができるようにする。

- SYMMPファイル内で指定される素材ファイルのパスは相対パスで記述される。

## サポートされるitemタイプ

SYMMPファイルでサポートされるitemタイプは以下の通り：

### voice
- 音声合成によるアイテム
- 必須プロパティ: line（テキスト）
- オプションプロパティ: character_name, voice_id, speed, pitch, volume
- 長さ自動計算: VOICEVOX APIによる音声生成時間を使用

### video
- 動画ファイルアイテム
- 必須プロパティ: file_path
- オプションプロパティ: x, y, z, zoom, rotation, alpha
- 長さ自動計算: ffprobeによる動画時間を使用

### image
- 画像ファイルアイテム
- 必須プロパティ: file_path
- オプションプロパティ: x, y, z, zoom, rotation, alpha
- 長さ自動計算: なし（数値指定必須）

### tachie
- 立ち絵アイテム
- 必須プロパティ: item（キャラクター名）
- オプションプロパティ: x, y, z, zoom, rotation, alpha, fade_in, fade_out
- 長さ自動計算: なし（数値指定必須）

### audio
- オーディオファイルアイテム
- 必須プロパティ: file_path
- オプションプロパティ: volume
- 長さ自動計算: ffprobeによる音声時間を使用

### SYMMPファイルのサンプル

```yaml
symmp_format_version: 0.1

# プロジェクト設定
project:
  name: サンプルプロジェクト
  output_dir: ./output
  
# デフォルト値設定
defaults:
  voice:
    character_name: ずんだもん
    length: auto
  tachie:
    fade_in: 0.5
    fade_out: 0.5

sequences:
- id: sequence1
  scenes:
  - id: scene1
    shots:
    - tachie:
        item: ずんだもん
        length: until:SEQUENCE_END
    - voice:
        character_name: ずんだもん
        line: |
          ずんだもんなのだ
        length: auto
  - id: scene2
    shots:
    - image:
        file_path: image/sample.png
        x: 160
        y: 65
        z: 10
        zoom: 1.0
        length: until:SCENE_END
    - voice:
        line: |
          今日はこの沼について語るよ
        length: auto
```


## SYMMP → YMMP 変換

- SYMMPファイルの構造はymmpファイルのItemsに出力される。
- ymmpファイルが持つべきであるが、上記に当てはまらないものはデフォルト値を用意し、必要に応じてSYMMPファイル内でオーバーライドできる。
  - デフォルト値の優先順位:
    1) SYMMPファイル内の個別指定, 2) SYMMPファイル内のdefaultsセクション, 3) configディレクトリ（起動時に指定）内の設定ファイル
  - 設定ファイル形式: YAML形式、ファイル名は各itemタイプに対応（voice.yaml, image.yaml等）
- item内のファイルパスは相対パスで記載される。相対パスの基準はYAMLファイルの配置ディレクトリとする。ymmpファイル出力時に、絶対パスに変更してから書き出される。
- SYMMPファイルのスキーマはバリデートされる。
  - 無効なYAML構造の場合、詳細なエラーメッセージを表示して終了
  - 必須フィールドの欠如時は、フィールド名と場所を明示
  - 参照IDが存在しない場合、参照元と参照先を明示してエラー



* 実行方法

プログラムは以下のコマンドラインインターフェースを提供する:

```
ymmp-compiler -i <input.yaml> -o <output.ymmp> [options]

必須引数:
  -i, --input <path>      入力YAMLファイルのパス
  -o, --output <path>     出力YMMPファイルのパス

オプション:
  --defaults-dir <path>   デフォルト値ディレクトリ (デフォルト: ./defaults)
  --validate-only         バリデーションのみ実行
  --verbose               詳細ログを出力
  --cache-dir <path>      VOICEVOX APIキャッシュディレクトリ
  --voicevox-url <url>    VOICEVOX APIエンドポイント (デフォルト: http://localhost:50021)
  --dry-run               実際の変換を行わず、処理内容を表示
```

* VOICEVOX連携

- 音声の長さを自動計算する場合、VOICEVOX APIを使用する
- APIエンドポイントは環境変数 VOICEVOX_API_URL または --voicevox-url オプションで指定（デフォルト: http://localhost:50021）
- 生成した音声データはキャッシュして、同じテキストの再計算を回避
- キャッシュファイルの形式: JSON形式でテキストハッシュをキー、音声時間（秒）を値として保存
- キャッシュの保存場所: --cache-dir オプションで指定（デフォルト: ~/.ymmp-compiler/cache）
- オフライン時は、キャッシュがある場合はそれを使用、なければエラー
- API通信タイムアウト: 30秒

* タイムライン計算ルール

1. 各要素の開始時刻は、前の要素の終了時刻から自動計算
2. shot内の複数itemは同じ開始時刻を持つ
3. 長さ指定の解決順序:
   - まず "auto" の長さを計算（外部API呼び出しを伴う）
   - 次に "until:SHOT_END" を最長のitemに合わせる
   - 次に "until:SCENE_END" を最長のshotに合わせる
   - 最後に "until:SEQUENCE_END" を解決
4. オーバーラップは基本的に許可しない（将来的にオプションで対応）
5. 依存関係の循環検出: 長さ計算時に循環参照を検出した場合はエラー
6. 最小長さ: 各itemの最小長さは0.1秒とする

* エラーハンドリング

- ファイルが見つからない場合: ファイルパスと種類を明示してエラー
  例: "Error: image file not found: /path/to/image.png"
- ID参照エラー: 参照元の位置（行番号）と参照先IDを明示
  例: "Error: reference ID 'scene2' not found (referenced at line 25)"
- 長さ計算エラー: 該当するitemの位置と理由を明示
  例: "Error: failed to calculate length for voice item at line 30: VOICEVOX API timeout"
- API通信エラー: リトライ処理（3回まで、間隔5秒）後、エラー詳細を表示
- バリデーションエラー: スキーマ違反の詳細と修正方法を表示
- 循環参照エラー: 依存関係の循環を検出した場合の詳細表示