# itertools

Generic, chainable iterators for Go slices, maps, `iter.Seq`, and `iter.Seq2`.
Filter, transform, sort, and collect values, or iterate directly with `for range`.

## Install

Requires Go 1.27 or later.

```sh
go get github.com/ertaquo/itertools
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/ertaquo/itertools"
)

func main() {
	evenSquares := itertools.
		ToIterator([]int{1, 2, 3, 4, 5}).
		Filter(func(n int) bool {
			return n%2 == 0
		}).
		Map(func(n int) int {
			return n * n
		})
	fmt.Println(evenSquares.Collect()) // [4 16]
}
```

Use `Iterator` for any values, `ComparableIterator` for equality and uniqueness,
`OrderedIterator` for ordered values, and `MapIterator` for key/value pairs.
Parallel operations accept `WithLimit(n)` to limit concurrency; their result
order may be unspecified.

See the [API reference](https://pkg.go.dev/github.com/ertaquo/itertools) for examples.

## AI assistance

AI tools were used to help write the documentation and tests.

## License

[MIT](LICENSE).
