# tenki-fuku-bot

## プロジェクト概要

天気予報に基づく服装提案を Discord に通知する Go 製 CLI。Google Apps Script 版を `gas/` に同梱。

## プロジェクト構造

```
.
├── cmd/cli/            # エントリーポイント (main.go)
├── internal/
│   ├── config/         # 設定読み込み (config/config.yaml)
│   ├── weather/        # 天気予報クライアント
│   ├── outfit/         # 服装提案ロジック
│   └── discord/        # Discord Webhook 送信
├── gas/                # Google Apps Script 版 (tenki-fuku-bot.gs)
├── config/config.yaml
├── .github/workflows/  # test.yml / notify.yml
└── Makefile
```

## 開発コマンド

```bash
make build            # ビルド (bin/tenki-fuku-bot)
make run              # 実行
make test             # go test ./internal/...
make test-coverage    # カバレッジレポート
make fmt              # go fmt
make lint             # golangci-lint run ./internal/...
```

## コーディング規約

- ロジックは `internal/` 以下のパッケージに配置し、各パッケージに `_test.go` を同梱する
