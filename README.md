# itertools

[![tag](https://img.shields.io/github/tag/ertaquo/itertools.svg)](https://github.com/ertaquo/itertools/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.27-%23007d9c)
[![GoDoc](https://pkg.go.dev/badge/github.com/ertaquo/itertools.svg)](https://pkg.go.dev/github.com/ertaquo/itertools)
![Build Status](https://github.com/ertaquo/itertools/actions/workflows/go.yml/badge.svg)
[![License](https://img.shields.io/github/license/ertaquo/itertools)](./LICENSE)

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
Prefer `OrderedIterator` when the value type satisfies `cmp.Ordered`.

Operations run sequentially by default. Pass `WithParallel()` to run callbacks
concurrently, or `WithParallelLimit(n)` to limit concurrency. Parallel results
and callback order are unspecified.

See the [API reference](https://pkg.go.dev/github.com/ertaquo/itertools) for examples.

## AI assistance

AI tools were used to help write the documentation and tests.

## License

[MIT](LICENSE).
