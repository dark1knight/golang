package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	acronyms := map[string]string{
		"lol": "laugh out loud",
		"dw":  "don't worry",
		"hf":  "have fun",
		"gg":  "good game",
		"brb": "be right back",
		"g2g": "got to go",
		"wtf": "what the fuck",
		"wp":  "well played",
		"gl":  "good luck",
		"imo": "in my opinion",
	}

	var reader = bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if nil != err {
		panic("unexpected error!")
	}

	tokens := strings.Split(line, " ")
	for i := 0; i < len(tokens); i++ {
		value, isPresent := acronyms[tokens[i]]
		if !isPresent {
			fmt.Printf("%s ", tokens[i])
			continue
		}
		fmt.Printf("%s ", value)
	}

}
