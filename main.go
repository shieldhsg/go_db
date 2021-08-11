package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"go_db/origin"
)

func main() {
	fmt.Println("开始")

	db := origin.InitDB()
	origin.Query(db)
}
