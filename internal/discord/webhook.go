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
	"men":   "成人男性",
	"women": "成人女性",
	"kids":  "子供",
}

func buildEmbed(advice outfit.OutfitAdvice, wd *weather.WeatherData) embed {
	emoji := categoryEmoji[advice.Category]
	label := categoryLabel[advice.Category]

	fields := []embedField{
		{Name: "服装", Value: advice.Outfit, Inline: false},
		{Name: "最高気温", Value: fmt.Sprintf("%.1f℃", advice.TempMax), Inline: true},
		{Name: "最低気温", Value: fmt.Sprintf("%.1f℃", advice.TempMin), Inline: true},
		{Name: "寒暖差", Value: fmt.Sprintf("%.1f℃", advice.TempDiff), Inline: true},
	}

	for _, tip := range advice.AllTips {
		fields = append(fields, embedField{
			Name:   "アドバイス",
			Value:  tip,
			Inline: false,
		})
	}

	return embed{
		Title:  fmt.Sprintf("%s %sの服装アドバイス", emoji, label),
		Color:  outfit.TempColor(advice.TempMax),
		Fields: fields,
	}
}

func SendWithURL(webhookURL string, advices []outfit.OutfitAdvice, wd *weather.WeatherData) error {
	embeds := make([]embed, 0, len(advices))
	for _, a := range advices {
		embeds = append(embeds, buildEmbed(a, wd))
	}

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
