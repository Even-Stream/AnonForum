package main

//loads configuation information
import (
    "log"
	"regexp"

    ini "gopkg.in/ini.v1"
)

var SiteName string
var BP string
var Max_request_size int64
var boards []*ini.Key
var Board_names []string
var Board_descs []string
var Board_map map[string]string
var HBoard_map = make(map[string]bool)
var Themes []string
var INV_INST string
var Word_filter = make(map[*regexp.Regexp]string)
var Forbidden = make(map[string]bool)
var Auto_phrases []string

func Load_conf() { 
    cfg, err := ini.LoadSources(
        ini.LoadOptions{AllowBooleanKeys: true,}, "af.ini")
    Err_check(err)

    SiteName = cfg.Section("").Key("site name").String()
    BP = cfg.Section("").Key("base path").String()
    Max_request_size, err = cfg.Section("").Key("max request size").Int64()
    Err_check(err)
    Max_upload_size = 1024 * 1024 * Max_request_size

	for word, replacement := range cfg.Section("filter").KeysHash() {
	    Word_filter[regexp.MustCompile(`(?i)` + word)] = replacement
	}

    Board_map = cfg.Section("boards").KeysHash()
    boards = cfg.Section("boards").Keys()

    for _, key := range boards {
        if !cfg.Section("hidden").HasKey(key.Name()) {
            Board_names = append(Board_names, key.Name())
            Board_descs = append(Board_descs, key.Value())
        } else {
            HBoard_map[key.Name()] = true}
    }

    if len(boards) == 0 {
        log.Fatal("Configuration error: No visible boards.")
    }
    if len(Board_names) != len(Board_descs) {
        log.Fatal("Configuration error: Not all boards have a description")
   }
   
   fhashes := cfg.Section("forbidden").KeyStrings()
   for _, h := range fhashes {
       Forbidden[`p:` + h] = true
   }

   Themes = cfg.Section("misc").Key("themes").Strings(" ")
   INV_INST = cfg.Section("misc").Key("invinst").String()

   Auto_phrases = cfg.Section("auto delete").KeyStrings()
   
   Conf_dependent()
}
