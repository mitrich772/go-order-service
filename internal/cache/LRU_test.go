package cache

import "testing"

// Тестирует добавление нового элемента и обновление существующего
func TestLRU_SetAndGet(t *testing.T) {
	cache := NewLru(2)

	// Добавление нового элемента
	exist := cache.Set("a", 1)
	if exist {
		t.Error("Ожидалось, что элемент новый, но вернуло true")
	}
	val, ok := cache.Get("a")
	if !ok || val != 1 {
		t.Error("Элемент не сохранился или неверное значение")
	}

	// Обновление существующего элемента
	exist = cache.Set("a", 2)
	if !exist {
		t.Error("Ожидалось, что элемент уже существовал")
	}
	val, _ = cache.Get("a")
	if val != 2 {
		t.Error("Значение элемента не обновилось")
	}
}

// Тестирует удаление наименее недавно использованного элемента при переполнении
func TestLRU_LRURemoval(t *testing.T) {
	cache := NewLru(2)

	cache.Set("a", 1)
	cache.Set("b", 2)

	// Добавляем третий элемент, должен удалиться "a"
	cache.Set("c", 3)
	_, ok := cache.Get("a")
	if ok {
		t.Error("Ожидалось удаление 'a', но он остался")
	}

	// "b" и "c" должны быть доступны
	if val, _ := cache.Get("b"); val != 2 {
		t.Error("Ожидалось значение 2 для 'b'")
	}
	if val, _ := cache.Get("c"); val != 3 {
		t.Error("Ожидалось значение 3 для 'c'")
	}
}

// Тестирует, что доступ к элементу делает его MRU
func TestLRU_AccessUpdatesMRU(t *testing.T) {
	cache := NewLru(2)

	cache.Set("a", 1)
	cache.Set("b", 2)

	// Доступ к "a" делает его MRU
	cache.Get("a")
	cache.Set("c", 3) // теперь "b" должен быть удалён

	_, ok := cache.Get("b")
	if ok {
		t.Error("Ожидалось удаление 'b', но он остался")
	}
	if val, _ := cache.Get("a"); val != 1 {
		t.Error("Элемент 'a' должен остаться")
	}
	if val, _ := cache.Get("c"); val != 3 {
		t.Error("Элемент 'c' должен остаться")
	}
}

// Тестирует поведение при попытке получения несуществующего ключа
func TestLRU_GetNonExistent(t *testing.T) {
	cache := NewLru(2)

	_, ok := cache.Get("x")
	if ok {
		t.Error("Ожидалось false для отсутствующего ключа")
	}
}
