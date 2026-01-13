package backend

import (
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"goterm/backend/internal/app"
)

type wailsEmitter struct {
	ctx context.Context
}

func (e *wailsEmitter) Emit(event string, payload any) {
	if e.ctx == nil {
		return
	}
	runtime.EventsEmit(e.ctx, event, payload)
}

func Run(assets embed.FS) {
	emitter := &wailsEmitter{}
	application, err := app.NewApp(app.Config{
		Emitter: emitter,
	})
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	err = wails.Run(&options.App{
		Title:  "GoTerm",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			emitter.ctx = ctx
			application.Startup(ctx)
		},
		Bind: []any{
			application,
		},
	})
	if err != nil {
		log.Fatalf("run app: %v", err)
	}
}
