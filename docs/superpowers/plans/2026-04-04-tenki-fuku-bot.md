# Tenki Fuku Bot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** OpenWeatherMap API から気温を取得し、カテゴリー別の服装アドバイスを Discord Webhook で通知するバッチツールを構築する。

**Architecture:** 3層のパイプライン構成。weather → outfit → discord のデータフロー。各パッケージは独立した関数として実装し、cmd/cli/main.go がオーケストレーションする。設定は YAML ファイル、シークレットは環境変数。

**Tech Stack:** Go 1.26, OpenWeatherMap API, Discord Webhook, GitHub Actions

---

## File Map

| Action | Path | Responsibility |
|--------|------|---------------|
| Delete | `cmd/app/` | 不要なバックグラウンドアプリ |
| Delete | `cmd/desktop/` | 不要なデスクトップアプリ |
| Delete | `internal/cli/` | 不要なCobra CLI |
| Delete | `internal/ui/` | 不要なWails UI |
| Delete | `internal/version/` | 不要なバージョンパッケージ |
| Delete | `internal/core/` | 不要なサービス層 |
| Delete | `internal/logger/` | 不要なZapロガー |
| Delete | `internal/config/` | 不要なViper設定 |
| Delete | `test/` | 不要なモック |
| Delete | `.github/workflows/build-cli.yml` | 不要なCLIビルド |
| Delete | `.github/workflows/build-desktop.yml` | 不要なDesktopビルド |
| Delete | `.github/workflows/security.yml` | 不要なセキュリティチェック |
| Delete | `air.toml` | 不要なHot Reload |
| Delete | `config.yaml.example` | 不要な設定例 |
| Delete | `env.example` | 不要な環境変数例 |
| Delete | `frontend/` | 不要なフロントエンド |
| Delete | `wails.json` | 不要なWails設定 |
| Modify | `go.mod` | モジュール名変更、依存関係整理 |
| Create | `cmd/cli/main.go` | エントリポイント |
| Create | `config/config.yaml` | 設定ファイル |
| Create | `internal/weather/client.go` | OpenWeatherMap API 通信 |
| Create | `internal/weather/client_test.go` | weather テスト |
| Create | `internal/outfit/advisor.go` | 服装判定ロジック |
| Create | `internal/outfit/advisor_test.go` | outfit テスト |
| Create | `internal/discord/webhook.go` | Discord Webhook 通知 |
| Create | `internal/discord/webhook_test.go` | discord テスト |
| Modify | `.github/workflows/test.yml` | テストパス更新 |
| Create | `.github/workflows/notify.yml` | 毎朝通知ワークフロー |
| Modify | `Makefile` | Bot 用に更新 |
| Modify | `.gitignore` | Bot 用に更新 |

---

### Task 1: Clean Up Template Code

**Files:**
- Delete: `cmd/app/`, `cmd/desktop/`, `internal/cli/`, `internal/ui/`, `internal/version/`, `internal/core/`, `internal/logger/`, `internal/config/`, `test/`, `.github/workflows/build-cli.yml`, `.github/workflows/build-desktop.yml`, `.github/workflows/security.yml`, `air.toml`, `config.yaml.example`, `env.example`, `frontend/` (if exists), `wails.json` (root level)

- [ ] **Step 1: Delete all template files**

Run:
```bash
rm -rf cmd/app cmd/desktop internal/cli internal/ui internal/version internal/core internal/logger internal/config test/ .github/workflows/build-cli.yml .github/workflows/build-desktop.yml .github/workflows/security.yml air.toml config.yaml.example env.example
```

Check for `frontend/` and `wails.json` at root:
```bash
rm -rf frontend/ wails.json
```

- [ ] **Step 2: Update go.mod**

Replace `go.mod` content:

```
module github.com/y-maeda1116/tenki-fuku-bot

go 1.26.0

require gopkg.in/yaml.v3 v3.0.4
```

Then run:
```bash
go mod tidy
```

- [ ] **Step 3: Verify clean state**

