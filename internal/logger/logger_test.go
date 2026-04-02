package logger

import (
	"testing"
)

func TestNew(t *testing.T) {
	// Infoレベルでロガーを作成
	log := New("info")

	if log == nil {
		t.Fatal("Expected non-nil logger")
	}
}

func TestLogger_Debug(t *testing.T) {
	log := New("debug")

	// パニックしないことを確認
	log.Debug("debug message", "key", "value")
}

func TestLogger_Info(t *testing.T) {
	log := New("info")

	// パニックしないことを確認
	log.Info("info message", "key", "value")
}

func TestLogger_Error(t *testing.T) {
	log := New("error")

	// パニックしないことを確認
	log.Error("error message", "key", "value")
}

func TestLogger_Sync(t *testing.T) {
	log := New("info")

	// パニックしないことを確認
	log.Sync()
}
