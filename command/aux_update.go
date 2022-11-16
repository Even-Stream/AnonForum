package main

import (
    "os"
    "text/template"
    "strconv"
    "errors"

    _ "github.com/mattn/go-sqlite3"
)

//structures used in templates
type Catalog struct {
    Name string
    Posts []*Post
    Subjects []string
    Header []string
}

type Hp {
    BoardN string
    Id int
    Content string
    Parent string
}

type Ht {
    BoardN string
    Id int
    Parent string
    Imgprev string
}

type Home struct {
    Latest []*Hp
    Thumbs []*Ht
    BList []string        //same as Header
    News string
    FAQ string
    Rules string
    Board_info string
}

//catalog template function for making new rows
var catfuncmap = template.FuncMap{
    "startrow": func(rowsize, index int) bool {
        if index % rowsize == 0 {
            return true
        }
        return false
    },
}


func get_cat_posts(board string) ([]*Post, []string) {
    stmts := Checkout()
    defer Checkin(stmts)

    stmt0 := stmts["thread_coll"]
    stmt := stmts["thread_head"]

    var cat_body []*Post
    var subjects []string

    parent_rows, err := stmt0.Query(board)
    Err_check(err)
    defer parent_rows.Close()

    for parent_rows.Next() {
        var cparent Post
        var filler int

        err = parent_rows.Scan(&cparent.Id, &filler)
        Err_check(err)

        err = stmt.QueryRow(cparent.Id, board).Scan(&cparent.Content, &cparent.Time, &cparent.File,
            &cparent.Filename, &cparent.Fileinfo, &cparent.Imgprev)
        Query_err_check(err)

        cat_body = append(cat_body, &cparent)
        subjects = append(subjects, Get_subject(strconv.Itoa(cparent.Id), board))
    }

    return cat_body, subjects
}


func Build_catalog(board string) {
    cattemp := template.New("catalog.html").Funcs(catfuncmap)
    cattemp, err := cattemp.ParseFiles(BP + "/templates/catalog.html")
    Err_check(err)

    path := BP + "head/" + board + "/"
    Dir_check(path)

    f, err := os.Create(path + "/catalog.html")
    Err_check(err)
    defer f.Close()

    posts, subjects := get_cat_posts(board)

    catalog := Catalog{Name: board, Posts: posts, Subjects: subjects, Header: Boards}
    cattemp.Execute(f, catalog)
}

func Build_home() {

}