Run:
```bash
go build ./...
```
Expected: no output (clean build with no errors — no main package yet, so it may say "no Go files" — that's OK)

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "chore: clean up template code for tenki-fuku-bot"
```

---

### Task 2: Config Package

**Files:**
- Create: `config/config.yaml`
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Create config/config.yaml**

```yaml
city: "Tokyo"
categories:
  men: true
  women: true
  kids: true
```

- [ ] **Step 2: Write the failing test for config loading**

`internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	content := []byte("city: Osaka\ncategories:\n  men: true\n  women: false\n  kids: true\n")
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.City != "Osaka" {
		t.Errorf("City = %q, want %q", cfg.City, "Osaka")
	}
	if !cfg.Categories.Men {
		t.Error("Categories.Men = false, want true")
	}
	if cfg.Categories.Women {
		t.Error("Categories.Women = true, want false")
	}
	if !cfg.Categories.Kids {
		t.Error("Categories.Kids = false, want true")
	}
}

func TestLoad_missingFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Load() expected error for missing file")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test -v ./internal/config/...`
Expected: FAIL — `Load` not defined

- [ ] **Step 4: Write implementation**

`internal/config/config.go`:

```go
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Categories struct {
	Men   bool `yaml:"men"`
	Women bool `yaml:"women"`
	Kids  bool `yaml:"kids"`
}

type Config struct {
	City       string     `yaml:"city"`
	Categories Categories `yaml:"categories"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}
	return &cfg, nil
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test -v ./internal/config/...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add config/config.yaml internal/config/config.go internal/config/config_test.go
git commit -m "feat: add config package with YAML loading"
```

---

### Task 3: Weather Package

**Files:**
- Create: `internal/weather/client.go`
- Create: `internal/weather/client_test.go`

- [ ] **Step 1: Write the failing test**

`internal/weather/client_test.go`:

```go
package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetch(t *testing.T) {
	response := map[string]interface{}{
		"name": "Tokyo",
		"weather": []map[string]interface{}{
			{"description": "clear sky"},
		},
		"main": map[string]interface{}{
			"temp_max": 22.5,
			"temp_min": 15.3,
		},
	}
	body, _ := json.Marshal(response)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "Tokyo" {
			t.Errorf("q param = %q, want %q", got, "Tokyo")
		}
		if got := r.URL.Query().Get("appid"); got != "test-key" {
			t.Errorf("appid param = %q, want %q", got, "test-key")
		}
		if got := r.URL.Query().Get("units"); got != "metric" {
			t.Errorf("units param = %q, want %q", got, "metric")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
	defer server.Close()

	data, err := FetchWithURL(server.URL, "Tokyo", "test-key")
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if data.City != "Tokyo" {
		t.Errorf("City = %q, want %q", data.City, "Tokyo")
	}
	if data.TempMax != 22.5 {
		t.Errorf("TempMax = %f, want 22.5", data.TempMax)
	}
	if data.TempMin != 15.3 {
		t.Errorf("TempMin = %f, want 15.3", data.TempMin)
	}
	if data.Description != "clear sky" {
		t.Errorf("Description = %q, want %q", data.Description, "clear sky")
	}
}

func TestFetch_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := FetchWithURL(server.URL, "Tokyo", "test-key")
	if err == nil {
		t.Error("Fetch() expected error for 500 response")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/weather/...`
Expected: FAIL — `FetchWithURL` and `WeatherData` not defined

- [ ] **Step 3: Write implementation**

`internal/weather/client.go`:

```go
package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type WeatherData struct {
	City        string
	TempMax     float64
	TempMin     float64
	Description string
}

type apiResponse struct {
	Name    string `json:"name"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		TempMax float64 `json:"temp_max"`
		TempMin float64 `json:"temp_min"`
	} `json:"main"`
}

const baseURL = "https://api.openweathermap.org/data/2.5/weather"

func Fetch(city, apiKey string) (*WeatherData, error) {
	return FetchWithURL(baseURL, city, apiKey)
}

func FetchWithURL(apiURL, city, apiKey string) (*WeatherData, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	q := req.URL.Query()
	q.Set("q", city)
	q.Set("appid", apiKey)
	q.Set("units", "metric")
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching weather: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding weather response: %w", err)
	}

	description := ""
	if len(apiResp.Weather) > 0 {
		description = apiResp.Weather[0].Description
	}

	return &WeatherData{
		City:        apiResp.Name,
		TempMax:     apiResp.Main.TempMax,
		TempMin:     apiResp.Main.TempMin,
		Description: description,
	}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/weather/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/weather/client.go internal/weather/client_test.go
git commit -m "feat: add weather package with OpenWeatherMap client"
```

---

### Task 4: Outfit Package

**Files:**
- Create: `internal/outfit/advisor.go`
- Create: `internal/outfit/advisor_test.go`

- [ ] **Step 1: Write the failing test**

`internal/outfit/advisor_test.go`:

```go
package outfit

import (
	"testing"

	"github.com/y-maeda1116/tenki-fuku-bot/internal/weather"
)

func TestAdvise_coldWeather(t *testing.T) {
	wd := &weather.WeatherData{City: "Tokyo", TempMax: 10.0, TempMin: 3.0}
	cats := map[string]bool{"men": true, "women": true, "kids": false}
	results := Advise(wd, cats)
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	for _, r := range results {
		if r.Outfit != "厚手のアウター（コート、ダウン）" {
			t.Errorf("Outfit = %q, want thick outerwear", r.Outfit)
		}
	}
}

func TestAdvise_warmWeather(t *testing.T) {
	wd := &weather.WeatherData{City: "Tokyo", TempMax: 28.0, TempMin: 22.0}
	cats := map[string]bool{"men": true}
	results := Advise(wd, cats)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Outfit != "半袖" {
		t.Errorf("Outfit = %q, want short sleeves", results[0].Outfit)
	}
}

func TestAdvise_mildWeather(t *testing.T) {
	wd := &weather.WeatherData{City: "Tokyo", TempMax: 18.0, TempMin: 12.0}
	cats := map[string]bool{"women": true}
	results := Advise(wd, cats)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Outfit != "薄手のジャケット、カーディガン" {
		t.Errorf("Outfit = %q, want light jacket", results[0].Outfit)
	}
}

func TestAdvise_pleasantWeather(t *testing.T) {
	wd := &weather.WeatherData{City: "Tokyo", TempMax: 23.0, TempMin: 18.0}
	cats := map[string]bool{"men": true}
	results := Advise(wd, cats)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Outfit != "長袖シャツ" {
		t.Errorf("Outfit = %q, want long sleeves", results[0].Outfit)
	}
}

func TestAdvise_largeTempDiff(t *testing.T) {
	wd := &weather.WeatherData{City: "Tokyo", TempMax: 25.0, TempMin: 12.0}
	cats := map[string]bool{"men": true}
	results := Advise(wd, cats)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Tips == "" {
		t.Error("Tips is empty, expected large temp diff advice")
	}
}

func TestAdvise_kidsExtraLayer(t *testing.T) {
	wd := &weather.WeatherData{City: "Tokyo", TempMax: 23.0, TempMin: 18.0}
	cats := map[string]bool{"kids": true}
	results := Advise(wd, cats)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	found := false
	for _, tip := range results[0].AllTips {
		if tip == "活動量を考慮して+1枚多めに着せるのがおすすめ" {
			found = true
		}
	}
	if !found {
		t.Errorf("AllTips = %v, want kids extra layer tip", results[0].AllTips)
	}
}

func TestAdvise_disabledCategory(t *testing.T) {
	wd := &weather.WeatherData{City: "Tokyo", TempMax: 20.0, TempMin: 15.0}
	cats := map[string]bool{"men": false, "women": false, "kids": false}
	results := Advise(wd, cats)
	if len(results) != 0 {
		t.Fatalf("len(results) = %d, want 0", len(results))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/outfit/...`
Expected: FAIL — `Advise`, `OutfitAdvice` not defined

- [ ] **Step 3: Write implementation**

`internal/outfit/advisor.go`:

```go
package outfit

import (
	"fmt"

	"github.com/y-maeda1116/tenki-fuku-bot/internal/weather"
)

type OutfitAdvice struct {
	Category string
	Outfit   string
	AllTips  []string
	TempMax  float64
	TempMin  float64
	TempDiff float64
}

var categoryLabels = map[string]string{
	"men":   "成人男性",
	"women": "成人女性",
	"kids":  "子供",
}

func selectOutfit(tempMax float64) string {
	switch {
	case tempMax < 15:
		return "厚手のアウター（コート、ダウン）"
	case tempMax < 20:
		return "薄手のジャケット、カーディガン"
	case tempMax < 25:
		return "長袖シャツ"
	default:
		return "半袖"
	}
}

func Advise(wd *weather.WeatherData, categories map[string]bool) []OutfitAdvice {
	var results []OutfitAdvice
	tempDiff := wd.TempMax - wd.TempMin

	for _, cat := range []string{"men", "women", "kids"} {
		if !categories[cat] {
			continue
		}

		outfit := selectOutfit(wd.TempMax)
		var tips []string

		if tempDiff >= 10 {
			tips = append(tips, "寒暖差が大きいです。脱ぎ着しやすい服装をおすすめします")
		}

		if cat == "kids" {
			tips = append(tips, "活動量を考慮して+1枚多めに着せるのがおすすめ")
		}

		results = append(results, OutfitAdvice{
			Category: cat,
			Outfit:   outfit,
			AllTips:  tips,
			TempMax:  wd.TempMax,
			TempMin:  wd.TempMin,
			TempDiff: tempDiff,
		})
	}

	return results
}

func TempColor(tempMax float64) int {
	switch {
	case tempMax < 15:
		return 0x3498DB
	case tempMax < 20:
		return 0x2ECC71
	case tempMax < 25:
		return 0xE67E22
	default:
		return 0xE74C3C
	}
}

func FormatAdvice(advice OutfitAdvice) string {
	label := categoryLabels[advice.Category]
	msg := fmt.Sprintf("**%sの服装アドバイス**\n👕 %s", label, advice.Outfit)
	for _, tip := range advice.AllTips {
		msg += fmt.Sprintf("\n💡 %s", tip)
	}
	msg += fmt.Sprintf("\n🌡️ 最高 %.1f℃ / 最低 %.1f℃（寒暖差 %.1f℃）", advice.TempMax, advice.TempMin, advice.TempDiff)
	return msg
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/outfit/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/outfit/advisor.go internal/outfit/advisor_test.go
git commit -m "feat: add outfit package with temperature-based advice logic"
```

---

### Task 5: Discord Package

**Files:**
- Create: `internal/discord/webhook.go`
- Create: `internal/discord/webhook_test.go`

- [ ] **Step 1: Write the failing test**

`internal/discord/webhook_test.go`:

```go
package discord

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/y-maeda1116/tenki-fuku-bot/internal/outfit"
	"github.com/y-maeda1116/tenki-fuku-bot/internal/weather"
)

func TestSend(t *testing.T) {
	var receivedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	wd := &weather.WeatherData{City: "Tokyo", TempMax: 18.0, TempMin: 12.0}
	advice := outfit.Advise(wd, map[string]bool{"men": true, "kids": true})

	err := SendWithURL(server.URL, advice, wd)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	embeds, ok := receivedBody["embeds"].([]interface{})
	if !ok {
		t.Fatal("embeds not found in response")
	}
	if len(embeds) != 2 {
		t.Fatalf("len(embeds) = %d, want 2", len(embeds))
	}
}

func TestSend_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	wd := &weather.WeatherData{City: "Tokyo", TempMax: 18.0, TempMin: 12.0}
	advice := outfit.Advise(wd, map[string]bool{"men": true})

	err := SendWithURL(server.URL, advice, wd)
	if err == nil {
		t.Error("Send() expected error for 500 response")
	}
}

func TestBuildEmbed(t *testing.T) {
	wd := &weather.WeatherData{City: "Tokyo", TempMax: 18.0, TempMin: 12.0}
	advice := outfit.Advise(wd, map[string]bool{"men": true})

	embed := buildEmbed(advice[0], wd)
	if embed.Title == "" {
		t.Error("embed Title is empty")
	}
	if embed.Color == 0 {
		t.Error("embed Color is 0")
	}
	if len(embed.Fields) == 0 {
		t.Error("embed has no Fields")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/discord/...`
Expected: FAIL — `SendWithURL`, `buildEmbed` not defined

- [ ] **Step 3: Write implementation**

`internal/discord/webhook.go`:

```go
package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/y-maeda1116/tenki-fuku-bot/internal/outfit"
	"github.com/y-maeda1116/tenki-fuku-bot/internal/weather"
)

type embedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type embed struct {
	Title  string       `json:"title"`
	Color  int          `json:"color"`
	Fields []embedField `json:"fields"`
}

type webhookPayload struct {
	Embeds []embed `json:"embeds"`
}

var categoryEmoji = map[string]string{
	"men":   "👔",
	"women": "👗",
	"kids":  "🧸",
}

var categoryLabel = map[string]string{
	"men":   "成人男性",
	"women": "成人女性",
	"kids":  "子供",
}

func buildEmbed(advice outfit.OutfitAdvice, wd *weather.WeatherData) embed {
	emoji := categoryEmoji[advice.Category]
	label := categoryLabel[advice.Category]

	fields := []embedField{
		{Name: "服装", Value: advice.Outfit, Inline: false},
		{Name: "最高気温", Value: fmt.Sprintf("%.1f℃", advice.TempMax), Inline: true},
		{Name: "最低気温", Value: fmt.Sprintf("%.1f℃", advice.TempMin), Inline: true},
		{Name: "寒暖差", Value: fmt.Sprintf("%.1f℃", advice.TempDiff), Inline: true},
	}

	for _, tip := range advice.AllTips {
		fields = append(fields, embedField{
			Name:   "アドバイス",
			Value:  tip,
			Inline: false,
		})
	}

	return embed{
		Title:  fmt.Sprintf("%s %sの服装アドバイス", emoji, label),
		Color:  outfit.TempColor(advice.TempMax),
		Fields: fields,
	}
}

func SendWithURL(webhookURL string, advices []outfit.OutfitAdvice, wd *weather.WeatherData) error {
	embeds := make([]embed, 0, len(advices))
	for _, a := range advices {
		embeds = append(embeds, buildEmbed(a, wd))
	}

	payload := webhookPayload{Embeds: embeds}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling webhook payload: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sending webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

func Send(webhookURL string, advices []outfit.OutfitAdvice, wd *weather.WeatherData) error {
	return SendWithURL(webhookURL, advices, wd)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/discord/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/discord/webhook.go internal/discord/webhook_test.go
git commit -m "feat: add discord package with webhook embeds notification"
```

---

### Task 6: Entry Point

**Files:**
- Create: `cmd/cli/main.go`

- [ ] **Step 1: Write entry point**

`cmd/cli/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/y-maeda1116/tenki-fuku-bot/internal/config"
	"github.com/y-maeda1116/tenki-fuku-bot/internal/discord"
	"github.com/y-maeda1116/tenki-fuku-bot/internal/outfit"
	"github.com/y-maeda1116/tenki-fuku-bot/internal/weather"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("WEATHER_API_KEY is not set")
	}
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("DISCORD_WEBHOOK_URL is not set")
	}

	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	wd, err := weather.Fetch(cfg.City, apiKey)
	if err != nil {
		return fmt.Errorf("fetching weather: %w", err)
	}

	cats := map[string]bool{
		"men":   cfg.Categories.Men,
		"women": cfg.Categories.Women,
		"kids":  cfg.Categories.Kids,
	}
	advices := outfit.Advise(wd, cats)
	if len(advices) == 0 {
		fmt.Println("No categories enabled, skipping notification")
		return nil
	}

	if err := discord.Send(webhookURL, advices, wd); err != nil {
		return fmt.Errorf("sending discord notification: %w", err)
	}

	fmt.Printf("Notification sent for %s (%.1f℃/%.1f℃)\n", wd.City, wd.TempMax, wd.TempMin)
	return nil
}
```

- [ ] **Step 2: Verify build succeeds**

Run: `go build ./cmd/cli/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add cmd/cli/main.go
git commit -m "feat: add entry point for tenki-fuku-bot"
```

---

### Task 7: GitHub Actions Workflows

**Files:**
- Create: `.github/workflows/notify.yml`
- Modify: `.github/workflows/test.yml`

- [ ] **Step 1: Create notify.yml**

`.github/workflows/notify.yml`:

```yaml
name: Daily Weather Notification

on:
  schedule:
    - cron: '0 21 * * *'  # JST 06:00 = UTC 21:00
  workflow_dispatch:

jobs:
  notify:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v6

    - name: Set up Go
      uses: actions/setup-go@v6
      with:
        go-version-file: 'go.mod'

    - name: Run tenki-fuku-bot
      env:
        WEATHER_API_KEY: ${{ secrets.WEATHER_API_KEY }}
        DISCORD_WEBHOOK_URL: ${{ secrets.DISCORD_WEBHOOK_URL }}
      run: go run cmd/cli/main.go
```

- [ ] **Step 2: Update test.yml**

`.github/workflows/test.yml`:

```yaml
name: Test

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v6

    - name: Set up Go
      uses: actions/setup-go@v6
      with:
        go-version-file: 'go.mod'

    - name: Install dependencies
      run: go mod download

    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./internal/...
```

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/notify.yml .github/workflows/test.yml
git commit -m "ci: add daily notification workflow and update test workflow"
```

---

### Task 8: Update Supporting Files

**Files:**
- Modify: `Makefile`
- Modify: `.gitignore`

- [ ] **Step 1: Rewrite Makefile**

```makefile
.PHONY: build run test test-coverage clean fmt lint help

APP_NAME := tenki-fuku-bot
BIN_DIR := bin

# --- Build ---

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/cli

# --- Run ---

run:
	@echo "Running $(APP_NAME)..."
	@go run ./cmd/cli

# --- Test ---

test:
	@echo "Running tests..."
	@go test -v ./internal/...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./internal/...
	@go tool cover -html=coverage.out -o coverage.html

# --- Format / Lint ---

fmt:
	@go fmt ./...

lint:
	@echo "Running linter..."
	@golangci-lint run ./internal/...

# --- Clean ---

clean:
	@rm -rf $(BIN_DIR) coverage.out coverage.html

# --- Help ---

help:
	@echo "Available targets:"
	@echo "  build          - Build the bot"
	@echo "  run            - Run the bot"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  fmt            - Format Go code"
	@echo "  lint           - Run linter"
	@echo "  clean          - Remove build artifacts"
```

- [ ] **Step 2: Update .gitignore**

Append to `.gitignore`:
```
# Config with secrets
config/config.yaml
```

Remove old template-specific entries that are no longer relevant (frontend, wails, etc). Final `.gitignore`:

```
# Binaries
bin/
*.exe
*.dll
*.so
*.dylib

# Test coverage
coverage.out
coverage.html

# Environment
.env
.env.*
!.env.example
*.log

# Config (may contain secrets — example only)
config/config.yaml

# Editor
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
```

- [ ] **Step 3: Commit**

```bash
git add Makefile .gitignore
git commit -m "chore: update Makefile and gitignore for tenki-fuku-bot"
```

---

### Task 9: Full Test Suite and Verification

**Files:**
- None (verification only)

- [ ] **Step 1: Run all tests**

Run: `go test -v ./internal/...`
Expected: All PASS

- [ ] **Step 2: Verify build**

Run: `go build ./cmd/cli/...`
Expected: no errors

- [ ] **Step 3: Verify go vet**

Run: `go vet ./...`
Expected: no output (no issues)

- [ ] **Step 4: Final commit (if any fixes needed)**

Fix any issues found and commit.
