package main

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	gokitsql "github.com/getbread/gokit/storage/sql"
)

type Entity struct {
	Id      string `db:"IDX"`
	Version uint32
	Name    string
}

var COLUMNS = gokitsql.StructDBFields(Entity{})

func main() {
	// Generating SQL using unsafe STAR operator
	genSql([]string {"*"})

	// Generating SQL gokit.sql utility method to convert struct fields into columns
	genSql(COLUMNS)
}

func genSql(columns []string) {
	fmt.Printf("Generating SQL for columns: [%s]\n", columns)
	params := make(map[string]interface{})
	params["ID"] = "123"
	params["VERSION"] = "5"
	queryBuilder := squirrel.Select(columns...).From("TABLE_X").Where(params).Offset(4)

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to construct SQL query: %w", err))
	}

	fmt.Printf("SQL: [%s]\n", sql)
	fmt.Printf("Params: [%v]\n", args)
}
