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
var quoreg = regexp.MustCompile(`&gt;(.+)`)
var spoilreg = regexp.MustCompile(`~~(.+)~~`)
var boldreg = regexp.MustCompile(`\*\*(.+)\*\*`)
var italicreg = regexp.MustCompile(`__(.+)__`)
var linkreg = regexp.MustCompile(`(http|ftp|https):\/\/(\S+)`)

const (	
	nlpost = "\n<br>"
	tagpost = "$1\n$2"
	reppost = `<ref hx-get="/im/ret/?p=$1" hx-trigger="mouseover once" hx-target="#p$1"><a href="#no$1">&#62;&#62;$1</a></ref><box id="p$1" class="prev"></box>`
)

var reprandpost = reppost
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

func process(rawline string) (string, []string) {

	repmatches := make([]string, 1)
	repmatchcon := repreg.FindAllStringSubmatch(rawline, -1) 
	if repmatchcon != nil {
		for _, match := range repmatchcon {
			repmatches = append(repmatches, match[1])
		}
	}
	
	postline := repreg.ReplaceAllString(rawline, reprandpost)
	postline = quoreg.ReplaceAllString(postline, quopost)
	postline = spoilreg.ReplaceAllString(postline, spoilpost)
	postline = boldreg.ReplaceAllString(postline, boldpost)
	postline = italicreg.ReplaceAllString(postline, italicpost)
	postline = linkreg.ReplaceAllString(postline, linkpost)

	return postline, repmatches  
}

func Format_post(input string) (string, []string) {

	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Scan()

	reprandpost = randreg.ReplaceAllString(reppost, `p$$1-` + Rand_gen())

	output, repmatches := process(scanner.Text())
	
	for scanner.Scan() {
		output = output + "\n"
		coutput, crepmatches := process(scanner.Text())	 
		output = output + coutput
		repmatches = append(repmatches, crepmatches...)
	}

	repmatches = removeDuplicates(repmatches)
	
	output = nlreg.ReplaceAllString(output, nlpost)
	output = tagreg.ReplaceAllString(output, tagpost)
	
	return output, repmatches
}