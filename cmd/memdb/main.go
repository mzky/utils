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
	dbInt.Append("numbers", "even", 2)
	dbInt.Append("numbers", "even", 4)
	dbString.Append("words", "greetings", "hello")
	dbString.Append("words", "greetings", "world")
	dbFloat.Append("decimals", "pi", 3.14)
	dbFloat.Append("decimals1", "e", 2.718)

	// Get data
	fmt.Println(dbInt.Get("numbers", "even"))       // Output: [2 4]
	fmt.Println(dbString.Get("words", "greetings")) // Output: [hello world]
	fmt.Println(dbFloat.Get("decimals", "pi"))      // Output: [3.14]
	fmt.Println(dbFloat.Get("decimals1", "e"))      // Output: [2.718]
	fmt.Println(dbFloat.GetList("decimals.pi"))     // Output: [3.14]

	fmt.Println(dbFloat.GetParentKeys())          //  Output: [decimals decimals1]
	fmt.Println(dbFloat.GetChildKeys("decimals")) // Output:  [pi]

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
