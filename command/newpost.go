package main

import (
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

var mime_ext = map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/gif": ".gif", "image/webp": ".webp"}

func gen_info(size int64) string {
	file_info := units.HumanSize(float64(size))
	return file_info
}

func New_post(w http.ResponseWriter, req *http.Request) {

	req.ParseMultipartForm(10 << 20)

	input := html.EscapeString(req.FormValue("newpost"))
	input = Format_post(input)
	parent := req.FormValue("parent")

	now := time.Now().In(nip)
	post_time := now.Format("1/2/06(Mon)15:04:05")

	stmts := Checkout()
  defer Checkin(stmts)

	//file uploading
	file, handler, file_err := req.FormFile("file")
	//never := false

	if file_err == nil {
		defer file.Close()

		mime_type := handler.Header["Content-Type"][0]
		ext, supp := mime_ext[mime_type]
		
		buffer := make([]byte, 512)
		_, err := file.Read(buffer)
		Err_check(err)

		_, supp2 := mime_ext[http.DetectContentType(buffer)]
		_, err = file.Seek(0, io.SeekStart)
		Err_check(err)

		if supp && supp2 {
			file_pre := strconv.FormatInt(time.Now().UnixNano(), 10)
			file_name := file_pre + ext
			file_path := BP + "/Files/"

			f, err := os.OpenFile(file_path + file_name, os.O_WRONLY|os.O_CREATE, 0666)
			Err_check(err)
			defer f.Close()

			io.Copy(f, file)

			//think about if a pdf is given 
			Make_thumb(file_path, file_pre, file_name, mime_type)
			file_info := gen_info(handler.Size)

			stmt := stmts["newpost_wf"]
			_, err = stmt.Exec(input, post_time, parent, file_name, handler.Filename, file_info, file_pre + "s.webp")
			Err_check(err)
		}

	} else {
		stmt := stmts["newpost_nf"]
		_, err := stmt.Exec(input, post_time, parent)
		Err_check(err)
	}
	
	Build_thread()
	
	http.Redirect(w, req, req.Header.Get("Referer"), 302)
}
