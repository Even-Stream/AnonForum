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
    Board string
    Subject string
    Posts []*Post
}


func get_posts(parent int) []*Post {

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

	return thread_body
}

func Build_thread() {
    threadtemp := template.New("thread.html")
    threadtemp, err := threadtemp.ParseFiles(BP + "/templates/thread.html")
    Err_check(err)

    f, err := os.Create(BP + "index.html")
    Err_check(err)
    defer f.Close()

    thread := Thread{Posts: get_posts(1), Subject: "Templates"}

    threadtemp.Execute(f, thread)
}
