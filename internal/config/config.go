package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 設定構造体
type Config struct {
	App     AppConfig
	Server  ServerConfig
	Logging LoggingConfig
	CLI     CLIConfig
	Desktop DesktopConfig
}

// AppConfig アプリケーション設定
type AppConfig struct {
	Name        string
	Version     string
	Environment string
}

// ServerConfig サーバー設定
type ServerConfig struct {
	Host string
	Port int
}

// LoggingConfig ログ設定
type LoggingConfig struct {
	Level  string
	Format string
	Output string
}

// CLIConfig CLI設定
type CLIConfig struct {
	Theme string
}

// DesktopConfig デスクトップ設定
type DesktopConfig struct {
	Window WindowConfig
}

// WindowConfig ウィンドウ設定
type WindowConfig struct {
	Width     int
	Height    int
	Resizable bool
}

// Load 設定ファイルを読み込む
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 設定ファイルを設定
	v.SetConfigFile(configPath)

	// 設定ファイルを読み込み
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// 構造体にマッピング
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
