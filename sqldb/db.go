package sqldb

import (
	"database/sql"
	"fmt"
)

var DB *sql.DB

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// ConnectDB opens a connection to the database
func ConnectDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./commitMessages.db")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Connected to DB")

	DB = db
	return db
}

// func addCommitToDB(message message) {

// 	stmt, err := db.Prepare("INSERT INTO commits (id, label, message) VALUES (?, ?, ?)")
// 	checkErr(err)
// 	stmt.Exec(nil, nil, message.Message)
// 	defer stmt.Close()

// 	fmt.Println("Added commit to DB")
// }