package core

import (
	"errors"
)

var (
	// 設定エラー
	ErrConfigNotFound = errors.New("configuration not found")
	ErrConfigInvalid  = errors.New("invalid configuration")

	// ビジネスロジックエラー
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrPermission   = errors.New("permission denied")
)

// AppError アプリケーションエラー
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError 新しいAppErrorを作成
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
