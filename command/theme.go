package main

import (
	"time"
	"net/http"
)

func Switch_theme(w http.ResponseWriter, req *http.Request) {

	cookie := &http.Cookie{
        	Name:   "theme",
        	Value:  req.FormValue("theme"),
		Expires: time.Now().AddDate(10, 0, 0),
        	Path: "/",
    	}

	http.SetCookie(w, cookie)	

	http.Redirect(w, req, req.Header.Get("Referer"), 302)
}