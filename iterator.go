package itertools

import (
	"iter"
	"reflect"
	"slices"
	"sync"

	"golang.org/x/exp/constraints"
	"golang.org/x/sync/errgroup"
)

type Iterator[V any] func(yield func(V) bool)

func ToIterator[Slice ~[]V, V any](s Slice) Iterator[V] {
	return Iterator[V](slices.Values(s))
}

func SeqToIterator[V any](s iter.Seq[V]) Iterator[V] {
	return Iterator[V](s)
}

func (i Iterator[V]) All(matchFunc func(v V) bool) bool {
	for v := range i {
		if !matchFunc(v) {
			return false
		}
	}

	return true
}

func (i Iterator[V]) Any(matchFunc func(v V) bool) bool {
	return i.ContainsFunc(matchFunc)
}

func (i Iterator[V]) Collect() []V {
	return slices.Collect(iter.Seq[V](i))
}

func (i Iterator[V]) CollectAs[Slice ~[]V]() Slice {
	return slices.Collect(iter.Seq[V](i))
}

func (i Iterator[V]) CollectSorted(compareFunc func(a, b V) int) []V {
	return slices.SortedFunc(iter.Seq[V](i), compareFunc)
}

func (i Iterator[V]) CollectSortedAs[Slice ~[]V](compareFunc func(a, b V) int) Slice {
	return slices.SortedFunc(iter.Seq[V](i), compareFunc)
}

func (i Iterator[V]) CollectSortedStable(compareFunc func(a, b V) int) []V {
	return slices.SortedStableFunc(iter.Seq[V](i), compareFunc)
}

func (i Iterator[V]) CollectSortedStableAs[Slice ~[]V](compareFunc func(a, b V) int) Slice {
	return slices.SortedStableFunc(iter.Seq[V](i), compareFunc)
}

func (i Iterator[V]) Contains(value V) bool {
	rv := reflect.ValueOf(value)
	if !rv.Type().Comparable() {
		panic("iterator type " + rv.Type().String() + " is not comparable")
	}

	for v := range i {
		if reflect.ValueOf(v).Equal(rv) {
			return true
		}
	}

	return false
}

func (i Iterator[V]) ContainsFunc(matchFunc func(v V) bool) bool {
	for v := range i {
		if matchFunc(v) {
			return true
		}
	}

	return false
}

func (i Iterator[V]) Count() int {
	var count int
	for range i {
		count++
	}
	return count
}

func (i Iterator[V]) Equal(another Iterator[V], compareFunc func(a, b V) int) bool {
	p1, p1stop := iter.Pull(iter.Seq[V](i))
	p2, p2stop := iter.Pull(iter.Seq[V](another))
	defer p1stop()
	defer p2stop()

	for {
		v1, ok1 := p1()
		v2, ok2 := p2()

		if !ok1 && !ok2 {
			return true
		}

		if !ok1 || !ok2 {
			return false
		}

		if c := compareFunc(v1, v2); c != 0 {
			return false
		}
	}
}

func (i Iterator[V]) Filter(filterFunc func(value V) bool) Iterator[V] {
	return func(yield func(V) bool) {
		for v := range i {
			if !filterFunc(v) {
				continue
			}

			if !yield(v) {
				return
			}
		}
	}
}

func (i Iterator[V]) FilterAndCollect(filterFunc func(value V) bool) []V {
	var result []V
	for v := range i {
		if !filterFunc(v) {
			continue
		}
		result = append(result, v)
	}
	return result
}

func (i Iterator[V]) FilterAndCollectWithError(filterFunc func(value V) (bool, error)) ([]V, error) {
	var result []V
	for v := range i {
		ok, err := filterFunc(v)
		if err != nil {
			return result, err
		}
		if !ok {
			continue
		}
		result = append(result, v)
	}
	return result, nil
}

func (i Iterator[V]) FilterAndCollectParallel(filterFunc func(value V) bool, parallelOptions ...ParallelOption) []V {
	var wg errgroup.Group
	var result []V
	var resultMutex sync.Mutex

	for _, option := range parallelOptions {
		option(&wg)
	}

	for v := range i {
		wg.Go(func() error {
			if !filterFunc(v) {
				return nil
			}

			resultMutex.Lock()
			defer resultMutex.Unlock()

			result = append(result, v)
			return nil
		})
	}

	wg.Wait()
	return result
}

func (i Iterator[V]) FilterAndCollectWithErrorParallel(filterFunc func(value V) (bool, error), parallelOptions ...ParallelOption) ([]V, error) {
	var wg errgroup.Group
	var result []V
	var resultMutex sync.Mutex

	for _, option := range parallelOptions {
		option(&wg)
	}

	for v := range i {
		wg.Go(func() error {
			ok, err := filterFunc(v)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			resultMutex.Lock()
			defer resultMutex.Unlock()

			result = append(result, v)
			return nil
		})
	}

	err := wg.Wait()
	return result, err
}

