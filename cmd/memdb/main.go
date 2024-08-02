package main

import (
	"fmt"
	"github.com/mzky/utils/memdb"
	"log"
)

func main() {
	db := memdb.New()

	keys := []string{"aa", "bb", "key.key2", "key.key2", "key.key2", "k.k2.k3", "k.k2.k3"}
	values := []interface{}{22, "11.1", 123, 31.2, 64.1, "v1", "v2"}
	// 同一个key后续插入的值转换为第一个值的类型，将所有数值类型转换为float64,避免int和float64类型转换错误
	for i, key := range keys {
		if err := db.Insert(key, values[i]); err != nil {
			log.Printf("Error setting value for key %s: %v", key, err)
		}
	}

	f, err := db.Get("key.key2")
	fmt.Println(f, err)
	s, err := db.Get("k.k2.k3")
	fmt.Println(s, err)
	a, err := db.Get("aa")
	fmt.Println(a, err)
	b, err := db.Get("bb")
	fmt.Println(a, err)

	fmt.Println(memdb.ToSlice[float64](f))
	fmt.Println(memdb.ToSlice[string](s))
	ss := memdb.ToFloat64(a)
	ii := memdb.ToString(b)
	bb := memdb.ToFloat64(b)
	aa := memdb.ToString(a)
	fmt.Println(ss, ii, bb, aa, "===")

	fmt.Println(db.GetKeys(""))
	fmt.Println(db.GetKeys("k"))
	fmt.Println(db.GetKeys("k.k2"))
	db.Save("data.json")
}
