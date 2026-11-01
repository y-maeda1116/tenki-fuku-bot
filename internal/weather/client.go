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
	TodayMax    float64
	TodayMin    float64
}

type forecastItem struct {
	DtTxt string `json:"dt_txt"`
	Main  struct {
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

// jst は日本標準時 (UTC+9)。日本にサマータイムはないため固定オフセットで足りる。
var jst = time.FixedZone("JST", 9*60*60)

// targetSlots は通知する JST の時間帯とラベル。OpenWeatherMap の予報は
// 3時間刻み (JST の 0/3/6/9/12/15/18/21時) のため、7時/17時ちょうどは取れない。
var targetSlots = []struct {
	hour  int
	label string
}{
	{9, "朝 (9時)"},
	{12, "昼 (12時)"},
	{18, "夕 (18時)"},
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

	// 今日/明日は JST 基準で判定する (dt_txt は UTC のため JST に変換して比較)
	now := time.Now().In(jst)
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
	today := now.Format("2006-01-02")

	var maxTemp float64 = -100
	var minTemp float64 = 100
	var todayMax float64 = -100
	var todayMin float64 = 100
	descCount := map[string]int{}
	topDesc := ""
	topCount := 0
	var slots []TimeSlot

	for _, item := range fcResp.List {
		jstTime, err := time.ParseInLocation("2006-01-02 15:04:05", item.DtTxt, time.UTC)
		if err != nil {
			continue
		}
		jstTime = jstTime.In(jst)
		datePart := jstTime.Format("2006-01-02")

		if datePart == today {
			if item.Main.Temp > todayMax {
				todayMax = item.Main.Temp
			}
			if item.Main.Temp < todayMin {
				todayMin = item.Main.Temp
			}
		}

		if datePart != tomorrow {
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

		for _, s := range targetSlots {
			if jstTime.Hour() == s.hour {
				desc := ""
				if len(item.Weather) > 0 {
					desc = item.Weather[0].Description
				}
				slots = append(slots, TimeSlot{
					Time:        s.label,
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
		TodayMax:    todayMax,
		TodayMin:    todayMin,
	}, nil
}
