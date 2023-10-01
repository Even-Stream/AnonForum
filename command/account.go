package main

import (
    "net/http"
    "database/sql"
    "time"
    //"fmt"

    "github.com/alexedwards/argon2id"
    _ "github.com/mattn/go-sqlite3"
    "github.com/google/uuid"
)

type Acc_type int64
const (
    Admin     Acc_type = iota
    Mod
    Maid
)

const (
    html_head = `<!DOCTYPE html>
    <html>
    <head>
        <style>
            body {
                background-color: #000c; 
                color: #ffffffdb;
            }

            a {
                color: #9dd1ff;
            }
        </style>`
    
    html_def_head = `
        <title>Administration</title>
    </head>
    <body><center><br>`

    html_tologin_head = `
        <title>Administration</title>
        <meta http-equiv="refresh" content="1; url=/login.html" />
    </head>
    <body><center><br>`	

    html_toentrance_head = `
        <title>Administration</title>
        <meta http-equiv="refresh" content="1; url=/entrance.html" />
    </head>
    <body><center><br>`

    html_toadministration_head = `        
        <title>Administration</title>
        <meta http-equiv="refresh" content="10; url=/entrance.html" />
    </head>
    <body><center><br>`    

    html_tohome_head = `
        <title>Administration</title>
        <meta http-equiv="refresh" content="1; url=/" />
    </head>
    <body><center><br>`

    html_foot = `</center></body>
    </html>`
)

var Argon_params = &argon2id.Params{
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

var Sessions = map[string]*session{}

func (s session) IsExpired() bool {
    return s.expiry.Before(time.Now().In(Nip))
}

func Admin_init() {
    conn, err := sql.Open("sqlite3", DB_uri)
    Err_check(err)
    defer conn.Close()
    add_token_stmt, err := conn.Prepare(Add_token_string)
    Err_check(err)

    add_token_stmt.Exec("500", Admin, time.Now().In(Nip).Add(time.Hour * 1).Format(time.UnixDate))
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

func Token_check (w http.ResponseWriter, req *http.Request) {
    ctx := req.Context()

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
    new_conn := WriteConnCheckout()
    defer WriteConnCheckin(new_conn)
    new_tx, err := new_conn.Begin()
    Err_check(err)
    defer new_tx.Rollback()

    var acc_type Acc_type
    err = new_tx.QueryRowContext(ctx, search_token_string, token).Scan(&acc_type)
    if err == sql.ErrNoRows {
        http.Error(w, "Invalid token.", http.StatusBadRequest)
        return
    } else {
        Err_check(err)
    }

    //look in database for username
    err = new_tx.QueryRowContext(ctx, search_user_string, username).Scan()
    if err != sql.ErrNoRows {
        http.Error(w, "Username already in use.", http.StatusBadRequest)
        return
    }

    //password length enforce
    pass_length := len([]rune(password))
    if pass_length > 30 || pass_length < 10 {
        http.Error(w, "Password not in valid range(10-30 characters)", http.StatusBadRequest)
        return 
    }

    //deleting token
    _, err = new_tx.ExecContext(ctx, delete_token_string, token)
    Err_check(err)

    hash, err := argon2id.CreateHash(password, Argon_params)
    Err_check(err)
    
    _, err = new_tx.ExecContext(ctx, new_user_string, username, hash, acc_type)
    Err_check(err)
    
    err = new_tx.Commit()
    Err_check(err)

    w.Write([]byte(html_head + html_tologin_head + `<p>Account created.</p>` + html_foot))
}

func Account_refresh(w http.ResponseWriter, sessionToken string) {
    expiresAt := time.Now().In(Nip).Add(10 * time.Minute)
    Sessions[sessionToken].expiry = expiresAt

    http.SetCookie(w, &http.Cookie{
        Name:    "session_token",
        Value:   sessionToken,
        Expires: expiresAt,
        Path: "/",
    })
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
    expiresAt := time.Now().In(Nip).Add(10 * time.Minute)

    Sessions[sessionToken] = &session{
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

    w.Write([]byte(html_head + html_toentrance_head + `<p>Welcome.</p>` + html_foot))
}

func Logged_in_check(w http.ResponseWriter, req *http.Request) *session {
    c, err := req.Cookie("session_token")

    if err != nil {
        if err == http.ErrNoCookie {
            http.Error(w, "Unauthorized.", http.StatusUnauthorized)
            return nil
        }
        w.WriteHeader(http.StatusBadRequest)
        return nil
    }

    sessionToken := c.Value
    userSession, exists := Sessions[sessionToken]
    if !exists {
        http.Error(w, "Unauthorized.", http.StatusUnauthorized)
        return nil
    }

    if userSession.IsExpired() {
        delete(Sessions, sessionToken)
        http.Error(w, "Session expired.", http.StatusUnauthorized)
        return nil
    } else {Account_refresh(w, sessionToken)}

    return userSession
}

//account exit 
func Logout(w http.ResponseWriter, req *http.Request) {
    c, err := req.Cookie("session_token")

    if err != nil {
        if err == http.ErrNoCookie {
            http.Error(w, "Unauthorized.", http.StatusUnauthorized)
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

    w.Write([]byte(html_head + html_tohome_head + `<p>Logged out.</p>` + html_foot))
}
