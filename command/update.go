package main 

import (
	"database/sql"
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
}

type Thread struct {
    Board string
    Subject string
    Posts []*Post
}


func get_posts(parent int) []*Post {
 	conn, err := sql.Open("sqlite3", BP + "command/post-coll.db")
  	Err_check(err)

	stmt, err := conn.Prepare(`SELECT Id, Content, Time, COALESCE(File, '') AS File, COALESCE(Filename, '') AS Filename, 
		COALESCE(Fileinfo, '') AS Fileinfo, COALESCE(Imgprev, '') Imgprev FROM posts WHERE parent = ?`)
	Err_check(err)

	rows, err := stmt.Query(parent)
	Err_check(err)
	defer rows.Close()
	
	var thread_body []*Post

	for rows.Next() {
        	var pst Post
        	err = rows.Scan(&pst.Id, &pst.Content, &pst.Time, &pst.File,
            		&pst.Filename, &pst.Fileinfo, &pst.Imgprev)
		Err_check(err)
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