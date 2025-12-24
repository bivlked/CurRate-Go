package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/models"
)

type fakeClock struct {
	now time.Time
}

func newFakeClock(start time.Time) *fakeClock {
	return &fakeClock{now: start}
}

func (f *fakeClock) Now() time.Time {
	return f.now
}

func (f *fakeClock) Since(t time.Time) time.Duration {
	return f.now.Sub(t)
}

func (f *fakeClock) Advance(d time.Duration) {
	f.now = f.now.Add(d)
}

func withFakeClock(t *testing.T, start time.Time) *fakeClock {
	t.Helper()

	clock := newFakeClock(start)
	originalNow := nowFunc
	originalSince := sinceFunc
	nowFunc = clock.Now
	sinceFunc = clock.Since
	t.Cleanup(func() {
		nowFunc = originalNow
		sinceFunc = originalSince
	})
	return clock
}

// Тесты базовой функциональности

func TestNewLRUCache(t *testing.T) {
	cache := NewLRUCache(100, 24*time.Hour)

	if cache == nil {
		t.Fatal("Cache не должен быть nil")
	}

	if cache.maxSize != 100 {
		t.Errorf("maxSize: ожидалось 100, получено %d", cache.maxSize)
	}

	if cache.ttl != 24*time.Hour {
		t.Errorf("ttl: ожидалось 24h, получено %v", cache.ttl)
	}

	if cache.Size() != 0 {
		t.Errorf("Начальный размер: ожидалось 0, получено %d", cache.Size())
	}
}

func TestNewLRUCache_PanicsOnZeroMaxSize(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("NewLRUCache должна паниковать при maxSize = 0")
		}
		expectedMsg := "cache: maxSize must be positive"
		if recovered != expectedMsg {
			t.Errorf("Сообщение паники: ожидалось %q, получено %q", expectedMsg, recovered)
		}
	}()

	NewLRUCache(0, 24*time.Hour)
}

func TestNewLRUCache_PanicsOnNegativeMaxSize(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("NewLRUCache должна паниковать при maxSize < 0")
		}
		expectedMsg := "cache: maxSize must be positive"
		if recovered != expectedMsg {
			t.Errorf("Сообщение паники: ожидалось %q, получено %q", expectedMsg, recovered)
		}
	}()

	NewLRUCache(-1, 24*time.Hour)
}

func TestLRUCache_SetAndGet(t *testing.T) {
	cache := NewLRUCache(100, 24*time.Hour)
	baseDate := testPastDateUTC()

	t.Run("Set и Get одной записи", func(t *testing.T) {
		date := baseDate
		cache.Set(models.USD, date, 80.5, date) // requestedDate и actualDate одинаковы

		rate, actualDate, found := cache.Get(models.USD, date)
		if !found {
			t.Fatal("Запись должна быть найдена")
		}

		if rate != 80.5 {
			t.Errorf("Курс: ожидалось 80.5, получено %v", rate)
		}

		if !actualDate.Equal(date) {
			t.Errorf("Фактическая дата: ожидалось %v, получено %v", date, actualDate)
		}

		if cache.Size() != 1 {
			t.Errorf("Размер: ожидалось 1, получено %d", cache.Size())
		}
	})

	t.Run("Get несуществующей записи", func(t *testing.T) {
		date := baseDate.AddDate(0, 0, 5)
		rate, actualDate, found := cache.Get(models.EUR, date)

		if found {
			t.Error("Запись не должна быть найдена")
		}

		if rate != 0 {
			t.Errorf("Rate должен быть 0, получено %v", rate)
		}

		if !actualDate.IsZero() {
			t.Errorf("ActualDate должен быть нулевым, получено %v", actualDate)
		}
	})

	t.Run("Set нескольких записей", func(t *testing.T) {
		cache := NewLRUCache(100, 24*time.Hour)

		date1 := baseDate
		date2 := baseDate.AddDate(0, 0, 1)

		cache.Set(models.USD, date1, 80.5, date1) // requestedDate и actualDate одинаковы
		cache.Set(models.EUR, date1, 94.2, date1)
		cache.Set(models.USD, date2, 81.0, date2)

		if cache.Size() != 3 {
			t.Errorf("Размер: ожидалось 3, получено %d", cache.Size())
		}

		// Проверяем все записи
		rate1, _, _ := cache.Get(models.USD, date1)
		if rate1 != 80.5 {
			t.Errorf("USD date1: ожидалось 80.5, получено %v", rate1)
		}

		rate2, _, _ := cache.Get(models.EUR, date1)
		if rate2 != 94.2 {
			t.Errorf("EUR date1: ожидалось 94.2, получено %v", rate2)
		}

		rate3, _, _ := cache.Get(models.USD, date2)
		if rate3 != 81.0 {
			t.Errorf("USD date2: ожидалось 81.0, получено %v", rate3)
		}
	})
}

