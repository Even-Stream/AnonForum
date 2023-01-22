package main

import (
    "net/http"
    "database/sql"
    "os"
    "strings"
    "time"
    
    _ "github.com/mattn/go-sqlite3"
)

const (
    get_files_string = `SELECT COALESCE(File, '') AS File, COALESCE(Imgprev, '') AS Imgprev FROM posts WHERE (Id = ?1 OR Parent = ?1) AND Board = ?2`
    delete_post_string = `DELETE FROM posts WHERE (Id = ?1 OR Parent = ?1) AND Board = ?2`
    ban_string = `INSERT INTO banned(Identifier, Expiry) VALUES ((SELECT Identifier FROM posts WHERE Id = ?1 AND Board = ?2), ?3)`
)

func Admin_actions(w http.ResponseWriter, req *http.Request) {
    c, err := req.Cookie("session_token")

    if err != nil {
        if err == http.ErrNoCookie {
            http.Error(w, "Unauthorized.", http.StatusUnauthorized)
            return
        }
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    sessionToken := c.Value
    userSession, exists := Sessions[sessionToken]
    if !exists {
        http.Error(w, "Unauthorized.", http.StatusUnauthorized)
        return
    }

    if userSession.IsExpired() {
        delete(Sessions, sessionToken)
        http.Error(w, "Session expired.", http.StatusUnauthorized)
        return
    }

    //use maps for these(no duplicates)
    actions := req.FormValue("actions")
    if Entry_check(w, req, "actions", actions) == 0 {return}
    ids := req.FormValue("ids")
    if Entry_check(w, req, "ids", ids) == 0 {return}
    boards := req.FormValue("boards")	
    if Entry_check(w, req, "boards", boards) == 0 {return}

        if strings.HasSuffix(actions, "Delete") {

            parents := req.FormValue("parents")
            if Entry_check(w, req, "parents", parents) == 0 {return}

            conn, err := sql.Open("sqlite3", DB_uri)
            Err_check(err)
            defer conn.Close()
            get_files_stmt, err := conn.Prepare(get_files_string)
            Err_check(err)

            //DO FOR ALL FILES
            file_rows, err := get_files_stmt.Query(ids, boards)
            Err_check(err)
            defer file_rows.Close()

            for file_rows.Next() {
                var file_name string
                var imgprev string

                err = file_rows.Scan(&file_name, &imgprev)
                Err_check(err)

                file_path := BP + "head/" + boards + "/Files/"
                if file_name != "" {
                    err = os.Remove(file_path + file_name)
                    Err_check(err)
                
                    if !strings.HasSuffix(imgprev, "image.webp") {
                        err = os.Remove(file_path + imgprev)
                        Err_check(err)
                    }
            }}

            conn2, err := sql.Open("sqlite3", DB_uri)
            Err_check(err)
            defer conn2.Close()
            delete_post_stmt, err := conn2.Prepare(delete_post_string)
            Err_check(err)

            delete_post_stmt.Exec(ids, boards)


            Build_thread(parents, boards)
            Build_board(boards)
            Build_catalog(boards)
            Build_home()
        }
        
        if strings.HasPrefix(actions, "Ban") {
            conn, err := sql.Open("sqlite3", DB_uri)
            Err_check(err)
            defer conn.Close()
            ban_stmt, err := conn.Prepare(ban_string)
            Err_check(err)

            duration := req.FormValue("duration")
            var ban_expiry time.Time
            if duration == "" {
                ban_expiry = time.Now().In(Loc).Add(time.Minute * 3) //.Hour * 24 * 5)
            }

            ban_stmt.Exec(ids, boards, ban_expiry.Format(time.UnixDate))
        }

        http.Redirect(w, req, req.Header.Get("Referer"), 302)
    
}