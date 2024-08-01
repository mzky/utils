package main

import (
	"fmt"
	"github.com/mzky/utils/memdb"
	"log"
)

func main() {
	db := memdb.New()

	keys := []string{"aa", "key.key2", "key.key2", "key.key2", "k.k2.k3", "k.k2.k3"}
	values := []interface{}{"1", 123.1, 34.1, 64.1, "v1", "v2"}

	for i, key := range keys {
		if err := db.Insert(key, values[i]); err != nil {
			log.Printf("Error setting value for key %s: %v", key, err)
		}
	}

	f, err := db.Get("key.key2")
	fmt.Println(f, err)
	s, err := db.Get("k.k2.k3")
	fmt.Println(s, err)

	fmt.Println(memdb.ToSlice[float64](f))
	fmt.Println(memdb.ToSlice[string](s))

	db.Save("data.json")
}
