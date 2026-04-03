# Tenki Fuku Bot - Design Document

## Overview

Go 言語製の天気服装アドバイスBot。OpenWeatherMap API から気温データを取得し、成人男性・成人女性・子供の各カテゴリーに最適化された服装アドバイスを Discord Webhook（Embeds形式）で通知する。GitHub Actions で毎朝 JST 06:00（UTC 21:00）に実行。

## Project Structure

```
tenki-fuku-bot/
├── cmd/cli/main.go             ← エントリポイント
├── config/
│   └── config.yaml             ← 設定ファイル（YAML形式）
├── internal/
│   ├── weather/
│   │   └── client.go           ← OpenWeatherMap API 通信
│   ├── outfit/
│   │   └── advisor.go          ← 気温ベース服装判定ロジック
│   └── discord/
│       └── webhook.go          ← Discord Webhook 通知（Embeds）
├── .github/workflows/
│   └── notify.yml              ← 毎朝 JST 06:00 実行
├── go.mod
└── go.sum
```

## Configuration

### config/config.yaml

```yaml
city: "Tokyo"
categories:
  men: true
  women: true
  kids: true
```

### Environment Variables

| 変数名 | 用途 | 必須 |
|--------|------|------|
| `WEATHER_API_KEY` | OpenWeatherMap API キー | Yes |
| `DISCORD_WEBHOOK_URL` | Discord Webhook URL | Yes |

環境変数は GitHub Actions Secrets から `env:` で渡す。

## Data Flow

```
main.go
  → config/config.yaml 読み込み
  → weather.Fetch(city, apiKey) → WeatherData{TempMax, TempMin, Description}
  → outfit.Advise(WeatherData, categories) → []OutfitAdvice
  → discord.Send(webhookURL, []OutfitAdvice) → Discord Embeds 通知
```

## Core Types

### WeatherData

```go
type WeatherData struct {
    City        string
    TempMax     float64
    TempMin     float64
    Description string
}
```

### OutfitAdvice

```go
type OutfitAdvice struct {
    Category   string // "men", "women", "kids"
    Outfit     string // 服装の説明
    Tips       string // 寒暖差アドバイス等
    TempMax    float64
    TempMin    float64
    TempDiff   float64
}
```

## Outfit Logic

### Temperature-Based Outfit Selection

| 気温帯（最高気温） | 服装 |
|---------------------|------|
| < 15℃ | 厚手のアウター（コート、ダウン） |
| 15℃〜20℃ | 薄手のジャケット、カーディガン |
| 20℃〜25℃ | 長袖シャツ |
| >= 25℃ | 半袖 |

### Special Rules

- **寒暖差アドバイス**: 最高気温 - 最低気温 >= 10℃ の場合、「脱ぎ着しやすい服装を」というアドバイスを全カテゴリーに追加
- **子供向け追加**: 子供カテゴリーには「活動量を考慮して+1枚」のアドバイスを常に追加

## Discord Notification Format

Embeds形式で各カテゴリーごとにEmbedを生成。

### Embed Structure

- **Title**: カテゴリー名（例: 「👔 成人男性の服装アドバイス」）
- **Color**: 気温に応じて変動
  - < 15℃: 青（0x3498DB）
  - 15-20℃: 緑（0x2ECC71）
  - 20-25℃: オレンジ（0xE67E22）
  - >= 25℃: 赤（0xE74C3C）
- **Fields**:
  - 服装アドバイス
  - 最高気温 / 最低気温
  - 寒暖差アドバイス（該当時）
  - 子供向け追加アドバイス（該当時）

## GitHub Actions Workflow

- **Schedule**: `cron: '0 21 * * *'`（JST 06:00 = UTC 21:00）
- **Steps**:
  1. Checkout
  2. Setup Go
  3. `go run cmd/cli/main.go`
- **Secrets**: `WEATHER_API_KEY`, `DISCORD_WEBHOOK_URL` を `env:` で渡す

## Error Handling

- API呼び出し失敗 → 標準エラー出力にログ、exit code 1 で終了
- 環境変数未設定 → 起動時にエラーメッセージ表示、exit code 1
- Webhook送信失敗 → 標準エラー出力にログ、exit code 1
- 設定ファイル読み込み失敗 → エラーメッセージ表示、exit code 1

## Dependencies

- `net/http` - 標準ライブラリ（API通信）
- `encoding/json` - 標準ライブラリ（JSON パース）
- `gopkg.in/yaml.v3` - YAML 設定ファイル読み込み
- 外部依存は最小限に抑える

## Out of Scope

- 複数都市対応（将来拡張として考慮）
- API レスポンスのキャッシュ
- リトライロジック
- ログファイル出力（標準出力/エラー出力のみ）
- テスト自動化（将来的に追加可能）
