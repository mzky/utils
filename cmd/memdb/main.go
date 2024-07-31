package main

import (
	"encoding/json"
	"fmt"
	"github.com/mzky/utils/memoryDB"
	"os"
)

// 必须Go 1.18以上
func main() {
	// Initialize databases for different types
	dbInt := memoryDB.New[int]()
	dbString := memoryDB.New[string]()
	dbFloat := memoryDB.New[float64]()

	// Append data
	dbInt.Append("numbers", "even", "a1", 2)
	dbInt.Append("numbers", "even", "a2", 4)
	dbString.Append("words", "greetings", "a1", "hello")
	dbString.Append("words", "greetings", "a2", "world")
	dbFloat.Append("decimals", "pi", "a1", 3.14)
	dbFloat.Append("decimals1", "e", "a2", 2.718)

	// Get data
	fmt.Println(dbInt.Get("numbers", "even", "a1"))
	fmt.Println(dbString.Get("words", "greetings", "a2"))
	fmt.Println(dbFloat.Get("decimals", "pi", "a1"))
	fmt.Println(dbFloat.Get("decimals1", "e", "a2"))
	fmt.Println(dbFloat.GetList("decimals.pi.a1"))
	fmt.Println(dbFloat.GetSeries("decimals", "pi"))

	fmt.Println(dbFloat.GetParentKeys())
	fmt.Println(dbFloat.GetChildKeys("decimals"))

	// Save data to JSON
	marshal, err := json.Marshal(dbInt.Data)
	if err != nil {
		fmt.Println("Error marshaling data:", err)
		return
	}
	fmt.Println(string(marshal)) // Output: {"numbers":{"even":[2,4]}}
	err = os.WriteFile("data.json", marshal, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
}
