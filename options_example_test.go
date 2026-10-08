package itertools_test

import (
	"fmt"

	"github.com/ertaquo/itertools"
)

func ExampleWithParallel() {
	values := itertools.ToIterator([]int{1, 2, 3})
	for _, square := range values.MapAndCollect(func(value int) int {
		return value * value
	}, itertools.WithParallel()) {
		fmt.Println(square)
	}
	// Unordered output:
	// 1
	// 4
	// 9
}

func ExampleWithParallelLimit() {
	values := itertools.ToIterator([]int{1, 2, 3})
	// Run at most two callbacks at a time.
	err := values.ForEach(func(value int) error {
		fmt.Println(value * value)
		return nil
	}, itertools.WithParallelLimit(2))
	fmt.Println(err)
	// Unordered output:
	// 1
	// 4
	// 9
	// <nil>
}
