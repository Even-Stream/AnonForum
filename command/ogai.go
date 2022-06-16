package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
)


var BP string
const rand_charset = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_"


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

func main() {

	flag.StringVar(&BP, "bp", "/mnt/c/server/data/content/media/toggle/", "the base path")
	flag.Parse()

	file, err := os.OpenFile(BP + "error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	Err_check(err)
	log.SetOutput(file)

	rand.Seed(time.Now().UnixNano())

	//New_db()
	Make_Conns()
	//Build_thread() 
	Listen()
}
