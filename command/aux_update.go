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

    thread_collstmt := stmts["thread_coll"]
    thread_headstmt := stmts["thread_head"]

    var cat_body []*Post
    var subjects []string

    parent_rows, err := thread_collstmt.Query(board)
    Err_check(err)
    defer parent_rows.Close()

    for parent_rows.Next() {
        var cparent Post
        var filler int

        err = parent_rows.Scan(&cparent.Id, &filler)
        Err_check(err)

        err = thread_headstmt.QueryRow(cparent.Id, board).Scan(&cparent.Content, &cparent.Time, &cparent.File,
            &cparent.Filename, &cparent.Fileinfo, &cparent.Imgprev)
        Query_err_check(err)

        cat_body = append(cat_body, &cparent)
        subjects = append(subjects, Get_subject(strconv.Itoa(cparent.Id), board))
    }

    return cat_body, subjects
}

func get_home() ([]*Hp, []*Ht) {
    stmts := Checkout()
    defer Checkin(stmts)

    hp_collstmt := stmts["hp_coll"]
    ht_collstmt := stmts["ht_coll"]

    var home_posts []*Hp
    var home_thumbs []*Ht

    hp_rows, err := hp_collstmt.Query()
    Err_check(err)
    defer hp_rows.Close()

    for hp_rows.Next() {
        var chp Hp
        
        err = hp_rows.Scan(&chp.BoardN, &chp.Id, &chp.Content, &chp.Parent)
        Err_check(err)

        home_posts = append(home_posts, &chp)
    }

    ht_rows, err := ht_collstmt.Query()
    Err_check(err)
    defer ht_rows.Close()
    
    for ht_rows.Next() {
        var cht Ht
        
        err = hp_rows.Scan(&cht.BoardN, &cht.Id, &cht.Parent, &cht.Imgprev)
        Err_check(err)

        home_thumbs = append(home_thumbs, &cht)
    }

    return home_posts, home_thumbs 
}


func Build_catalog(board string) {
    cattemp := template.New("catalog.html").Funcs(catfuncmap)
    cattemp, err := cattemp.ParseFiles(BP + "/templates/catalog.html")
    Err_check(err)

    path := BP + "head/" + board + "/"
    Dir_check(path)

    f, err := os.Create(path + "catalog.html")
    Err_check(err)
    defer f.Close()

    posts, subjects := get_cat_posts(board)

    catalog := Catalog{Name: board, Posts: posts, Subjects: subjects, Header: Boards}
    cattemp.Execute(f, catalog)
}

func Build_home() {
    hometemp := template.New("home.html")
    hometemp, err := hometemp.ParseFiles(BP + "/templates/home.html")
    Err_check(err)

    path := BP + "head/"
    Dir_check(path)

    f, err := os.Create(path + "index.html")
    Err_check(err)
    defer f.Close()

    hps, hts := get_home()

    home := Home{Latest: hps, Thumbs: hts}
    hometemp.Execute(f, home)
}
