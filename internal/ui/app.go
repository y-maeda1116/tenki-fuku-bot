// internal/ui/app.go
package ui

import "context"

// App Wailsアプリ
type App struct {
	version string
}

// NewApp 新しいAppを作成
func NewApp(version string) *App {
	return &App{version: version}
}

// Startup アプリ起動時に呼ばれる
func (a *App) Startup(ctx context.Context) {
	// 初期化処理
}

// Shutdown アプリ終了時に呼ばれる
func (a *App) Shutdown(ctx context.Context) {
	// クリーンアップ処理
}

// Greet あいさつを返す
func (a *App) Greet(name string) string {
	if name == "" {
		name = "World"
	}
	return "Hello, " + name + "!"
}

// Version バージョンを返す
func (a *App) Version() string {
	return a.version
}
