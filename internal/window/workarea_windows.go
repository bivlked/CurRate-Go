//go:build windows

package window

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                = windows.NewLazySystemDLL("user32.dll")
	procFindWindowEx      = user32.NewProc("FindWindowExW")
	procMonitorFromWindow = user32.NewProc("MonitorFromWindow")
	procGetMonitorInfo    = user32.NewProc("GetMonitorInfoW")
	procGetDpiForWindow   = user32.NewProc("GetDpiForWindow")
)

const monitorDefaultToNearest = 2 // MONITOR_DEFAULTTONEAREST

type rect struct{ Left, Top, Right, Bottom int32 }

type monitorInfo struct {
	CbSize    uint32
	RcMonitor rect
	RcWork    rect
	DwFlags   uint32
}

// WorkArea находит окно этого процесса с классом className и возвращает
// высоту рабочей области его монитора (экран без панели задач)
// в физических пикселях и DPI самого окна: по этому DPI Wails переводит
// размеры окна из логических пикселей в физические
func WorkArea(className string) (heightPx, dpi int, err error) {
	hwnd, err := findOwnWindow(className)
	if err != nil {
		return 0, 0, err
	}

	mon, err := call(procMonitorFromWindow, hwnd, monitorDefaultToNearest)
	if err != nil {
		return 0, 0, err
	}
	mi := monitorInfo{CbSize: uint32(unsafe.Sizeof(monitorInfo{}))}
	if _, err := call(procGetMonitorInfo, mon, uintptr(unsafe.Pointer(&mi))); err != nil {
		return 0, 0, err
	}

	d, err := call(procGetDpiForWindow, hwnd)
	if err != nil {
		return 0, 0, err
	}

	return int(mi.RcWork.Bottom - mi.RcWork.Top), int(d), nil
}

// findOwnWindow ищет окно верхнего уровня с классом className,
// принадлежащее этому процессу: класс мог зарегистрировать
// и второй запущенный экземпляр приложения
func findOwnWindow(className string) (uintptr, error) {
	cls, err := windows.UTF16PtrFromString(className)
	if err != nil {
		return 0, err
	}
	pid := windows.GetCurrentProcessId()

	var after uintptr
	for {
		hwnd, err := call(procFindWindowEx, 0, after, uintptr(unsafe.Pointer(cls)), 0)
		if err != nil {
			return 0, fmt.Errorf("окно %s этого процесса не найдено: %w", className, err)
		}
		var owner uint32
		if _, err := windows.GetWindowThreadProcessId(windows.HWND(hwnd), &owner); err != nil {
			return 0, err
		}
		if owner == pid {
			return hwnd, nil
		}
		after = hwnd
	}
}

// call вызывает функцию Win32, у которой ноль означает неудачу.
// Ошибку LazyProc.Call (GetLastError) смотрит только в этом случае,
// а нулевой код ошибки заменяет понятным текстом: часть функций его не выставляет.
// Find заранее, потому что Call на отсутствующей функции паникует.
// uintptrescapes: аргументы-указатели живут до конца вызова, как в самом LazyProc.Call
//
//go:uintptrescapes
func call(p *windows.LazyProc, args ...uintptr) (uintptr, error) {
	if err := p.Find(); err != nil {
		return 0, err
	}
	r, _, err := p.Call(args...)
	if r != 0 {
		return r, nil
	}
	if errors.Is(err, windows.ERROR_SUCCESS) {
		return 0, fmt.Errorf("%s: неудача без кода ошибки", p.Name)
	}
	return 0, fmt.Errorf("%s: %w", p.Name, err)
}
