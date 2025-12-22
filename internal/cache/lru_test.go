package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/bivlked/currate-go/internal/models"
)

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

func TestLRUCache_SetAndGet(t *testing.T) {
	cache := NewLRUCache(100, 24*time.Hour)

	t.Run("Set и Get одной записи", func(t *testing.T) {
		date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)
		cache.Set(models.USD, date, 80.5)

		rate, found := cache.Get(models.USD, date)
		if !found {
			t.Fatal("Запись должна быть найдена")
		}

		if rate != 80.5 {
			t.Errorf("Курс: ожидалось 80.5, получено %v", rate)
		}

		if cache.Size() != 1 {
			t.Errorf("Размер: ожидалось 1, получено %d", cache.Size())
		}
	})

	t.Run("Get несуществующей записи", func(t *testing.T) {
		date := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)
		rate, found := cache.Get(models.EUR, date)

		if found {
			t.Error("Запись не должна быть найдена")
		}

		if rate != 0 {
			t.Errorf("Rate должен быть 0, получено %v", rate)
		}
	})

	t.Run("Set нескольких записей", func(t *testing.T) {
		cache := NewLRUCache(100, 24*time.Hour)

		date1 := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2025, 12, 21, 0, 0, 0, 0, time.UTC)

		cache.Set(models.USD, date1, 80.5)
		cache.Set(models.EUR, date1, 94.2)
		cache.Set(models.USD, date2, 81.0)

		if cache.Size() != 3 {
			t.Errorf("Размер: ожидалось 3, получено %d", cache.Size())
		}

		// Проверяем все записи
		rate1, _ := cache.Get(models.USD, date1)
		if rate1 != 80.5 {
			t.Errorf("USD date1: ожидалось 80.5, получено %v", rate1)
		}

		rate2, _ := cache.Get(models.EUR, date1)
		if rate2 != 94.2 {
			t.Errorf("EUR date1: ожидалось 94.2, получено %v", rate2)
		}

		rate3, _ := cache.Get(models.USD, date2)
		if rate3 != 81.0 {
			t.Errorf("USD date2: ожидалось 81.0, получено %v", rate3)
		}
	})
}

func TestLRUCache_Update(t *testing.T) {
	cache := NewLRUCache(100, 24*time.Hour)

	date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

	// Первоначальное значение
	cache.Set(models.USD, date, 80.5)

	// Обновление
	cache.Set(models.USD, date, 85.0)

	rate, found := cache.Get(models.USD, date)
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
		date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

		// Добавляем 3 записи (заполняем кэш)
		cache.Set(models.USD, date.AddDate(0, 0, 0), 80.0)
		cache.Set(models.USD, date.AddDate(0, 0, 1), 81.0)
		cache.Set(models.USD, date.AddDate(0, 0, 2), 82.0)

		if cache.Size() != 3 {
			t.Errorf("Размер: ожидалось 3, получено %d", cache.Size())
		}

		// Добавляем 4-ю запись - должна вытеснить первую
		cache.Set(models.USD, date.AddDate(0, 0, 3), 83.0)

		if cache.Size() != 3 {
			t.Errorf("Размер: ожидалось 3 (после вытеснения), получено %d", cache.Size())
		}

		// Первая запись должна быть вытеснена
		_, found := cache.Get(models.USD, date.AddDate(0, 0, 0))
		if found {
			t.Error("Первая запись должна быть вытеснена")
		}

		// Остальные записи должны остаться
		_, found = cache.Get(models.USD, date.AddDate(0, 0, 1))
		if !found {
			t.Error("Вторая запись должна остаться")
		}

		_, found = cache.Get(models.USD, date.AddDate(0, 0, 2))
		if !found {
			t.Error("Третья запись должна остаться")
		}

		_, found = cache.Get(models.USD, date.AddDate(0, 0, 3))
		if !found {
			t.Error("Четвертая запись должна быть добавлена")
		}
	})

	t.Run("LRU порядок обновляется при Get", func(t *testing.T) {
		cache := NewLRUCache(3, 24*time.Hour)
		date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

		// Добавляем 3 записи
		cache.Set(models.USD, date.AddDate(0, 0, 0), 80.0) // oldest
		cache.Set(models.USD, date.AddDate(0, 0, 1), 81.0)
		cache.Set(models.USD, date.AddDate(0, 0, 2), 82.0) // newest

		// Обращаемся к oldest записи - она становится newest
		cache.Get(models.USD, date.AddDate(0, 0, 0))

		// Добавляем новую запись - должна вытеснить вторую (теперь она oldest)
		cache.Set(models.USD, date.AddDate(0, 0, 3), 83.0)

		// Первая запись должна остаться (была перемещена в конец)
		_, found := cache.Get(models.USD, date.AddDate(0, 0, 0))
		if !found {
			t.Error("Первая запись должна остаться (была обновлена через Get)")
		}

		// Вторая запись должна быть вытеснена (стала oldest)
		_, found = cache.Get(models.USD, date.AddDate(0, 0, 1))
		if found {
			t.Error("Вторая запись должна быть вытеснена")
		}

		// Третья и четвертая должны остаться
		_, found = cache.Get(models.USD, date.AddDate(0, 0, 2))
		if !found {
			t.Error("Третья запись должна остаться")
		}

		_, found = cache.Get(models.USD, date.AddDate(0, 0, 3))
		if !found {
			t.Error("Четвертая запись должна быть добавлена")
		}
	})
}

// Тесты TTL

