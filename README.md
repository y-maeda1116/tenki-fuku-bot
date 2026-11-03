# Tenki Fuku Bot

前日20時に気象庁 (JMA) の予報API から翌日の気温情報を取得し、成人男性・成人女性・子供それぞれに最適化された服装アドバイスを Discord Webhook で通知するツールです。

## 仕組み

1. 気象庁予報API（`/bosai/forecast/data/forecast/{areaCode}.json`、APIキー不要）から翌日の天気と今日/明日の最高・最低気温を取得
2. 気温に基づいてカテゴリー別の服装アドバイスを生成
3. Discord Webhook の Embeds 形式で通知

```
[20:00 JST]  Fetch Tomorrow's Forecast  →  Generate Advice  →  Send to Discord
```

## 服装ロジック

| 最高気温 | 服装 |
|----------|------|
| < 15℃ | 厚手のアウター（コート、ダウン） |
| 15℃〜20℃ | 薄手のジャケット、カーディガン |
| 20℃〜25℃ | 長袖シャツ |
| >= 25℃ | 半袖 |

追加ルール:
- 寒暖差（最高-最低）が 10℃以上 → 「脱ぎ着しやすい服装を" アドバイスを全カテゴリーに表示
- 子供カテゴリー → "活動量を考慮して+1枚" を常に表示

Discord 通知の Embed カラー:
- < 15℃: 青（`0x3498DB`）
- 15-20℃: 緑（`0x2ECC71`）
- 20-25℃: オレンジ（`0xE67E22`）
- >= 25℃: 赤（`0xE74C3C`）

## プロジェクト構成

```
.
├── cmd/cli/main.go             # エントリポイント
├── config/config.yaml          # 設定ファイル（都市・カテゴリ）
├── internal/
│   ├── config/config.go        # YAML設定読み込み
│   ├── weather/client.go       # 気象庁 (JMA) 予報API クライアント
│   ├── outfit/advisor.go       # 服装判定ロジック
│   └── discord/webhook.go      # Discord Webhook 通知（Embeds）
├── gas/
│   └── tenki-fuku-bot.gs       # Google Apps Script 版（シングルファイル）
├── .github/workflows/
│   ├── notify.yml              # 毎晩 JST 20:00 に翌日通知
│   └── test.yml                # テスト CI
├── Makefile
└── go.mod
```

## セットアップ

### 前提条件

- Go 1.26 or later
- [Discord Webhook URL](https://support.discord.com/hc/ja/articles/228383949-Webhook%E3%82%E8%E3%82)

### インストール

```bash
git clone https://github.com/y-maeda1116/tenki-fuku-bot.git
cd tenki-fuku-bot
go mod download
```

### 設定

1. `config/config.yaml` で JMA の地方コードとエリア名、カテゴリーを設定:

```yaml
areaCode: "130000"       # 東京
weatherArea: "東京地方"
tempArea: "東京"
categories:
  men: true
  women: true
  kids: true
```

2. 環境変数を設定:

```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
```

## 実行

```bash
# 直接実行（翌日の予報を取得して Discord に通知）
go run cmd/cli/main.go

# Make を使用
make run

# ビルド
make build
```

## テスト

```bash
make test              # テスト実行
make test-coverage    # カバレッジ付き
make fmt               # フォーマット
make lint              # Lint
```

## CI/CD

### GitHub Actions

- **テスト**: push/PR 時に自動実行（`.github/workflows/test.yml`）
- **通知**: 毎晩 JST 20:00（UTC 11:00）に翌日の予報を自動通知（`.github/workflows/notify.yml`、`workflow_dispatch` での手動実行も可能）

GitHub リポジトリの **Settings > Secrets and variables > Actions** で以下を設定:

- `DISCORD_WEBHOOK_URL` — Discord Webhook URL

### Google Apps Script 版

GitHub Actions 版の代替手段。ロジックは Go 版と同一（気象庁 API 使用、API キー不要）。

1. [Google Apps Script](https://script.google.com/) で新規プロジェクトを作成
2. `gas/tenki-fuku-bot.gs` の内容をコピペ
3. **スクリプトプロパティ**（プロジェクトの設定 > スクリプトプロパティ）に設定:
   - `DISCORD_WEBHOOK_URL` — Discord Webhook URL
4. **トリガー設定**: 関数 `main` を時間主導型 > 日付ベースのタイマー > 午後 8時〜9時

> 注意: GitHub Actions 版と GAS 版の両方でトリガーを有効にすると、Discord に同じ通知が2回届きます。どちらか一方のみ有効にしてください。

## License

MIT
