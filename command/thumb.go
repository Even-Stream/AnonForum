package main 

import (
    "bytes"
    "image"
    _ "golang.org/x/image/webp"
    
    gih "github.com/corona10/goimagehash"
    "github.com/h2non/bimg"
)

func Make_thumb(file_path, file_pre string, file_buffer []byte, max_dim int) (int, int, string, error) {

    var selected bimg.Options
    selected.Type = bimg.WEBP

    file_dim, err := bimg.Size(file_buffer) 
    if err != nil {return 0, 0, "", err}

    if file_dim.Height >= file_dim.Width && file_dim.Height > max_dim {
        selected.Height = max_dim
    } else if file_dim.Width > file_dim.Height && file_dim.Width > max_dim {
        selected.Width = max_dim
    }

    newImage, err := bimg.NewImage(file_buffer).Process(selected)
    Err_check(err)
    
    imageReader := bytes.NewReader(newImage)
    img, _, err := image.Decode(imageReader)
    Err_check(err)
    
    hash, err := gih.PerceptionHash(img)
    Err_check(err)

    bimg.Write(file_path + file_pre + "s.webp", newImage)

    return file_dim.Width, file_dim.Height, hash.ToString(), nil
}
