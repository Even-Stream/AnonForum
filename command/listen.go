package main

import (
    "net/http"
    "strings"
    "time"

    "golang.org/x/time/rate"
)

var rarr = []rate.Limit{20, .04, .5, 1, .1}
var barr = []int{30, 1, 1, 1, 4}
var limiter = NewIPRateLimiter(rarr, barr)

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
    mux.HandleFunc("/im/adm/", Console_enter)
    mux.HandleFunc("/im/login/", Credential_check)
    http.ListenAndServe(":81", hongMeiling(mux))
}

func hongMeiling(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {    

        var sel int
        url := r.URL.String()
        switch {
            case strings.Contains(url, "ret"):
                sel = 0
            case strings.Contains(url, "post"):
                sel = 1
            case strings.Contains(url, "theme"):
                sel = 2
            case strings.Contains(url, "adm"):
                sel = 3
            case strings.Contains(url, "login"):
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
