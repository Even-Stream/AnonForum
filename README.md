# Ogai

An image board engine written in golang.

Live instance:

- https://sayachan.pl/ 
- https://[200:c5b0:cfeb:5db:c054:d66d:eb6f:7412]:4443/ (fastest, version 0.5 or above)
- https://s4taqq6ysw2wde4rcmq6xk5vvgmab3cuqtykwfq2padiwh2rncjfxcyd.onion:4443/
- https://mwmrm4yaihoyk7acurfnmfiuucn1ozeyaewo15zdttfeuek5rsto.loki:4443/

## Advantages

Uses CSS instead of JS for some common features

- Expanding thumbnails

- Linked-to post highlighting 

Takes advantage of Nginx capabilities

- Banners without JS

- Theme picker without JS

Other

- Editing or deleting your last post without JS

- Files keep their original name when downloaded 

- Video embedding without JS(uses ytp-dl) 

- Webp thumbnails 

- Uses Sqlite by default 

- No PHP or Perl

## Compile Instructions
sudo apt install build-essential cmake git libvips-dev libavformat-dev libswresample-dev libavcodec-dev libavutil-dev libavformat-dev libswscale-dev

sudo apt install golang-1.21/bookworm-backports
`Or compile the latest version of Go`

sudo apt install yt-dlp/bookworm-backports

git clone https://gitgud.io/nvtelen/ogai

cd ogai/command

go mod init modules

go mod tidy 

go build -o ogai *.go

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

## Project status
As of 10/1/23, Ogai is considered feature complete. 

Future updates are possible, but no longer a priority for me.
