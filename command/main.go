package main

import (
    "database/sql"
    "log"
    "os"
    "math/rand"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

const rand_charset = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_"

var DB_uri string 

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

func Rand_gen() string {
    result := ""

    for i := 0; i < 6; i++ {
        c := rand_charset[rand.Intn(len(rand_charset))]
    result += string(c)
    }

    return result
}

func Time_report(entry string) {
    log.Printf(entry)
}

func main() {

    Load_conf()

    file, err := os.OpenFile(BP + "error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    Err_check(err)
    defer file.Close()

    log.SetOutput(file)
    log.SetFlags(log.LstdFlags | log.Lmicroseconds)

    rand.Seed(time.Now().UnixNano())

    if _, err = os.Stat(BP + "command/post-coll.db"); err != nil {
        New_db()
    }

    db_path := BP + "command/post-coll.db"
    DB_uri = "file://" + db_path + "?cache=private&_synchronous=NORMAL&_journal_mode=WAL"
    Make_Conns() 

    for _, board := range Board_names{
        Build_home()
        Build_board(board)
        Build_catalog(board)
    }
    
    Listen()
}
