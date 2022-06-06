package main

import "net/http"

func Listen() {

	//listen mux
	mux := http.NewServeMux()
	mux.HandleFunc("/im/ret/", Get_prev)
	mux.HandleFunc("/im/post/", New_post)
	http.ListenAndServe(":81", mux)
}