func (i Iterator[V]) Find(matchFunc func(v V) bool) *V {
	for v := range i {
		if matchFunc(v) {
			return &v
		}
	}

	return nil
}

func (i Iterator[V]) FindOr(matchFunc func(v V) bool, defaultValue V) V {
	if v := i.Find(matchFunc); v != nil {
		return *v
	}

	return defaultValue
}

func (i Iterator[V]) FindOrNone(matchFunc func(v V) bool) V {
	if v := i.Find(matchFunc); v != nil {
		return *v
	}

	var none V
	return none
}

func (i Iterator[V]) First() *V {
	for v := range i {
		return &v
	}

	return nil
}

func (i Iterator[V]) FirstOr(defaultValue V) V {
	if v := i.First(); v != nil {
		return *v
	}

	return defaultValue
}

func (i Iterator[V]) FirstOrNone() V {
	if v := i.First(); v != nil {
		return *v
	}

	var none V
	return none
}

func (i Iterator[V]) ForEach(callback func(v V) error) error {
	for v := range i {
		if err := callback(v); err != nil {
			return err
		}
	}

	return nil
}

func (i Iterator[V]) ForEachParallel(callback func(v V) error, parallelOptions ...ParallelOption) error {
	var wg errgroup.Group

	for _, option := range parallelOptions {
		option(&wg)
	}

	for v := range i {
		wg.Go(func() error {
			return callback(v)
		})
	}

	return wg.Wait()
}

func (i Iterator[V]) Get[N constraints.Signed | constraints.Unsigned](n N) *V {
	if n < 0 {
		return nil
	}

	for v := range i {
		if n == 0 {
			return &v
		}

		n--
	}

	return nil
}

func (i Iterator[V]) GetOr[N constraints.Signed | constraints.Unsigned](n N, defaultValue V) V {
	if v := i.Get(n); v != nil {
		return *v
	}

	return defaultValue
}

func (i Iterator[V]) GetOrNone[N constraints.Signed | constraints.Unsigned](n N) V {
	if v := i.Get(n); v != nil {
		return *v
	}

	var none V
	return none
}

func (i Iterator[V]) Concat(another Iterator[V]) Iterator[V] {
	return func(yield func(V) bool) {
		for v := range i {
			if !yield(v) {
				return
			}
		}

		for v := range another {
			if !yield(v) {
				return
			}
		}
	}
}

func (i Iterator[V]) Last() *V {
	var last *V
	for v := range i {
		last = &v
	}

	return last
}

func (i Iterator[V]) LastOr(defaultValue V) V {
	if v := i.Last(); v != nil {
		return *v
	}

	return defaultValue
}

func (i Iterator[V]) LastOrNone() V {
	if v := i.Last(); v != nil {
		return *v
	}

	var none V
	return none
}

func (i Iterator[V]) Limit[N constraints.Signed | constraints.Unsigned](n N) Iterator[V] {
	return func(yield func(V) bool) {
		if n <= 0 {
			return
		}

		for v := range i {
			if !yield(v) {
				return
			}

			n--
			if n <= 0 {
				return
			}
		}
	}
}

func (i Iterator[V]) Map[V2 any](mapFunc func(value V) V2) Iterator[V2] {
	return func(yield func(V2) bool) {
		for v := range i {
			if !yield(mapFunc(v)) {
				return
			}
		}
	}
}

func (i Iterator[V]) MapAndCollect[V2 any](mapFunc func(value V) V2) []V2 {
	var result []V2
	for v := range i {
		result = append(result, mapFunc(v))
	}
	return result
}

func (i Iterator[V]) MapAndCollectParallel[V2 any](mapFunc func(value V) V2, parallelOptions ...ParallelOption) []V2 {
	var wg errgroup.Group
	var result []V2
	var resultMutex sync.Mutex

	for _, option := range parallelOptions {
		option(&wg)
	}

	for v := range i {
		wg.Go(func() error {
			v2 := mapFunc(v)

			resultMutex.Lock()
			defer resultMutex.Unlock()
			result = append(result, v2)
			return nil
		})
	}

	wg.Wait()
	return result
}

func (i Iterator[V]) MapAndCollectWithError[V2 any](mapFunc func(value V) (V2, error)) ([]V2, error) {
	var result []V2
	for v := range i {
		v2, err := mapFunc(v)
		if err != nil {
			return result, err
		}
		result = append(result, v2)
	}
	return result, nil
}

func (i Iterator[V]) MapAndCollectWithErrorParallel[V2 any](mapFunc func(value V) (V2, error), parallelOptions ...ParallelOption) ([]V2, error) {
	var wg errgroup.Group
	var result []V2
	var resultMutex sync.Mutex

	for _, option := range parallelOptions {
		option(&wg)
	}

	for v := range i {
		wg.Go(func() error {
			v2, err := mapFunc(v)
			if err != nil {
				return err
			}

			resultMutex.Lock()
			defer resultMutex.Unlock()
			result = append(result, v2)
			return nil
		})
	}

	err := wg.Wait()
	return result, err
}

