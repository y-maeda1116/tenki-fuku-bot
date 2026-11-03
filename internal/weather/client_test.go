package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// jmaFixture は JMA forecast API のレスポンス構造を模した JSON を返す
func jmaFixture(t *testing.T, today, tomorrow string) []byte {
	t.Helper()
	resp := []map[string]interface{}{
		{
			"timeSeries": []map[string]interface{}{
				{
					// 天気: [今日, 明日]
					"timeDefines": []string{
						today + "T05:00:00+09:00",
						tomorrow + "T00:00:00+09:00",
					},
					"areas": []map[string]interface{}{
						{
							"area":     map[string]interface{}{"name": "東京地方"},
							"weathers": []string{"晴れ 昼過ぎ から くもり", "くもり　夜遅く　雨"},
						},
					},
				},
				{
					// 降水確率 (この実装では未使用)
					"timeDefines": []string{tomorrow + "T00:00:00+09:00"},
					"areas": []map[string]interface{}{
						{
							"area": map[string]interface{}{"name": "東京地方"},
							"pops": []string{"50"},
						},
					},
				},
				{
					// 気温: [今日 最高(09時), 今日 最低(00時), 明日 最低(00時), 明日 最高(09時)]
					"timeDefines": []string{
						today + "T09:00:00+09:00",
						today + "T00:00:00+09:00",
						tomorrow + "T00:00:00+09:00",
						tomorrow + "T09:00:00+09:00",
					},
					"areas": []map[string]interface{}{
						{
							"area":  map[string]interface{}{"name": "東京"},
							"temps": []string{"19", "11", "12", "18"},
						},
					},
				},
			},
		},
	}
	body, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshaling fixture: %v", err)
	}
	return body
}

func jmaTestServer(t *testing.T, body []byte) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
	t.Cleanup(server.Close)
	return server
}

func tomorrowDates() (today, tomorrow string) {
	now := time.Now().In(jst)
	return now.Format("2006-01-02"), now.AddDate(0, 0, 1).Format("2006-01-02")
}

func TestFetchTomorrow(t *testing.T) {
	today, tomorrow := tomorrowDates()
	server := jmaTestServer(t, jmaFixture(t, today, tomorrow))

	data, err := FetchTomorrowWithURL(server.URL, "130000", "東京地方", "東京")
	if err != nil {
		t.Fatalf("FetchTomorrow() error = %v", err)
	}
	if data.City != "東京地方" {
		t.Errorf("City = %q, want 東京地方", data.City)
	}
	if data.Date != tomorrow {
		t.Errorf("Date = %q, want %q", data.Date, tomorrow)
	}
	// 全角スペースは半角に正規化される
	if data.WeatherText != "くもり 夜遅く 雨" {
		t.Errorf("WeatherText = %q, want %q", data.WeatherText, "くもり 夜遅く 雨")
	}
	if data.TempMax != 18.0 {
		t.Errorf("TempMax = %f, want 18.0", data.TempMax)
	}
	if data.TempMin != 12.0 {
		t.Errorf("TempMin = %f, want 12.0", data.TempMin)
	}
	if data.TodayMax != 19.0 {
		t.Errorf("TodayMax = %f, want 19.0", data.TodayMax)
	}
	if data.TodayMin != 11.0 {
		t.Errorf("TodayMin = %f, want 11.0", data.TodayMin)
	}
}

func TestFetchTomorrow_missingTodayTempEntry(t *testing.T) {
	// 気温の一部が空文字の場合はその項目をスキップする (今日の最低が欠損)
	today, tomorrow := tomorrowDates()

	body := jmaFixture(t, today, tomorrow)
	var resp []map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	temps := resp[0]["timeSeries"].([]interface{})[2].(map[string]interface{})["areas"].([]interface{})[0].(map[string]interface{})
	temps["temps"] = []string{"19", "", "12", "18"}
	mutated, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("re-marshaling fixture: %v", err)
	}
	body = mutated

	server := jmaTestServer(t, body)
	data, err := FetchTomorrowWithURL(server.URL, "130000", "東京地方", "東京")
	if err != nil {
		t.Fatalf("FetchTomorrow() error = %v", err)
	}
	if data.TempMax != 18.0 {
		t.Errorf("TempMax = %f, want 18.0", data.TempMax)
	}
	if data.TempMin != 12.0 {
		t.Errorf("TempMin = %f, want 12.0", data.TempMin)
	}
	if data.TodayMax != 19.0 {
		t.Errorf("TodayMax = %f, want 19.0", data.TodayMax)
	}
	if data.TodayMin != 100 {
		t.Errorf("TodayMin = %f, want 100 (欠損)", data.TodayMin)
	}
}

func TestFetchTomorrow_serverError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := FetchTomorrowWithURL(server.URL, "130000", "東京地方", "東京")
	if err == nil {
		t.Error("FetchTomorrow() expected error for 500 response")
	}
}

func TestFetchTomorrow_noDataForTomorrow(t *testing.T) {
	// 明日のデータが一切ないレスポンス
	now := time.Now().In(jst)
	yst := now.AddDate(0, 0, -1).Format("2006-01-02")
	today := now.Format("2006-01-02")

	body := jmaFixture(t, yst, today)
	server := jmaTestServer(t, body)

	_, err := FetchTomorrowWithURL(server.URL, "130000", "東京地方", "東京")
	if err == nil {
		t.Error("FetchTomorrow() expected error when no data for tomorrow")
	}
}
