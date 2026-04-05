package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TimeSlot struct {
	Time        string
	Description string
	Temp        float64
}

type WeatherData struct {
	City        string
	TempMax     float64
	TempMin     float64
	Description string
	Date        string
	TimeSlots   []TimeSlot
}

type forecastItem struct {
	DtTxt   string `json:"dt_txt"`
	Main    struct {
		Temp float64 `json:"temp"`
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

var targetTimes = []string{"06:00:00", "12:00:00", "15:00:00"}

var timeLabels = map[string]string{
	"06:00:00": "朝 (7時)",
	"12:00:00": "昼 (12時)",
	"15:00:00": "夕 (17時)",
}

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
	topDesc := ""
	topCount := 0
	var slots []TimeSlot

	for _, item := range fcResp.List {
		if len(item.DtTxt) < 10 || item.DtTxt[:10] != tomorrow {
			continue
		}
		if item.Main.Temp > maxTemp {
			maxTemp = item.Main.Temp
		}
		if item.Main.Temp < minTemp {
			minTemp = item.Main.Temp
		}
		if len(item.Weather) > 0 {
			d := item.Weather[0].Description
			descCount[d]++
			if descCount[d] > topCount {
				topCount = descCount[d]
				topDesc = d
			}
		}

		timePart := item.DtTxt[11:]
		for _, tt := range targetTimes {
			if timePart == tt {
				desc := ""
				if len(item.Weather) > 0 {
					desc = item.Weather[0].Description
				}
				slots = append(slots, TimeSlot{
					Time:        timeLabels[tt],
					Description: desc,
					Temp:        item.Main.Temp,
				})
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
		TimeSlots:   slots,
	}, nil
}
