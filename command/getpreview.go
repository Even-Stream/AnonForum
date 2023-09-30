package main

import (
    "bytes"
    "net/http"
    "text/template"

    _ "github.com/mattn/go-sqlite3"
)

type Prev struct {
    Id string
    Board string
    Content string
    Time string
    Filename string
    Fileinfo string
    Filemime string
    Imgprev string 
    Option string
}

const (
    prev_body = `<label class="{{.Option}}"><time>{{.Time}}</time> <selfref>No. {{.Id}}</selfref></label>
        <br>
        {{if .Filemime}}<label>File <a href="/">{{.Filename}}</a> - ({{.Fileinfo}})</label><br>{{end}}
        <div class="prevcontent">
        {{if .Imgprev}}{{if audiocheck .Filemime}}
        <img class="imspec" src="/resources/audio_image.webp">
        {{else}}<img class="imspec" src="/{{.Board}}/Files/{{.Imgprev}}">{{end}}{{end}}
        {{.Content}}</div>`
)

//retrieves post request
func Get_prev(w http.ResponseWriter, req *http.Request) {
    ctx := req.Context()

    if Request_filter(w, req, "GET", 1 << 13) == 0 {return}

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
    prv.Board = board
    prv.Id = id

    row := stmt.QueryRowContext(ctx, id, board)

    err := row.Scan(&prv.Content, &prv.Time, &prv.Filename, &prv.Fileinfo, &prv.Filemime, &prv.Imgprev, &prv.Option)
    Query_err_check(err)

    Prev_body, err := template.New("todos").Funcs(Filefuncmap).Parse(prev_body)
    Err_check(err)
    Prev_body.Execute(&temp, prv)

    data = temp.String()    

    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)

    w.Write([]byte(data))

}