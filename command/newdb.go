package main

import (
    "database/sql"
    "os"
    "strings"

    _ "github.com/mattn/go-sqlite3"
)

const (
    createPostsTableSQL = `CREATE TABLE posts (
        "Board" TEXT NOT NULL,
        "Id" INTEGER NOT NULL,
        "Content" TEXT,
        "Time" TEXT,
        "Parent" INTEGER,
        "Password" TEXT NOT NULL,
        "Identifier" TEXT,
        "File" TEXT,
        "Filename" TEXT,
        "Fileinfo" TEXT,
        "Filemime" TEXT,
        "Imgprev" TEXT,
        "Hash" TEXT,
        "Option" TEXT,
        "Calendar" INTEGER NOT NULL,
        "Clock" INTEGER NOT NULL,
        "Pinned" INTEGER NOT NULL,
        "Locked" INTEGER NOT NULL,
        "Anchored" INTEGER NOT NULL,
        PRIMARY KEY (Board, Id)
    );`

    createRepliesTableSQL = `CREATE TABLE replies (
        "Board" TEXT NOT NULL,
        "Source" INTEGER NOT NULL,
        "Replier" INTEGER NOT NULL,
        "Password" TEXT NOT NULL,
        FOREIGN KEY ("Board", "Replier") REFERENCES posts("Board", "Id") ON DELETE CASCADE
    );`


    createSubjectsTableSQL = `CREATE TABLE subjects (
        "Board" TEXT NOT NULL,
        "Parent" INTEGER NOT NULL,
        "Subject" TEXT NOT NULL
    );`

    createLatestIdTableSQL = `CREATE TABLE latest (
        "Board" TEXT PRIMARY KEY,
        "Id" INTEGER NOT NULL
    );`

    createHomePostTableSQL = `CREATE TABLE homepost (
        "Board" TEXT NOT NULL,
        "Id" INTEGER NOT NULL,
        "Content" TEXT NOT NULL,
        "TrunContent" TEXT NOT NULL,
        "Parent" INTEGER NOT NULL,
        "Password" TEXT NOT NULL,
        FOREIGN KEY ("Board", "Id") REFERENCES posts("Board", "Id") ON DELETE CASCADE
    );`

    createHomeThumbTableSQL = `CREATE TABLE homethumb (
        "Board" TEXT NOT NULL,
        "Id" INTEGER NOT NULL,
        "Parent" TEXT NOT NULL,
        "Imgprev" TEXT NOT NULL,
        "Password" TEXT NOT NULL,
        FOREIGN KEY ("Board", "Id") REFERENCES posts("Board", "Id") ON DELETE CASCADE
    );`

    createCredTableSQL = `CREATE TABLE credentials (
        "Username" TEXT NOT NULL,
        "Hash" TEXT NOT NULL,
        "Type" INTEGER NOT NULL
    );`

    createTokenTableSQL = `CREATE TABLE tokens (
        "Token" TEXT NOT NULL,
        "Type" TEXT NOT NULL,
        "Time" TEXT NOT NULL
    );`

    createBannedTableSQL = `CREATE TABLE banned (
        "Identifier" TEXT NOT NULL,
        "Expiry" TEXT NOT NULL,
        "Mod" TEXT NOT NULL,
        "Content" TEXT,
        "Reason" TEXT
    );`

    createDeletedTableSQL = `CREATE TABLE deleted (
        "Identifier" TEXT NOT NULL,
        "Time" TEXT NOT NULL,
        "Mod" TEXT NOT NULL,
        "Content" TEXT,
        "Reason" TEXT
    );`

    //triggers
    createLatestTriggerSQL = `CREATE TRIGGER latest_update
        AFTER INSERT ON posts
        BEGIN
            UPDATE latest 
            SET Id = Id + 1
            WHERE Board = NEW.Board;
        END;`
        
    clearRepsTriggerSQL = `CREATE TRIGGER rep_clear
        AFTER UPDATE ON posts
        BEGIN
            DELETE FROM replies WHERE Replier = OLD.Id AND Board = OLD.Board;
            DELETE FROM homethumb WHERE Imgprev = OLD.Imgprev AND NEW.Imgprev = 'deleted';
        END;
    `
        
    anchorCheckSQL = `CREATE TRIGGER anchor_check
        AFTER INSERT ON posts
        BEGIN
            UPDATE posts
            SET Anchored = IIF((SELECT COUNT(Id) FROM posts WHERE Parent = NEW.Parent AND Board = NEW.Board AND Pinned <> 1) > 200, 1, 0)
            WHERE Id = NEW.Parent AND Board = NEW.Board;
        END;
    `
        
    trimHomePostStack = `CREATE TRIGGER homepost_trim
        AFTER INSERT ON homepost
        BEGIN
            DELETE FROM homepost WHERE ROWID =
                IIF((SELECT COUNT(Id) FROM homepost) > 20,
                (SELECT min(ROWID) from homepost), NULL);
        END;`

    trimHomeThumbStack = `CREATE TRIGGER homethumb_trim
        AFTER INSERT ON homethumb
        BEGIN
            DELETE FROM homethumb WHERE ROWID =
                IIF((SELECT COUNT(Id) FROM homethumb) > 10,
                (SELECT min(ROWID) from homethumb), NULL);
        END;`
        
        
    //how new posts know what their id is 
    latestseedSQL = `INSERT OR IGNORE INTO latest (Board, Id) VALUES (cb, 1);`
)

func create_table(db *sql.DB) {

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

    statement, err = db.Prepare(createBannedTableSQL)
        Err_check(err)
    statement.Exec()

    statement, err = db.Prepare(createDeletedTableSQL)
        Err_check(err)
    statement.Exec()

    statement, err = db.Prepare(createLatestTriggerSQL)
        Err_check(err)
    statement.Exec()
    
    statement, err = db.Prepare(clearRepsTriggerSQL)
        Err_check(err)
    statement.Exec()
    
    statement, err = db.Prepare(anchorCheckSQL)
        Err_check(err)
    statement.Exec()

    statement, err = db.Prepare(trimHomePostStack)
        Err_check(err)
    statement.Exec()

    statement, err = db.Prepare(trimHomeThumbStack)
        Err_check(err)
    statement.Exec()

}


func LatestSeed() {
    conn, err := sql.Open("sqlite3", DB_path)
    Err_check(err)
    defer conn.Close()
    
    for board := range Board_map {
        statement, err := conn.Prepare(strings.Replace(latestseedSQL, "cb", `'` + board + `'`, 1))
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
