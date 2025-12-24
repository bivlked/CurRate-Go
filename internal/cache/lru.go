// Package cache предоставляет функциональность кэширования курсов валют
package cache

import (
	"container/list"
	"sync"
	"time"

	"github.com/bivlked/currate-go/internal/models"
)

var nowFunc = time.Now
var sinceFunc = time.Since

// LRUCache - потокобезопасный LRU кэш с поддержкой TTL
// Использует комбинацию map и двусвязного списка для O(1) операций
type LRUCache struct {
	mu      sync.RWMutex
	cache   map[string]*list.Element // Хэш-таблица для быстрого доступа
	lru     *list.List               // Двусвязный список для LRU порядка
	maxSize int                      // Максимальный размер кэша
	ttl     time.Duration            // Время жизни записи
}

// Entry - запись в кэше
type Entry struct {
	key        string
	rate       float64
	timestamp  time.Time
	actualDate time.Time // Фактическая дата курса из XML (может отличаться от запрошенной)
}

// NewLRUCache создает новый LRU кэш с заданным размером и TTL
// maxSize - максимальное количество записей в кэше (должно быть > 0)
// ttl - время жизни записи (например, 24 часа)
//
// Паникует, если maxSize <= 0 (программистская ошибка).
//
// Пример использования:
//
//	cache := cache.NewLRUCache(100, 24*time.Hour)
func NewLRUCache(maxSize int, ttl time.Duration) *LRUCache {
	if maxSize <= 0 {
		panic("cache: maxSize must be positive")
	}
	return &LRUCache{
		cache:   make(map[string]*list.Element),
		lru:     list.New(),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// Get получает курс из кэша для указанной валюты и даты
// Возвращает (rate, actualDate, true) если запись найдена и не истекла
// Возвращает (0, time.Time{}, false) если запись не найдена или истекла
//
// Метод thread-safe и обновляет LRU порядок при успешном доступе
func (c *LRUCache) Get(currency models.Currency, date time.Time) (float64, time.Time, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.makeKey(currency, date)
	elem, exists := c.cache[key]
	if !exists {
		return 0, time.Time{}, false
	}

	entry := elem.Value.(*Entry)

	// Проверка TTL
	if sinceFunc(entry.timestamp) > c.ttl {
		// TTL истек - удаляем запись
		c.lru.Remove(elem)
		delete(c.cache, key)
		return 0, time.Time{}, false
	}

	// Переместить в конец списка (most recently used)
	c.lru.MoveToBack(elem)
	return entry.rate, entry.actualDate, true
}

// Set сохраняет курс в кэш для указанной валюты и запрошенной даты
// requestedDate - запрошенная дата (используется как ключ кэша)
// actualDate - фактическая дата курса из XML (сохраняется в Entry, может отличаться от запрошенной)
// Если запись уже существует - обновляет её и обновляет timestamp
// Если кэш переполнен - вытесняет наименее используемую запись (LRU)
//
// Метод thread-safe
func (c *LRUCache) Set(currency models.Currency, requestedDate time.Time, rate float64, actualDate time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.makeKey(currency, requestedDate)

	// Если уже существует - обновить
	if elem, exists := c.cache[key]; exists {
		entry := elem.Value.(*Entry)
		entry.rate = rate
		entry.timestamp = nowFunc()
		entry.actualDate = actualDate // Обновляем фактическую дату
		c.lru.MoveToBack(elem)
		return
	}

	// Вытеснение если переполнен
	if c.lru.Len() >= c.maxSize {
		oldest := c.lru.Front()
		if oldest != nil {
			c.lru.Remove(oldest)
			delete(c.cache, oldest.Value.(*Entry).key)
		}
	}

	// Добавить новую запись
	entry := &Entry{
		key:        key,
		rate:       rate,
		timestamp:  nowFunc(),
		actualDate: actualDate, // Сохраняем фактическую дату (может отличаться от requestedDate)
	}
	elem := c.lru.PushBack(entry)
	c.cache[key] = elem
}

// makeKey создает уникальный ключ для валюты и даты
// Формат: "USD:2025-12-20"
func (c *LRUCache) makeKey(currency models.Currency, date time.Time) string {
	return string(currency) + ":" + date.Format("2006-01-02")
}

// Size возвращает текущий размер кэша (количество записей)
// Метод thread-safe
func (c *LRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lru.Len()
}

// Clear очищает весь кэш
// Метод thread-safe
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*list.Element)
	c.lru.Init()
}
