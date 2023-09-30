package main

import (
    "os"
    "text/template"
    "strconv"
    "errors"
    "strings"

    _ "github.com/mattn/go-sqlite3"
)

//structures used in templates
type Post struct {
    BoardN string
    Id int
    Content string
    Time string
    Parent int
    Identifier string
    File string
    Filename string
    Fileinfo string
    Filemime string
    Imgprev string
    Option string
    Pinned bool
    Locked bool
	Anchored bool
    Replies []int
}

type Thread struct {
    BoardN string
    TId string
    BoardDesc string
    Subject string
    Posts []*Post
    Header []string
    HeaderDescs []string
    OmittedPosts int
    OmittedFiles int
    SThemes []string
}

type Board struct {
    Name string
    Desc string
    Threads []*Thread
    Header []string
    HeaderDescs []string
    SThemes []string
}

//getting kind of file 
var Filefuncmap = template.FuncMap {
    "imagecheck": func(filemime string) bool {
        if strings.HasPrefix(filemime, "image") {return true}
        return false
    },
    "avcheck": func(filemime string) bool {
        if strings.HasPrefix(filemime, "audio") {return true}
        if strings.HasPrefix(filemime, "video") {return true}
        return false
    },
    "audiocheck": func(filemime string) bool {
        if strings.HasPrefix(filemime, "audio") {return true}
        return false
    },
    "videocheck": func(filemime string) bool {
        if strings.HasPrefix(filemime, "video") {return true}
        return false
    },
}

func Dir_check(path string) {

    if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
        err := os.Mkdir(path, os.ModePerm)
        Err_check(err)
        err = os.Mkdir(path + "Files/", os.ModePerm)
        Err_check(err)
    }
}

func Get_subject(parent, board string) string {
    stmts := Checkout()
    defer Checkin(stmts)

    var subject string

    subject_look_stmt := stmts["subject_look"]
    err := subject_look_stmt.QueryRow(parent, board).Scan(&subject)
    Query_err_check(err)

    return subject
}

func Get_omitted(parent, board string) (int, int) {
    stmts := Checkout()
    defer Checkin(stmts)

    var total_posts int
    var total_files int
    var shown_posts int
    var shown_files int

    total_countstmt := stmts["total_count"]
    err := total_countstmt.QueryRow(board, parent).Scan(&total_posts, &total_files)
    Query_err_check(err)

    shown_countstmt := stmts["shown_count"]
    err = shown_countstmt.QueryRow(board, parent).Scan(&shown_posts, &shown_files)
    Query_err_check(err)

    return (total_posts - shown_posts), (total_files - shown_files)
}

