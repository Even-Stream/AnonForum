package main

import (
    "net/http"
    "database/sql"
    "time"
"fmt"

    "github.com/alexedwards/argon2id"
    _ "github.com/mattn/go-sqlite3"
    "github.com/google/uuid"
)

type Acc_type int64
const (
    Admin     Acc_type = iota
    Moderator
    Maid
)

const (
    html_head = `<!DOCTYPE html>
    <html>
    <head>
        <title>Administration</title>
    </head>
    <body><center><br>`

    html_foot = `</center></body>
    </html>`

    entry_form = `
    <form action="/im/admf_login/" enctype="multipart/form-data" method="Post">
        <label>Username:</label><br><br>
        <input name="username" type="text" value=""><br><br>
        <label>Password:</label><br><br>
        <input name="password" type="password" value=""><br><br>
        <input type="submit" value="Enter">
    </form>`

    create_form = `
    <form action="/im/admf_verify/" enctype="multipart/form-data" method="Post">
        <label>Username:</label><br><br>
        <input name="username" type="text" value=""><br><br>
        <label>Password:</label><br><br>
        <input name="password" type="password" value=""><br><br>
        <label>Confirm Password:</label><br><br>
        <input name="passwordcopy" type="password" value=""><br><br>
        <label>Token:</label><br><br>
        <input name="token" type="text" value=""><br><br>
        <input type="submit" value="Enter">
    </form>`

    add_token_string = `INSERT INTO tokens(Token, Type) VALUES (?, ?)`
    search_token_string = `SELECT Type FROM tokens WHERE Token = ?`
    delete_token_string = `DELETE FROM tokens where Token = ?`
    new_user_string = `INSERT INTO credentials(Username, Hash, Type) VALUES (?, ?, ?)`
    search_user_string = `SELECT Hash, Type FROM credentials WHERE Username = ?`
)

var params = &argon2id.Params{
	Memory:      128 * 1024,
	Iterations:  4,
	Parallelism: 4,
	SaltLength:  16,
	KeyLength:   32,
}

type session struct {
    username string
    expiry   time.Time
    acc_type Acc_type
}

var Sessions = map[string]session{}

func (s session) IsExpired() bool {
    return s.expiry.Before(time.Now())
}

func Admin_init() {
    conn, err := sql.Open("sqlite3", DB_uri)
    Err_check(err)
    defer conn.Close()
    add_token_stmt, err := conn.Prepare(add_token_string)
    Err_check(err)

    add_token_stmt.Exec("500", Admin)
}

func Request_filter(w http.ResponseWriter, req *http.Request, method string, max_size int64) int {
    if req.Method != method {
        http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
        return 0
    }

    req.Body = http.MaxBytesReader(w, req.Body, max_size)
    if err := req.ParseForm(); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return 0
    }

    return 1
}

func Entry_check(w http.ResponseWriter, req *http.Request, entry string, value string) int {
    if value == "" {
        http.Error(w, entry + " not specified", http.StatusBadRequest)
        return 0
    }

    return 1    
}

//listened to

//account creation
func Create_account(w http.ResponseWriter, req *http.Request) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(html_head + create_form + html_foot))
}

