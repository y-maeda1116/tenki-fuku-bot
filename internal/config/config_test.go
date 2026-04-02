package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// 一時的な設定ファイルを作成
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	content := `
app:
  name: test-app
  version: 1.0.0
logging:
  level: debug
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.App.Name != "test-app" {
		t.Errorf("Expected 'test-app', got '%s'", cfg.App.Name)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")

	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}
