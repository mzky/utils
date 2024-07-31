package memoryDB

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// SliceOf is a type parameter that represents any slice type.
type SliceOf[T any] []T

// MemoryDB defines an in-memory database type with type-aware slices.
type MemoryDB[T any] struct {
	Data map[string]map[string]SliceOf[T]
	mu   sync.RWMutex // Adds read-write lock for concurrent safety
}

// New initializes an in-memory database.
func New[T any]() *MemoryDB[T] {
	return &MemoryDB[T]{
		Data: make(map[string]map[string]SliceOf[T]),
	}
}

// Set sets a value associated with a parent key and a child key.
func (db *MemoryDB[T]) Set(parentKey string, childKey string, value T) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.Data[parentKey]; !ok {
		db.Data[parentKey] = make(map[string]SliceOf[T])
	}
	db.Data[parentKey][childKey] = append(db.Data[parentKey][childKey], value)
}

// Get retrieves a slice of values associated with a parent key and a child key.
func (db *MemoryDB[T]) Get(parentKey string, childKey string) SliceOf[T] {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if data, ok := db.Data[parentKey]; ok {
		if values, ok := data[childKey]; ok {
			return values
		}
	}
	return nil
}

// GetList retrieves a slice of values associated with a parent key and a child key.
func (db *MemoryDB[T]) GetList(Key string) SliceOf[T] {
	db.mu.RLock()
	defer db.mu.RUnlock()

	k := strings.Split(Key, ".")
	if len(k) != 2 {
		return nil
	}
	parentKey := k[0]
	childKey := k[1]
	if data, ok := db.Data[parentKey]; ok {
		if values, ok := data[childKey]; ok {
			return values
		}
	}
	return nil
}

func (db *MemoryDB[T]) GetParentKeys() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var list []string
	for s, _ := range db.Data {
		list = append(list, s)
	}
	return list
}
func (db *MemoryDB[T]) GetChildKeys(parentKey string) []string {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if data, ok := db.Data[parentKey]; ok {
		var list []string
		for s, _ := range data {
			list = append(list, s)
		}
		return list
	}
	return nil
}

// Append appends a value to the slice associated with a parent key and a child key.
func (db *MemoryDB[T]) Append(parentKey string, childKey string, value T) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.Data[parentKey]; !ok {
		db.Data[parentKey] = make(map[string]SliceOf[T])
	}
	db.Data[parentKey][childKey] = append(db.Data[parentKey][childKey], value)
}

// Delete removes the slice associated with a parent key and a child key.
func (db *MemoryDB[T]) Delete(parentKey string, childKey string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if data, ok := db.Data[parentKey]; ok {
		delete(data, childKey)
		if len(data) == 0 {
			delete(db.Data, parentKey)
		}
	}
}

// Clear clears the entire database.
func (db *MemoryDB[T]) Clear() {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.Data = make(map[string]map[string]SliceOf[T])
}

// GetTypeKey returns a unique identifier for a type.
func GetTypeKey(t reflect.Type) string {
	return fmt.Sprintf("%s", t)
}
