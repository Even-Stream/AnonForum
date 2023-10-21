package main

import (
    "database/sql"
    "log"
    "os"
    "math/rand"
    "time"
    "errors"
    "strings"
    "io/fs"

    _ "github.com/mattn/go-sqlite3"
)

var DB_uri string
var DB_path string 

func Err_check(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

func Query_err_check(err error) {
    if err != nil {

        if err == sql.ErrNoRows {
            // there were no rows, but otherwise no error occurred
        } else {
                log.Fatal(err)
            }

    }
}

func Time_report(entry string) {
    log.Printf(entry)
}

func Delete_file(file_path, file_name, imgprev string) {
    err := os.Remove(file_path + file_name)
    if !errors.Is(err, fs.ErrNotExist) {Err_check(err)} 
                
    if imgprev != "" && !strings.HasSuffix(imgprev, "image.webp") {
        err = os.Remove(file_path + imgprev)
        if !errors.Is(err, fs.ErrNotExist) {Err_check(err)}
    }
}

func main() {

    Load_conf()

    file, err := os.OpenFile(BP + "error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    Err_check(err)
    defer file.Close()

    log.SetOutput(file)
    log.SetFlags(log.LstdFlags | log.Lmicroseconds)
    rand.Seed(time.Now().UnixNano())

    DB_path = BP + "command/post-coll.db"
    DB_uri = "file://" + DB_path + "?_foreign_keys=on&cache=private&_synchronous=NORMAL&_journal_mode=WAL"
    if _, err = os.Stat(DB_path); err != nil {
        New_db()
        Admin_init()
    }
    Make_Conns()
    go Clean(40 * time.Hour, "get_deleted", "delete_remove")
    go Clean(10 * time.Minute, "get_expired_tokens", "delete_expired_token")
 
    for board, _ := range Board_map{
        Build_home()
        Build_board(board)
        Build_catalog(board)
    }
    Listen()
}
