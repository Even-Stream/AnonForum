package main 

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

const Max_conns = 5
var readConns = make(chan map[string]*sql.Stmt, Max_conns)
var writeStrings = make(chan map[string]string, 1)
var writeConn = make(chan *sql.DB, 1) 

//statement strings
const (
    prev_string = `SELECT Content, Time, COALESCE(Filemime, '') Filemime,
            COALESCE(Imgprev, '') Imgprev, Option FROM posts WHERE Id = ? AND Board = ?`
    prev_parentstring = `SELECT Parent FROM posts WHERE Id = ? AND Board = ?`
    updatestring = `SELECT Id, Content, Time, COALESCE(File, '') AS File, COALESCE(Filename, '') AS Filename, 
                COALESCE(Fileinfo, '') AS Fileinfo, COALESCE(Filemime, '') AS Filemime, COALESCE(Imgprev, '') Imgprev, Option FROM posts
                WHERE Parent = ? AND Board = ?`
    update_repstring = `SELECT Replier FROM replies WHERE Source = ? AND Board = ?`
    parent_collstring = `SELECT Parent, MAX(Id) FROM posts WHERE (instr(Option, 'sage') = 0 OR Id = Parent) AND Board = ? 
        GROUP BY Parent ORDER BY MAX(Id) DESC LIMIT 15`
    thread_headstring = `SELECT Content, Time, Parent, COALESCE(File, '') AS File, COALESCE(Filename, '') AS Filename, 
                COALESCE(Fileinfo, '') AS Fileinfo, COALESCE(Filemime, '') AS Filemime, COALESCE(Imgprev, '') Imgprev, Option
                FROM posts
                WHERE Id = ? AND Board = ?`
    thread_bodystring = `SELECT * FROM (
                SELECT Id, Content, Time, Parent, COALESCE(File, '') AS File, COALESCE(Filename, '') AS Filename, 
                COALESCE(Fileinfo, '') AS Fileinfo, COALESCE(Filemime, '') AS Filemime, COALESCE(Imgprev, '') Imgprev, Option FROM posts 
                WHERE Parent = ? AND Board = ? AND Id != Parent ORDER BY Id DESC LIMIT 5)
                ORDER BY Id ASC`
    thread_collstring = `SELECT Parent, MAX(Id) FROM posts WHERE (instr(Option, 'sage') = 0 OR Id = Parent) AND Board = ? 
        GROUP BY Parent ORDER BY MAX(Id) DESC`
    subject_lookstring = `SELECT Subject FROM subjects WHERE Parent = ? AND Board = ?`
    shown_countstring = `Select COUNT(*), COUNT(File) FROM 
      (SELECT *	FROM posts WHERE Board = ?1 AND Parent = ?2 AND Id <> ?2 ORDER BY Id DESC LIMIT 5)`
    total_countstring = `Select COUNT(*), COUNT(File) FROM posts WHERE Board = ?1 AND Parent = ?2 AND Id <> ?2`

    //all inserts(and necessary queries) are preformed in one transaction 
    newpost_wfstring = `INSERT INTO posts(Board, Id, Content, Time, Parent, Identifier, File, Filename, Fileinfo, Filemime, Imgprev, 
        Option, Calendar, Clock) 
        VALUES (?1, (SELECT Id FROM latest WHERE Board = ?1), ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, ?11, ?12, ?13)`
    newpost_nfstring = `INSERT INTO posts(Board, Id, Content, Time, Parent, Identifier, Option, Calendar, Clock) 
        VALUES (?1, (SELECT Id FROM latest WHERE Board = ?1), ?2, ?3, ?4, ?5, ?6, ?7, ?8)`
    repadd_string = `INSERT INTO replies(Board, Source, Replier) VALUES (?1, ?2, (SELECT Id FROM latest WHERE Board = ?1) - 1)`
    subadd_string = `INSERT INTO subjects(Board, Parent, Subject) VALUES (?, ?, ?)`
    hpadd_string = `INSERT INTO homepost(Board, Id, Content, TrunContent, Parent)
        VALUES (?1, (SELECT Id FROM latest WHERE Board = ?1) - 1, ?2, ?3, ?4)`
    htadd_string = `INSERT into homethumb(Board, Id, Parent, Imgprev)
        VALUES (?1, (SELECT Id FROM latest WHERE Board = ?1), ?2, ?3)`
    parent_checkstring = `SELECT COUNT(*)
                FROM posts
                WHERE Parent = ? AND Board = ?`
    threadid_string = `SELECT Id FROM latest WHERE Board = ?`

    add_token_string = `INSERT INTO tokens(Token, Type) VALUES (?, ?)`
    search_token_string = `SELECT Type FROM tokens WHERE Token = ?`
    delete_token_string = `DELETE FROM tokens where Token = ?`
    new_user_string = `INSERT INTO credentials(Username, Hash, Type) VALUES (?, ?, ?)`
    search_user_string = `SELECT Hash, Type FROM credentials WHERE Username = ?`

    ban_search_string = `SELECT Expiry, Reason FROM banned WHERE Identifier = ? ORDER BY ROWID ASC`
    ban_remove_string = `DELETE FROM banned WHERE Identifier = ? AND Expiry = ?`

    get_files_string = `SELECT COALESCE(File, '') AS File, COALESCE(Imgprev, '') AS Imgprev FROM posts WHERE (Id = ?1 OR Parent = ?1) AND Board = ?2`
    delete_post_string = `DELETE FROM posts WHERE (Id = ?1 OR Parent = ?1) AND Board = ?2`
    ban_string = `INSERT INTO banned(Identifier, Expiry, Mod, Content, Reason) VALUES ((SELECT Identifier FROM posts WHERE Id = ?1 AND Board = ?2), 
        ?3, ?4, (SELECT Content FROM posts WHERE Id = ?1 AND Board = ?2), ?5)`
    delete_log_string = `INSERT INTO deleted(Identifier, Time, Mod, Content, Reason) VALUES ((SELECT Identifier FROM posts WHERE Id = ?1 AND Board = ?2),
        ?3, ?4, replace(replace((SELECT Content FROM posts WHERE Id = ?1 AND Board = ?2), '<', '&lt;'), '>', '&gt;'), ?5)`
    ban_message_string = `UPDATE posts SET Content = Content || '<br><br><div class="banmessage">(' || ? || ')</div>' WHERE Id = ? AND Board = ?`
)

