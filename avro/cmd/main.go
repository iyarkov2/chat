package main

import (
	"encoding/json"
	"fmt"
	"github.com/iyarkov2/chat/avro/schema/generated"
)

func main() {
	record := generated.NewTestRecord()

	fmt.Println("Test record: ", record)

	var f interface{}
	err := json.Unmarshal([]byte(record.Schema()), &f)
	if err != nil {
		panic(fmt.Errorf("schema unmarshar error, %v", err))
	}
	m := f.(map[string]interface{})

	fmt.Println("Schema Version: ", m["version"])
}
