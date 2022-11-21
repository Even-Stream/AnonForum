package main

import (
    "net/http"
    "time"
    "html"
    "io"
    "os"
    "strings"
    slices "golang.org/x/exp/slices"
    "strconv"

    _ "github.com/mattn/go-sqlite3"
    units "github.com/docker/go-units"
)

var nip, _ = time.LoadLocation("Asia/Tokyo")

var mime_ext = map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/gif": ".gif", "image/webp": ".webp"}

const (
    max_upload_size = 20 << 20   //12MB
    max_post_length = 10000
)

func gen_info(size int64, width int, height int) string {
    file_info := units.HumanSize(float64(size))
    file_info = file_info + ", " + strconv.Itoa(width) + "x" + strconv.Itoa(height)
    return file_info
}

func New_post(w http.ResponseWriter, req *http.Request) {
    //bad request filtering 

    if req.Method != "POST" {
        http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
        return
    }

    req.Body = http.MaxBytesReader(w, req.Body, max_upload_size)
    if err := req.ParseMultipartForm(10 << 20); err != nil {
        http.Error(w, "Request size exceeds limit(10MB).", http.StatusBadRequest)
        return
    }

    file, handler, file_err := req.FormFile("file")

    no_text := (strings.TrimSpace(req.FormValue("newpost")) == "")
    if file_err != nil && no_text {
        http.Error(w, "Empty post.", http.StatusBadRequest)
        return
    }

    post_length := len([]rune(req.FormValue("newpost")))
    if post_length > max_post_length {
        http.Error(w, "Post exceeds character limit(10000). Post length: " + strconv.Itoa(post_length), http.StatusBadRequest)
        return 
    }

    parent := req.FormValue("parent")
    board := req.FormValue("board")
    subject := req.FormValue("subject")
    option := req.FormValue("option")

    if parent == "" || board == "" {
        http.Error(w, "Board or parent thread not specified.", http.StatusBadRequest)
        return
    }

    if !(slices.Contains(Boards, board)) {
        http.Error(w, "Board is invalid.", http.StatusBadRequest)
        return
    }

    input := html.EscapeString(req.FormValue("newpost"))
    input, repmatches := Format_post(input, board)

    now := time.Now().In(nip)
    post_time := now.Format("1/2/06(Mon)15:04:05")

    wstmts := writeCheckout()
      defer writeCheckin(wstmts)

    rstmts := Checkout()
    defer Checkin(rstmts)

    //new thread testing 
    stmt0 := rstmts["parent_check"]
    var parent_result int

    err := stmt0.QueryRow(parent, board).Scan(&parent_result)
    Query_err_check(err)

    if parent_result == 0 {
        stmt01 := rstmts["lastid"]
        var latestid int

        err = stmt01.QueryRow(board).Scan(&latestid)
        Err_check(err)

        latestid++
        if parent != strconv.Itoa(latestid) {
            http.Error(w, "Parent thread is invalid.", http.StatusBadRequest)
            return
        } else {
            //subject insert
            if subject != "" {
                stmt0A := wstmts["subadd"]
                _, err = stmt0A.Exec(board, parent, subject)
                Err_check(err)
            }
        }
    }

    hpadd_stmt := wstmts["hpadd"]

    //file present
    if file_err == nil {
        defer file.Close()

        mime_type := handler.Header["Content-Type"][0]
        ext, supp := mime_ext[mime_type]

        buffer := make([]byte, 512)
        _, err := file.Read(buffer)

        if err == io.EOF {
            buffer = []byte("ts")
            err = nil
        }
        Err_check(err)

        _, supp2 := mime_ext[http.DetectContentType(buffer)]
        _, err = file.Seek(0, io.SeekStart)
        Err_check(err)

        if supp && supp2 {

            file_pre := strconv.FormatInt(time.Now().UnixNano(), 10)
            file_name := file_pre + ext
            file_path := BP + "head/" + board + "/Files/"

            f, err := os.OpenFile(file_path + file_name, os.O_WRONLY|os.O_CREATE, 0666)
            Err_check(err)
            defer f.Close()

            io.Copy(f, file)

            //think about if a pdf is given 
            width, height := Make_thumb(file_path, file_pre, file_name, mime_type)
            file_info := gen_info(handler.Size, width, height)

            stmt := wstmts["newpost_wf"]
            htadd_stmt := wstmts["htadd"]

            ofname := []rune(handler.Filename)
            rem := len(ofname) - 20
            if rem < 0 {
                rem = 0
            }
            ffname := string(ofname[rem:])

            if !no_text { 
                _, err = hpadd_stmt.Exec(board, input, parent)
                Err_check(err)
            }
            _, err = htadd_stmt.Exec(board, parent, file_pre + "s.webp")
            Err_check(err)
            _, err = stmt.Exec(board, input, post_time, parent, file_name, ffname, file_info, file_pre + "s.webp", option)
            Err_check(err)
        }
    //file not present 
    } else {
        _, err = hpadd_stmt.Exec(board, input, parent)
        Err_check(err)
        stmt := wstmts["newpost_nf"]
        _, err := stmt.Exec(board, input, post_time, parent, option)
        Err_check(err)
    }


    //reply insert
    if len(repmatches) > 0 {
        stmt := wstmts["repadd"]
        for _, match := range repmatches {
            match_id, err := strconv.ParseUint(match, 10, 64)
            Err_check(err)
            _, err = stmt.Exec(board, match_id)
            Err_check(err)
        }    
    }

    Build_thread(parent, board)
    http.Redirect(w, req, req.Header.Get("Referer"), 302)

    Build_board(board)
    Build_catalog(board)
    Build_home()
}
