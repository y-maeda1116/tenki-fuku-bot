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
				"dt_txt": tomorrow + " 09:00:00",
				"main":   map[string]interface{}{"temp_max": 22.5, "temp_min": 18.0},
				"weather": []map[string]interface{}{
					{"description": "scattered clouds"},
				},
			},
			{
				"dt_txt": tomorrow + " 12:00:00",
				"main":   map[string]interface{}{"temp_max": 26.3, "temp_min": 22.1},
				"weather": []map[string]interface{}{
					{"description": "clear sky"},
				},
			},
			{
				"dt_txt": tomorrow + " 18:00:00",
				"main":   map[string]interface{}{"temp_max": 20.0, "temp_min": 15.3},
				"weather": []map[string]interface{}{
					{"description": "clear sky"},
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
	if data.TempMax != 26.3 {
		t.Errorf("TempMax = %f, want 26.3", data.TempMax)
	}
	if data.TempMin != 15.3 {
		t.Errorf("TempMin = %f, want 15.3", data.TempMin)
	}
	if data.Description != "clear sky" {
		t.Errorf("Description = %q, want %q", data.Description, "clear sky")
	}
	if data.Date != tomorrow {
		t.Errorf("Date = %q, want %q", data.Date, tomorrow)
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
				"main":   map[string]interface{}{"temp_max": 20.0, "temp_min": 15.0},
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
