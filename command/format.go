package main 

import (
    "strings"
    "bufio"
    "regexp"
)

var nlreg = regexp.MustCompile("\n")
var tagreg = regexp.MustCompile("(br>)(<)")

var repreg = regexp.MustCompile(`&gt;&gt;(\d+)\b`)
var randreg = regexp.MustCompile(`p\$1`)
var hashreg = regexp.MustCompile(`#/2/3.html`)
var prevreg = regexp.MustCompile(`#board`)
var quoreg = regexp.MustCompile(`&gt;(.+)`)

var spoilreg = regexp.MustCompile(`~~([^<]+)~~`)
var boldreg = regexp.MustCompile(`\*\*([^<])\*\*`)
var italicreg = regexp.MustCompile(`__([^<]+)__`)
var linkreg = regexp.MustCompile(`(http|ftp|https):\/\/(\S+)`)

const (    
    nlpost = "\n<br>"
    tagpost = "$1\n$2"
    reppost = `<ref hx-get="/im/ret/?p=$1&board=#board" hx-trigger="mouseover once" hx-target="#p$1"><a href="#/2/3.html#no$1">&#62;&#62;$1</a></ref><box id="p$1" class="prev"></box>`
)

var reprandpost = reppost
var rand_gen string
const (
    quopost = `<quo>&#62;$1</quo>`
    spoilpost = `<spoil>$1</spoil>`
    boldpost = `<b>$1</b>`
    italicpost = `<i>$1</i>`
    linkpost = `<a href="$1://$2">$1://$2</a>`
)


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

func process(rawline, board string) (string, []string) {

    stmts := Checkout()
    defer Checkin(stmts)
    stmt := stmts["prev_parent"]

    repmatches := make([]string, 0)
    repparents := make([]string, 0)

    repmatchcon := repreg.FindAllStringSubmatch(rawline, -1) 
    if repmatchcon != nil {
        for _, match := range repmatchcon {
            repmatches = append(repmatches, match[1])

            var parent string
            err := stmt.QueryRow(match[1], board).Scan(&parent)
            Query_err_check(err)

            repparents = append(repparents, parent)
        }
    }

    postline := repreg.ReplaceAllString(rawline, reprandpost)

    i := 0
    postline = hashreg.ReplaceAllStringFunc(postline, func(match string) string {
        cparent := repparents[i]
        response := `/` + board + `/` + cparent + `.html`
        i++
        return response  
    })
    postline = prevreg.ReplaceAllString(postline, board)

    postline = quoreg.ReplaceAllString(postline, quopost)
    postline = spoilreg.ReplaceAllString(postline, spoilpost)
    postline = boldreg.ReplaceAllString(postline, boldpost)
    postline = italicreg.ReplaceAllString(postline, italicpost)
    postline = linkreg.ReplaceAllString(postline, linkpost)

    return postline, repmatches  
}

func Format_post(input, board string) (string, []string) {

    scanner := bufio.NewScanner(strings.NewReader(input))
    scanner.Scan()

    //flexible statement
    rand_gen = Rand_gen()

    reprandpost = randreg.ReplaceAllString(reppost, `p$$1-` + rand_gen)

    output, repmatches := process(scanner.Text(), board)

    for scanner.Scan() {
        output = output + "\n"
        coutput, crepmatches := process(scanner.Text(), board)     
        output = output + coutput
        repmatches = append(repmatches, crepmatches...)
    }

    repmatches = removeDuplicates(repmatches)

    output = nlreg.ReplaceAllString(output, nlpost)
    output = tagreg.ReplaceAllString(output, tagpost)

    return output, repmatches
}
