package itertools_test

import (
	"fmt"

	"github.com/ertaquo/itertools"
)

func ExampleWithLimit() {
	values := itertools.ToIterator([]int{1, 2, 3})
	// Run at most two callbacks at a time.
	err := values.ForEachParallel(func(value int) error {
		fmt.Println(value * value)
		return nil
	}, itertools.WithLimit(2))
	fmt.Println(err)
	// Unordered output:
	// 1
	// 4
	// 9
	// <nil>
}
