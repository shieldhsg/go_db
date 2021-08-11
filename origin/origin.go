package origin

import (
	"database/sql"
	"fmt"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Sex   int    `json:"sex"`
	Phone string `json:"phone"`
}

func InitDB() (db *sql.DB) {
	DB, _ := sql.Open("mysql", "root:123@tcp(127.0.0.1:3306)/test")
	// 设置数据库最大连接数
	DB.SetConnMaxLifetime(100)
	// 设置数据库最大闲置连接数
	DB.SetConnMaxIdleTime(10)
	// 验证连接
	if err := DB.Ping(); err != nil {
		fmt.Println("open database fail")
		return
	}
	fmt.Println("connect success")
	return DB
}

func Query(db *sql.DB) {
	user := &User{}
	rows, err := db.Query("select * from user")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for rows.Next() {
		err = rows.Scan(&user.ID, &user.Name, &user.Age, &user.Sex, &user.Phone)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Printf("%#+v", user)
		fmt.Println("")
	}
	rows.Close()
}
