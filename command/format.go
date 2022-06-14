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
	
var nlpost = "\n<br>"
var tagpost = "$1\n$2"

var reppost = `<ref hx-get="/im/ret/?p=$1" hx-trigger="mouseover once" hx-target="#p$1"><a href="#no$1">&#62;&#62;$1</a></ref><box id="p$1" class="prev"></box>`
var reprandpost = reppost
var quopost = `<quo>&#62;$1</quo>`
var spoilpost = `<spoil>$1</spoil>`
var boldpost = `<b>$1</b>`
var italicpost = `<i>$1</i>`
var linkpost = `<a href="$1://$2">$1://$2</a>`


func process(rawline string) string {

	postline := repreg.ReplaceAllString(rawline, reprandpost)
	postline = quoreg.ReplaceAllString(postline, quopost)
	postline = spoilreg.ReplaceAllString(postline, spoilpost)
	postline = boldreg.ReplaceAllString(postline, boldpost)
	postline = italicreg.ReplaceAllString(postline, italicpost)
	postline = linkreg.ReplaceAllString(postline, linkpost)

	return postline
}

func Format_post(input string) string {

	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Scan()

	reprandpost = randreg.ReplaceAllString(reppost, `p$$1-` + Rand_gen())

	output := process(scanner.Text())
	
	for scanner.Scan() {
		output = output + "\n" 
		output = output + process(scanner.Text())	
	}

	output = nlreg.ReplaceAllString(output, nlpost)
	output = tagreg.ReplaceAllString(output, tagpost)
	
	return output
}
