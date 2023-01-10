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
        "Filemime" TEXT,
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

    createHomePostTableSQL := `CREATE TABLE homepost (
        "Board" TEXT NOT NULL,
        "Id" INTEGER NOT NULL,
        "Content" TEXT NOT NULL,
        "TrunContent" TEXT NOT NULL,
        "Parent" INTEGER NOT NULL
    );`

    createHomeThumbTableSQL := `CREATE TABLE homethumb (
        "Board" TEXT NOT NULL,
        "Id" INTEGER NOT NULL,
        "Parent" TEXT NOT NULL,
        "Imgprev" TEXT NOT NULL
    );`

    createCredTableSQL := `CREATE TABLE credentials (
        "Username" TEXT NOT NULL,
        "Hash" TEXT NOT NULL,
        "Type" INTEGER NOT NULL
    );`

    createTokenTableSQL := `CREATE TABLE tokens (
        "Token" TEXT NOT NULL,
        "Type" TEXT NOT NULL
    );`

    //triggers
    createLatestTriggerSQL := `CREATE TRIGGER latest_update
        AFTER INSERT ON posts
        BEGIN
            UPDATE latest 
            SET Id = Id + 1
            WHERE Board = NEW.Board;
        END;`
        
    trimHomePostStack := `CREATE TRIGGER homepost_trim
        AFTER INSERT ON homepost
        BEGIN
            DELETE FROM homepost WHERE ROWID =
                IIF((SELECT COUNT(Id) FROM homepost) > 20,
                (SELECT min(ROWID) from homepost), NULL);
        END;`

    trimHomeThumbStack := `CREATE TRIGGER homethumb_trim
        AFTER INSERT ON homethumb
        BEGIN
            DELETE FROM homethumb WHERE ROWID =
                IIF((SELECT COUNT(Id) FROM homethumb) > 5,
                (SELECT min(ROWID) from homethumb), NULL);
        END;`
        
    //how new posts know what their id is 
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

    statement, err = db.Prepare(createHomePostTableSQL)
        Err_check(err)
    statement.Exec()

    statement, err = db.Prepare(createHomeThumbTableSQL)
        Err_check(err)
    statement.Exec()


    statement, err = db.Prepare(createCredTableSQL)
        Err_check(err)
    statement.Exec()

    statement, err = db.Prepare(createTokenTableSQL)
        Err_check(err)
    statement.Exec()

    statement, err = db.Prepare(createLatestTriggerSQL)
        Err_check(err)
    statement.Exec()

    statement, err = db.Prepare(trimHomePostStack)
        Err_check(err)
    statement.Exec()

    statement, err = db.Prepare(trimHomeThumbStack)
        Err_check(err)
    statement.Exec()

    for _, board := range Board_names {
        statement, err = db.Prepare(strings.Replace(latestseedSQL, "cb", `'` + board + `'`, 1))
            Err_check(err)
        statement.Exec()
    }

}

func New_db() {

    file, err := os.Create(DB_path)
    Err_check(err)

    file.Close()

    conn, err := sql.Open("sqlite3", DB_path)
    Err_check(err)
    defer conn.Close()

    create_table(conn)
}
