# Ogai

An image board engine written in golang. 

live instance(url subject to change):
https://[200:c5b0:cfeb:5db:c054:d66d:eb6f:7412]/content/media/toggle/

## Compile Instructions
sudo apt install build-essential cmake git

`Compile the latest version of Go`

git clone https://gitgud.io/nvtelen/ogai

cd ogai/command

go mod init modules

go mod tidy 

go build -o ogai *.go

`To use nginx, compile it with this configuration`

sudo ./configure --with-http_ssl_module --with-http_v2_module --with-http_sub_module

## Project status
6/11/22

ongoing
