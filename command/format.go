package main 

import (
    "strings"
    "bufio"
    "regexp"
)

var nlreg = regexp.MustCompile("\n")
var tagreg = regexp.MustCompile("(br>)(<)")

var repreg = regexp.MustCompile(`(?i)&gt;&gt;(/(\D+)/)?(\d+)\b`)
var hashreg = regexp.MustCompile(`#/2/3.html`)
var prevreg = regexp.MustCompile(`#board`)
var quoreg = regexp.MustCompile(`&gt;(\S.+)`)

var spoilreg = regexp.MustCompile(`~~([^<]+)~~`)
var boldreg = regexp.MustCompile(`\*\*([^<])\*\*`)
var italicreg = regexp.MustCompile(`__([^<]+)__`)
var linkreg = regexp.MustCompile(`([^>"]|\A)(http|ftp|https):\/\/(\S+)`)

var vidreg *regexp.Regexp

const (    
    nlpost = "\n<br>"
    tagpost = "$1\n$2"
    reppost = `<a class="preview" prev-get="/im/ret/?p=$3&board=#board" href="#/2/3.html#no$3">&#62;&#62;$1$3</a>`
)
var vidpost string

const (
    quopost = `<quo>&#62;$1</quo>`
    spoilpost = `<spoil>$1</spoil>`
    boldpost = `<b>$1</b>`
    italicpost = `<i>$1</i>`
    linkpost = `$1<a href="$2://$3" rel="noopener noreferrer nofollow">$2://$3</a>`
)

func Conf_dependent() {
    vidreg = regexp.MustCompile(`(https:\/\/|https:\/\/www\.)` +
                                `(youtube.com\/watch\?v=|youtu.be\/|` + INV_INST + `\/watch\?v=)(\S+)`)

    vidpost = `<details><summary>$1$2$3 <a href="$1$2$3" rel="noopener noreferrer nofollow">[link]</a></summary><iframe src="https://` +
    INV_INST + `/embed/$3?autoplay=0" allowfullscreen="" width="560" height="315" frameborder="0" loading="lazy" sandbox></iframe>` +
    `</details>` 
}


func removeDuplicates(strSlice []string) []string {
    allKeys := make(map[string]bool)
    list := []string{}
    for _, item := range strSlice {
        if _, value := allKeys[item]; !value {
            allKeys[item] = true
            list = append(list, item)
        }
    }
    return list
}

func process(rawline, board, orig_parent string) (string, []string) {

    stmts := Checkout()
    defer Checkin(stmts)
    stmt := stmts["prev_parent"]

    repmatches := make([]string, 0)
    repparents := make([]string, 0)
    repboards := make([]string, 0)

    repmatchcon := repreg.FindAllStringSubmatch(rawline, -1) 
    if repmatchcon != nil {
        for _, match := range repmatchcon {

            repmatches = append(repmatches, match[3])

            var sboard string

            if match[2] == "" {
                sboard = board
                repboards = append(repboards, board)
            } else {
                sboard = match[2]
                repboards = append(repboards, match[2])
            }

            var parent string
            err := stmt.QueryRow(match[3], sboard).Scan(&parent)
            Query_err_check(err)

            if parent == "" {parent = orig_parent}

            repparents = append(repparents, parent)
        }
    }

    postline := repreg.ReplaceAllString(rawline, reppost)

    rpi := 0
    postline = hashreg.ReplaceAllStringFunc(postline, func(match string) string {
        cboard := repboards[rpi]
        cparent := repparents[rpi]
        response := `/` + cboard + `/` + cparent + `.html`
        rpi++
        return response  
    })
    
    rbi := 0
    postline = prevreg.ReplaceAllStringFunc(postline, func(match string) string {
        cboard :=  repboards[rbi]
        rbi++
        return cboard
    })

    postline = spoilreg.ReplaceAllString(postline, spoilpost)
    postline = boldreg.ReplaceAllString(postline, boldpost)
    postline = italicreg.ReplaceAllString(postline, italicpost)
    postline = vidreg.ReplaceAllString(postline, vidpost)
    postline = linkreg.ReplaceAllString(postline, linkpost)
    postline = quoreg.ReplaceAllString(postline, quopost)

    return postline, repmatches  
}

func Format_post(input, board, orig_parent string) (string, []string) {

    scanner := bufio.NewScanner(strings.NewReader(input))
    scanner.Scan()

    output, repmatches := process(scanner.Text(), board, orig_parent)

    for scanner.Scan() {
        output += "\n"
        coutput, crepmatches := process(scanner.Text(), board, orig_parent)     
        output += coutput
        repmatches = append(repmatches, crepmatches...)
    }

    repmatches = removeDuplicates(repmatches)

    output = nlreg.ReplaceAllString(output, nlpost)
    output = tagreg.ReplaceAllString(output, tagpost)

    return output, repmatches
}

func hprocess(rawline string) string {
    postline := spoilreg.ReplaceAllString(rawline, "~~SPOILER~~")
    postline = boldreg.ReplaceAllString(postline, `$1`)
    postline = italicreg.ReplaceAllString(postline, `$1`)
    return postline
}

func HProcess_post(input string) (string, string) {
    scanner := bufio.NewScanner(strings.NewReader(input))
    scanner.Scan()

    output := hprocess(scanner.Text())

    for scanner.Scan() {
        output += "\n"
        coutput := hprocess(scanner.Text())
        output += coutput
    }

    //truncate output
    ofpost := []rune(output)
    var trunoutput string
    plen := len(ofpost)
    
    if plen > 70 {
        plen = 70 
        trunoutput = string(ofpost[:plen])
        trunoutput += "..."
    } else {
        trunoutput = string(ofpost)
    }
    trunoutput = nlreg.ReplaceAllString(trunoutput, " ")

    return output, trunoutput 
} 
