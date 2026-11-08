//go:build windows

package window

import "testing"

// Окна с таким классом у тестового процесса нет: WorkArea должна вернуть
// ошибку, а не чужое окно и не панику
func TestWorkAreaNoWindow(t *testing.T) {
	if _, _, err := WorkArea("CurRateNoSuchWindowClass"); err == nil {
		t.Fatal("WorkArea: ожидалась ошибка для отсутствующего окна")
	}
}
