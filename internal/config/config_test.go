package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	content := []byte("areaCode: \"270000\"\nweatherArea: \"大阪地方\"\ntempArea: \"大阪\"\ncategories:\n  men: true\n  women: false\n  kids: true\n")
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AreaCode != "270000" {
		t.Errorf("AreaCode = %q, want %q", cfg.AreaCode, "270000")
	}
	if cfg.WeatherArea != "大阪地方" {
		t.Errorf("WeatherArea = %q, want %q", cfg.WeatherArea, "大阪地方")
	}
	if cfg.TempArea != "大阪" {
		t.Errorf("TempArea = %q, want %q", cfg.TempArea, "大阪")
	}
	if !cfg.Categories.Men {
		t.Error("Categories.Men = false, want true")
	}
	if cfg.Categories.Women {
		t.Error("Categories.Women = true, want false")
	}
	if !cfg.Categories.Kids {
		t.Error("Categories.Kids = false, want true")
	}
}

func TestLoad_missingFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Load() expected error for missing file")
	}
}
