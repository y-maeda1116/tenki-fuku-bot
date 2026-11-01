package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var testJST = time.FixedZone("JST", 9*60*60)

// dtTxtUTC は JST 時刻を OpenWeatherMap の dt_txt 形式 (UTC) に変換する
func dtTxtUTC(t time.Time) string {
	return t.In(time.UTC).Format("2006-01-02 15:04:05")
}

func TestFetchTomorrow(t *testing.T) {
	now := time.Now().In(testJST)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, testJST)
	tomorrow := today.AddDate(0, 0, 1)
	dayAfter := today.AddDate(0, 0, 2)

	// dt_txt は UTC で返るため、JST の各時刻に対応する UTC 文字列を渡す
	response := map[string]interface{}{
		"city": map[string]interface{}{
			"name": "Tokyo",
		},
		"list": []map[string]interface{}{
			{
				// 今日 JST 15時 → 今日の最高/最低気温にのみ反映
				"dt_txt":  dtTxtUTC(time.Date(today.Year(), today.Month(), today.Day(), 15, 0, 0, 0, testJST)),
				"main":    map[string]interface{}{"temp": 10.0},
				"weather": []map[string]interface{}{{"description": "晴れ"}},
			},
			{
				// 明日 JST 9時 → 朝スロット
				"dt_txt":  dtTxtUTC(time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 9, 0, 0, 0, testJST)),
				"main":    map[string]interface{}{"temp": 15.0},
				"weather": []map[string]interface{}{{"description": "薄い雲"}},
			},
			{
				// 明日 JST 12時 → 昼スロット
				"dt_txt":  dtTxtUTC(time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 12, 0, 0, 0, testJST)),
				"main":    map[string]interface{}{"temp": 22.5},
				"weather": []map[string]interface{}{{"description": "晴れ"}},
			},
			{
				// 明日 JST 18時 → 夕スロット
				"dt_txt":  dtTxtUTC(time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 18, 0, 0, 0, testJST)),
				"main":    map[string]interface{}{"temp": 18.3},
				"weather": []map[string]interface{}{{"description": "くもり"}},
			},
			{
				// 明日 JST 21時 → スロット外だが最高/最低気温には反映
				"dt_txt":  dtTxtUTC(time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 21, 0, 0, 0, testJST)),
				"main":    map[string]interface{}{"temp": 13.0},
				"weather": []map[string]interface{}{{"description": "くもり"}},
			},
			{
				// 明日 JST 24時 (= 明日 UTC 15時) → JST では明後日なので明日の集計に含まれない
				"dt_txt":  dtTxtUTC(time.Date(dayAfter.Year(), dayAfter.Month(), dayAfter.Day(), 0, 0, 0, 0, testJST)),
				"main":    map[string]interface{}{"temp": 30.0},
				"weather": []map[string]interface{}{{"description": "雨"}},
			},
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

	data, err := FetchTomorrowWithURL(server.URL, "Tokyo", "test-key")
	if err != nil {
		t.Fatalf("FetchTomorrow() error = %v", err)
	}
	if data.City != "Tokyo" {
		t.Errorf("City = %q, want %q", data.City, "Tokyo")
	}
	if data.TempMax != 22.5 {
		t.Errorf("TempMax = %f, want 22.5", data.TempMax)
	}
	if data.TempMin != 13.0 {
		t.Errorf("TempMin = %f, want 13.0", data.TempMin)
	}
	if data.TodayMax != 10.0 {
		t.Errorf("TodayMax = %f, want 10.0", data.TodayMax)
	}
	if data.TodayMin != 10.0 {
		t.Errorf("TodayMin = %f, want 10.0", data.TodayMin)
	}

	wantSlots := []TimeSlot{
		{Time: "朝 (9時)", Description: "薄い雲", Temp: 15.0},
		{Time: "昼 (12時)", Description: "晴れ", Temp: 22.5},
		{Time: "夕 (18時)", Description: "くもり", Temp: 18.3},
	}
	if len(data.TimeSlots) != len(wantSlots) {
		t.Fatalf("len(TimeSlots) = %d, want %d", len(data.TimeSlots), len(wantSlots))
	}
	for i, want := range wantSlots {
		got := data.TimeSlots[i]
		if got != want {
			t.Errorf("TimeSlots[%d] = %+v, want %+v", i, got, want)
		}
	}
}

func TestFetchTomorrow_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := FetchTomorrowWithURL(server.URL, "Tokyo", "test-key")
	if err == nil {
		t.Error("FetchTomorrow() expected error for 500 response")
	}
}

func TestFetchTomorrow_noDataForTomorrow(t *testing.T) {
	now := time.Now().In(testJST)
	yesterday := now.AddDate(0, 0, -1)
	response := map[string]interface{}{
		"city": map[string]interface{}{
			"name": "Tokyo",
		},
		"list": []map[string]interface{}{
			{
				"dt_txt":  dtTxtUTC(time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 12, 0, 0, 0, testJST)),
				"main":    map[string]interface{}{"temp": 20.0},
				"weather": []map[string]interface{}{{"description": "clear sky"}},
			},
		},
	}
	body, _ := json.Marshal(response)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
	defer server.Close()

	_, err := FetchTomorrowWithURL(server.URL, "Tokyo", "test-key")
	if err == nil {
		t.Error("FetchTomorrow() expected error when no data for tomorrow")
	}
}
