package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchTomorrow(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	response := map[string]interface{}{
		"city": map[string]interface{}{
			"name": "Tokyo",
		},
		"list": []map[string]interface{}{
			{
				"dt_txt": tomorrow + " 06:00:00",
				"main":   map[string]interface{}{"temp": 15.0},
				"weather": []map[string]interface{}{
					{"description": "薄い雲"},
				},
			},
			{
				"dt_txt": tomorrow + " 12:00:00",
				"main":   map[string]interface{}{"temp": 22.5},
				"weather": []map[string]interface{}{
					{"description": "晴れ"},
				},
			},
			{
				"dt_txt": tomorrow + " 15:00:00",
				"main":   map[string]interface{}{"temp": 18.3},
				"weather": []map[string]interface{}{
					{"description": "くもり"},
				},
			},
			{
				"dt_txt": tomorrow + " 18:00:00",
				"main":   map[string]interface{}{"temp": 13.0},
				"weather": []map[string]interface{}{
					{"description": "くもり"},
				},
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
	if len(data.TimeSlots) != 3 {
		t.Fatalf("len(TimeSlots) = %d, want 3", len(data.TimeSlots))
	}
	if data.TimeSlots[0].Temp != 15.0 {
		t.Errorf("TimeSlots[0].Temp = %f, want 15.0", data.TimeSlots[0].Temp)
	}
	if data.TimeSlots[0].Time != "朝 (7時)" {
		t.Errorf("TimeSlots[0].Time = %q, want %q", data.TimeSlots[0].Time, "朝 (7時)")
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
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	response := map[string]interface{}{
		"city": map[string]interface{}{
			"name": "Tokyo",
		},
		"list": []map[string]interface{}{
			{
				"dt_txt": yesterday + " 12:00:00",
				"main":   map[string]interface{}{"temp": 20.0},
				"weather": []map[string]interface{}{
					{"description": "clear sky"},
				},
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
