package main

import (
    "database/sql"
    "os"
    "strings"

    _ "github.com/mattn/go-sqlite3"
)

func create_table(db *sql.DB) {

    createPostsTableSQL := `CREATE TABLE posts (
        "Board" TEXT NOT NULL,
        "Id" INTEGER NOT NULL,
        "Content" TEXT,
        "Time" TEXT,
        "Parent" INTEGER,
        "Password" TEXT,
        "Identifier" TEXT,
        "File" TEXT,
        "Filename" TEXT,
        "Fileinfo" TEXT,
        "Imgprev" TEXT,
        "Phash" TEXT,
        "Option" TEXT
    );`

    createRepliesTableSQL := `CREATE TABLE replies (
        "Board" TEXT NOT NULL,
        "Source" INTEGER NOT NULL,
        "Replier" INTEGER NOT NULL
    );`


    createSubjectsTableSQL := `CREATE TABLE subjects (
        "Board" TEXT NOT NULL,
        "Parent" INTEGER NOT NULL,
        "Subject" TEXT NOT NULL
    );`

    createLatestIdTableSQL := `CREATE TABLE latest (
        "Board" TEXT NOT NULL,
        "Id" INTEGER NOT NULL
    );`

    createLatestTriggerSQL := `CREATE TRIGGER latest_update
        AFTER INSERT ON posts
        BEGIN
            UPDATE latest 
            SET Id = Id + 1
            WHERE Board = NEW.Board;
        END;`

    latestseedSQL := `INSERT INTO latest (Board, Id) VALUES (cb, 1);`


    statement, err := db.Prepare(createPostsTableSQL)
    Err_check(err)
    statement.Exec()

    statement, err = db.Prepare(createRepliesTableSQL)
        Err_check(err)
    statement.Exec()

    statement, err = db.Prepare(createSubjectsTableSQL)
        Err_check(err)
    statement.Exec()

    statement, err = db.Prepare(createLatestIdTableSQL)
        Err_check(err)
    statement.Exec()

    statement, err = db.Prepare(createLatestTriggerSQL)
        Err_check(err)
    statement.Exec()


    for _, board := range Boards {
        statement, err = db.Prepare(strings.Replace(latestseedSQL, "cb", `'` + board + `'`, 1))
            Err_check(err)
        statement.Exec()
    }

}

func New_db() {

    file, err := os.Create(BP + "command/post-coll.db")
    Err_check(err)

    file.Close()

    conn, err := sql.Open("sqlite3", BP + "command/post-coll.db")
    Err_check(err)
    defer conn.Close()

    create_table(conn)
}
