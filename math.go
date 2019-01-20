package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Necroforger/dgrouter/exrouter"
)

func Add(ctx *exrouter.Context) {
	tokens := strings.Split(ctx.Msg.Content, " ")
	nums := tokens[1:]

	total := 0
	for _, num := range nums {
		n, err := strconv.Atoi(num)
		if err != nil {
			fmt.Println("error casting string to int,", err)
		}
		total += n
	}

	ctx.Reply(strconv.Itoa(total))
}
