package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	WeatherText string
	Date        string
	TimeSlots   []TimeSlot
	TodayMax    float64
	TodayMin    float64
}

// jst は日本標準時 (UTC+9)。日本にサマータイムはないため固定オフセットで足りる。
var jst = time.FixedZone("JST", 9*60*60)

const forecastURLBase = "https://www.jma.go.jp/bosai/forecast/data/forecast"

// JMA forecast API のレスポンス (必要なフィールドのみ)
type jmaReport struct {
	TimeSeries []jmaTimeSeries `json:"timeSeries"`
}

type jmaTimeSeries struct {
	TimeDefines []string  `json:"timeDefines"`
	Areas       []jmaArea `json:"areas"`
}

type jmaArea struct {
	Area struct {
		Name string `json:"name"`
	} `json:"area"`
	Weathers []string `json:"weathers"`
	Temps    []string `json:"temps"`
}

// FetchTomorrow fetches tomorrow's weather forecast for the given JMA area code.
// 天気は weatherArea (例: 東京地方)、気温は tempArea (例: 東京) のエリア名で取得する。
func FetchTomorrow(areaCode, weatherArea, tempArea string) (*WeatherData, error) {
	apiURL := fmt.Sprintf("%s/%s.json", forecastURLBase, areaCode)
	return FetchTomorrowWithURL(apiURL, weatherArea, tempArea)
}

// FetchTomorrowWithURL fetches tomorrow's forecast using a custom API URL (for testing).
func FetchTomorrowWithURL(apiURL, weatherArea, tempArea string) (*WeatherData, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching forecast: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("forecast API returned status %d", resp.StatusCode)
	}

	var reports []jmaReport
	if err := json.NewDecoder(resp.Body).Decode(&reports); err != nil {
		return nil, fmt.Errorf("decoding forecast response: %w", err)
	}
	if len(reports) == 0 {
		return nil, fmt.Errorf("empty forecast response")
	}

	// 今日/明日は JST 基準で判定する
	now := time.Now().In(jst)
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
	today := now.Format("2006-01-02")

	ts := reports[0].TimeSeries
	if len(ts) < 3 {
		return nil, fmt.Errorf("unexpected forecast response: got %d timeSeries, want >= 3", len(ts))
	}

	weatherText, err := extractWeather(ts[0], weatherArea, tomorrow)
	if err != nil {
		return nil, err
	}

	maxTemp, minTemp, todayMax, todayMin, err := extractTemps(ts[2], tempArea, today, tomorrow)
	if err != nil {
		return nil, err
	}

	// センチネル値が残っている場合は明日のデータが欠損している
	if maxTemp == -100 || minTemp == 100 {
		return nil, fmt.Errorf("no forecast data available for tomorrow (%s)", tomorrow)
	}

	return &WeatherData{
		City:        weatherArea,
		TempMax:     maxTemp,
		TempMin:     minTemp,
		WeatherText: weatherText,
		Date:        tomorrow,
		TodayMax:    todayMax,
		TodayMin:    todayMin,
	}, nil
}

// extractWeather は天気シリーズ (timeSeries[0]) から明日の天気文を取り出す。
// 全角スペースは半角に正規化する。
func extractWeather(ts jmaTimeSeries, weatherArea, tomorrow string) (string, error) {
	area, err := findArea(ts.Areas, weatherArea)
	if err != nil {
		return "", fmt.Errorf("weather series: %w", err)
	}
	for i, td := range ts.TimeDefines {
		if len(td) >= 10 && td[:10] == tomorrow && i < len(area.Weathers) {
			return strings.Join(strings.Fields(area.Weathers[i]), " "), nil
		}
	}
	return "", fmt.Errorf("no weather data for tomorrow (%s)", tomorrow)
}

// extractTemps は気温シリーズ (timeSeries[2]) から今日/明日の最高・最低気温を取り出す。
// JMA の気温は timeDefines が 00時 (最低) と 09時 (最高) の組で並ぶ。空文字は欠損としてスキップする。
func extractTemps(ts jmaTimeSeries, tempArea, today, tomorrow string) (maxTemp, minTemp, todayMax, todayMin float64, err error) {
	maxTemp, minTemp, todayMax, todayMin = -100, 100, -100, 100

	area, err := findArea(ts.Areas, tempArea)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("temp series: %w", err)
	}

	for i, td := range ts.TimeDefines {
		if len(td) < 16 || i >= len(area.Temps) {
			continue
		}
		datePart := td[:10]
		hour := td[11:13]
		if area.Temps[i] == "" {
			continue
		}
		temp, convErr := parseFloat(area.Temps[i])
		if convErr != nil {
			continue
		}
		switch {
		case datePart == today && hour == "00":
			todayMin = temp
		case datePart == today && hour == "09":
			todayMax = temp
		case datePart == tomorrow && hour == "00":
			minTemp = temp
		case datePart == tomorrow && hour == "09":
			maxTemp = temp
		}
	}
	return maxTemp, minTemp, todayMax, todayMin, nil
}

func findArea(areas []jmaArea, name string) (*jmaArea, error) {
	for i := range areas {
		if areas[i].Area.Name == name {
			return &areas[i], nil
		}
	}
	return nil, fmt.Errorf("area %q not found", name)
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%g", &f)
	return f, err
}
