package main

import (
    //"time"
	//"net/http"
)

//for users to edit and delete their posts

/*
func User_actions (w http.ResponseWriter, req *http.Request) {
    ctx := req.Context()
	
	c, err := req.Cookie("post_pass")

    if err != nil {
        if err == http.ErrNoCookie {
            http.Error(w, "Unauthorized.", http.StatusUnauthorized)
            return
        }
        w.WriteHeader(http.StatusBadRequest)
        return
    }
	
	post_pass := c.Value
	
	actiontype := req.FormValue("actiontype")
	board := req.FormValue("board")
	id := req.FormValue("id")
	
	now := time.Now().In(Nip)
	then := now.Add(time.Duration(-30) * time.Hour)
    sdate := then.Format("20060102")
	
	//begin transaction
    new_conn := WriteConnCheckout()
    defer WriteConnCheckin(new_conn)
    new_tx, err := new_conn.Begin()
    Err_check(err)
    defer new_tx.Rollback()
	
	err = new_tx.Commit()
    Err_check(err)
	
	
	Build_thread(parents, board)
    Build_board(board)
    go Build_catalog(board)
    go Build_home()
	
	//error if no rows are affected: This post is too old, has replies, or doesn't exist. 
	
	http.Redirect(w, req, req.Header.Get("Referer"), 302)
}
*/