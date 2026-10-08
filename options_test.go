package itertools_test

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/ertaquo/itertools"
)

type optionIterator struct {
	name    string
	forEach func(func(int) error, ...itertools.Option) error
}

func optionIterators() []optionIterator {
	values := []int{1, 2, 3}
	return []optionIterator{
		{"Iterator", itertools.ToIterator(values).ForEach},
		{"ComparableIterator", itertools.ToComparableIterator(values).ForEach},
		{"OrderedIterator", itertools.ToOrderedIterator(values).ForEach},
		{"MapIterator", func(callback func(int) error, options ...itertools.Option) error {
			return mapSeq(mapPair{"a", 1}, mapPair{"b", 2}, mapPair{"c", 3}).ForEach(func(_ string, value int) error {
				return callback(value)
			}, options...)
		}},
	}
}

func TestParallelOptions(t *testing.T) {
	for _, iterator := range optionIterators() {
		t.Run(iterator.name, func(t *testing.T) {
			for _, tt := range []struct {
				name    string
				options []itertools.Option
				want    int32
			}{
				{"default", nil, 1},
				{"parallel", []itertools.Option{itertools.WithParallel()}, 3},
				{"limit one", []itertools.Option{itertools.WithParallelLimit(1)}, 1},
				{"limit two", []itertools.Option{itertools.WithParallelLimit(2)}, 2},
				{"zero limit", []itertools.Option{itertools.WithParallelLimit(0)}, 3},
				{"negative limit", []itertools.Option{itertools.WithParallelLimit(-1)}, 3},
				{"keep limit", []itertools.Option{itertools.WithParallelLimit(1), itertools.WithParallel()}, 1},
				{"replace limit", []itertools.Option{itertools.WithParallelLimit(1), itertools.WithParallelLimit(2)}, 2},
				{"clear limit", []itertools.Option{itertools.WithParallelLimit(2), itertools.WithParallelLimit(0)}, 3},
			} {
				t.Run(tt.name, func(t *testing.T) {
					synctest.Test(t, func(t *testing.T) {
						var active, peak, completed atomic.Int32
						err := iterator.forEach(func(int) error {
							now := active.Add(1)
							for old := peak.Load(); now > old; old = peak.Load() {
								if peak.CompareAndSwap(old, now) {
									break
								}
							}
							// Virtual time advances when the callbacks are blocked.
							time.Sleep(time.Millisecond)
							active.Add(-1)
							completed.Add(1)
							return nil
						}, tt.options...)
						if err != nil {
							t.Fatal(err)
						}
						if got := peak.Load(); got != tt.want {
							t.Errorf("peak concurrent callbacks = %d, want %d", got, tt.want)
						}
						if active.Load() != 0 || completed.Load() != 3 {
							t.Errorf("after return active = %d, completed = %d, want 0, 3", active.Load(), completed.Load())
						}
					})
				})
			}
		})
	}
}

func TestParallelOptionsWithError(t *testing.T) {
	bad := errors.New("bad value")
	for _, iterator := range optionIterators() {
		t.Run(iterator.name, func(t *testing.T) {
			for _, tt := range []struct {
				name    string
				options []itertools.Option
				want    []int
			}{
				{"default", nil, []int{1, 2}},
				{"parallel", []itertools.Option{itertools.WithParallel()}, []int{1, 2, 3}},
				{"limited", []itertools.Option{itertools.WithParallelLimit(1)}, []int{1, 2, 3}},
			} {
				t.Run(tt.name, func(t *testing.T) {
					var mutex sync.Mutex
					var called []int
					err := iterator.forEach(func(value int) error {
						mutex.Lock()
						called = append(called, value)
						mutex.Unlock()
						if value == 2 {
							return bad
						}
						return nil
					}, tt.options...)
					if !errors.Is(err, bad) {
						t.Errorf("error = %v, want %v", err, bad)
					}
					slices.Sort(called)
					if !slices.Equal(called, tt.want) {
						t.Errorf("callback values = %v, want %v", called, tt.want)
					}
				})
			}
		})
	}
}
