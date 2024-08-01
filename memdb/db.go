package memdb

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type DB struct {
	Data map[string]interface{}
}

func New() *DB {
	return &DB{Data: make(map[string]interface{})}
}

func (db *DB) Append(value interface{}, key ...string) error {
	k := strings.Join(key, ".")
	return db.Insert(k, value)
}

func (db *DB) Insert(key string, value interface{}) error {
	keys := strings.Split(key, ".")
	current := db.Data
	for i, k := range keys {
		if i == len(keys)-1 {
			// 最后一个键，需要处理值的追加
			if existingValue, exists := current[k]; exists {
				if reflect.TypeOf(existingValue).Kind() == reflect.Slice {
					slice := reflect.ValueOf(existingValue)
					newSlice := reflect.Append(slice, reflect.ValueOf(value))
					current[k] = newSlice.Interface()
				} else {
					// 如果现有值不是切片，将其转换为切片并追加新值
					slice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(existingValue)), 0, 2)
					slice = reflect.Append(slice, reflect.ValueOf(existingValue))
					slice = reflect.Append(slice, reflect.ValueOf(value))
					current[k] = slice.Interface()
				}
			} else {
				// 键不存在，直接赋值
				current[k] = value
			}
		} else {
			if current[k] == nil {
				current[k] = make(map[string]interface{})
			}
			current = current[k].(map[string]interface{})
		}
	}
	return nil
}

func (db *DB) Get(key string) (interface{}, error) {
	keys := strings.Split(key, ".")
	current := db.Data
	for _, k := range keys {
		if val, ok := current[k]; ok {
			if next, ok := val.(map[string]interface{}); ok {
				current = next
			} else {
				return val, nil // 返回非map类型的值
			}
		} else {
			// 提供更详细的错误信息，这里简化为直接返回key未找到的错误
			return nil, fmt.Errorf("key %s not found", key)
		}
	}
	return current, nil
}

// Save 保存到json文件
func (db *DB) Save(filePath string) error {
	jBytes, err := json.Marshal(db.Data)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, jBytes, 0644)
}

// SliceOf 泛型切片类型
type SliceOf[T any] []T

// ToSlice T 泛型类型
func ToSlice[T any](iFace interface{}) SliceOf[T] {
	if iFace == nil {
		// 对于空 Data 或空接口值，直接返回一个空的目标类型切片，而不是nil，
		// 更好地区分处理空切片和nil切片的情况
		return SliceOf[T]{}
	}
	v := reflect.ValueOf(iFace)
	if v.Kind() != reflect.Slice {
		return SliceOf[T]{}
	}

	// 创建目标类型的切片。这里不能预先分配长度为0的切片，因为v.Len()在v.Len()==0时仍会执行。
	result := make(SliceOf[T], v.Len())

	for i := 0; i < v.Len(); i++ {
		val := v.Index(i)
		// 检查类型兼容性
		if !val.Type().AssignableTo(reflect.TypeOf((*T)(nil)).Elem()) {
			// 当类型不匹配时，返回一个空切片而非nil，以保持函数行为的一致性
			return SliceOf[T]{}
		}
		// 安全的类型断言。这里使用了val.Convert而不是直接的类型断言，以避免可能的panic。
		// 注意，这种方法在类型不匹配时会返回零值，因此前面的类型检查仍然必要。
		result[i] = val.Convert(reflect.TypeOf((*T)(nil)).Elem()).Interface().(T)
	}

	return result
}
