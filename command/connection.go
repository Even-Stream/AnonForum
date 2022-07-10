package main 

import (
	"database/sql"
	"strings"
	_ "github.com/mattn/go-sqlite3"
)

var Max_conns = 5
var readConns = make(chan map[string]map[string]*sql.Stmt, Max_conns)
var writeConns = make(chan map[string]map[string]*sql.Stmt, 1)

//statement strings
const (
	prev_string = `SELECT Content, 
			COALESCE(Imgprev, '') Imgprev FROM board_posts WHERE id = ?`
	prev_parentstring = `SELECT Parent FROM board_posts WHERE id = ?`
	updatestring = `SELECT Id, Content, Time, COALESCE(File, '') AS File, COALESCE(Filename, '') AS Filename, 
				COALESCE(Fileinfo, '') AS Fileinfo, COALESCE(Imgprev, '') Imgprev FROM board_posts WHERE Parent = ?`
	update_repstring = `SELECT Replier FROM board_replies WHERE Source = ?`
	parent_collstring = `SELECT DISTINCT Parent FROM board_posts ORDER BY Id DESC LIMIT 10`
	thread_headstring = `SELECT Content, Time, COALESCE(File, '') AS File, COALESCE(Filename, '') AS Filename, 
				COALESCE(Fileinfo, '') AS Fileinfo, COALESCE(Imgprev, '') Imgprev
				FROM board_posts
				WHERE Id = ?`
	thread_bodystring = `SELECT * FROM (
				SELECT Id, Content, Time, COALESCE(File, '') AS File, COALESCE(Filename, '') AS Filename, 
				COALESCE(Fileinfo, '') AS Fileinfo, COALESCE(Imgprev, '') Imgprev FROM board_posts 
				WHERE Parent = ? AND Id != Parent ORDER BY Id DESC LIMIT 5)
				ORDER BY Id ASC`
	lastid_string = `SELECT IFNULL ((SELECT MAX(Id) FROM board_posts), 0)`
	parent_checkstring = `SELECT COUNT(*)
				FROM board_posts
				WHERE Parent = ?`
	thread_collstring = `SELECT DISTINCT Parent FROM board_posts ORDER BY Id DESC`
	subject_lookstring = `SELECT Subject FROM board_subjects WHERE Parent = ?`

	newpost_wfstring = `INSERT INTO board_posts(Content, Time, Parent, File, Filename, Fileinfo, Imgprev) VALUES (?, ?, ?, ?, ?, ?, ?)`
	newpost_nfstring = `INSERT INTO board_posts(Content, Time, Parent) VALUES (?, ?, ?)`
	repadd_string = `INSERT INTO board_replies(Source, Replier) VALUES (?, ?)`
	subadd_string = `INSERT INTO board_subjects(Parent, Subject) VALUES (?, ?)`
)

func Checkout() map[string]map[string]*sql.Stmt {
  return <-readConns
}
func Checkin(c map[string]map[string]*sql.Stmt) {
  readConns <- c
}

func writeCheckout() map[string]map[string]*sql.Stmt {
  return <-writeConns
}
func writeCheckin(c map[string]map[string]*sql.Stmt) {
  writeConns <- c
}


