package main

import (
	"log"
	"strconv"

	"github.com/Necroforger/dgrouter/exrouter"
)

func Add(ctx *exrouter.Context) {
	total := 0
	for _, num := range ctx.Args.After(1) {
		n, err := strconv.Atoi(num)
		if err != nil {
			log.Printf("error casting string to int: %v", err)
		}
		total += n
	}

	ctx.Reply(strconv.Itoa(total))
}
