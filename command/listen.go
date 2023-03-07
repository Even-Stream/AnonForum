package main

import (
    "net/http"
    "time"
    "strings"

    "golang.org/x/time/rate"
)

var rarr = []rate.Limit{20, .04, .5, 1, 20}
var barr = []int{30, 1, 1, 1, 10}
var limiter = NewIPRateLimiter(rarr, barr)

var admf_map = map[string]bool {
    "login": true,
    "verify": true,
    "logout": true,
    "console": true,
    "mod": true,}

func Listen() {

    go func() {
        for range time.Tick(time.Hour) {
            limiter = NewIPRateLimiter(rarr, barr)
    }}()

    //listen mux
    mux := http.NewServeMux()
    mux.HandleFunc("/im/ret/", Get_prev)
    mux.HandleFunc("/im/post/", New_post)
    mux.HandleFunc("/im/theme/", Switch_theme)
    mux.HandleFunc("/im/login/", Credential_check)
    mux.HandleFunc("/im/verify/", Token_check)
    mux.HandleFunc("/im/logout/", Logout)
    mux.HandleFunc("/im/console/", Load_console)
    mux.HandleFunc("/im/mod/", Admin_actions)

    srv := &http.Server {
        Addr: ":1024",
        IdleTimeout: 5 * time.Second,
        ReadTimeout: 3 * time.Second,
        ReadHeaderTimeout: 3 * time.Second,
        WriteTimeout: 5 * time.Second,
        Handler: hongMeiling(mux),
    }
    srv.ListenAndServe()
}

func hongMeiling(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {    

        var sel int
        fullurl := r.URL.String()
        url := strings.Split(fullurl, "/")[2]

        switch {
            case url == "ret":
                sel = 0
            case url == "post":
                sel = 1
            case url == "theme":
                sel = 2
            case url == "adm":
                sel = 3
            case admf_map[url]:
                sel = 4
        }

        climiter := limiter.GetLimiter(r.Header.Get("X-Real-IP"), sel)


            if !climiter.Allow() {
                    http.Error(w, "Request limit exceeded. Please wait.", http.StatusTooManyRequests)
                    return
            }

        next.ServeHTTP(w, r)    
    })
}
