package main

import (
    "net/http"
    "time"
    "html"
    "io"
    "os"
    "strings"
    "strconv"
    "bytes"
    "context"
    "image/png"
    "database/sql"

    "github.com/zergon321/reisen"
    _ "github.com/mattn/go-sqlite3"
    "github.com/dustin/go-humanize"
    "github.com/gabriel-vasile/mimetype" 
)

var Nip, _ = time.LoadLocation("Asia/Tokyo")
var Max_upload_size int64

var mime_ext = map[string]string{"image/png": ".png", "image/jpeg": ".jpg", 
    "image/gif": ".gif", "image/webp": ".webp", "image/avif": ".avif", "image/vnd.mozilla.apng": ".apng",
    "image/svg+xml": ".svg",
    "audio/mpeg": ".mp3", "audio/ogg": ".ogg", "audio/flac": ".flac", "audio/opus": ".opus", "audio/x-m4a": ".m4a",
    "video/webm": ".webm", "video/mp4": ".mp4"}

const (
    max_post_length = 10000
)

func image_gen_info(size int64, width int, height int) string {
    file_info := humanize.Bytes(uint64(size))
    file_info = file_info + ", " + strconv.Itoa(width) + "x" + strconv.Itoa(height)
    return file_info
}

func generic_gen_info(size int64) string {
    file_info := humanize.Bytes(uint64(size))
    return file_info
}

func Ban_check(w http.ResponseWriter, req *http.Request, new_tx *sql.Tx, ctx context.Context, identity string) bool {
    ban_searchstmt := WriteStrings["ban_search"]
    ban_rows, err := new_tx.QueryContext(ctx, ban_searchstmt, identity)
    Err_check(err)
    defer ban_rows.Close()
	
    for ban_rows.Next() {
    //user was banned
        var ban_result string
        var ban_reason string
        err = ban_rows.Scan(&ban_result, &ban_reason) 

        if ban_result == "-1" {
            http.Error(w, "You are permanently banned. Reason: " + ban_reason, http.StatusBadRequest)
            return true
        }
        ban_expiry, err := time.Parse(time.RFC1123, ban_result)
        Err_check(err)

        if time.Now().In(Nip).Before(ban_expiry) {
        //user is still banned
            http.Error(w, "You are banned until: " + ban_result + " Reason: " + ban_reason, http.StatusBadRequest)
            return true
        } else {
            ban_removestmt := WriteStrings["ban_remove"]
            _, err = new_tx.ExecContext(ctx, ban_removestmt, identity, ban_result)
            Err_check(err)
        }

        err = new_tx.QueryRowContext(ctx, ban_searchstmt, identity).Scan(&ban_result)
        Query_err_check(err)
    }
    return false
}

