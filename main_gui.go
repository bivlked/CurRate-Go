package main

import (
	"context"
	"embed"
	"log"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/bivlked/currate-go/internal/app"
	"github.com/bivlked/currate-go/internal/cache"
	"github.com/bivlked/currate-go/internal/converter"
	"github.com/bivlked/currate-go/internal/models"
	"github.com/bivlked/currate-go/internal/parser"
)

//go:embed all:frontend
var assets embed.FS

// rateProviderAdapter адаптирует функцию parser.FetchRates к интерфейсу converter.RateProvider
type rateProviderAdapter struct{}

func (r *rateProviderAdapter) FetchRates(date time.Time) (*models.RateData, error) {
	return parser.FetchRates(date)
}

func main() {
	// Создаем кэш для курсов валют
	cacheStorage := cache.NewLRUCache(100, 24*time.Hour)

	// Создаем адаптер для parser.FetchRates
	rateProvider := &rateProviderAdapter{}

	// Создаем конвертер с парсером ЦБ РФ и кэшем
	conv := converter.NewConverter(rateProvider, cacheStorage)

	// Создаем App instance для GUI
	appInstance := app.NewApp(conv)

	// Запускаем Wails приложение
	err := wails.Run(&options.App{
		Title:         "Конвертер валют (c) BiV",
		Width:         360,
		Height:        758,
		MinWidth:      360,
		MaxWidth:      360,
		MinHeight:     758,
		MaxHeight:     758,
		DisableResize: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup: func(ctx context.Context) {
			appInstance.Startup(ctx)
		},
		Bind: []interface{}{
			appInstance,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false, // Включаем иконку окна
		},
	})

	if err != nil {
		log.Fatal("Ошибка запуска приложения:", err)
	}
}