var  WriteStrings = map[string]string{"newpost_wf": newpost_wfstring, "newpost_nf": newpost_nfstring,
        "repadd": repadd_string, "subadd": subadd_string, "hpadd": hpadd_string,
        "htadd": htadd_string, 
        "parent_check": parent_checkstring, "threadid" : threadid_string,
        "add_token":  add_token_string, "search_token": search_token_string, 
        "ban_search": ban_search_string, "ban_remove": ban_remove_string, "delete_token": delete_token_string,
        "new_user": new_user_string, "search_user": search_user_string,
        "get_files": get_files_string, "delete_post": delete_post_string, "ban": ban_string, "delete_log": delete_log_string, 
        "ban_message": ban_message_string}

func Checkout() map[string]*sql.Stmt {
        return <-readConns
}
func Checkin(c map[string]*sql.Stmt) {
        readConns <- c
}

func WriteConnCheckout() *sql.DB {
    return <- writeConn
}

func WriteConnCheckin(c *sql.DB) {
    writeConn <- c
}


func Make_Conns() {
    for i := 0; i < Max_conns; i++ {

        //preview statements
        conn1, err := sql.Open("sqlite3", DB_uri)
        Err_check(err)

        prev_stmt, err := conn1.Prepare(prev_string)
        Err_check(err)

        conn2, err := sql.Open("sqlite3", DB_uri)
        Err_check(err)

        prev_parentstmt, err := conn2.Prepare(prev_parentstring)
        Err_check(err)


        //thread update statements
        conn3, err := sql.Open("sqlite3", DB_uri)
        Err_check(err)

        updatestmt, err := conn3.Prepare(updatestring)
        Err_check(err)

        conn4, err := sql.Open("sqlite3", DB_uri)
        Err_check(err)

        update_repstmt, err := conn4.Prepare(update_repstring)
        Err_check(err)


        //board upate statements
        conn5, err := sql.Open("sqlite3", DB_uri)
        Err_check(err)

        parent_collstmt, err := conn5.Prepare(parent_collstring)
        Err_check(err)

        conn6, err := sql.Open("sqlite3", DB_uri)
        Err_check(err)

        thread_headstmt, err := conn6.Prepare(thread_headstring)
        Err_check(err)

        conn7, err := sql.Open("sqlite3", DB_uri)
        Err_check(err)

        thread_bodystmt, err := conn7.Prepare(thread_bodystring)
        Err_check(err)

        //catalog update statement
        conn10, err := sql.Open("sqlite3", DB_uri)
        Err_check(err)    

        thread_collstmt, err := conn10.Prepare(thread_collstring)
        Err_check(err)
       
        //subject lookup
        conn11, err := sql.Open("sqlite3", DB_uri)
        Err_check(err)    

        subject_lookstmt, err := conn11.Prepare(subject_lookstring)
        Err_check(err)

        conn10a, err := sql.Open("sqlite3", DB_uri)
        Err_check(err)

        hp_collstmt, err := conn10a.Prepare("SELECT * FROM homepost ORDER BY ROWID DESC")
        Err_check(err)

        conn10b, err := sql.Open("sqlite3", DB_uri)
        Err_check(err)

        ht_collstmt, err := conn10b.Prepare("SELECT * FROM homethumb ORDER BY ROWID DESC")
        Err_check(err)

        conn12, err := sql.Open("sqlite3", DB_uri)
        Err_check(err)

        shown_countstmt, err := conn12.Prepare(shown_countstring)
        Err_check(err)

        conn13, err := sql.Open("sqlite3", DB_uri)
        Err_check(err)

        total_countstmt, err := conn13.Prepare(total_countstring)
        Err_check(err)

        
        read_stmts := map[string]*sql.Stmt{"prev": prev_stmt, "prev_parent": prev_parentstmt,
            "update": updatestmt, "update_rep": update_repstmt, "parent_coll": parent_collstmt,
            "thread_head": thread_headstmt, "thread_body": thread_bodystmt,
            "thread_coll": thread_collstmt,"subject_look": subject_lookstmt,
            "hp_coll": hp_collstmt, "ht_coll": ht_collstmt,
            "shown_count": shown_countstmt, "total_count": total_countstmt}

        readConns <- read_stmts
    }
  
    new_conn, err := sql.Open("sqlite3", DB_uri)
    Err_check(err)
    writeConn <- new_conn
}
