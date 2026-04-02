# Go Template Project

## プロジェクト概要

Mac、Windows、Linux対応のクロスプラットフォームGoアプリケーションテンプレート（CLI + Desktop）。

## プロジェクト構造

```
.
├── .github/workflows/    # GitHub Actions CI/CD
│   ├── build-cli.yml     # CLI クロスビルド
│   ├── build-desktop.yml # Desktop クロスビルド（Wails）
│   └── test.yml          # テスト + カバレッジ
├── .claude/              # Claude Code 設定
├── cmd/                  # アプリケーションエントリーポイント
│   ├── app/
│   │   └── main.go       # バックグラウンドアプリ（signal 待機付き）
│   ├── cli/
│   │   └── main.go       # CLI アプリ（Cobra ベース）
│   └── desktop/
│       ├── main.go       # Desktop アプリ（Wails ベース）
│       ├── wails.json    # Wails 設定
│       └── frontend/     # フロントエンド（React + Vite）
├── internal/             # 内部パッケージ（外部からインポート不可）
│   ├── cli/              # CLI コマンド定義（root, hello, version）
│   ├── config/           # 設定読み込み（Viper ベース）
│   ├── core/             # ビジネスロジック（Service）
│   ├── logger/           # ログ設定（Zap ベース）
│   └── ui/               # Desktop UI ロジック
├── frontend/             # Wails フロントエンド（React + TypeScript）
├── test/mocks/           # モック生成先
├── bin/                  # ビルド出力ディレクトリ
├── Makefile              # ビルド・テスト・Lint ターゲット
├── config.yaml.example   # 設定ファイルテンプレート
├── env.example           # 環境変数テンプレート
├── go.mod / go.sum       # Go モジュール
├── wails.json            # Wails ルート設定
└── air.toml              # Hot Reload 設定（CLI 用）
```

## 開発ルール

### 依存関係管理

- 設定は `github.com/spf13/viper` を使用（YAML ファイル + 環境変数）
- ログは `go.uber.org/zap` を使用
- CLI は `github.com/spf13/cobra` を使用
- Desktop は `github.com/wailsapp/wails/v2` を使用
- 依存パッケージは `go mod tidy` で管理する

### ビルド

```bash
make build-cli        # CLI ビルド
make build-desktop    # Desktop ビルド（Wails）
make build-all        # 両方ビルド
make run-cli          # CLI 実行
make run-desktop      # Desktop 開発モード（wails dev）
make clean            # クリーンアップ
```

### テスト

```bash
go test ./...                # 全テスト
go test -v -race ./...       # レース検出付き
make test-coverage           # カバレッジレポート生成
make mocks                   # モック生成（mockgen）
```

### コーディング規約

- 標準的なGoプロジェクト構造（`cmd/`, `internal/`）を維持する
- CLI コードは `cmd/cli/` + `internal/cli/`
- Desktop コードは `cmd/desktop/` + `internal/ui/`
- ビジネスロジックは `internal/core/`
- 内部パッケージは `internal/` 以下に配置
- signal処理を使用して Ctrl+C で安全に終了させる

### GitHub Actions

- `test.yml`: Ubuntu 22.04 でテスト + レース検出 + カバレッジ
- `build-cli.yml`: Linux/macOS/Windows 向け CLI クロスビルド
- `build-desktop.yml`: Linux(Ubuntu 22.04)/macOS/Windows 向け Desktop ビルド
