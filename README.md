An anonymous message board engine written in golang.

## Compile Instructions
sudo apt install cmake git build-essential libvips-dev libavformat-dev libswresample-dev libavcodec-dev libavutil-dev libavformat-dev libswscale-dev

sudo apt install golang-1.21/bookworm-backports
`Or compile the latest version of Go`

sudo apt install yt-dlp/bookworm-backports

git clone https://github.com/Even-Stream/AnonForum

cd AnonForum/command

go mod init modules

go mod tidy 

go build -o forum *.go

`To use nginx`

wget (current version from http://nginx.org/en/download.html)

tar -xzvf nginx-(current version).tar.gz

git clone https://github.com/vision5/ngx_devel_kit

git clone https://github.com/openresty/set-misc-nginx-module

git clone https://github.com/yaoweibin/ngx_http_substitutions_filter_module

cd nginx-(current version)

sudo ./configure --add-module=../ngx_devel_kit --add-module=../set-misc-nginx-module --add-module=../ngx_http_substitutions_filter_module

sudo make & make install 

## Post Formatting
quote: >example

reply: >>1

cross-board reply: >>/board/1

spoiler: \~\~example\~\~

bold: \*\*example\*\*

italics: \_\_example\_\_