package main

import (
    "net/http"
    "os/exec"
    "text/template"
)

type vidresult struct {
    VidUrl string
}

func Vidget(w http.ResponseWriter, req *http.Request) {
    vidid := req.FormValue("id")
    if vidid == "" {
        http.Error(w, "Video Id Not Given", http.StatusBadRequest)
        return
    }

    vidurlget := exec.Command("/usr/bin/yt-dlp", `--get-url`, `-f`, `b`, `https://youtu.be/` + vidid)
    vidurl, _ := vidurlget.Output()

    video_temp := template.New("video.html")
    video_temp, err := video_temp.ParseFiles(BP + "/templates/video.html")

    result := vidresult{VidUrl: string(vidurl)}
    err = video_temp.Execute(w, result)
    Err_check(err)
}