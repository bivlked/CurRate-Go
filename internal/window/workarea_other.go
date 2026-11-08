//go:build !windows

package window

import "errors"

// WorkArea на других системах данных не отдаёт: окно остаётся
// заданного размера
func WorkArea(_ string) (heightPx, dpi int, err error) {
	return 0, 0, errors.ErrUnsupported
}
