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

`To use nginx`

wget (current version from http://nginx.org/en/download.html)

tar -xzvf nginx-(current version).tar.gz

git clone https://github.com/vision5/ngx_devel_kit

git clone https://github.com/openresty/set-misc-nginx-module

cd nginx-(current version)

sudo ./configure --with-http_sub_module --add-module=../ngx_devel_kit --add-module=../set-misc-nginx-module

sudo make & make install 

## Post Formatting
quote: >example

reply: >>1

spoiler: \~\~example\~\~

bold: \*\*example\*\*

italics: \_\_example\_\_

## Project status
12/4/22

ongoing
