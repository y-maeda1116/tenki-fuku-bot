package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/y-maeda1116/tenki-fuku-bot/internal/outfit"
	"github.com/y-maeda1116/tenki-fuku-bot/internal/weather"
)

type embedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type embed struct {
	Title  string       `json:"title"`
	Color  int          `json:"color"`
	Fields []embedField `json:"fields"`
}

type webhookPayload struct {
	Embeds []embed `json:"embeds"`
}

var categoryEmoji = map[string]string{
	"men":   "👔",
	"women": "👗",
	"kids":  "🧸",
}

var categoryLabel = map[string]string{
	"men":   "男性",
	"women": "女性",
	"kids":  "子ども",
}

func buildWeatherEmbed(wd *weather.WeatherData) embed {
	fields := []embedField{}
	for _, slot := range wd.TimeSlots {
		fields = append(fields, embedField{
			Name:   slot.Time,
			Value:  fmt.Sprintf("%s  %.1f℃", slot.Description, slot.Temp),
			Inline: true,
		})
	}
	fields = append(fields,
		embedField{Name: "最高", Value: fmt.Sprintf("%.1f℃", wd.TempMax), Inline: true},
		embedField{Name: "最低", Value: fmt.Sprintf("%.1f℃", wd.TempMin), Inline: true},
		embedField{Name: "寒暖差", Value: fmt.Sprintf("%.1f℃", wd.TempMax-wd.TempMin), Inline: true},
	)

	return embed{
		Title:  fmt.Sprintf("🌤 明日の天気（%s）", wd.City),
		Color:  outfit.TempColor(wd.TempMax),
		Fields: fields,
	}
}

func buildOutfitEmbed(advice outfit.OutfitAdvice) embed {
	emoji := categoryEmoji[advice.Category]
	label := categoryLabel[advice.Category]

	fields := []embedField{
		{Name: "服装", Value: advice.Outfit, Inline: false},
	}

	for _, tip := range advice.AllTips {
		fields = append(fields, embedField{
			Name:   "アドバイス",
			Value:  tip,
			Inline: false,
		})
	}

	return embed{
		Title:  fmt.Sprintf("%s %s", emoji, label),
		Color:  outfit.TempColor(advice.TempMax),
		Fields: fields,
	}
}

func buildEmbeds(advices []outfit.OutfitAdvice, wd *weather.WeatherData) []embed {
	embeds := []embed{buildWeatherEmbed(wd)}
	for _, a := range advices {
		embeds = append(embeds, buildOutfitEmbed(a))
	}
	return embeds
}

func SendWithURL(webhookURL string, advices []outfit.OutfitAdvice, wd *weather.WeatherData) error {
	embeds := buildEmbeds(advices, wd)

	payload := webhookPayload{Embeds: embeds}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling webhook payload: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sending webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

func Send(webhookURL string, advices []outfit.OutfitAdvice, wd *weather.WeatherData) error {
	return SendWithURL(webhookURL, advices, wd)
}
