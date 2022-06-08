package main 

import (
  "github.com/h2non/bimg"
)


var taller = bimg.Options{
  		Height: 200,
  		Type: bimg.WEBP,
	}

var wider = bimg.Options{
  		Width: 200,
  		Type: bimg.WEBP,
	}

var small = bimg.Options{
  		Type: bimg.WEBP,
	}

var options = [3]bimg.Options{taller, wider, small}


func Make_thumb(file_path, file_pre, file_name, mime_type string) {

	var selected bimg.Options
	file_full := file_path + file_name

	buffer, err := bimg.Read(file_full)
	Err_check(err)

	file_dim, err := bimg.Size(buffer) 
	Err_check(err)

	if file_dim.Height >= file_dim.Width && file_dim.Height > 200 {
		selected = options[0]
	} else if file_dim.Width > file_dim.Height && file_dim.Width > 200 {
		selected = options[1]
	} else {
		selected = options[2]
	}

	newImage, err := bimg.NewImage(buffer).Process(selected)
	Err_check(err)

	bimg.Write(file_path + file_pre + "s.webp", newImage)
}
