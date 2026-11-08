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
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/bivlked/currate-go/internal/app"
	"github.com/bivlked/currate-go/internal/cache"
	"github.com/bivlked/currate-go/internal/converter"
	"github.com/bivlked/currate-go/internal/parser"
	"github.com/bivlked/currate-go/internal/window"
)

//go:embed all:frontend
var assets embed.FS

// Размер окна в Wails - внешний, с рамкой и заголовком: рамка Windows 11
// забирает около 15 px по ширине и 38 px по высоте. Внутри остаётся
// около 361x784 - столько нужно странице в худшем случае (месяц в 6 недель,
// выходной, большая сумма и сравнение с референсом)
const (
	windowWidth  = 376
	windowHeight = 822
	// Свой класс окна, чтобы fitWindowToScreen нашла именно его
	windowClassName = "CurRateWindow"
)

func main() {
	// Создаем кэш для курсов валют
	cacheStorage := cache.NewLRUCache(100, 24*time.Hour)

	// Создаем конвертер с парсером ЦБ РФ и кэшем
	conv := converter.NewConverter(converter.FetchRatesFunc(parser.FetchRates), cacheStorage)

	// Создаем App instance для GUI
	appInstance := app.NewApp(conv)

	// Запускаем Wails приложение
	err := wails.Run(&options.App{
		Title:         "Конвертер валют (c) BiV",
		Width:         windowWidth,
		Height:        windowHeight,
		MinWidth:      windowWidth,
		MaxWidth:      windowWidth,
		MinHeight:     windowHeight,
		MaxHeight:     windowHeight,
		DisableResize: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnStartup: func(ctx context.Context) {
			appInstance.Startup(ctx)
			fitWindowToScreen(ctx)
		},
		Bind: []interface{}{
			appInstance,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false, // Включаем иконку окна
			WindowClassName:      windowClassName,
		},
	})

	if err != nil {
		log.Fatal("Ошибка запуска приложения:", err)
	}
}

// fitWindowToScreen уменьшает окно по высоте, если оно не помещается
// в рабочую область своего монитора (экран без панели задач). Так бывает
// на низких экранах: 1280x720 или 1920x1080 при масштабе 150%. Размер окна
// менять нельзя, поэтому без подгонки низ с результатом оказался бы
// за краем экрана, а так страница просто прокручивается.
// Монитор берётся у самого окна, а не у курсора: Windows может открыть
// окно и на другом мониторе. Если система данные не отдала, окно остаётся
// полного размера, как раньше
func fitWindowToScreen(ctx context.Context) {
	workPx, dpi, err := window.WorkArea(windowClassName)
	if err != nil {
		return
	}
	height := window.FitHeight(windowHeight, workPx, dpi)
	if height == windowHeight {
		return
	}
	// Сначала максимум: Wails вслед за ним опускает и минимум
	runtime.WindowSetMaxSize(ctx, windowWidth, height)
	runtime.WindowSetMinSize(ctx, windowWidth, height)
	runtime.WindowSetSize(ctx, windowWidth, height)
	runtime.WindowCenter(ctx)
}
