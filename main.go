package main

import (
	"MarkdownDesk/backend"
	"MarkdownDesk/backend/configure"
	"embed"

	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

func init() {}

func main() {

	app := application.New(application.Options{
		Name:        "Markdown Desk",
		Description: "編寫Markdown的簡易Editor",
		Services: []application.Service{
			application.NewService(&backend.ConfigService{}),
			application.NewServiceWithOptions(
				&backend.FileService{},
				application.ServiceOptions{
					Route: configure.ImageWebRoute,
				},
			)},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:          "MarkdownDesk",
		Width:          1000,
		Height:         618,
		DisableResize:  false,
		EnableFileDrop: true,
		URL:            "/",
	}).OnWindowEvent(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {

		files := event.Context().DroppedFiles()

		if len(files) > 0 {
			emitKey := configure.EmitKey
			emitData := map[string]any{"path": files[0]}
			application.Get().Event.Emit(emitKey, emitData)
		}
	})

	err := app.Run()

	if err != nil {
		log.Fatal(err)
	}
}