func New_post(w http.ResponseWriter, req *http.Request) {
    ctx := req.Context()

    if Request_filter(w, req, "POST", Max_upload_size) == 0 {return}
    if err := req.ParseMultipartForm(Max_upload_size); err != nil {
        http.Error(w, "Request size exceeds limit: " + strconv.FormatInt(Max_request_size - 1, 10) + " MB", http.StatusBadRequest)
        return
    }
    defer req.MultipartForm.RemoveAll()

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
    identity := req.Header.Get("X-Real-IP")

    if identity == "" {
        http.Error(w, "No IP?", http.StatusBadRequest)
        return
    }

    if board == "" {
        http.Error(w, "Board not specified.", http.StatusBadRequest)
        return
    }

    if _, board_check := Board_map[board]; !board_check {
        http.Error(w, "Board is invalid.", http.StatusBadRequest)
        return
    }

    c, err := req.Cookie("session_token")

    if err == nil {
        sessionToken := c.Value
        userSession, exists := Sessions[sessionToken]
        if exists {
            if userSession.IsExpired() {
                delete(Sessions, sessionToken)
            } else {
                Account_refresh(w, sessionToken)
                switch {
                    case userSession.acc_type == Admin:
                        option += " admin"
                    case userSession.acc_type == Mod:
                        option += " moderator"
                    case userSession.acc_type == Maid:
                        option += " maid"
                }   
    }}}

    //begin transaction
    new_conn := WriteConnCheckout()
    defer WriteConnCheckin(new_conn)
    new_tx, err := new_conn.Begin()
    Err_check(err)
    defer new_tx.Rollback()
    
    if Ban_check(w, req, new_tx, ctx, identity) {return}

    //new thread if no parent is specified
    if parent != "" {
        parent_checkstmt := WriteStrings["parent_check"]
        var parent_result int

        err = new_tx.QueryRowContext(ctx, parent_checkstmt, parent, board).Scan(&parent_result)
        Query_err_check(err)

        if parent_result == 0 {
            http.Error(w, "Parent thread is invalid.", http.StatusBadRequest)
            return
        }

        var lock_result int
        lock_checkstmt := WriteStrings["lock_check"]
        err = new_tx.QueryRowContext(ctx, lock_checkstmt, parent, board).Scan(&lock_result)
        Query_err_check(err)

        if lock_result == 1 {
            http.Error(w, "This thread is locked.", http.StatusBadRequest)
            return
        }
    } else {
    //new thread logic
        if file_err != nil {
            http.Error(w, "Please upload a file.", http.StatusBadRequest)
            return
        }
    
        threadid_stmt := WriteStrings["threadid"]

        err = new_tx.QueryRowContext(ctx, threadid_stmt, board).Scan(&parent)
        Query_err_check(err)
        
        //subject insert
        if trimmed_subject := strings.TrimSpace(subject); trimmed_subject != "" {
            subadd_stmt := WriteStrings["subadd"]
            _, err = new_tx.ExecContext(ctx, subadd_stmt, board, parent, trimmed_subject)
            Err_check(err)
        }
    }

    input := html.EscapeString(req.FormValue("newpost"))
	//word filter
    for re, replacement := range Word_filter {
        input = re.ReplaceAllString(input, replacement)
	}
    home_content, home_truncontent := HProcess_post(input)
    input, repmatches := Format_post(input, board, parent)

    now := time.Now().In(Nip)
    post_time := now.Format("1/2/06(Mon)15:04:05")
    calendar := now.Format("20060102")
    clock := now.Format("1504")

    hpadd_stmt := WriteStrings["hpadd"]
	
	post_pass := Rand_gen()

    //file present
    if file_err == nil {
        defer file.Close()

        buffer := make([]byte, 512)
        _, err = file.Read(buffer)

        if err == io.EOF {
            buffer = []byte("ts")
            err = nil
        }
        Err_check(err)

        mime_type := mimetype.Detect(buffer).String()
        ext, supported_file := mime_ext[mime_type]

        _, err = file.Seek(0, io.SeekStart)
        Err_check(err)
		
        if supported_file {
            htadd_cond := false

            file_pre := strconv.FormatInt(time.Now().UnixNano(), 10)
            file_name := file_pre + ext
            file_path := BP + "head/" + board + "/Files/"
            hash := ""

            f, err := os.OpenFile(file_path + file_name, os.O_WRONLY|os.O_CREATE, 0666)
            Err_check(err)
            defer f.Close()

            //test type
            var file_info string

            if strings.HasPrefix(mime_type, "image") {
                htadd_cond = true
                file_buffer := bytes.NewBuffer(nil)
                io.Copy(file_buffer, file)
 
                width, height, ihash, cerr := Make_thumb(file_path, file_pre, file_buffer.Bytes(), 200)
                hash = ihash
                if cerr != nil {
                    //delete empty file
                    Delete_file(file_path, file_name, "")
                    http.Error(w, "Corrupted image.", http.StatusBadRequest)
                    return
                }
                file_pre += "s.webp"
                
                var dupt, dupid string
                dupcheck_stmt := WriteStrings["dupcheck"]
                err = new_tx.QueryRowContext(ctx, dupcheck_stmt, hash, board).Scan(&dupt, &dupid)
                if err != sql.ErrNoRows {
                    Delete_file(file_path, file_name, file_pre)
                    http.Error(w, `Duplicate image. Original: /` + board + `/` + dupt + `.html#no` + dupid, http.StatusUnauthorized)
                    return
                }
                if Forbidden[hash] {
                    Delete_file(file_path, file_name, file_pre)
                    http.Redirect(w, req, req.Header.Get("Referer"), 302)
                    return
                }

                file_info = image_gen_info(handler.Size, width, height)
                io.Copy(f, file_buffer)
            } else { 
                io.Copy(f, file)
                file_info = generic_gen_info(handler.Size)

                media, err := reisen.NewMedia(file_path + file_name)
	            Err_check(err)
	            defer media.Close()
                err = media.OpenDecode()
                Err_check(err)
				
                vss := media.VideoStreams()
                if len(vss) == 1 {
	                err = vss[0].Open()
                    Err_check(err)

                    for {
                        packet, gotPacket, err := media.ReadPacket()
                        Err_check(err)
                        if !gotPacket {break}
                            
                        if packet.Type() == reisen.StreamVideo {
						    cstream := vss[packet.StreamIndex()]
                            videoFrame, gotFrame, err := cstream.ReadVideoFrame()
                            Err_check(err)
                            if !gotFrame {break}
                            if videoFrame == nil{continue}
							
                            frimg := videoFrame.Image()
                              
                            cover_buffer := new(bytes.Buffer)
                            err = png.Encode(cover_buffer, frimg.SubImage(frimg.Rect))
                            Err_check(err)

                            _, _, _, cerr := Make_thumb(file_path, file_pre, cover_buffer.Bytes(), 300)
                            if cerr != nil {
                                file_pre = "audio_image.webp"
                            } else {
                                file_pre += "s.webp"
                            }
                            htadd_cond = true
                            break
                        }
                    }
                } else if len(vss) > 1 {
				    http.Error(w, "Multiple video streams detected. File could not be processed.", http.StatusBadRequest)
                    return
				} else {file_pre = "audio_image.webp"}
            }
            
            newpst_wfstmt := WriteStrings["newpost_wf"]

            ofname := []rune(handler.Filename)
            rem := len(ofname) - 50
            if rem < 0 {
                rem = 0
            }
            ffname := string(ofname[rem:])

            _, err = new_tx.ExecContext(ctx, newpst_wfstmt, board, input, post_time, parent, identity, 
                file_name, ffname, file_info, mime_type, file_pre, hash,
                option, calendar, clock, post_pass)
            Err_check(err)
			
            if !HBoard_map[board] && htadd_cond {
                htadd_stmt := WriteStrings["htadd"]
                _, err = new_tx.ExecContext(ctx, htadd_stmt, board, parent, file_pre, post_pass)
                Err_check(err)
			}
        } else {
              http.Error(w, "Unsupported file type.", http.StatusBadRequest)
              return
        }
    //file not present 
    } else {
        newpost_nfstmt := WriteStrings["newpost_nf"]
        _, err := new_tx.ExecContext(ctx, newpost_nfstmt, board, input, post_time, parent, identity, option, calendar, clock, post_pass)
        Err_check(err)
    }

    if !no_text { 
        if !HBoard_map[board] {
            _, err = new_tx.ExecContext(ctx, hpadd_stmt, board, home_content, home_truncontent, parent, post_pass)
            Err_check(err)
    }}

    //reply insert
    if len(repmatches) > 0 {
        repadd_stmt := WriteStrings["repadd"]
        for _, match := range repmatches {
            match_id, err := strconv.ParseUint(match, 10, 64)
            Err_check(err)
            _, err = new_tx.ExecContext(ctx, repadd_stmt, board, match_id, post_pass)
            Err_check(err)
        }    
    }

    err = new_tx.Commit()
    Err_check(err)
	
	http.SetCookie(w, &http.Cookie{
        Name:    "post_pass",
        Value:   post_pass,
        Path: "/",
    })

    Build_thread(parent, board)
    http.Redirect(w, req, req.Header.Get("Referer"), 302)

    go Build_board(board)
    go Build_catalog(board)
    go Build_home()
}
