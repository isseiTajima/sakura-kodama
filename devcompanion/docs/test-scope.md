# テストスコープ

## Wails 開発ワークフロー回帰テスト

`internal/wails_dev_workflow_test.go` は Wails dev モードのハンドシェイクを回帰テストする。

### テスト対象

- `wails.json` の `frontend:dev:serverUrl` が期待値と一致すること
- フロントエンド Dev サーバーが `npm run dev -- --host 127.0.0.1 --port 5173` で起動されること
- `make dev` コマンドでバックエンドと UI が同時に起動すること

### 背景

Wails は `frontend:dev:serverUrl` に設定した URL へリクエストをプロキシする。
このフィールドが正しく設定されていないと `AssetServer options invalid` エラーになる。