//for board pages
func get_threads(board string) []*Thread {
    stmts := Checkout()
    defer Checkin(stmts)

    parent_coll_stmt := stmts["parent_coll"]
    thread_head_stmt := stmts["thread_head"]
    thread_body_stmt := stmts["thread_body"]
    update_rep_stmt := stmts["update_rep"]

    var board_body []*Thread

    //tables will be called a board 
    parent_rows, err := parent_coll_stmt.Query(board)
    Err_check(err)
    defer parent_rows.Close()

    for parent_rows.Next() {
        var fstpst Post
        var filler int
        var pst_coll []*Post

        err = parent_rows.Scan(&fstpst.Id, &filler)
        Err_check(err)
        err = thread_head_stmt.QueryRow(fstpst.Id, board).Scan(&fstpst.Content, &fstpst.Time, &fstpst.Parent, &fstpst.File,
            &fstpst.Filename, &fstpst.Fileinfo, &fstpst.Filemime, &fstpst.Imgprev, &fstpst.Option, &fstpst.Pinned, &fstpst.Locked, &fstpst.Anchored)
        Query_err_check(err)

        pst_coll = append(pst_coll, &fstpst)

        thread_rows, err := thread_body_stmt.Query(fstpst.Id, board)
        Err_check(err)
        defer thread_rows.Close()

        for thread_rows.Next() {
            var cpst Post
			cpst.BoardN = board

            err = thread_rows.Scan(&cpst.Id, &cpst.Content, &cpst.Time, &cpst.Parent, &cpst.File,
                &cpst.Filename, &cpst.Fileinfo, &cpst.Filemime, &cpst.Imgprev, &cpst.Option)
            Err_check(err)

            pst_coll = append(pst_coll, &cpst)
        }

        for _, pst := range pst_coll[1:] {
            rep_rows, err := update_rep_stmt.Query(pst.Id, board)
            Err_check(err)

            for rep_rows.Next() {
        var replier int
        rep_rows.Scan(&replier)
        pst.Replies = append(pst.Replies, replier)
            }
            rep_rows.Close()
        }

        fstpstid := strconv.Itoa(fstpst.Id)

        sub := Get_subject(fstpstid, board)
        omitted_posts, omitted_files := Get_omitted(fstpstid, board)

        var thr Thread
        if sub != "" {
            thr = Thread{Posts: pst_coll, Subject: sub, OmittedPosts: omitted_posts, OmittedFiles: omitted_files}
        } else {
            thr = Thread{Posts: pst_coll, OmittedPosts: omitted_posts, OmittedFiles: omitted_files}
        }

        board_body = append(board_body, &thr)
    }

    return board_body
}

//for individual threads
func get_posts(parent string, board string) ([]*Post, error) {

    stmts := Checkout()
    defer Checkin(stmts)

    update_stmt := stmts["update"]
    update_rep_stmt := stmts["update_rep"]

    rows, err := update_stmt.Query(parent, board)
    Err_check(err)
    defer rows.Close()

    var thread_body []*Post

    for rows.Next() {
        var pst Post
		pst.BoardN = board
		
        err = rows.Scan(&pst.Id, &pst.Content, &pst.Time, &pst.File,
            &pst.Filename, &pst.Fileinfo, &pst.Filemime, &pst.Imgprev, &pst.Option, &pst.Pinned, &pst.Locked, &pst.Anchored)
        Err_check(err)

        rep_rows, err := update_rep_stmt.Query(pst.Id, board)
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
    boardtemp := template.New("board.html").Funcs(Filefuncmap)
    boardtemp, err := boardtemp.ParseFiles(BP + "/templates/board.html", BP + "/templates/snippet.html")
    Err_check(err)

    path := BP + "head/" + board + "/"
    Dir_check(path)

    f, err := os.Create(path + "index.html")
    Err_check(err)
    defer f.Close()

    threads := get_threads(board)

    cboard := Board{Name: board,  Desc: Board_map[board],Threads: threads,
        Header: Board_names, HeaderDescs: Board_descs, SThemes: Themes}
    boardtemp.Execute(f, cboard)

}

func Build_thread(parent string, board string) { //will accept argument for board and thread number
    posts, err := get_posts(parent, board)
    if len(posts) == 0 {return} 
    sub := Get_subject(parent, board)

    threadtemp := template.New("thread.html").Funcs(Filefuncmap)
    threadtemp, err = threadtemp.ParseFiles(BP + "/templates/thread.html", BP + "/templates/snippet.html")
    Err_check(err)

    path := BP + "head/" + board + "/"
    Dir_check(path)

    f, err := os.Create(path + parent + ".html")
    Err_check(err)
    defer f.Close()



    if err == nil {
        var thr Thread

        if sub != "" {
            thr = Thread{BoardN: board, TId: parent, BoardDesc: Board_map[board],
                Posts: posts, Subject: sub,
                Header: Board_names, HeaderDescs: Board_descs, SThemes: Themes}
        } else {
            thr = Thread{BoardN: board, TId: parent, BoardDesc: Board_map[board], Posts: posts, 
            Header: Board_names, HeaderDescs: Board_descs, SThemes: Themes}
        }
        threadtemp.Execute(f, thr)
    }

}
