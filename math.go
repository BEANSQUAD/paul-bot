package main

import (
	"fmt"
	"strconv"

	"github.com/Necroforger/dgrouter/exrouter"
)

func Add(ctx *exrouter.Context) {
	total := 0
	for _, num := range ctx.Args {
		n, err := strconv.Atoi(num)
		if err != nil {
			fmt.Println("error casting string to int,", err)
		}
		total += n
	}

	ctx.Reply(strconv.Itoa(total))
}
