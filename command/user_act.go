package main

import (
    "time"
    "strings"
    "strconv"
    "html"
    "net/http"
    "database/sql"
    //"fmt"
)

//for users to edit and delete their posts

func User_actions(w http.ResponseWriter, req *http.Request) {
    ctx := req.Context()
    var post_pass string
    
    c, err := req.Cookie("post_pass")

    if err != nil {
        if err != http.ErrNoCookie {
            Err_check(err)
        }
    } else {
        post_pass = c.Value
    }
    
    if pwd := req.FormValue("pwd"); pwd != "password" && pwd != "" {
        post_pass = pwd
    }
    
    option := req.FormValue("option")
    if Entry_check(w, req, "option", option) == 0 {return}
    board := req.FormValue("board")
    if Entry_check(w, req, "board", board) == 0 {return}
    
    if _, board_check := Board_map[board]; !board_check {
        http.Error(w, "Board is invalid.", http.StatusBadRequest)
        return
    }
    
    identity := req.Header.Get("X-Real-IP")
    if Entry_check(w, req, "IP", identity) == 0 {return}
    
    now := time.Now().In(Nip)
    then := now.Add(time.Duration(-30) * time.Hour)
    sdate := then.Format("20060102")
    
    //begin transaction
    new_conn := WriteConnCheckout()
    defer WriteConnCheckin(new_conn)
    new_tx, err := new_conn.Begin()
    Err_check(err)
    defer new_tx.Rollback()
    
    if Ban_check(w, req, new_tx, ctx, identity) {return}
    
    var parent string
    parent_row := new_tx.QueryRowContext(ctx, `SELECT Parent FROM posts WHERE Password = ? AND Board = ? LIMIT 1`, post_pass, board)
    err = parent_row.Scan(&parent)
    if err == sql.ErrNoRows {
        http.Error(w, "Post not found on this board.", http.StatusUnauthorized)
        return
    } else {
        Err_check(err)
    }
    
    var file_name string
    var imgprev string
            
    user_get_file_stmt := WriteStrings["user_get_file"]
    file_row := new_tx.QueryRowContext(ctx, user_get_file_stmt, post_pass, board)

    err = file_row.Scan(&file_name, &imgprev)
    Query_err_check(err)
    
    //setup done
    if option == "Delete" {
        file_deletion := func() {
            if file_name != "" {
                file_path := BP + "head/" + board + "/Files/"
                Delete_file(file_path, file_name, imgprev)
        }}
        
        if req.FormValue("onlyimgdel") == "on" {
            file_deletion()
            user_filedelete_stmt := WriteStrings["user_filedelete"]
            _, err := new_tx.ExecContext(ctx, user_filedelete_stmt, post_pass, board)
            Err_check(err)
        } else {
            user_delete_stmt := WriteStrings["user_delete"]
            isparent_stmt := WriteStrings["isparent2"]
            
            var pcheck bool
            var id string
            pcheck_row := new_tx.QueryRowContext(ctx, isparent_stmt, post_pass, board)
            pcheck_row.Scan(&pcheck, &id)
            
            if pcheck {
                file_path := BP + "head/" + board + "/"
                Delete_file(file_path, id + ".html", "")
            }
            
            res, err := new_tx.ExecContext(ctx, user_delete_stmt, sdate, post_pass, board)
            Err_check(err)
        
            rowsaffected, err := res.RowsAffected()
            Err_check(err)
        
            if rowsaffected == 0 {
                http.Error(w, "This post is too old, has replies, or doesn't exist.", http.StatusUnauthorized)
                return
            } else {
                file_deletion()
            }
    }}
    
    if option == "Edit" {
        if Request_filter(w, req, "POST", Max_upload_size) == 0 {return}
       
        no_text := (strings.TrimSpace(req.FormValue("newpost")) == "")
        if no_text {
            http.Error(w, "Empty post.", http.StatusBadRequest)
            return
        }
       
        post_length := len([]rune(req.FormValue("newpost")))
        if post_length > max_post_length {
            http.Error(w, "Post exceeds character limit(10000). Post length: " + strconv.Itoa(post_length), http.StatusBadRequest)
            return 
        }
        
        //cleaning
        input := html.EscapeString(req.FormValue("newpost"))
        //word filter
        for re, replacement := range Word_filter {
            input = re.ReplaceAllString(input, replacement)
        }
        
        home_content, home_truncontent := HProcess_post(input)
        input, repmatches := Format_post(input, board, parent)
        
        //updating
        user_edit_stmt := WriteStrings["user_edit"]
        edit_message := `Post edited on ` + now.Format("2 Jan 2006, 3:04pm") 
        
        res, err := new_tx.ExecContext(ctx, user_edit_stmt, input, edit_message, sdate, post_pass, board)
        Err_check(err)
        
        rowsaffected, err := res.RowsAffected()
        Err_check(err)
        
        if rowsaffected == 0 {
            http.Error(w, "This post is too old or doesn't exist.", http.StatusUnauthorized)
            return
        }
        
        if len(repmatches) > 0 {
            repupdate_stmt := WriteStrings["repupdate"]
            for _, match := range repmatches {
                match_id, err := strconv.ParseUint(match, 10, 64)
                Err_check(err)
                _, err = new_tx.ExecContext(ctx, repupdate_stmt, board, match_id, post_pass)
                Err_check(err)
            }    
        }
        
        hpupdate_stmt := WriteStrings["hpupdate"]
        _, err = new_tx.ExecContext(ctx, hpupdate_stmt, home_content, home_truncontent, post_pass, board)
        Err_check(err)
    }
    
    err = new_tx.Commit()
    Err_check(err)
    
    
    Build_thread(parent, board)
    Build_board(board)
    go Build_catalog(board)
    go Build_home()
    
    //error if no rows are affected: This post is too old, has replies, or doesn't exist. 
    
    http.Redirect(w, req, req.Header.Get("Referer"), 302)
}
