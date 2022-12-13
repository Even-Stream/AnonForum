package main

//loads configuation information
import (
    "log"

    ini "gopkg.in/ini.v1"
)

var BP string
var boards []*ini.Key
var Board_names []string
var Board_descs []string
var Board_map map[string]string

func Load_conf() {
    cfg, err := ini.Load("/etc/ogai.ini")
    Err_check(err)

    BP = cfg.Section("").Key("base path").String()

    Board_map = cfg.Section("boards").KeysHash()
    boards = cfg.Section("boards").Keys()

    for _, key := range boards {
        Board_names = append(Board_names, key.Name())
        Board_descs = append(Board_descs, key.Value())
    }

    if len(boards) == 0 {
        log.Fatal("Configuration error: No visible boards.")
    }
    if len(Board_names) != len(Board_descs) {
        log.Fatal("Configuration error: Not all boards have a description")
   }
}
