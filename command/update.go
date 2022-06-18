package main

import (
	"os"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)


type Post struct {
    Id int
    Content string
    Time string
    File string
    Filename string
    Fileinfo string
    Imgprev string
    Replies []int
}

type Thread struct {
    Parent string
    Subject string
    Posts []*Post
}

type Board struct {
		Name string 
		Threads []*Thread
}


func get_threads(board string) ([]*Thread, error) {
	stmts := Checkout()
	defer Checkin(stmts)
	stmt := stmts["update_board"]

//tables will be called a board 	
	rows, err := stmt.Query()
	Err_check(err)
	defer rows.Close()

	var board_body []*Thread

	for rows.Next() {
		var thr Thread
		err = rows.Scan(&thr.Parent)
		Err_check(err)

		board_body = append(board_body, &thr)
	}	

	return board_body, err
}

func get_posts(parent string) ([]*Post, error) {

	stmts := Checkout()
  defer Checkin(stmts)

  stmt := stmts["update"]
	stmt2 := stmts["update_rep"]

	rows, err := stmt.Query(parent)
	Err_check(err)
	defer rows.Close()

	var thread_body []*Post

	for rows.Next() {
		var pst Post
		err = rows.Scan(&pst.Id, &pst.Content, &pst.Time, &pst.File,
			&pst.Filename, &pst.Fileinfo, &pst.Imgprev)
		Err_check(err)

		rep_rows, err := stmt2.Query(pst.Id)
		Err_check(err)

		for rep_rows.Next() {
			var replier int
			rep_rows.Scan(&replier)
			pst.Replies = append(pst.Replies, replier)
		}
		
		rep_rows.Close()
		thread_body = append(thread_body, &pst)
	}

	return thread_body, err
}

func Build_board(name string) {
	boardtemp := template.New("board.html")
	boardtemp, err := boardtemp.ParseFiles(BP + "/templates/board.html")
	Err_check(err)

	f, err := os.Create(BP + "head/" + name + "/index.html")
	Err_check(err)
	defer f.Close()

	threads, err := get_threads(name)

	if err == nil {
		board := Board{Name: name, Threads: threads}
		boardtemp.Execute(f, board)
	}
}

func Build_thread(parent string) { //will accept argument for board and thread number
    threadtemp := template.New("thread.html")
    threadtemp, err := threadtemp.ParseFiles(BP + "/templates/thread.html")
    Err_check(err)

    f, err := os.Create(BP + "head/ot/" + parent + ".html")
    Err_check(err)
    defer f.Close()

    posts, err := get_posts(parent)
    
    if err == nil {
        thread := Thread{Posts: posts, Subject: "Templates", Parent: parent}
        threadtemp.Execute(f, thread)
    }
}
