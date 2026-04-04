package main

import (
	"fmt"
	"os"

	"github.com/y-maeda1116/tenki-fuku-bot/internal/config"
	"github.com/y-maeda1116/tenki-fuku-bot/internal/discord"
	"github.com/y-maeda1116/tenki-fuku-bot/internal/outfit"
	"github.com/y-maeda1116/tenki-fuku-bot/internal/weather"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("WEATHER_API_KEY is not set")
	}
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("DISCORD_WEBHOOK_URL is not set")
	}

	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	wd, err := weather.FetchTomorrow(cfg.City, apiKey)
	if err != nil {
		return fmt.Errorf("fetching weather: %w", err)
	}

	cats := map[string]bool{
		"men":   cfg.Categories.Men,
		"women": cfg.Categories.Women,
		"kids":  cfg.Categories.Kids,
	}
	advices := outfit.Advise(wd, cats)
	if len(advices) == 0 {
		fmt.Println("No categories enabled, skipping notification")
		return nil
	}

	if err := discord.Send(webhookURL, advices, wd); err != nil {
		return fmt.Errorf("sending discord notification: %w", err)
	}

	fmt.Printf("Notification sent for %s tomorrow %s (%.1f℃/%.1f℃)\n", wd.City, wd.Date, wd.TempMax, wd.TempMin)
	return nil
}
