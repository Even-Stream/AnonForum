package main

import (
	"os"
	"text/template"
	"strconv"

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
	BoardN string
	Parent string
	Subject string
	Posts []*Post
}

type Board struct {
	Name string 
	Threads []*Thread
	Latest int
}

type Catalog struct {
	Name string
	Posts []*Post
	Subjects []string
}


var catfuncmap = template.FuncMap{
	"startrow": func(rowsize, index int) bool {
		if index % rowsize == 0 {
			return true
		}
		return false 
	},
}

func get_cat_posts(board string) ([]*Post, []string) {
	stmts := Checkout()
	defer Checkin(stmts)

	stmt0 := stmts["thread_coll"]
	stmt := stmts["thread_head"]

	var cat_body []*Post
	var subjects []string

	parent_rows, err := stmt0.Query()
	Err_check(err)
	defer parent_rows.Close()

	for parent_rows.Next() {
		var cparent Post
		
		err = parent_rows.Scan(&cparent.Id)
		Err_check(err)
		err = stmt.QueryRow(cparent.Id).Scan(&cparent.Content, &cparent.Time, &cparent.File,
			&cparent.Filename, &cparent.Fileinfo, &cparent.Imgprev)
		Query_err_check(err)

		cat_body = append(cat_body, &cparent)
		subjects = append(subjects, "Template")
	}

	return cat_body, subjects
}

func get_threads(board string) ([]*Thread, int) {
	stmts := Checkout()
	defer Checkin(stmts)
	
	stmt0 := stmts["parent_coll"]
	stmt := stmts["thread_head"]
	stmt2 := stmts["thread_body"]
	stmt3 := stmts["update_rep"]
	stmt4 := stmts["lastid"]

	var board_body []*Thread

	//tables will be called a board 
	parent_rows, err := stmt0.Query()
	Err_check(err)
	defer parent_rows.Close()
	
	for parent_rows.Next() {
		var fstpst Post
		var pst_coll []*Post
		
		err = parent_rows.Scan(&fstpst.Id)
		Err_check(err)
		err = stmt.QueryRow(fstpst.Id).Scan(&fstpst.Content, &fstpst.Time, &fstpst.File,
			&fstpst.Filename, &fstpst.Fileinfo, &fstpst.Imgprev)
		Query_err_check(err)

		pst_coll = append(pst_coll, &fstpst)

		thread_rows, err := stmt2.Query(fstpst.Id)
		Err_check(err)
		defer thread_rows.Close()

		for thread_rows.Next() {
			var cpst Post

			err = thread_rows.Scan(&cpst.Id, &cpst.Content, &cpst.Time, &cpst.File,
			&cpst.Filename, &cpst.Fileinfo, &cpst.Imgprev)
			Err_check(err)
			
			pst_coll = append(pst_coll, &cpst)
		}

		for _, pst := range pst_coll {
			rep_rows, err := stmt3.Query(pst.Id)
			Err_check(err)
			
			for rep_rows.Next() {
				var replier int
				rep_rows.Scan(&replier)
				pst.Replies = append(pst.Replies, replier)
			}
			rep_rows.Close()
		}

		thr := Thread{BoardN: board, Posts: pst_coll, Subject: "Templates", Parent: strconv.Itoa(fstpst.Id)}
		board_body = append(board_body, &thr)
	}
	
	var latestid int
	err = stmt4.QueryRow().Scan(&latestid)
	Query_err_check(err)
	//latestid will equal 0 when there are no posts yet
	latestid++

	return board_body, latestid
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

func Build_catalog(name string) {
	cattemp := template.New("catalog.html").Funcs(catfuncmap)
	cattemp, err := cattemp.ParseFiles(BP + "/templates/catalog.html")
	Err_check(err)

	f, err := os.Create(BP + "head/" + name + "/catalog.html")
	Err_check(err)
	defer f.Close()

	posts, subjects := get_cat_posts(name)
    
	catalog := Catalog{Name: name, Posts: posts, Subjects: subjects}
	cattemp.Execute(f, catalog)
}

func Build_board(name string) {
	boardtemp := template.New("board.html")
	boardtemp, err := boardtemp.ParseFiles(BP + "/templates/board.html")
	Err_check(err)

	f, err := os.Create(BP + "head/" + name + "/index.html")
	Err_check(err)
	defer f.Close()

	threads, latestid := get_threads(name)

	board := Board{Name: name, Threads: threads, Latest: latestid}
	boardtemp.Execute(f, board)
	
}

func Build_thread(parent string, boardn string) { //will accept argument for board and thread number
	threadtemp := template.New("thread.html")
	threadtemp, err := threadtemp.ParseFiles(BP + "/templates/thread.html")
	Err_check(err)

	f, err := os.Create(BP + "head/ot/" + parent + ".html")
	Err_check(err)
	defer f.Close()

	posts, err := get_posts(parent)
    
	if err == nil {
		thread := Thread{BoardN: boardn, Posts: posts, Subject: "Templates", Parent: parent}
		threadtemp.Execute(f, thread)
	}
}