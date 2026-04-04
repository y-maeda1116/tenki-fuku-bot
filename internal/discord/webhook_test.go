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
	advice := []outfit.OutfitAdvice{
		{
			Category: "men",
			Outfit:   "薄手のジャケット、カーディガン",
			AllTips:  []string{"寒暖差が大きいです。脱ぎ着しやすい服装をおすすめします"},
			TempMax:  18.0,
			TempMin:  12.0,
			TempDiff: 6.0,
		},
		{
			Category: "kids",
			Outfit:   "薄手のジャケット、カーディガン",
			AllTips:  []string{"活動量を考慮して+1枚多めに着せるのがおすすめ"},
			TempMax:  18.0,
			TempMin:  12.0,
			TempDiff: 6.0,
		},
	}

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
	advice := []outfit.OutfitAdvice{
		{
			Category: "men",
			Outfit:   "薄手のジャケット、カーディガン",
			AllTips:  nil,
			TempMax:  18.0,
			TempMin:  12.0,
			TempDiff: 6.0,
		},
	}

	err := SendWithURL(server.URL, advice, wd)
	if err == nil {
		t.Error("Send() expected error for 500 response")
	}
}

func TestBuildEmbed(t *testing.T) {
	wd := &weather.WeatherData{City: "Tokyo", TempMax: 18.0, TempMin: 12.0, Description: "薄い雲"}
	advice := outfit.OutfitAdvice{
		Category: "men",
		Outfit:   "薄手のジャケット、カーディガン",
		AllTips:  nil,
		TempMax:  18.0,
		TempMin:  12.0,
		TempDiff: 6.0,
	}

	embed := buildEmbed(advice, wd)
	if embed.Title == "" {
		t.Error("embed Title is empty")
	}
	if embed.Color == 0 {
		t.Error("embed Color is 0")
	}
	if len(embed.Fields) == 0 {
		t.Error("embed has no Fields")
	}

	// Verify weather description field exists
	found := false
	for _, f := range embed.Fields {
		if f.Name == "天気" {
			found = true
			if f.Value != "薄い雲" {
				t.Errorf("天気 field = %q, want %q", f.Value, "薄い雲")
			}
		}
	}
	if !found {
		t.Error("embed missing 天気 field")
	}
}