func TestLRUCache_Update(t *testing.T) {
	cache := NewLRUCache(100, 24*time.Hour)

	date := testPastDateUTC()

	// Первоначальное значение
	cache.Set(models.USD, date, 80.5, date) // requestedDate и actualDate одинаковы

	// Обновление
	cache.Set(models.USD, date, 85.0, date)

	rate, _, found := cache.Get(models.USD, date)
	if !found {
		t.Fatal("Запись должна быть найдена")
	}

	if rate != 85.0 {
		t.Errorf("Курс: ожидалось 85.0 (обновленное значение), получено %v", rate)
	}

	// Размер не должен измениться
	if cache.Size() != 1 {
		t.Errorf("Размер: ожидалось 1, получено %d", cache.Size())
	}
}

// Тесты LRU вытеснения

func TestLRUCache_Eviction(t *testing.T) {
	t.Run("Вытеснение при переполнении", func(t *testing.T) {
		cache := NewLRUCache(3, 24*time.Hour)
		date := testPastDateUTC()

		// Добавляем 3 записи (заполняем кэш)
		date0 := date.AddDate(0, 0, 0)
		date1 := date.AddDate(0, 0, 1)
		date2 := date.AddDate(0, 0, 2)
		date3 := date.AddDate(0, 0, 3)
		cache.Set(models.USD, date0, 80.0, date0) // requestedDate и actualDate одинаковы
		cache.Set(models.USD, date1, 81.0, date1)
		cache.Set(models.USD, date2, 82.0, date2)

		if cache.Size() != 3 {
			t.Errorf("Размер: ожидалось 3, получено %d", cache.Size())
		}

		// Добавляем 4-ю запись - должна вытеснить первую
		cache.Set(models.USD, date3, 83.0, date3)

		if cache.Size() != 3 {
			t.Errorf("Размер: ожидалось 3 (после вытеснения), получено %d", cache.Size())
		}

		// Первая запись должна быть вытеснена
		_, _, found := cache.Get(models.USD, date0)
		if found {
			t.Error("Первая запись должна быть вытеснена")
		}

		// Остальные записи должны остаться
		_, _, found = cache.Get(models.USD, date1)
		if !found {
			t.Error("Вторая запись должна остаться")
		}

		_, _, found = cache.Get(models.USD, date.AddDate(0, 0, 2))
		if !found {
			t.Error("Третья запись должна остаться")
		}

		_, _, found = cache.Get(models.USD, date.AddDate(0, 0, 3))
		if !found {
			t.Error("Четвертая запись должна быть добавлена")
		}
	})

	t.Run("LRU порядок обновляется при Get", func(t *testing.T) {
		cache := NewLRUCache(3, 24*time.Hour)
		date := testPastDateUTC()

		// Добавляем 3 записи
		date0 := date.AddDate(0, 0, 0)
		date1 := date.AddDate(0, 0, 1)
		date2 := date.AddDate(0, 0, 2)
		date3 := date.AddDate(0, 0, 3)
		cache.Set(models.USD, date0, 80.0, date0) // oldest, requestedDate и actualDate одинаковы
		cache.Set(models.USD, date1, 81.0, date1)
		cache.Set(models.USD, date2, 82.0, date2) // newest

		// Обращаемся к oldest записи - она становится newest
		cache.Get(models.USD, date0)

		// Добавляем новую запись - должна вытеснить вторую (теперь она oldest)
		cache.Set(models.USD, date3, 83.0, date3)

		// Первая запись должна остаться (была перемещена в конец)
		_, _, found := cache.Get(models.USD, date0)
		if !found {
			t.Error("Первая запись должна остаться (была обновлена через Get)")
		}

		// Вторая запись должна быть вытеснена (стала oldest)
		_, _, found = cache.Get(models.USD, date1)
		if found {
			t.Error("Вторая запись должна быть вытеснена")
		}

		// Третья и четвертая должны остаться
		_, _, found = cache.Get(models.USD, date2)
		if !found {
			t.Error("Третья запись должна остаться")
		}

		_, _, found = cache.Get(models.USD, date3)
		if !found {
			t.Error("Четвертая запись должна быть добавлена")
		}
	})
}

// Тесты TTL

