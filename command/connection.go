package main 

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var Max_conns = 5
var readConns = make(chan map[string]*sql.Stmt, Max_conns)
var writeConns = make(chan map[string]*sql.Stmt, 1)

func Checkout() map[string]*sql.Stmt {
  return <-readConns
}
func Checkin(c map[string]*sql.Stmt) {
  readConns <- c
}

func writeCheckout() map[string]*sql.Stmt {
  return <-writeConns
}
func writeCheckin(c map[string]*sql.Stmt) {
  writeConns <- c
}


func Make_Conns() {
	db_path := BP + "command/post-coll.db" 
	db_uri := "file://" + db_path + "?cache=private&_synchronous=NORMAL&_journal_mode=WAL"

	for i := 0; i < Max_conns; i++ {
	
		conn1, err := sql.Open("sqlite3", db_uri)
		Err_check(err)
		
		prev_stmt, err := conn1.Prepare(`SELECT Content, 
			COALESCE(Imgprev, '') Imgprev FROM posts WHERE id = ?`)
		Err_check(err)	
		
		conn4, err := sql.Open("sqlite3", db_uri)
		Err_check(err)

		updatestmt, err := conn4.Prepare(`SELECT Id, Content, Time, COALESCE(File, '') AS File, COALESCE(Filename, '') AS Filename, 
				COALESCE(Fileinfo, '') AS Fileinfo, COALESCE(Imgprev, '') Imgprev FROM posts WHERE Parent = ?`)
		Err_check(err)


		conn5, err := sql.Open("sqlite3", db_uri)
		Err_check(err)

		update_repstmt, err := conn5.Prepare(`Select Replier FROM replies WHERE Source = ?`)
		Err_check(err)
		

		stmts := map[string]*sql.Stmt{"prev": prev_stmt, "update": updatestmt, "update_rep": update_repstmt}
		readConns <- stmts
	}

	conn2, err := sql.Open("sqlite3", db_uri)
	Err_check(err)

	newpost_wfstmt, err := conn2.Prepare(`INSERT INTO posts(Content, Time, Parent, File, Filename, Fileinfo, Imgprev) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	Err_check(err)


	conn3, err := sql.Open("sqlite3", db_uri)
	Err_check(err)

	newpost_nfstmt, err := conn3.Prepare(`INSERT INTO posts(Content, Time, Parent) VALUES (?, ?, ?)`)
	Err_check(err)

	conn6, err := sql.Open("sqlite3", db_uri)
	Err_check(err)

	repadd_stmt, err := conn6.Prepare(`INSERT INTO replies(Source, Replier) VALUES (?, ?)`)
	Err_check(err)

	stmts := map[string]*sql.Stmt{"newpost_wf": newpost_wfstmt, "newpost_nf": newpost_nfstmt, "repadd": repadd_stmt}
	writeConns <- stmts
}