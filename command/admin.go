package main

import (
    "net/http"
    "os"
    "strings"
    "time"
    "strconv"
    //"fmt"
    "database/sql"
    "text/template"
)

const (
    base_query_string = `SELECT ROWID, Board, Id, Content, Time, Parent, Identifier, COALESCE(File, '') AS File, 
            COALESCE(Filename, '') AS Filename,
            COALESCE(Fileinfo, '') AS Fileinfo, COALESCE(Filemime, '') AS Filemime, COALESCE(Imgprev, '') AS Imgprev, Option FROM posts 
            WHERE Parent <> 0`
    query_cap = ` ORDER BY ROWID DESC`
)

type Query_results struct {
    Posts []*Post
    Auth Acc_type
}

func Admin_actions(w http.ResponseWriter, req *http.Request) {
    ctx := req.Context()

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
    } else {Account_refresh(w, sessionToken)}

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
            //fmt.Println("confirmation")

            hours := req.FormValue("hours")
            days := req.FormValue("days")
             
            duration := 0
            dint, err := strconv.Atoi(days)
            if err == nil {duration += (dint * 24)}
            hint, err := strconv.Atoi(hours)
            if err == nil {duration += hint}
 
            var ban_expiry time.Time
            if duration == 0 {
                ban_expiry = time.Now().In(Loc).Add(time.Hour * 96)
            } else {
                ban_expiry = time.Now().In(Loc).Add(time.Hour * time.Duration(duration))
            }

            ban_stmt := WriteStrings["ban"]
            if dint > 0 {
                _, err = new_tx.ExecContext(ctx, ban_stmt, ids, boards, ban_expiry.Format(time.UnixDate))
            } else { //permaban
                _, err = new_tx.ExecContext(ctx, ban_stmt, ids, boards, -1)
            }
            Err_check(err)

            ban_message := req.FormValue("banmessage")
            if ban_message != "" {
                ban_message_stmt := WriteStrings["ban_message"]
                _, err = new_tx.ExecContext(ctx, ban_message_stmt, ban_message, ids, boards)
                Err_check(err)
                update_posts = true
            }
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

//the console
func Load_console(w http.ResponseWriter, req *http.Request) {
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

    //put this in a function, with the query string being an input. Every query will return an array of posts
    conn, err := sql.Open("sqlite3", DB_uri)
    Err_check(err)
    defer conn.Close()

    query_string := base_query_string

    //time control
    sdate :=  strings.ReplaceAll(req.FormValue("sdate"), "-", "")
    if sdate != "" {
        _, err := strconv.Atoi(sdate)
        if err != nil {
            http.Error(w, "Invalid start date.", http.StatusBadRequest)
            return
        }

        query_string += " AND Calendar >= " + sdate
    }

    edate :=  strings.ReplaceAll(req.FormValue("edate"), "-", "")
    if edate != "" {
        _, err := strconv.Atoi(edate)
        if err != nil {
            http.Error(w, "Invalid end date.", http.StatusBadRequest)
            return
        }

        query_string += " AND Calendar <= " + edate
    }

    stime :=  strings.ReplaceAll(req.FormValue("stime"), ":", "")
    if stime != "" {
        _, err := strconv.Atoi(stime)
        if err != nil {
            http.Error(w, "Invalid start time.", http.StatusBadRequest)
            return
        }

        query_string += " AND Clock >= " + stime
    }

    etime :=  strings.ReplaceAll(req.FormValue("etime"), ":", "")
    if etime != "" {
        _, err := strconv.Atoi(etime)
        if err != nil {
            http.Error(w, "Invalid end time.", http.StatusBadRequest)
            return
        }

        query_string += " AND Clock <= " + etime
    }

    //location control
    board :=  req.FormValue("board")
    if board != "" {query_string += ` AND Board = "` + board + `"`}

    parent :=  req.FormValue("parent")
    if parent != "" {
        _, err := strconv.Atoi(parent)
        if err != nil {
            http.Error(w, "Invalid parent.", http.StatusBadRequest)
            return
        }

        query_string += " AND Parent = " + parent
    }

    //identifier
    identifier :=  req.FormValue("identifier")
    if identifier != "" {query_string += ` AND Identifier = "` + identifier + `"`}

    query_string += query_cap

    limit := req.FormValue("limit")
    if limit == "" {
        query_string += " LIMIT 10"
    } else {
        intval, err := strconv.Atoi(limit)
        if err != nil {
            http.Error(w, "Invalid limit.", http.StatusBadRequest)
            return
        }
        
        if intval > 0 {query_string += " LIMIT " + limit}
    }

    query_stmt, err := conn.Prepare(query_string)
    Err_check(err)


    rows, err := query_stmt.Query()
    Err_check(err)
    defer rows.Close()

    var most_recent []*Post
    var filler int

    for rows.Next() {
        var pst Post
        err = rows.Scan(&filler, &pst.BoardN, &pst.Id, &pst.Content, &pst.Time, &pst.Parent, &pst.Identifier, &pst.File,
                        &pst.Filename, &pst.Fileinfo, &pst.Filemime, &pst.Imgprev, &pst.Option)
        Err_check(err)
        most_recent = append(most_recent, &pst)
    }

    if err == nil {
        mostrecent_temp := template.New("console.html").Funcs(Filefuncmap)
        mostrecent_temp, err := mostrecent_temp.ParseFiles(BP + "/templates/console.html")
        Err_check(err)

        results := Query_results{Posts: most_recent, Auth: userSession.acc_type}
	err = mostrecent_temp.Execute(w, results)
	Err_check(err)
    }
}