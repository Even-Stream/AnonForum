package main 

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var Max_conns = 5
var Conns = make(chan *sql.Stmt, Max_conns)

func Checkout() *sql.Stmt {
  return <-Conns
}

func Checkin(c *sql.Stmt) {
  Conns <- c
}

func Make_Conns() {
	for i := 0; i < Max_conns; i++ {
		conn, err := sql.Open("sqlite3", BP + "command/post-coll.db")
		Err_check(err)
		stmt, err := conn.Prepare(`SELECT Content, 
			COALESCE(Imgprev, '') Imgprev FROM posts WHERE id = ?`)
		Err_check(err)	
	
		Conns <- stmt
	}
}