func Make_Conns() {
	db_path := BP + "command/post-coll.db" 
	db_uri := "file://" + db_path + "?cache=private&_synchronous=NORMAL&_journal_mode=WAL"
	
	for i := 0; i < Max_conns; i++ {

		read_stmts := map[string]map[string]*sql.Stmt{}

		for _, board := range Boards {

			//preview statements
			conn1, err := sql.Open("sqlite3", db_uri)
			Err_check(err)
		
			prev_stmt, err := conn1.Prepare(strings.Replace(prev_string, "board", board, 1))
			Err_check(err)

			conn2, err := sql.Open("sqlite3", db_uri)
			Err_check(err)

			prev_parentstmt, err := conn2.Prepare(strings.Replace(prev_parentstring, "board", board, 1))
			Err_check(err)


			//thread update statements
			conn3, err := sql.Open("sqlite3", db_uri)
			Err_check(err)

			updatestmt, err := conn3.Prepare(strings.Replace(updatestring, "board", board, 1))
			Err_check(err)

			conn4, err := sql.Open("sqlite3", db_uri)
			Err_check(err)

			update_repstmt, err := conn4.Prepare(strings.Replace(update_repstring, "board", board, 1))
			Err_check(err)

		
			//board upate statements
			conn5, err := sql.Open("sqlite3", db_uri)
			Err_check(err)

			parent_collstmt, err := conn5.Prepare(strings.Replace(parent_collstring, "board", board, 1))
			Err_check(err)

			conn6, err := sql.Open("sqlite3", db_uri)
			Err_check(err)

			thread_headstmt, err := conn6.Prepare(strings.Replace(thread_headstring, "board", board, 1))
			Err_check(err)

			conn7, err := sql.Open("sqlite3", db_uri)
			Err_check(err)

			thread_bodystmt, err := conn7.Prepare(strings.Replace(thread_bodystring, "board", board, 1))
			Err_check(err)

			conn8, err := sql.Open("sqlite3", db_uri)
			Err_check(err)

			lastid_stmt, err := conn8.Prepare(strings.Replace(lastid_string, "board", board, 1))
			Err_check(err)

			conn9, err := sql.Open("sqlite3", db_uri)
			Err_check(err)

			parent_checkstmt, err := conn9.Prepare(strings.Replace(parent_checkstring, "board", board, 1))
			Err_check(err)

			//catalog update statement
			conn10, err:= sql.Open("sqlite3", db_uri)
			Err_check(err)	

			thread_collstmt, err := conn10.Prepare(strings.Replace(thread_collstring, "board", board, 1))
			Err_check(err)

			//subject lookup
			conn11, err:= sql.Open("sqlite3", db_uri)
			Err_check(err)	

			subject_lookstmt, err := conn11.Prepare(strings.Replace(subject_lookstring, "board", board, 1))
			Err_check(err)

			read_stmts[board] = map[string]*sql.Stmt{"prev": prev_stmt, "prev_parent": prev_parentstmt, "update": updatestmt, "update_rep": update_repstmt, 
			"parent_coll": parent_collstmt, "thread_head": thread_headstmt, "thread_body": thread_bodystmt, 
			"lastid": lastid_stmt, "parent_check": parent_checkstmt, "thread_coll": thread_collstmt, "subject_look": subject_lookstmt}
		}
		readConns <- read_stmts
	}

	write_stmts := map[string]map[string]*sql.Stmt{}

	for _, board := range Boards {
	
		conn12, err := sql.Open("sqlite3", db_uri)
			Err_check(err)

		newpost_wfstmt, err := conn12.Prepare(strings.Replace(newpost_wfstring, "board", board, 1))
		Err_check(err)


		conn13, err := sql.Open("sqlite3", db_uri)
		Err_check(err)

		newpost_nfstmt, err := conn13.Prepare(strings.Replace(newpost_nfstring, "board", board, 1))
		Err_check(err)

		conn14, err := sql.Open("sqlite3", db_uri)
		Err_check(err)

		repadd_stmt, err := conn14.Prepare(strings.Replace(repadd_string, "board", board, 1))
		Err_check(err)

		conn15, err := sql.Open("sqlite3", db_uri)
		Err_check(err)

		subadd_stmt, err := conn15.Prepare(strings.Replace(subadd_string, "board", board, 1))
		Err_check(err)

		write_stmts[board] = map[string]*sql.Stmt{"newpost_wf": newpost_wfstmt, "newpost_nf": newpost_nfstmt, "repadd": repadd_stmt, "subadd": subadd_stmt}
	}
	writeConns <- write_stmts
}