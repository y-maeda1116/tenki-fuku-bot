package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	content := []byte("city: Osaka\ncategories:\n  men: true\n  women: false\n  kids: true\n")
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.City != "Osaka" {
		t.Errorf("City = %q, want %q", cfg.City, "Osaka")
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
