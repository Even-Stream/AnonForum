package main

import (
    "net/http"
    "time"
    "html"
    "io"
    "os"
    "strings"
    "strconv"

    _ "github.com/mattn/go-sqlite3"
    units "github.com/docker/go-units"
)

var nip, _ = time.LoadLocation("Asia/Tokyo")

var mime_ext = map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/gif": ".gif", "image/webp": ".webp"}

const (
    max_upload_size = 20 << 20   //20MB
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

    if board == "" {
        http.Error(w, "Board not specified.", http.StatusBadRequest)
        return
    }

    if _, board_check := Boards[board]; !board_check {
        http.Error(w, "Board is invalid.", http.StatusBadRequest)
        return
    }

    input := html.EscapeString(req.FormValue("newpost"))
    input, repmatches := Format_post(input, board)

    now := time.Now().In(nip)
    post_time := now.Format("1/2/06(Mon)15:04:05")

    //begin transaction
    new_conn := WriteConnCheckout()
    defer WriteConnCheckin(new_conn)
    new_tx, err := new_conn.Begin()
    Err_check(err)
    defer new_tx.Rollback()
    

    //new thread if no parent is specified
    if parent != "" {
        parent_checkstmt := WriteStrings["parent_check"]
        var parent_result int

        err = new_tx.QueryRow(parent_checkstmt, parent, board).Scan(&parent_result)
        Query_err_check(err)

        if parent_result == 0 {
            http.Error(w, "Parent thread is invalid.", http.StatusBadRequest)
            return
        }
    } else {
    //new thread logic
        threadid_stmt := WriteStrings["threadid"]

        err = new_tx.QueryRow(threadid_stmt, board).Scan(&parent)
        Query_err_check(err)
        
        //subject insert
        if subject != "" {
            subadd_stmt := WriteStrings["subadd"]
            _, err = new_tx.Exec(subadd_stmt, board, parent, subject)
            Err_check(err)
        }
    }
    

    hpadd_stmt := WriteStrings["hpadd"]

    //file present
    if file_err == nil {
        defer file.Close()

        mime_type := handler.Header["Content-Type"][0]
        ext, supp := mime_ext[mime_type]

        buffer := make([]byte, 512)
        _, err = file.Read(buffer)

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

            newpst_wfstmt := WriteStrings["newpost_wf"]
            htadd_stmt := WriteStrings["htadd"]

            ofname := []rune(handler.Filename)
            rem := len(ofname) - 20
            if rem < 0 {
                rem = 0
            }
            ffname := string(ofname[rem:])

            if !no_text { 
                _, err = new_tx.Exec(hpadd_stmt, board, input, parent)
                Err_check(err)
            }
            _, err = new_tx.Exec(htadd_stmt, board, parent, file_pre + "s.webp")
            Err_check(err)
            _, err = new_tx.Exec(newpst_wfstmt, board, input, post_time, parent, file_name, ffname, file_info, file_pre + "s.webp", option)
            Err_check(err)
        }
    //file not present 
    } else {
        _, err = new_tx.Exec(hpadd_stmt, board, input, parent)
        Err_check(err)
        newpost_nfstmt := WriteStrings["newpost_nf"]
        _, err := new_tx.Exec(newpost_nfstmt, board, input, post_time, parent, option)
        Err_check(err)
    }


    //reply insert
    if len(repmatches) > 0 {
        repadd_stmt := WriteStrings["repadd"]
        for _, match := range repmatches {
            match_id, err := strconv.ParseUint(match, 10, 64)
            Err_check(err)
            _, err = new_tx.Exec(repadd_stmt, board, match_id)
            Err_check(err)
        }    
    }

    err = new_tx.Commit()
    Err_check(err)

    Build_thread(parent, board)
    http.Redirect(w, req, req.Header.Get("Referer"), 302)

    Build_board(board)
    Build_catalog(board)
    Build_home()
}
