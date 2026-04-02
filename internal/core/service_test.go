package core

import (
	"testing"
)

func TestService_SayHello(t *testing.T) {
	s := NewService()

	result, err := s.SayHello("World")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", result)
	}
}

func TestService_SayHello_EmptyName(t *testing.T) {
	s := NewService()

	_, err := s.SayHello("")

	if err == nil {
		t.Error("Expected error for empty name")
	}

	if err != ErrInvalidInput {
		t.Errorf("Expected ErrInvalidInput, got %v", err)
	}
}

func TestService_GetVersion(t *testing.T) {
	s := NewService()

	version := s.GetVersion()

	if version == "" {
		t.Error("Expected non-empty version")
	}
}
