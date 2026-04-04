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
