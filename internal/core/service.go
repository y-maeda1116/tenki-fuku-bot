package core

// Service コアビジネスロジック
type Service struct {
	version string
}

// NewService 新しいServiceを作成
func NewService() *Service {
	return &Service{
		version: "dev",
	}
}

// SayHello あいさつメッセージを生成
func (s *Service) SayHello(name string) (string, error) {
	if name == "" {
		return "", ErrInvalidInput
	}
	return "Hello, " + name + "!", nil
}

// GetVersion バージョンを取得
func (s *Service) GetVersion() string {
	return s.version
}

// SetVersion バージョンを設定（テスト用）
func (s *Service) SetVersion(version string) {
	s.version = version
}
