# DevCompanion 技術仕様書 (v2.1)

## 1. アーキテクチャ概要

DevCompanionは、OSや開発環境からの微細な活動を観測し、開発者の状況を推定して最適なフィードバックを行うイベントパイプライン構造を採用しています。

### イベントパイプライン
1. **Sensors**: OSやファイルシステムの事実のみを観測。
2. **Signals**: 意味を持たない低レベルイベント。
3. **Context Engine**: シグナルを組み合わせ、確率ベースで開発者の状態を推定。
4. **Persona Engine**: 状態とキャラクター設定を組み合わせ、セリフのトーンを決定。
5. **Message Output**: LLMを介して最終的なセリフを生成。

---

## 2. 構成レイヤー

### Layer 1: Sensors (観測)
- **ProcessSensor**: 実行中のプロセス（AI Agent, IDE等）を監視。
- **FSSensor**: ソースコードや設定ファイルの変更を監視。
- **GitSensor**: コミットやブランチ操作を監視。
- **IdleSensor**: キーボード/マウス入力の不在を監視。

### Layer 2: Signals (イベント)
Sensorsから生成される最小単位のイベント。
- `process_started`, `file_modified`, `git_commit` 等。

### Layer 3: Context Engine (状況推定)
複数のSignalsの時間密度と重み付けにより、以下の状態を確率的に判定。
- `CODING`: 通常のコーディング中。
- `AI_PAIR_PROGRAMMING`: AIエージェントと協力して開発中。
- `DEEP_WORK`: 高い集中状態で作業中。
- `STUCK`: エラーが頻発し、行き詰まっている状態。

### Layer 4: Persona Engine (人格)
- **Character Core**: 「応援してくれる後輩（サクラ）」としての基本人格。
- **Persona Style**: `soft` (優しく), `energetic` (元気に), `strict` (厳しく) の3つの表現スタイル。

---

## 3. デバッグと観測性 (Observability)

開発者の行動を正確にシミュレーションし、Context Engine の精度を向上させるための機構を備えています。

### 3.1 Signal Recorder
`monitor` 層で受信した全シグナルを JSONL 形式で永続化します。
- **保存場所**: `~/.devcompanion/signals/signals_YYYYMMDD_HHMMSS.jsonl`
- **目的**: 本番環境で発生した複雑なイベントシーケンスの記録。

### 3.2 Signal Replay Engine
記録されたシグナルログを読み込み、パイプラインに再注入します。
- **再生モード**: 
  - `RealTime`: 元のイベント間隔を再現。
  - `Fast`: ウェイトなしで即座に全イベントを処理（ユニットテスト用）。

### 3.3 Context Viewer
リプレイ中の内部状態（Confidence, State遷移, Persona Style）をリアルタイムに可視化する CLI ツール。
- **実行**: `go run cmd/contextviewer/main.go -f <log_path>`

---

## 4. テスト戦略 (Testing Strategy)

### 4.1 決定論的テスト (Deterministic Testing)
LLM のセリフ生成（Fallback モード）にシード固定オプションを導入し、ランダム性に依存しない期待値検証を可能にしています。

### 4.2 長時間安定性テスト (Stability Test)
1万件以上のシグナルを高速リプレイし、以下の項目を自動検証します。
- goroutine のリークがないこと。
- メモリ消費が一定範囲内に収まること。
- パイプライン内で panic が発生しないこと。

---

## 5. ディレクトリ構造

```text
cmd/
└── contextviewer/   # 状態遷移可視化ツール
internal/
├── agent/           # Agent Adapter
├── behavior/        # (Legacy) 行動推論
├── config/          # 設定管理
├── context/         # Context Engine (状況推定)
├── debug/           # デバッグ機構
│   ├── recorder/    # シグナル記録
│   └── replay/      # シグナル再生
├── llm/             # 生成AI / セリフ生成
├── monitor/         # パイプライン・オーケストレーター
├── sensor/          # Sensors (観測)
├── session/         # (Legacy) セッション管理
├── types/           # アーキテクチャ共有型
└── ws/              # フロントエンド通信
docs/
└── spec.md          # 本仕様書
```

---

## 6. 拡張性
...（以下、既存の拡張性セクション）