func TestLRUCache_TTL(t *testing.T) {
	t.Run("Запись с истекшим TTL удаляется", func(t *testing.T) {
		// Короткий TTL для теста
		start := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)
		clock := withFakeClock(t, start)
		cache := NewLRUCache(100, 100*time.Millisecond)
		date := testPastDateUTC()

		cache.Set(models.USD, date, 80.5, date) // requestedDate и actualDate одинаковы

		// Проверяем что запись есть
		rate, _, found := cache.Get(models.USD, date)
		if !found {
			t.Fatal("Запись должна быть найдена")
		}
		if rate != 80.5 {
			t.Errorf("Курс: ожидалось 80.5, получено %v", rate)
		}

		// Двигаем время за пределы TTL
		clock.Advance(150 * time.Millisecond)

		// Запись должна быть удалена
		_, _, found = cache.Get(models.USD, date)
		if found {
			t.Error("Запись с истекшим TTL должна быть удалена")
		}

		// Размер должен уменьшиться
		if cache.Size() != 0 {
			t.Errorf("Размер: ожидалось 0 (после истечения TTL), получено %d", cache.Size())
		}
	})

	t.Run("Запись с актуальным TTL остается", func(t *testing.T) {
		start := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)
		clock := withFakeClock(t, start)
		cache := NewLRUCache(100, 1*time.Second)
		date := testPastDateUTC()

		cache.Set(models.USD, date, 80.5, date) // requestedDate и actualDate одинаковы

		// Небольшой шаг времени (меньше TTL)
		clock.Advance(100 * time.Millisecond)

		// Запись должна быть доступна
		rate, _, found := cache.Get(models.USD, date)
		if !found {
			t.Error("Запись с актуальным TTL должна быть найдена")
		}
		if rate != 80.5 {
			t.Errorf("Курс: ожидалось 80.5, получено %v", rate)
		}
	})

	t.Run("Update обновляет timestamp", func(t *testing.T) {
		start := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)
		clock := withFakeClock(t, start)
		cache := NewLRUCache(100, 200*time.Millisecond)
		date := testPastDateUTC()

		cache.Set(models.USD, date, 80.5, date) // requestedDate и actualDate одинаковы

		// Двигаем время вперед
		clock.Advance(120 * time.Millisecond)

		// Обновляем запись
		cache.Set(models.USD, date, 81.0, date)

		// Двигаем время вперед
		clock.Advance(120 * time.Millisecond)

		// Запись должна быть доступна (timestamp был обновлен)
		rate, _, found := cache.Get(models.USD, date)
		if !found {
			t.Error("Обновленная запись должна быть найдена (timestamp обновлен)")
		}
		if rate != 81.0 {
			t.Errorf("Курс: ожидалось 81.0, получено %v", rate)
		}
	})
}

// Тесты thread-safety

func TestLRUCache_ThreadSafety(t *testing.T) {
	t.Run("Конкурентный доступ (100 goroutines)", func(t *testing.T) {
		cache := NewLRUCache(100, 24*time.Hour)
		date := testPastDateUTC()

		var wg sync.WaitGroup
		goroutineCount := 100

		// Запускаем 100 goroutines
		for i := 0; i < goroutineCount; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				currency := models.USD
				if id%2 == 0 {
					currency = models.EUR
				}

				// Set
				reqDate := date.AddDate(0, 0, id%10)
				cache.Set(currency, reqDate, float64(80+id), reqDate) // requestedDate и actualDate одинаковы

				// Get
				cache.Get(currency, date.AddDate(0, 0, id%10)) // Игнорируем возвращаемые значения

				// Size
				cache.Size()
			}(i)
		}

		wg.Wait()

		// Проверяем что кэш не сломался
		size := cache.Size()
		if size > 100 {
			t.Errorf("Размер превысил максимум: %d > 100", size)
		}

		// Проверяем что можем получить записи
		_, _, found := cache.Get(models.USD, date)
		_ = found // Результат не критичен, главное что нет паники
	})

	t.Run("Конкурентный Set одной записи", func(t *testing.T) {
		cache := NewLRUCache(100, 24*time.Hour)
		date := testPastDateUTC()

		var wg sync.WaitGroup
		goroutineCount := 50

		// Все goroutines пишут в одну и ту же запись
		for i := 0; i < goroutineCount; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				cache.Set(models.USD, date, float64(80+id), date) // requestedDate и actualDate одинаковы
			}(i)
		}

		wg.Wait()

		// Проверяем что запись существует
		rate, _, found := cache.Get(models.USD, date)
		if !found {
			t.Error("Запись должна быть найдена после конкурентных Set")
		}

		// Курс должен быть одним из установленных значений (80-129)
		if rate < 80 || rate >= 130 {
			t.Errorf("Курс вне ожидаемого диапазона: %v", rate)
		}

		// Размер должен быть 1 (все пишут в одну запись)
		if cache.Size() != 1 {
			t.Errorf("Размер: ожидалось 1, получено %d", cache.Size())
		}
	})

	t.Run("Конкурентный Get и Set", func(t *testing.T) {
		cache := NewLRUCache(50, 24*time.Hour)
		date := testPastDateUTC()

		var wg sync.WaitGroup

		// 50 goroutines делают Set
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				reqDate := date.AddDate(0, 0, id)
				cache.Set(models.USD, reqDate, float64(80+id), reqDate) // requestedDate и actualDate одинаковы
			}(i)
		}

		// 50 goroutines делают Get
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				cache.Get(models.USD, date.AddDate(0, 0, id)) // Игнорируем возвращаемые значения
			}(i)
		}

		wg.Wait()

		// Проверяем что кэш не сломался
		size := cache.Size()
		if size > 50 {
			t.Errorf("Размер превысил максимум: %d > 50", size)
		}
	})
}

