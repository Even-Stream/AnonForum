package main

import (
    "time"
	"os"
	"errors"
	"strings"
	"net/http"
	"database/sql"
	"io/fs"
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
	board := req.FormValue("board")
	
	now := time.Now().In(Nip)
	then := now.Add(time.Duration(-30) * time.Hour)
    sdate := then.Format("20060102")
	
	//begin transaction
    new_conn := WriteConnCheckout()
    defer WriteConnCheckin(new_conn)
    new_tx, err := new_conn.Begin()
    Err_check(err)
    defer new_tx.Rollback()
	
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
	    user_delete_stmt := WriteStrings["user_delete"]
		res, err := new_tx.ExecContext(ctx, user_delete_stmt, sdate, post_pass, board)
		Err_check(err)
		
		rowsaffected, err := res.RowsAffected()
		Err_check(err)
		
		if rowsaffected == 0 {
		    http.Error(w, "This post is too old, has replies, or doesn't exist.", http.StatusUnauthorized)
            return
		} else {
            file_path := BP + "head/" + board + "/Files/"
			if file_name != "" {
                err = os.Remove(file_path + file_name)
                if !errors.Is(err, fs.ErrNotExist) {Err_check(err)}
                
                if !strings.HasSuffix(imgprev, "image.webp") {
                    err = os.Remove(file_path + imgprev)
                    if !errors.Is(err, fs.ErrNotExist) {Err_check(err)}
                }
        }}
	} 
	
	if option == "Edit" {
	
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