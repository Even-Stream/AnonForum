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

const max_upload_size = 1024 * 1024 * 11	//11MB

func gen_info(size int64, width int, height int) string {
	file_info := units.HumanSize(float64(size))
	file_info = file_info + ", " + strconv.Itoa(width) + "x" + strconv.Itoa(height)
	return file_info
}

func New_post(w http.ResponseWriter, req *http.Request) {

	if req.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

 	req.Body = http.MaxBytesReader(w, req.Body, max_upload_size)
	if err := req.ParseMultipartForm(max_upload_size); err != nil {
		http.Error(w, "Request size exceeds limit.", http.StatusBadRequest)
		return
	}

	input := html.EscapeString(req.FormValue("newpost"))
	input = Format_post(input)
	parent := req.FormValue("parent")

	now := time.Now().In(nip)
	post_time := now.Format("1/2/06(Mon)15:04:05")

	stmts := writeCheckout()
  	defer writeCheckin(stmts)

	//file uploading
	file, handler, file_err := req.FormFile("file")

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
			file_path := BP + "/Files/"

			f, err := os.OpenFile(file_path + file_name, os.O_WRONLY|os.O_CREATE, 0666)
			Err_check(err)
			defer f.Close()

			io.Copy(f, file)

			//think about if a pdf is given 
			width, height := Make_thumb(file_path, file_pre, file_name, mime_type)
			file_info := gen_info(handler.Size, width, height)

			stmt := stmts["newpost_wf"]

			ofname := []rune(handler.Filename)
			rem := len(ofname) - 20
			if rem < 0 {
				rem = 0
			}
			ffname := string(ofname[rem:])

			_, err = stmt.Exec(input, post_time, parent, file_name, ffname, file_info, file_pre + "s.webp")
			Err_check(err)
		}
	//file not present 
	} else {
		stmt := stmts["newpost_nf"]
		_, err := stmt.Exec(input, post_time, parent)
		Err_check(err)
	}
	
	Build_thread()
	
	http.Redirect(w, req, req.Header.Get("Referer"), 302)
}
