package main

import (
	"database/sql"
	"net/http"
	"time"
	"html"
	"io"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	units "github.com/docker/go-units" 
)

var nip, _ = time.LoadLocation("Asia/Tokyo")

var ext = map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/jxl": ".jxl", "image/gif": ".gif"}

func gen_info(size int64) string {
	file_info := units.HumanSize(float64(size))  
	return file_info
}

func New_post(w http.ResponseWriter, req *http.Request) {

	req.ParseMultipartForm(10 << 20)

	input := html.EscapeString(req.FormValue("newpost"))
	parent := req.FormValue("parent")

	now := time.Now().In(nip)
	post_time := now.Format("1/2/06(Mon)15:04:05")

	conn, err := sql.Open("sqlite3", BP + "command/post-coll.db") 
	Err_check(err)
	defer conn.Close() 

	//file uploading
    	file, handler, file_err := req.FormFile("file")

	if file_err == nil {	
		defer file.Close()

		mime_type := handler.Header["Content-Type"][0]
		file_pre := strconv.FormatInt(time.Now().UnixNano(), 10)
		file_name := file_pre + ext[mime_type]
		file_path := BP + "/Files/" 

		f, err := os.OpenFile(file_path + file_name, os.O_WRONLY|os.O_CREATE, 0666)
		Err_check(err)
		defer f.Close()
	
		io.Copy(f, file)

		//think about if a pdf is given 
		Make_thumb(file_path, file_pre, file_name, mime_type)
		file_info := gen_info(handler.Size)

		stmt, err := conn.Prepare(`INSERT INTO posts(Content, Time, Parent, File, Filename, Fileinfo, Imgprev) VALUES (?, ?, ?, ?, ?, ?, ?)`)
		Err_check(err)
		_, err = stmt.Exec(input, post_time, parent, file_name, handler.Filename, file_info, file_pre + "s.webp")

	} else {
		stmt, err := conn.Prepare(`INSERT INTO posts(Content, Time, Parent) VALUES (?, ?, ?)`)
		Err_check(err)
		_, err = stmt.Exec(input, post_time, parent)
	}

	Err_check(err)
	Build_thread()

	http.Redirect(w, req, req.Header.Get("Referer"), 302)
}
