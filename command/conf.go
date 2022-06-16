package main

//loads configuation information
import (
	"os"

	toml "github.com/BurntSushi/toml"
)

var BP string

type Config struct {
	Base_path	string
}

func Load_conf() {
	var conf Config 
	
	tomlData, err := os.ReadFile("/etc/ogai.toml")
	Err_check(err)

	_, err = toml.Decode(string(tomlData), &conf)
	Err_check(err)
	BP = conf.Base_path
}