func Token_check (w http.ResponseWriter, req *http.Request) {

    if Request_filter(w, req, "POST", 1 << 10) == 0 {return}
    if err := req.ParseMultipartForm(1 << 10); err != nil {
        http.Error(w, "Request size exceeds limit.", http.StatusBadRequest)
        return
    }

    token := req.FormValue("token")
    if Entry_check(w, req, "token", token) == 0 {return}
    username := req.FormValue("username")
    if Entry_check(w, req, "username", username) == 0 {return}
    password := req.FormValue("password")
    if Entry_check(w, req, "password", password) == 0 {return}
    passwordcopy := req.FormValue("passwordcopy")
    if Entry_check(w, req, "passwordcopy", passwordcopy) == 0 {return}

    if password != passwordcopy {
        http.Error(w, "Passwords don't match.", http.StatusBadRequest)
        return
    }

    //look in database for token, if there, delete token, create account 
    conn, err := sql.Open("sqlite3", DB_uri)
    Err_check(err)
    defer conn.Close()
    search_token_stmt, err := conn.Prepare(search_token_string)
    Err_check(err)

    var acc_type Acc_type
    err = search_token_stmt.QueryRow(token).Scan(&acc_type)
    if err == sql.ErrNoRows {
        http.Error(w, "Invalid token.", http.StatusBadRequest)
        return
    }

    //password length enforce
    pass_length := len([]rune(password))
    if pass_length > 30 || pass_length < 10 {
        http.Error(w, "Password not in valid range(10-30 characters)", http.StatusBadRequest)
        return 
    }

    //deleting token
    conn2, err := sql.Open("sqlite3", DB_uri)
    Err_check(err)
    defer conn2.Close()
    delete_token_stmt, err := conn2.Prepare(delete_token_string)
    Err_check(err)
    delete_token_stmt.Exec(token)


    conn3, err := sql.Open("sqlite3", DB_uri)
    Err_check(err)
    defer conn3.Close()
    new_user_stmt, err := conn3.Prepare(new_user_string)
    Err_check(err)

    hash, err := argon2id.CreateHash(password, params)
    Err_check(err)

    new_user_stmt.Exec(username, hash, acc_type)

    w.Write([]byte(html_head + `<p>Account created.</p>` + html_foot))
}


//account enter
func Console_enter(w http.ResponseWriter, req *http.Request) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(html_head + entry_form + html_foot))
}

func Credential_check (w http.ResponseWriter, req *http.Request) {

    if Request_filter(w, req, "POST", 1 << 9) == 0 {return}
    if err := req.ParseMultipartForm(1 << 9); err != nil {
        http.Error(w, "Request size exceeds limit.", http.StatusBadRequest)
        return
    }

    password := req.FormValue("password")
    if Entry_check(w, req, "password", password) == 0 {return}
    username := req.FormValue("username")
    if Entry_check(w, req, "username", username) == 0 {return}

    pass_length := len([]rune(req.FormValue("password")))
    if pass_length > 30 || pass_length < 10 {
        http.Error(w, "Password not in valid range(10-30 characters)", http.StatusBadRequest)
        return 
    }

    //database check
    conn, err := sql.Open("sqlite3", DB_uri)
    Err_check(err)
    defer conn.Close()
    search_user_stmt, err := conn.Prepare(search_user_string)
    Err_check(err)

    var found_hash string
    var acc_type Acc_type

    err = search_user_stmt.QueryRow(username).Scan(&found_hash, &acc_type)
    if err == sql.ErrNoRows {
        http.Error(w, "Invalid credentials.", http.StatusBadRequest)
        return
    }

    //match check
    match, err := argon2id.ComparePasswordAndHash(password, found_hash)
    Err_check(err)

    if !match {
        http.Error(w, "Invalid credentials.", http.StatusBadRequest)
        return
    }

    sessionToken := uuid.NewString()
    expiresAt := time.Now().Add(20 * time.Minute)

    Sessions[sessionToken] = session{
        username: username,
        expiry:   expiresAt,
        acc_type: acc_type,
    }

    http.SetCookie(w, &http.Cookie{
        Name:    "session_token",
        Value:   sessionToken,
        Expires: expiresAt,
        Path: "/",
    })

    w.Write([]byte(html_head + `<p>Welcome.</p>` + html_foot))
}

//account exit 
func Logout(w http.ResponseWriter, req *http.Request) {
    c, err := req.Cookie("session_token")

    if err != nil {
    fmt.Println(err)
        if err == http.ErrNoCookie {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    sessionToken := c.Value
    delete(Sessions, sessionToken)

    http.SetCookie(w, &http.Cookie{
        Name:    "session_token",
        Value:   "",
        Expires: time.Now(),
        Path: "/",
    })

    w.Write([]byte(html_head + `<p>Logged out.</p>` + html_foot))
}


//creating token(requires admin account)
