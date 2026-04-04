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
	if results[0].AllTips == nil || len(results[0].AllTips) == 0 {
		t.Error("AllTips is empty, expected large temp diff advice")
	}
	found := false
	for _, tip := range results[0].AllTips {
		if tip == "寒暖差が大きいです。脱ぎ着しやすい服装をおすすめします" {
			found = true
		}
	}
	if !found {
		t.Errorf("AllTips = %v, want large temp diff tip", results[0].AllTips)
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
