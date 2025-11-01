package cache

import (
	"container/list"
)

// представляет одну запись в LRU кеше.
type item struct {
	Key   string
	Value interface{}
}

// LRU — структура кеша.
type LRU struct {
	capacity int
	items    map[string]*list.Element
	queue    *list.List
}

// NewLru создаёт новый LRU кеш с указанной вместимостью.
func NewLru(capacity int) *LRU {
	return &LRU{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		queue:    list.New(),
	}
}

// Set добавляет ключ со значением в кеш или обновляет существующий.
// Если ключ уже есть, его элемент перемещается в начало очереди
// Возвращает true, если ключ уже существовал.
func (c *LRU) Set(key string, value interface{}) (exist bool) {
	if element, exists := c.items[key]; exists {
		c.queue.MoveToFront(element)
		element.Value.(*item).Value = value
		return true
	}

	if c.queue.Len() == c.capacity {
		c.purge()
	}

	item := &item{
		Key:   key,
		Value: value,
	}

	element := c.queue.PushFront(item)
	c.items[item.Key] = element

	return false
}

// purge удаляет самый старый элемент
func (c *LRU) purge() {
	if element := c.queue.Back(); element != nil {
		item := c.queue.Remove(element).(*item)
		delete(c.items, item.Key)
	}
}

// Get возвращает значение по ключу из кеша и перемещает элемент в начало очереди.
// Возвращает false, если ключа нет в кеше.
func (c *LRU) Get(key string) (interface{}, bool) {
	element, exists := c.items[key]
	if !exists {
		return nil, false
	}
	c.queue.MoveToFront(element)
	return element.Value.(*item).Value, true
}
