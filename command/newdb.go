package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func create_table(db *sql.DB) {
	//SQL command 
	createPostsTableSQL := `CREATE TABLE posts (
		"Id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"Content" TEXT,
		"Time" TEXT,
		"Parent" INTEGER,
		"Password" TEXT,
		"Identifier" TEXT,
		"File" TEXT,
		"Filename" TEXT,
		"Fileinfo" TEXT,
		"Imgprev" TEXT
	);` 

	log.Println("Creating posts table...")
	statement, err := db.Prepare(createPostsTableSQL) //"prepares sql statement"
	Err_check(err)

	statement.Exec() //executes the statements
	log.Println("posts table created")
}

func New_db() {

	file, err := os.Create(BP + "command/post-coll.db") 
	Err_check(err)	

	file.Close()

	conn, err := sql.Open("sqlite3", BP + "command/post-coll.db") 
	Err_check(err)
	defer conn.Close() 
	
	create_table(conn) 
}
