package main

import (
    "net/http"
    "database/sql"
    "os"
    "strings"
    
    _ "github.com/mattn/go-sqlite3"
)

const (
    get_files_string = `SELECT COALESCE(File, '') AS File, COALESCE(Imgprev, '') AS Imgprev FROM posts WHERE (Id = ?1 OR Parent = ?1) AND Board = ?2`
    delete_post_string = `DELETE FROM posts WHERE (Id = ?1 OR Parent = ?1) AND Board = ?2`
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
    parents := req.FormValue("parents")
    if Entry_check(w, req, "parents", parents) == 0 {return}

    switch {
        case actions == "delete":

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
            http.Redirect(w, req, req.Header.Get("Referer"), 302)

            Build_thread(parents, boards)
            Build_board(boards)
            Build_catalog(boards)
            Build_home()
        default:
            http.Error(w, "Invalid action.", http.StatusUnauthorized)
            return
    }
}