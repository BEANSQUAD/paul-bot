package main

import (
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/Necroforger/dgrouter/exrouter"
)

// Add takes a variable amount of inputs from an in-client chat command and returns the sum of them in the same channel.
// The numbers are split by spaces, and because the function is int-based, it ignores anything that cannot be parsed as
// an int, such as letters or floats.
func Add(ctx *exrouter.Context) {
	total := 0
	for _, num := range ctx.Args[1:] {
		n, err := strconv.Atoi(num)
		if err != nil {
			log.Printf("error casting string to int: %v", err)
		}
		total += n
	}

	ctx.Reply(strconv.Itoa(total))
}
