package main

import (
    "os"
    "text/template"
    "strconv"
    "errors"

    _ "github.com/mattn/go-sqlite3"
)

//structures used in templates
type Post struct {
    Id int
    Content string
    Time string
    File string
    Filename string
    Fileinfo string
    Imgprev string
    Option string
    Replies []int
}

type Thread struct {
    BoardN string
    Parent string
    Subject string
    Posts []*Post
    Header []string
}

type Board struct {
    Name string
    Desc string
    Threads []*Thread
    Header []string
}

func Dir_check(path string) {

    if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
        err := os.Mkdir(path, os.ModePerm)
        Err_check(err)
        err = os.Mkdir(path + "Files/", os.ModePerm)
        Err_check(err)
    }
}

func Get_subject(parent string, board string) string {
    stmts := Checkout()
    defer Checkin(stmts)

    var subject string

    stmt := stmts["subject_look"]
    err := stmt.QueryRow(parent, board).Scan(&subject)
    Query_err_check(err)

    return subject
}

//for board pages
func get_threads(board string) []*Thread {
    stmts := Checkout()
    defer Checkin(stmts)

    stmt0 := stmts["parent_coll"]
    stmt := stmts["thread_head"]
    stmt2 := stmts["thread_body"]
    stmt3 := stmts["update_rep"]

    var board_body []*Thread

    //tables will be called a board 
    parent_rows, err := stmt0.Query(board)
    Err_check(err)
    defer parent_rows.Close()

    for parent_rows.Next() {
        var fstpst Post
        var filler int
        var pst_coll []*Post

        err = parent_rows.Scan(&fstpst.Id, &filler)
        Err_check(err)
        err = stmt.QueryRow(fstpst.Id, board).Scan(&fstpst.Content, &fstpst.Time, &fstpst.File,
            &fstpst.Filename, &fstpst.Fileinfo, &fstpst.Imgprev)
        Query_err_check(err)

        pst_coll = append(pst_coll, &fstpst)

        thread_rows, err := stmt2.Query(fstpst.Id, board)
        Err_check(err)
        defer thread_rows.Close()

        for thread_rows.Next() {
            var cpst Post

            err = thread_rows.Scan(&cpst.Id, &cpst.Content, &cpst.Time, &cpst.File,
            &cpst.Filename, &cpst.Fileinfo, &cpst.Imgprev, &cpst.Option)
            Err_check(err)

            pst_coll = append(pst_coll, &cpst)
        }

        for _, pst := range pst_coll {
            rep_rows, err := stmt3.Query(pst.Id, board)
            Err_check(err)

            for rep_rows.Next() {
        var replier int
        rep_rows.Scan(&replier)
        pst.Replies = append(pst.Replies, replier)
            }
            rep_rows.Close()
        }

        sub := Get_subject(strconv.Itoa(fstpst.Id), board)
        var thr Thread
        if sub != "" {
            thr = Thread{BoardN: board, Posts: pst_coll, Subject: sub, Parent: strconv.Itoa(fstpst.Id)}
        } else {
            thr = Thread{BoardN: board, Posts: pst_coll, Parent: strconv.Itoa(fstpst.Id)}
        }

        board_body = append(board_body, &thr)
    }

    return board_body
}

//for individual threads
func get_posts(parent string, board string) ([]*Post, error) {

    stmts := Checkout()
    defer Checkin(stmts)

    stmt := stmts["update"]
    stmt2 := stmts["update_rep"]

    rows, err := stmt.Query(parent, board)
    Err_check(err)
    defer rows.Close()

    var thread_body []*Post

    for rows.Next() {
        var pst Post
        err = rows.Scan(&pst.Id, &pst.Content, &pst.Time, &pst.File,
            &pst.Filename, &pst.Fileinfo, &pst.Imgprev, &pst.Option)
        Err_check(err)

        rep_rows, err := stmt2.Query(pst.Id, board)
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

func Build_board(board string) {
    boardtemp := template.New("board.html")
    boardtemp, err := boardtemp.ParseFiles(BP + "/templates/board.html")
    Err_check(err)

    path := BP + "head/" + board + "/"
    Dir_check(path)

    f, err := os.Create(path + "index.html")
    Err_check(err)
    defer f.Close()

    threads := get_threads(board)

    cboard := Board{Name: board, Threads: threads, Header: Boards}
    boardtemp.Execute(f, cboard)

}

func Build_thread(parent string, board string) { //will accept argument for board and thread number
    threadtemp := template.New("thread.html")
    threadtemp, err := threadtemp.ParseFiles(BP + "/templates/thread.html")
    Err_check(err)


    path := BP + "head/" + board + "/"
    Dir_check(path)

    f, err := os.Create(path + parent + ".html")
    Err_check(err)
    defer f.Close()


    posts, err := get_posts(parent, board)


    sub := Get_subject(parent, board)


    if err == nil {
        var thr Thread

        if sub != "" {
            thr = Thread{BoardN: board, Posts: posts, Subject: sub, Parent: parent, Header: Boards}
        } else {
            thr = Thread{BoardN: board, Posts: posts, Parent: parent, Header: Boards}
        }
        threadtemp.Execute(f, thr)
    }

}
