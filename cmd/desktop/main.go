// cmd/desktop/main.go
package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/y-maeda1116/template-go-cross/internal/ui"
	"github.com/y-maeda1116/template-go-cross/internal/version"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// アプリインスタンスを作成
	app := ui.NewApp(version.Version)

	// デスクトップアプリを開始
	err := wails.Run(&options.App{
		Title:  "My App",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.Startup,
		OnShutdown: app.Shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
		os.Exit(1)
	}
}
