package main

//loads configuation information
import (
    "os"
    "log"    

    toml "github.com/BurntSushi/toml"
)

var BP string
var Boards []string
var Descs []string

type Config struct {
    Base_path    string
    Boards        []string
    Descs        []string
}

func Load_conf() {
    var conf Config 

    tomlData, err := os.ReadFile("/etc/ogai.toml")
    Err_check(err)

    _, err = toml.Decode(string(tomlData), &conf)
    Err_check(err)
    BP = conf.Base_path
    Boards = conf.Boards
    Descs = conf.Descs

    if len(Boards) == 0 {
        log.Fatal("Configuration error: No visible boards.")
    }

    if len(Boards) != len(Descs) {
        log.Fatal("Configuration error: # of boards and descriptions do not match.")
    }
}
