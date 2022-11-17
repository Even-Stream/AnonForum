package main 

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

var Max_conns = 5
var readConns = make(chan map[string]*sql.Stmt, Max_conns)
var writeConns = make(chan map[string]*sql.Stmt, 1)

//statement strings
const (
    prev_string = `SELECT Content, 
            COALESCE(Imgprev, '') Imgprev FROM posts WHERE Id = ? AND Board = ?`
    prev_parentstring = `SELECT Parent FROM posts WHERE Id = ? AND Board = ?`
    updatestring = `SELECT Id, Content, Time, COALESCE(File, '') AS File, COALESCE(Filename, '') AS Filename, 
                COALESCE(Fileinfo, '') AS Fileinfo, COALESCE(Imgprev, '') Imgprev, Option FROM posts
                WHERE Parent = ? AND Board = ?`
    update_repstring = `SELECT Replier FROM replies WHERE Source = ? AND Board = ?`
    parent_collstring = `SELECT Parent, MAX(Id) FROM posts WHERE (Option <> "sage" OR Id = Parent) AND Board = ? 
        GROUP BY Parent ORDER BY MAX(Id) DESC LIMIT 15`
    thread_headstring = `SELECT Content, Time, COALESCE(File, '') AS File, COALESCE(Filename, '') AS Filename, 
                COALESCE(Fileinfo, '') AS Fileinfo, COALESCE(Imgprev, '') Imgprev
                FROM posts
                WHERE Id = ? AND Board = ?`
    thread_bodystring = `SELECT * FROM (
                SELECT Id, Content, Time, COALESCE(File, '') AS File, COALESCE(Filename, '') AS Filename, 
                COALESCE(Fileinfo, '') AS Fileinfo, COALESCE(Imgprev, '') Imgprev, Option FROM posts 
                WHERE Parent = ? AND Board = ? AND Id != Parent ORDER BY Id DESC LIMIT 5)
                ORDER BY Id ASC`
    lastid_string = `SELECT IFNULL ((SELECT MAX(Id) FROM posts WHERE Board = ?), 0)`
    parent_checkstring = `SELECT COUNT(*)
                FROM posts
                WHERE Parent = ? AND Board = ?`
    thread_collstring = `SELECT Parent, MAX(Id) FROM posts WHERE (Option <> "sage" OR Id = Parent) AND Board = ? 
        GROUP BY Parent ORDER BY MAX(Id) DESC`
    subject_lookstring = `SELECT Subject FROM subjects WHERE Parent = ? AND Board = ?`

    newpost_wfstring = `INSERT INTO posts(Board, Id, Content, Time, Parent, File, Filename, Fileinfo, Imgprev, Option) 
        VALUES (?1, (SELECT Id FROM latest WHERE Board = ?1), ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9)`
    newpost_nfstring = `INSERT INTO posts(Board, Id, Content, Time, Parent, Option) 
        VALUES (?1, (SELECT Id FROM latest WHERE Board = ?1), ?2, ?3, ?4, ?5)`
    repadd_string = `INSERT INTO replies(Board, Source, Replier) VALUES (?1, ?2, (SELECT Id FROM latest WHERE Board = ?1) - 1)`
    subadd_string = `INSERT INTO subjects(Board, Parent, Subject) VALUES (?, ?, ?)`
    hpadd_string = `INSERT INTO homepost(Board, Id, Content, Parent)
        VALUES (?1, (SELECT Id FROM latest WHERE Board = ?1), ?2, ?3)`
    htadd_string = `INSERT into homethumb(Board, Id, Parent, Imgprev)
        VALUES (?1, (SELECT Id FROM latest WHERE Board = ?1), ?2, ?3)`
)

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

        //preview statements
        conn1, err := sql.Open("sqlite3", db_uri)
        Err_check(err)

        prev_stmt, err := conn1.Prepare(prev_string)
        Err_check(err)

        conn2, err := sql.Open("sqlite3", db_uri)
        Err_check(err)

        prev_parentstmt, err := conn2.Prepare(prev_parentstring)
        Err_check(err)


        //thread update statements
        conn3, err := sql.Open("sqlite3", db_uri)
        Err_check(err)

        updatestmt, err := conn3.Prepare(updatestring)
        Err_check(err)

        conn4, err := sql.Open("sqlite3", db_uri)
        Err_check(err)

        update_repstmt, err := conn4.Prepare(update_repstring)
        Err_check(err)


        //board upate statements
        conn5, err := sql.Open("sqlite3", db_uri)
        Err_check(err)

        parent_collstmt, err := conn5.Prepare(parent_collstring)
        Err_check(err)

        conn6, err := sql.Open("sqlite3", db_uri)
        Err_check(err)

        thread_headstmt, err := conn6.Prepare(thread_headstring)
        Err_check(err)

        conn7, err := sql.Open("sqlite3", db_uri)
        Err_check(err)

        thread_bodystmt, err := conn7.Prepare(thread_bodystring)
        Err_check(err)

        conn8, err := sql.Open("sqlite3", db_uri)
        Err_check(err)

        lastid_stmt, err := conn8.Prepare(lastid_string)
        Err_check(err)

        conn9, err := sql.Open("sqlite3", db_uri)
        Err_check(err)

        parent_checkstmt, err := conn9.Prepare(parent_checkstring)
        Err_check(err)

        //catalog update statement
        conn10, err := sql.Open("sqlite3", db_uri)
        Err_check(err)    

        thread_collstmt, err := conn10.Prepare(thread_collstring)
        Err_check(err)
       
        //subject lookup
        conn11, err := sql.Open("sqlite3", db_uri)
        Err_check(err)    

        subject_lookstmt, err := conn11.Prepare(subject_lookstring)
        Err_check(err)

        conn10a, err := sql.Open("sqlite3", db_uri)
        Err_check(err)

        hp_collstmt, err := conn10a.Prepare("SELECT * FROM homepost ORDER BY Id DESC")
        Err_check(err)

        conn10b, err := sql.Open("sqlite3", db_uri)
        Err_check(err)

        ht_collstmt, err := conn10b.Prepare("SELECT * FROM homethumb ORDER BY Id DESC")
        Err_check(err)
        
        read_stmts := map[string]*sql.Stmt{"prev": prev_stmt, "prev_parent": prev_parentstmt,
            "update": updatestmt, "update_rep": update_repstmt, "parent_coll": parent_collstmt,
            "thread_head": thread_headstmt, "thread_body": thread_bodystmt, "lastid": lastid_stmt,
            "parent_check": parent_checkstmt, "thread_coll": thread_collstmt,"subject_look": subject_lookstmt,
            "hp_coll": hp_collstmt, "ht_coll": ht_collstmt}

        readConns <- read_stmts
    }


    conn12, err := sql.Open("sqlite3", db_uri)
    Err_check(err)

    newpost_wfstmt, err := conn12.Prepare(newpost_wfstring)
    Err_check(err)


    conn13, err := sql.Open("sqlite3", db_uri)
    Err_check(err)

    newpost_nfstmt, err := conn13.Prepare(newpost_nfstring)
    Err_check(err)

    conn14, err := sql.Open("sqlite3", db_uri)
    Err_check(err)

    repadd_stmt, err := conn14.Prepare(repadd_string)
    Err_check(err)

    conn15, err := sql.Open("sqlite3", db_uri)
    Err_check(err)

    subadd_stmt, err := conn15.Prepare(subadd_string)
    Err_check(err)

    conn16, err := sql.Open("sqlite3", db_uri)
    Err_check(err)

    hpadd_stmt, err := conn16.Prepare(hpadd_string)
    Err_check(err)

    conn17, err := sql.Open("sqlite3", db_uri)
    Err_check(err)

    htadd_stmt, err := conn17.Prepare(htadd_string)
    Err_check(err)
    
    write_stmts := map[string]*sql.Stmt{"newpost_wf": newpost_wfstmt, "newpost_nf": newpost_nfstmt,
        "repadd": repadd_stmt, "subadd": subadd_stmt,
        "hpadd": hpadd_stmt, "htadd": htadd_stmt}

    writeConns <- write_stmts
}
