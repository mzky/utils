package main

import (
	"fmt"
	"github.com/mzky/utils/memoryDB"
	"math/rand"
)

// 必须Go 1.18以上

// 生成随机浮点数
func randomFloat() float64 {
	return 1 + rand.Float64()*(100-1)
}
func main() {
	// Initialize databases for different types
	dbInt := memoryDB.New[int]()
	dbString := memoryDB.New[string]()
	dbFloat := memoryDB.New[float64]()

	// Append data
	dbInt.Append("numbers.even.a1", 2)
	dbInt.Set("numbers", "even", "a2", 4)

	dbString.Append("words.greetings.a1", "hello")
	dbString.Set("words", "greetings", "a2", "world")

	dbFloat.Append("decimals.pi.a1", 3.14)
	dbFloat.Set("decimals1", "e", "a2", 2.718)

	// Get data
	fmt.Println(dbInt.Get("numbers", "even", "a1"))
	fmt.Println(dbString.Get("words", "greetings", "a2"))
	fmt.Println(dbFloat.Get("decimals", "pi", "a1"))

	fmt.Println(dbFloat.GetList("decimals.pi.a1"))
	fmt.Println(dbFloat.GetSeriesList("decimals.pi"))
	fmt.Println(dbFloat.GetSeries("decimals", "pi"))

	fmt.Println(dbFloat.GetParentKeys())
	fmt.Println(dbFloat.GetChildKeys("decimals"))

	for i := 0; i < 100; i++ {
		dbFloat.Append("decimals.pi.a1", randomFloat())
	}
	fmt.Println(dbFloat.GetList("decimals.pi.a1"))
	dbFloat.Save("data.json")
	var df = memoryDB.New[float64]()
	df.Load("data.json")
	fmt.Println(df.Data)
}
