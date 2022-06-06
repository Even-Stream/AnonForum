package main

import (
	"database/sql"
	"log"
	"os"
	"io/fs"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func find(root, ext string) []string {
   var a []string
   filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
      if e != nil { return e }
      if filepath.Ext(d.Name()) == ext {
         a = append(a, s)
      }
      return nil
   })
   return a
}

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

func insert_post(db *sql.DB, content string) {
	log.Println("Inserting post...")
	insertPostSQL := `INSERT INTO posts(Content) VALUES (?)`
	statement, err := db.Prepare(insertPostSQL) 

	Err_check(err)
	_, err = statement.Exec(content)
	Err_check(err)
}

func New_db() {

	log.Println("creating post-coll.db...")
	file, err := os.Create(BP + "command/post-coll.db") 
	Err_check(err)	

	file.Close()
	log.Println("post-coll.db created")

	conn, err := sql.Open("sqlite3", BP + "command/post-coll.db") 
	Err_check(err)
	defer conn.Close() 
	
	create_table(conn) 

	//Inserts content from text files
	for _, p := range find(BP, ".txt") {

		content, err := os.ReadFile(p)
		Err_check(err)

		insert_post(conn, string(content))
	}
}
