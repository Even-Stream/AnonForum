package main

import (
	"database/sql"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func create_table(db *sql.DB) {

	createPostsTableSQL := `CREATE TABLE board_posts (
		"Id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"Content" TEXT,
		"Time" TEXT,
		"Parent" INTEGER,
		"Password" TEXT,
		"Identifier" TEXT,
		"File" TEXT,
		"Filename" TEXT,
		"Fileinfo" TEXT,
		"Imgprev" TEXT,
		"Phash" TEXT,
		"Option" TEXT
	);`

	createRepliesTableSQL := `CREATE TABLE board_replies (
		"Source" INTEGER NOT NULL,
		"Replier" INTEGER NOT NULL
	);`


	createSubjectsTableSQL := `CREATE TABLE board_subjects (
		"Parent" INTEGER NOT NULL,
		"Subject" TEXT NOT NULL
	);`

	for _, board := range Boards {
		//fmt.Println(board)
		statement, err := db.Prepare(strings.Replace(createPostsTableSQL, "board", board, 1))
		Err_check(err)
		statement.Exec()

		statement, err = db.Prepare(strings.Replace(createRepliesTableSQL, "board", board, 1))
        	Err_check(err)
		statement.Exec()

		statement, err = db.Prepare(strings.Replace(createSubjectsTableSQL, "board", board, 1))
        	Err_check(err)
		statement.Exec()
	}
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