func (i Iterator[V]) Max(compareFunc func(a, b V) int) *V {
	var max *V
	for v := range i {
		if max == nil {
			max = &v
		} else if compareFunc(v, *max) > 0 {
			max = &v
		}
	}
	return max
}

func (i Iterator[V]) MaxOr(compareFunc func(a, b V) int, defaultValue V) V {
	if max := i.Max(compareFunc); max != nil {
		return *max
	}
	return defaultValue
}

func (i Iterator[V]) MaxOrNone(compareFunc func(a, b V) int) V {
	if max := i.Max(compareFunc); max != nil {
		return *max
	}

	var none V
	return none
}

func (i Iterator[V]) Min(compareFunc func(a, b V) int) *V {
	var min *V
	for v := range i {
		if min == nil {
			min = &v
		} else if compareFunc(v, *min) < 0 {
			min = &v
		}
	}
	return min
}

func (i Iterator[V]) MinMax(compareFunc func(a, b V) int) (min *V, max *V) {
	for v := range i {
		if min == nil {
			min = &v
			max = &v
		} else {
			if compareFunc(v, *min) < 0 {
				min = &v
			}
			if compareFunc(v, *max) > 0 {
				max = &v
			}
		}
	}
	return min, max
}

func (i Iterator[V]) MinMaxOr(compareFunc func(a, b V) int, defaultValue V) (min V, max V) {
	if minP, maxP := i.MinMax(compareFunc); minP != nil && maxP != nil {
		return *minP, *maxP
	}

	return defaultValue, defaultValue
}

func (i Iterator[V]) MinMaxOrNone(compareFunc func(a, b V) int) (min V, max V) {
	if minP, maxP := i.MinMax(compareFunc); minP != nil && maxP != nil {
		return *minP, *maxP
	}

	var none V
	return none, none
}

func (i Iterator[V]) MinOr(compareFunc func(a, b V) int, defaultValue V) V {
	if min := i.Min(compareFunc); min != nil {
		return *min
	}

	return defaultValue
}

func (i Iterator[V]) MinOrNone(compareFunc func(a, b V) int) V {
	if min := i.Min(compareFunc); min != nil {
		return *min
	}

	var none V
	return none
}

func (i Iterator[V]) Pull() (next func() (V, bool), stop func()) {
	return iter.Pull(iter.Seq[V](i))
}

func (i Iterator[V]) Reduce(startValue V, accumulator func(v V) V) V {
	curr := startValue
	for v := range i {
		curr = accumulator(v)
	}
	return curr
}

func (i Iterator[V]) Reverse() Iterator[V] {
	s := i.Collect()
	slices.Reverse(s)
	return ToIterator(s)
}

func (i Iterator[V]) Skip[N constraints.Signed | constraints.Unsigned](n N) Iterator[V] {
	return func(yield func(V) bool) {
		for v := range i {
			if n > 0 {
				n--
				continue
			}

			if !yield(v) {
				return
			}
		}
	}
}

func (i Iterator[V]) SkipWhile(matchFunc func(v V) bool) Iterator[V] {
	return func(yield func(V) bool) {
		wasSkipped := false

		for v := range i {
			if !wasSkipped && matchFunc(v) {
				continue
			}

			wasSkipped = true

			if !yield(v) {
				return
			}
		}
	}
}

func (i Iterator[V]) Sorted(compareFunc func(a, b V) int) Iterator[V] {
	return ToIterator(i.CollectSorted(compareFunc))
}

func (i Iterator[V]) SortedStable(compareFunc func(a, b V) int) Iterator[V] {
	return ToIterator(i.CollectSortedStable(compareFunc))
}

func (i Iterator[V]) TakeWhile(matchFunc func(v V) bool) Iterator[V] {
	return func(yield func(V) bool) {
		for v := range i {
			if !matchFunc(v) {
				return
			}

			if !yield(v) {
				return
			}
		}
	}
}

func (i Iterator[V]) ToSeq() iter.Seq[V] {
	return iter.Seq[V](i)
}

func (i Iterator[V]) Unique(compareFunc func(a, b V) int) Iterator[V] {
	var past []V
	return func(yield func(V) bool) {
		for v := range i {
			metBefore := false

			for _, p := range past {
				if compareFunc(p, v) == 0 {
					metBefore = true
					break
				}
			}

			if !metBefore {
				past = append(past, v)

				if !yield(v) {
					return
				}
			}
		}
	}
}

func (i Iterator[V]) Zip[K comparable](keys Iterator[K]) MapIterator[K, V] {
	values := i.Collect()
	return func(yield func(K, V) bool) {
		if len(values) == 0 {
			return
		}

		for k := range keys {
			if !yield(k, values[0]) {
				return
			}

			values = values[1:]
			if len(values) == 0 {
				return
			}
		}
	}
}
