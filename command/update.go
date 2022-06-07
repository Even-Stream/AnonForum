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
    Replies []int
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
		COALESCE(Fileinfo, '') AS Fileinfo, COALESCE(Imgprev, '') Imgprev FROM posts WHERE Parent = ?`)
	Err_check(err)

	//make another statment that searches the replies table(every board will have one) 
	//where the source equals the given id, add the replier to the post's replies array 
	
	conn2, err := sql.Open("sqlite3", BP + "command/post-coll.db")
	Err_check(err)

	stmt2, err := conn2.Prepare(`Select Replier FROM replies WHERE Source = ?`)
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
