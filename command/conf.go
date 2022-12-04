package main

//loads configuation information
import (
    "log"    

    ini "gopkg.in/ini.v1"
)

var BP string
var Boards = make(map[string]string)

func Load_conf() {
    cfg, err := ini.Load("/etc/ogai.ini")
    Err_check(err)

    BP = cfg.Section("").Key("base path").String()
    Boards = cfg.Section("boards").KeysHash()

    if len(Boards) == 0 {
        log.Fatal("Configuration error: No visible boards.")
    }
}
