package core

import (
	"errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	err := errors.New("original error")
	appErr := NewAppError("CODE001", "test message", err)

	if appErr.Error() != "test message" {
		t.Errorf("Expected 'test message', got '%s'", appErr.Error())
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := NewAppError("CODE001", "test message", originalErr)

	if errors.Unwrap(appErr) != originalErr {
		t.Error("Unwrap should return the original error")
	}
}

func TestAppError_Code(t *testing.T) {
	appErr := NewAppError("CODE001", "test message", nil)

	if appErr.Code != "CODE001" {
		t.Errorf("Expected 'CODE001', got '%s'", appErr.Code)
	}
}
