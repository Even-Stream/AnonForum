package main

import "net/http"

func Listen() {

	//listen mux
	mux := http.NewServeMux()
	mux.HandleFunc("/im/ret/", Get_prev)
	mux.HandleFunc("/im/post/", New_post)
	mux.HandleFunc("/im/theme/", Switch_theme)
	http.ListenAndServe(":81", mux)
}