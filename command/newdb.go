package main

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func create_table(db *sql.DB) {

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

	createRepliesTableSQL := `CREATE TABLE replies (
		"Source" INTEGER NOT NULL,
		"Replier" INTEGER NOT NULL
	);`

	statement, err := db.Prepare(createPostsTableSQL)
	Err_check(err)
	statement.Exec()

	statement, err = db.Prepare(createRepliesTableSQL)
        Err_check(err)
	statement.Exec()
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
