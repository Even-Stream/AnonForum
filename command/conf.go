package main

//loads configuation information
import (
    "log"
    "os"

    ini "gopkg.in/ini.v1"
)

var BP string
var boards []*ini.Key
var Board_names []string
var Board_descs []string
var Board_map map[string]string
var Themes []string
var INV_INST string

func Load_conf() {
    homedir, err := os.UserHomeDir()
    Err_check(err)
    
    cfg, err := ini.Load(homedir +  "/.config/ogai.ini")
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

   Themes = cfg.Section("misc").Key("themes").Strings(" ")
   INV_INST = cfg.Section("misc").Key("invinst").String()
   Conf_dependent()
}
