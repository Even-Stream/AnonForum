package main

import (
	"net/http"
	"strings"

	"golang.org/x/time/rate"
)

var rarr = []rate.Limit{10, .02, .5}
var barr = []int{30, 1, 1}
var limiter = NewIPRateLimiter(rarr, barr)

func Listen() {

	//listen mux
	mux := http.NewServeMux()
	mux.HandleFunc("/im/ret/", Get_prev)
	mux.HandleFunc("/im/post/", New_post)
	mux.HandleFunc("/im/theme/", Switch_theme)
	http.ListenAndServe(":81", hongMeiling(mux))
}

func hongMeiling(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println(next)
		//fmt.Println(r.URL)	
				
		var sel int
		url := r.URL.String()
		switch {
			case strings.Contains(url, "ret"):
				sel = 0
			case strings.Contains(url, "post"):
				sel = 1
			case strings.Contains(url, "theme"):
				sel = 2
		}

		climiter := limiter.GetLimiter(r.Header.Get("X-Real-IP"), sel)

        	if !climiter.Allow() {
            		http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
            		return
        	}

		next.ServeHTTP(w, r)	
	})
}