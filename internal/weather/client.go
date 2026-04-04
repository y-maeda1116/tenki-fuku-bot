package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WeatherData struct {
	City        string
	TempMax     float64
	TempMin     float64
	Description string
	Date        string
}

type forecastItem struct {
	DtTxt   string `json:"dt_txt"`
	Main    struct {
		TempMax float64 `json:"temp_max"`
		TempMin float64 `json:"temp_min"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
}

type forecastResponse struct {
	City struct {
		Name string `json:"name"`
	} `json:"city"`
	List []forecastItem `json:"list"`
}

const forecastURL = "https://api.openweathermap.org/data/2.5/forecast"

// FetchTomorrow fetches tomorrow's weather forecast for the given city.
func FetchTomorrow(city, apiKey string) (*WeatherData, error) {
	return FetchTomorrowWithURL(forecastURL, city, apiKey)
}

// FetchTomorrowWithURL fetches tomorrow's forecast using a custom API URL (for testing).
func FetchTomorrowWithURL(apiURL, city, apiKey string) (*WeatherData, error) {
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
		return nil, fmt.Errorf("fetching forecast: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("forecast API returned status %d", resp.StatusCode)
	}

	var fcResp forecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&fcResp); err != nil {
		return nil, fmt.Errorf("decoding forecast response: %w", err)
	}

	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	var maxTemp float64 = -100
	var minTemp float64 = 100
	descCount := map[string]int{}
	var topDesc string
	topCount := 0

	for _, item := range fcResp.List {
		if len(item.DtTxt) < 10 || item.DtTxt[:10] != tomorrow {
			continue
		}
		if item.Main.TempMax > maxTemp {
			maxTemp = item.Main.TempMax
		}
		if item.Main.TempMin < minTemp {
			minTemp = item.Main.TempMin
		}
		if len(item.Weather) > 0 {
			d := item.Weather[0].Description
			descCount[d]++
			if descCount[d] > topCount {
				topCount = descCount[d]
				topDesc = d
			}
		}
	}

	if maxTemp == -100 {
		return nil, fmt.Errorf("no forecast data available for tomorrow (%s)", tomorrow)
	}

	return &WeatherData{
		City:        fcResp.City.Name,
		TempMax:     maxTemp,
		TempMin:     minTemp,
		Description: topDesc,
		Date:        tomorrow,
	}, nil
}