// Тесты Clear

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(100, 24*time.Hour)
	date := testPastDateUTC()

	// Добавляем записи
	cache.Set(models.USD, date, 80.5, date) // requestedDate и actualDate одинаковы
	cache.Set(models.EUR, date, 94.2, date)
	cache.Set(models.USD, date.AddDate(0, 0, 1), 81.0, date.AddDate(0, 0, 1))

	if cache.Size() != 3 {
		t.Errorf("Размер до Clear: ожидалось 3, получено %d", cache.Size())
	}

	// Очищаем
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Размер после Clear: ожидалось 0, получено %d", cache.Size())
	}

	// Проверяем что все записи удалены
	_, _, found := cache.Get(models.USD, date)
	if found {
		t.Error("Запись должна быть удалена после Clear")
	}

	_, _, found = cache.Get(models.EUR, date)
	if found {
		t.Error("Запись должна быть удалена после Clear")
	}

	// Проверяем что можем добавить новые записи
	cache.Set(models.USD, date, 85.0, date) // requestedDate и actualDate одинаковы
	if cache.Size() != 1 {
		t.Errorf("Размер после добавления: ожидалось 1, получено %d", cache.Size())
	}
}

// Тесты makeKey

func TestLRUCache_MakeKey(t *testing.T) {
	cache := NewLRUCache(100, 24*time.Hour)
	baseDate := testPastDateUTC()

	tests := []struct {
		name     string
		currency models.Currency
		date     time.Time
		expected string
	}{
		{
			name:     "USD базовая дата",
			currency: models.USD,
			date:     baseDate,
			expected: "USD:" + baseDate.Format("2006-01-02"),
		},
		{
			name:     "EUR базовая дата минус 30 дней",
			currency: models.EUR,
			date:     baseDate.AddDate(0, 0, -30),
			expected: "EUR:" + baseDate.AddDate(0, 0, -30).Format("2006-01-02"),
		},
		{
			name:     "USD с временем (должно игнорироваться)",
			currency: models.USD,
			date:     time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), 15, 30, 45, 0, time.UTC),
			expected: "USD:" + baseDate.Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cache.makeKey(tt.currency, tt.date)
			if got != tt.expected {
				t.Errorf("makeKey: ожидалось %s, получено %s", tt.expected, got)
			}
		})
	}
}

// Бенчмарки

func BenchmarkLRUCache_Set(b *testing.B) {
	cache := NewLRUCache(1000, 24*time.Hour)
	date := testPastDateUTC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reqDate := date.AddDate(0, 0, i%100)
		cache.Set(models.USD, reqDate, float64(80+i), reqDate) // requestedDate и actualDate одинаковы
	}
}

func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(1000, 24*time.Hour)
	date := testPastDateUTC()

	// Предварительно заполняем кэш
	for i := 0; i < 100; i++ {
		reqDate := date.AddDate(0, 0, i)
		cache.Set(models.USD, reqDate, float64(80+i), reqDate) // requestedDate и actualDate одинаковы
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(models.USD, date.AddDate(0, 0, i%100))
	}
}

func BenchmarkLRUCache_Concurrent(b *testing.B) {
	cache := NewLRUCache(1000, 24*time.Hour)
	date := testPastDateUTC()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				reqDate := date.AddDate(0, 0, i%100)
				cache.Set(models.USD, reqDate, float64(80+i), reqDate) // requestedDate и actualDate одинаковы
			} else {
				cache.Get(models.USD, date.AddDate(0, 0, i%100))
			}
			i++
		}
	})
}
