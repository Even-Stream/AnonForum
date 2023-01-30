package main

import (
    "net/http"
    "os"
    "strings"
    "time"
    "fmt"
    "context"
)

func Admin_actions(w http.ResponseWriter, req *http.Request) {
    ctx, cancel := context.WithTimeout(req.Context(), 10 * time.Second)
    defer cancel()

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

    update_posts := false

    //begin transaction
    new_conn := WriteConnCheckout()
    defer WriteConnCheckin(new_conn)
    new_tx, err := new_conn.Begin()
    Err_check(err)
    defer new_tx.Rollback()

        if strings.HasPrefix(actions, "Ban") {
            fmt.Println("confirmation")

            duration := req.FormValue("duration")
            var ban_expiry time.Time
            if duration == "" {
                ban_expiry = time.Now().In(Loc).Add(time.Minute * 3) //.Hour * 24 * 5)
            }

            fmt.Println(ban_expiry)
            ban_stmt := WriteStrings["ban"]
            _, err = new_tx.ExecContext(ctx, ban_stmt, ids, boards, ban_expiry.Format(time.UnixDate))
            Err_check(err)
        }

        if strings.HasSuffix(actions, "Delete") {
            if Entry_check(w, req, "parents", parents) == 0 {return}

            get_files_stmt := WriteStrings["get_files"]

            //DO FOR ALL FILES
            file_rows, err := new_tx.QueryContext(ctx, get_files_stmt, ids, boards)
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

            delete_post_stmt := WriteStrings["delete_post"]
            _, err = new_tx.ExecContext(ctx, delete_post_stmt, ids, boards)
            Err_check(err)

            update_posts = true
        }

        err = new_tx.Commit()
        Err_check(err)

        if update_posts {
            Build_thread(parents, boards)
            Build_board(boards)
            Build_catalog(boards)
            Build_home()
        }

        http.Redirect(w, req, req.Header.Get("Referer"), 302)
}