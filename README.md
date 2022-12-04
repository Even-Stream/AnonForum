# Ogai

An image board engine written in golang. 

live instance(url subject to change):
https://[200:c5b0:cfeb:5db:c054:d66d:eb6f:7412]/content/media/toggle/

## Advantages

Uses CSS instead of JS for some common features

- Expanding thumbnails

- Linked-to post highlighting 

Takes advantage of Nginx features

- Theme picker with no JS

Other

- Files keep their original name when downloaded 

- Embeded thumbnailer

- Webp thumbnails 

- Uses Sqlite by default 

- No PHP or Perl

## Compile Instructions
sudo apt install build-essential cmake git libvips-dev

`Compile the latest version of Go`

git clone https://gitgud.io/nvtelen/ogai

cd ogai/command

go mod init modules

go mod tidy 

go build -o ogai *.go

`To use nginx, compile it with this configuration`

sudo ./configure --with-http_sub_module

## Project status
7/29/22

ongoing
