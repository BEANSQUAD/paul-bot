package math

import (
	"fmt"
	"strconv"
)

func Add(nums []string) int {
	total := 0
	for _, num := range nums {
		n, err := strconv.Atoi(num)
		if err != nil {
			fmt.Println("error casting string to int,", err)
		}
		total += n
	}
	return total
}