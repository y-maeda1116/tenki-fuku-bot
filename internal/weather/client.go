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
