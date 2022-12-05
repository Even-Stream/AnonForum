package main

/*
this program will receive a request for a post, 
retrieve that post from a database, 
and send that post back
*/

import (
    "bytes"
    "net/http"
    "text/template"
    "time"
    "context"

    _ "github.com/mattn/go-sqlite3"
)

type Prev struct {
    Id int
    Board string
    Content string
    Imgprev string 
}

//retrieves post request
func Get_prev(w http.ResponseWriter, req *http.Request) {
    //time out
    ctx, cancel := context.WithTimeout(req.Context(), 10 * time.Millisecond)
    defer cancel()

    id := req.FormValue("p")
    board := req.FormValue("board")

    if id == "" || board == "" {
        http.Error(w, "Invalid preview request.", http.StatusBadRequest)
        return
    }

    stmts := Checkout()
    defer Checkin(stmts)
    stmt := stmts["prev"]

    var data string
    var temp bytes.Buffer
    var prv Prev

    row := stmt.QueryRowContext(ctx, id, board)

    err := row.Scan(&prv.Content, &prv.Imgprev)
    Query_err_check(err)

    //Prev_body, err := template.New("todos").Parse("{{if .Imgprev}}<img src=\"/{{.Board}}/{{.Imgprev}}\">{{end}}{{.Content}}")
    Prev_body, err := template.New("todos").Parse("{{if .Imgprev}}<img class=\"imspec\" src=\"Files/{{.Imgprev}}\">{{end}}{{.Content}}")
    Err_check(err)
    Prev_body.Execute(&temp, prv)

    data = temp.String()    

    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)

    w.Write([]byte(data))
}
