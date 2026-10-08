package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ertaquo/itertools"
)

func main() {
	result, err := itertools.
		ToIterator([]int{1, 2, 3, 4, 5}).
		Filter(func(n int) bool {
			return n%2 == 0
		}).
		Map(func(n int) int {
			return n * n
		}).
		Map(func(n int) string {
			return strconv.Itoa(n)
		}).
		Concat(itertools.ToIterator([]string{"6", "seven", "8"})).
		MapAndCollectWithError(func(s string) (int, error) {
			time.Sleep(1 * time.Second)
			return strconv.Atoi(s)
		}, itertools.WithParallel())

	fmt.Println(result, err)
}
