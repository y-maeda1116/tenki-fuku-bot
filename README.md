# Tenki Fuku Bot

前日20時に翌日の天気予報を取得し、成人男性・成人女性・子供それぞれのカテゴリーに最適化された服装アドバイスを Discord Webhook で通知するツールです。

## 機能

- OpenWeatherMap 5日予報API から翌日の気温データを取得
- 気温に応じた服装アドバイスを自動生成
- 成人男性・成人女性・子供の3カテゴリーに対応
- 寒暖差が大きい日の脱ぎ着アドバイス
- 子供向けの「+1枚」アドバイス
- Discord Embeds 形式の見やすい通知
- GitHub Actions または Google Apps Script で毎晩20時自動実行

## 服装ロジック

| 最高気温 | 服装 |
|----------|------|
| < 15℃ | 厚手のアウター（コート、ダウン） |
| 15℃〜20℃ | 薄手のジャケット、カーディガン |
| 20℃〜25℃ | 長袖シャツ |
| >= 25℃ | 半袖 |

- 寒暖差（最高-最低）が 10℃以上 → 「脱ぎ着しやすい服装を」強調
- 子供カテゴリー → 「活動量を考慮して+1枚」を追加

## プロジェクト構成

```
.
├── cmd/cli/main.go             # エントリポイント
├── config/config.yaml          # 設定ファイル（都市・カテゴリ）
├── internal/
│   ├── config/                 # YAML設定読み込み
│   ├── weather/                # OpenWeatherMap API クライアント
│   ├── outfit/                 # 服装判定ロジック
│   └── discord/                # Discord Webhook 通知
├── gas/                        # Google Apps Script 版
│   └── tenki-fuku-bot.gs
├── .github/workflows/
│   ├── notify.yml              # 毎晩 JST 20:00 に翌日通知
│   └── test.yml                # テスト CI
├── Makefile
└── go.mod
```

## セットアップ

### 前提条件

- Go 1.26 or later

### インストール

```bash
git clone https://github.com/y-maeda1116/tenki-fuku-bot.git
cd tenki-fuku-bot
go mod download
```

### 設定

1. `config/config.yaml` で都市とカテゴリーを設定:

```yaml
city: "Tokyo"
categories:
  men: true
  women: true
  kids: true
```

2. 環境変数を設定:

```bash
export WEATHER_API_KEY="your-openweathermap-api-key"
export DISCORD_WEBHOOK_URL="your-discord-webhook-url"
```

## 実行

```bash
# 直接実行
go run cmd/cli/main.go

# または Make を使用
make run

# ビルド
make build
```

## テスト

```bash
# テスト実行
make test

# カバレッジ付き
make test-coverage

# フォーマット
make fmt

# Lint
make lint
```

## CI/CD

### GitHub Actions

- **テスト**: push/PR 時に自動実行
- **通知**: 毎日 JST 20:00（UTC 11:00）に翌日の予報を自動通知

GitHub リポジトリの **Settings > Secrets and variables > Actions** で以下を設定:

- `WEATHER_API_KEY`
- `DISCORD_WEBHOOK_URL`

### Google Apps Script（cron が不安定な場合の代替）

1. [Google Apps Script](https://script.google.com/) で新規プロジェクトを作成
2. `gas/tenki-fuku-bot.gs` の内容をコピペ
3. **スクリプトプロパティ** に設定:
   - `WEATHER_API_KEY`
   - `DISCORD_WEBHOOK_URL`
4. **トリガー設定**: 関数 `main` を時間主導型 > 日付ベースのタイマー > 午後 8時〜9時

## License

MIT