func TestLRUCache_TTL(t *testing.T) {
	t.Run("Запись с истекшим TTL удаляется", func(t *testing.T) {
		// Короткий TTL для теста
		cache := NewLRUCache(100, 100*time.Millisecond)
		date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

		cache.Set(models.USD, date, 80.5)

		// Проверяем что запись есть
		rate, found := cache.Get(models.USD, date)
		if !found {
			t.Fatal("Запись должна быть найдена")
		}
		if rate != 80.5 {
			t.Errorf("Курс: ожидалось 80.5, получено %v", rate)
		}

		// Ждем пока TTL истечет
		time.Sleep(150 * time.Millisecond)

		// Запись должна быть удалена
		_, found = cache.Get(models.USD, date)
		if found {
			t.Error("Запись с истекшим TTL должна быть удалена")
		}

		// Размер должен уменьшиться
		if cache.Size() != 0 {
			t.Errorf("Размер: ожидалось 0 (после истечения TTL), получено %d", cache.Size())
		}
	})

	t.Run("Запись с актуальным TTL остается", func(t *testing.T) {
		cache := NewLRUCache(100, 1*time.Second)
		date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

		cache.Set(models.USD, date, 80.5)

		// Небольшая задержка (меньше TTL)
		time.Sleep(100 * time.Millisecond)

		// Запись должна быть доступна
		rate, found := cache.Get(models.USD, date)
		if !found {
			t.Error("Запись с актуальным TTL должна быть найдена")
		}
		if rate != 80.5 {
			t.Errorf("Курс: ожидалось 80.5, получено %v", rate)
		}
	})

	t.Run("Update обновляет timestamp", func(t *testing.T) {
		cache := NewLRUCache(100, 200*time.Millisecond)
		date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

		cache.Set(models.USD, date, 80.5)

		// Ждем половину TTL
		time.Sleep(120 * time.Millisecond)

		// Обновляем запись
		cache.Set(models.USD, date, 81.0)

		// Ждем еще половину TTL
		time.Sleep(120 * time.Millisecond)

		// Запись должна быть доступна (timestamp был обновлен)
		rate, found := cache.Get(models.USD, date)
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
		date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

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
				cache.Set(currency, date.AddDate(0, 0, id%10), float64(80+id))

				// Get
				cache.Get(currency, date.AddDate(0, 0, id%10))

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
		_, found := cache.Get(models.USD, date)
		_ = found // Результат не критичен, главное что нет паники
	})

	t.Run("Конкурентный Set одной записи", func(t *testing.T) {
		cache := NewLRUCache(100, 24*time.Hour)
		date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

		var wg sync.WaitGroup
		goroutineCount := 50

		// Все goroutines пишут в одну и ту же запись
		for i := 0; i < goroutineCount; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				cache.Set(models.USD, date, float64(80+id))
			}(i)
		}

		wg.Wait()

		// Проверяем что запись существует
		rate, found := cache.Get(models.USD, date)
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
		date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

		var wg sync.WaitGroup

		// 50 goroutines делают Set
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				cache.Set(models.USD, date.AddDate(0, 0, id), float64(80+id))
			}(i)
		}

		// 50 goroutines делают Get
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				cache.Get(models.USD, date.AddDate(0, 0, id))
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
	date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

	// Добавляем записи
	cache.Set(models.USD, date, 80.5)
	cache.Set(models.EUR, date, 94.2)
	cache.Set(models.USD, date.AddDate(0, 0, 1), 81.0)

	if cache.Size() != 3 {
		t.Errorf("Размер до Clear: ожидалось 3, получено %d", cache.Size())
	}

	// Очищаем
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Размер после Clear: ожидалось 0, получено %d", cache.Size())
	}

	// Проверяем что все записи удалены
	_, found := cache.Get(models.USD, date)
	if found {
		t.Error("Запись должна быть удалена после Clear")
	}

	_, found = cache.Get(models.EUR, date)
	if found {
		t.Error("Запись должна быть удалена после Clear")
	}

	// Проверяем что можем добавить новые записи
	cache.Set(models.USD, date, 85.0)
	if cache.Size() != 1 {
		t.Errorf("Размер после добавления: ожидалось 1, получено %d", cache.Size())
	}
}

// Тесты makeKey

func TestLRUCache_MakeKey(t *testing.T) {
	cache := NewLRUCache(100, 24*time.Hour)

	tests := []struct {
		name     string
		currency models.Currency
		date     time.Time
		expected string
	}{
		{
			name:     "USD 20 декабря 2025",
			currency: models.USD,
			date:     time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC),
			expected: "USD:2025-12-20",
		},
		{
			name:     "EUR 1 января 2024",
			currency: models.EUR,
			date:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "EUR:2024-01-01",
		},
		{
			name:     "USD с временем (должно игнорироваться)",
			currency: models.USD,
			date:     time.Date(2025, 12, 20, 15, 30, 45, 0, time.UTC),
			expected: "USD:2025-12-20",
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
	date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(models.USD, date.AddDate(0, 0, i%100), float64(80+i))
	}
}

func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(1000, 24*time.Hour)
	date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

	// Предварительно заполняем кэш
	for i := 0; i < 100; i++ {
		cache.Set(models.USD, date.AddDate(0, 0, i), float64(80+i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(models.USD, date.AddDate(0, 0, i%100))
	}
}

func BenchmarkLRUCache_Concurrent(b *testing.B) {
	cache := NewLRUCache(1000, 24*time.Hour)
	date := time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				cache.Set(models.USD, date.AddDate(0, 0, i%100), float64(80+i))
			} else {
				cache.Get(models.USD, date.AddDate(0, 0, i%100))
			}
			i++
		}
	})
}
