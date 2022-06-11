package main

import (
	"database/sql"
	"flag"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var BP string

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


func main() {

	flag.StringVar(&BP, "BP", "/mnt/c/server/data/content/media/toggle/", "the base path")
	flag.Parse()

	//New_db() 
	Make_Conns()
	Listen()
	//Build_thread()
}
