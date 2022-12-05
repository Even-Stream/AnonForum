package main

import (
    "time"
    "net/http"
    "context"
)

func Switch_theme(w http.ResponseWriter, req *http.Request) {
    //time out
    _, cancel := context.WithTimeout(req.Context(), 10 * time.Millisecond)
    defer cancel()

    cookie := &http.Cookie{
            Name:   "theme",
            Value:  req.FormValue("theme"),
        Expires: time.Now().AddDate(10, 0, 0),
            Path: "/",
        }

    http.SetCookie(w, cookie)    

    http.Redirect(w, req, req.Header.Get("Referer"), 302)
}
