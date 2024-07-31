package memoryDB

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
)

// SliceOf is a type parameter that represents any slice type.
type SliceOf[T any] []T

type MapOfSliceOf[T any] map[string]SliceOf[T]

// MemoryDB defines an in-memory database type with type-aware slices.
type MemoryDB[T any] struct {
	Data map[string]map[string]MapOfSliceOf[T]
	mu   sync.RWMutex // Adds read-write lock for concurrent safety
}

// New initializes an in-memory database.
func New[T any]() *MemoryDB[T] {
	return &MemoryDB[T]{
		Data: make(map[string]map[string]MapOfSliceOf[T]),
	}
}

// Set sets a value associated with a parent key and a child key.
func (db *MemoryDB[T]) Set(parentKey, childKey, series string, value T) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.Data[parentKey]; !ok {
		db.Data[parentKey] = make(map[string]MapOfSliceOf[T])
	}
	if _, ok := db.Data[parentKey][childKey]; !ok {
		db.Data[parentKey][childKey] = make(MapOfSliceOf[T])
	}
	db.Data[parentKey][childKey][series] = append(db.Data[parentKey][childKey][series], value)
}

// Get retrieves a slice of values associated with a parent key and a child key.
func (db *MemoryDB[T]) Get(parentKey, childKey, series string) SliceOf[T] {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if data, ok := db.Data[parentKey]; ok {
		if values, ok := data[childKey]; ok {
			if v, ok := values[series]; ok {
				return v
			}
		}
	}
	return []T{}
}

// GetSeries retrieves a map of values associated with a parent key and a child key.
func (db *MemoryDB[T]) GetSeries(parentKey, childKey string) MapOfSliceOf[T] {
	db.mu.RLock()
	defer db.mu.RUnlock()

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

	db.Data = make(map[string]map[string]MapOfSliceOf[T])
}

// GetTypeKey returns a unique identifier for a type.
func GetTypeKey(t reflect.Type) string {
	return fmt.Sprintf("%s", t)
}

// GetList retrieves a slice of values associated with a parent key and a child key.
func (db *MemoryDB[T]) GetList(Key string) SliceOf[T] {
	k := strings.Split(Key, ".")
	if len(k) != 3 {
		return []T{}
	}
	return db.Get(k[0], k[1], k[2])
}

// GetSeriesList retrieves a map of values associated with a parent key and a child key.
func (db *MemoryDB[T]) GetSeriesList(Key string) MapOfSliceOf[T] {
	k := strings.Split(Key, ".")
	if len(k) != 2 {
		return nil
	}
	return db.GetSeries(k[0], k[1])
}

// Append appends a value to the slice associated with a parent key and a child key.
func (db *MemoryDB[T]) Append(key string, value T) {
	k := strings.Split(key, ".")
	if len(k) != 3 {
		return
	}
	db.Set(k[0], k[1], k[2], value)
}

func (db *MemoryDB[T]) Save(fileName string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	bytes, err := json.Marshal(db.Data)
	if err != nil {
		return err
	}
	return os.WriteFile(fileName, bytes, 0644) // Adjusted file permissions for better security
}

func (db *MemoryDB[T]) Load(fileName string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &db.Data)
}
