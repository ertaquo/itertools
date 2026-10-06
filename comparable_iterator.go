package itertools

import (
	"iter"
	"slices"
	"sync"

	"golang.org/x/exp/constraints"
	"golang.org/x/sync/errgroup"
)

// ComparableIterator yields comparable values of type V. It has the same
// signature as [iter.Seq] and can be used directly in a for range loop.
//
// Prefer [OrderedIterator] when V satisfies its type constraint.
type ComparableIterator[V comparable] func(yield func(V) bool)

// ToComparableIterator returns an iterator over the elements of s in slice order.
//
// Prefer [ToOrderedIterator] when V satisfies its type constraint.
func ToComparableIterator[Slice ~[]V, V comparable](s Slice) ComparableIterator[V] {
	return ComparableIterator[V](slices.Values(s))
}

// SeqToComparableIterator converts s to a ComparableIterator without collecting
// its values.
//
// Prefer [SeqToOrderedIterator] when V satisfies its type constraint.
func SeqToComparableIterator[V comparable](s iter.Seq[V]) ComparableIterator[V] {
	return ComparableIterator[V](s)
}

// Concat returns an iterator over the values of i followed by those of another.
func (i ComparableIterator[V]) Concat[A Iterator[V] | ComparableIterator[V]](another A) ComparableIterator[V] {
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

// Contains reports whether the iterator contains value. It stops at the first
// match.
func (i ComparableIterator[V]) Contains(value V) bool {
	for v := range i {
		if v == value {
			return true
		}
	}

	return false
}

// Equal reports whether two iterators yield equal values in the same order. It
// stops at the first difference.
func (i ComparableIterator[V]) Equal[A Iterator[V] | ComparableIterator[V]](another A) bool {
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

		if v1 != v2 {
			return false
		}
	}
}

// EqualFunc reports whether two iterators yield equal values in the same order
// according to compareFunc. It stops at the first difference. Use Equal when
// ordinary equality suffices; it avoids a comparison function.
func (i ComparableIterator[V]) EqualFunc[A Iterator[V] | ComparableIterator[V]](another A, compareFunc func(a, b V) int) bool {
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

// FilterNone returns an iterator over values that are not the zero value of V. It
// evaluates each value as it is requested.
func (i ComparableIterator[V]) FilterNone() ComparableIterator[V] {
	var none V
	return func(yield func(V) bool) {
		for v := range i {
			if v == none {
				continue
			}

			if !yield(v) {
				return
			}
		}
	}
}

// Unique returns an iterator that yields only the first occurrence of each value.
// It uses a map to track seen values and retains that state across traversals.
func (i ComparableIterator[V]) Unique() ComparableIterator[V] {
	past := make(map[V]struct{})

	return func(yield func(V) bool) {
		for v := range i {
			if _, exists := past[v]; exists {
				continue
			}
			past[v] = struct{}{}

			if !yield(v) {
				return
			}
		}
	}
}

// UniqueFunc returns an iterator that yields only the first value in each
// compareFunc equivalence class. It compares each new value with all previously
// yielded values, so it can be slow for large iterators. Use Unique when ordinary
// equality suffices; it uses a map. The returned iterator retains its seen values
// across traversals.
func (i ComparableIterator[V]) UniqueFunc(compareFunc func(a, b V) int) Iterator[V] {
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

// ToIterator returns i as an Iterator without collecting its values.
func (i ComparableIterator[V]) ToIterator() Iterator[V] {
	return Iterator[V](i)
}

// Zip returns an iterator pairing keys with the values of i until either side is
// exhausted. It collects all values of i before reading keys and consumes those
// values across traversals.
//
// Collecting the values first may be slow or use substantial memory for large
// inputs.
func (i ComparableIterator[V]) Zip[K comparable, Keys Iterator[K] | ComparableIterator[K]](keys Keys) MapIterator[K, V] {
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

// --- From Iterator[V] ---

// All reports whether matchFunc returns true for every value. It returns true for
// an empty iterator and stops at the first nonmatching value.
func (i ComparableIterator[V]) All(matchFunc func(v V) bool) bool {
	for v := range i {
		if !matchFunc(v) {
			return false
		}
	}

	return true
}

// Any reports whether matchFunc returns true for any value. It stops at the first
// match.
func (i ComparableIterator[V]) Any(matchFunc func(v V) bool) bool {
	return i.ContainsFunc(matchFunc)
}

// Collect returns the values in iteration order as a slice. The result is nil if
// the iterator is empty.
func (i ComparableIterator[V]) Collect() []V {
	return slices.Collect(iter.Seq[V](i))
}

// CollectAs returns the values in iteration order as a slice of type Slice. The
// result is nil if the iterator is empty.
func (i ComparableIterator[V]) CollectAs[Slice ~[]V]() Slice {
	return slices.Collect(iter.Seq[V](i))
}

// CollectSorted collects the values, sorts them using compareFunc, and returns
// the slice. The sort is not guaranteed to be stable. For ordered values,
// [OrderedIterator.CollectSorted] avoids a custom comparison function.
func (i ComparableIterator[V]) CollectSorted(compareFunc func(a, b V) int) []V {
	return slices.SortedFunc(iter.Seq[V](i), compareFunc)
}

// CollectSortedAs is like CollectSorted but returns a slice of type Slice.
//
// For ordered values, [OrderedIterator.CollectSortedAs] avoids a custom
// comparison function.
func (i ComparableIterator[V]) CollectSortedAs[Slice ~[]V](compareFunc func(a, b V) int) Slice {
	return slices.SortedFunc(iter.Seq[V](i), compareFunc)
}

// CollectSortedStable collects the values and sorts them using compareFunc while
// preserving the order of values that compare equal. For ordered values,
// [OrderedIterator.CollectSortedStable] avoids a custom comparison function.
func (i ComparableIterator[V]) CollectSortedStable(compareFunc func(a, b V) int) []V {
	return slices.SortedStableFunc(iter.Seq[V](i), compareFunc)
}

// CollectSortedStableAs is like CollectSortedStable but returns a slice of type
// Slice.
//
// For ordered values, [OrderedIterator.CollectSortedStableAs] avoids a custom
// comparison function.
func (i ComparableIterator[V]) CollectSortedStableAs[Slice ~[]V](compareFunc func(a, b V) int) Slice {
	return slices.SortedStableFunc(iter.Seq[V](i), compareFunc)
}

// ContainsFunc reports whether matchFunc returns true for any value. It stops at
// the first match.
func (i ComparableIterator[V]) ContainsFunc(matchFunc func(v V) bool) bool {
	for v := range i {
		if matchFunc(v) {
			return true
		}
	}

	return false
}

// Count returns the number of values yielded by the iterator.
func (i ComparableIterator[V]) Count() int {
	var count int
	for range i {
		count++
	}
	return count
}

// Filter returns an iterator over values for which filterFunc returns true. It
// evaluates filterFunc as values are requested.
func (i ComparableIterator[V]) Filter(filterFunc func(value V) bool) Iterator[V] {
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

// FilterAndCollect returns a slice of values for which filterFunc returns true,
// in iteration order. The result is nil if no values match.
func (i ComparableIterator[V]) FilterAndCollect(filterFunc func(value V) bool) []V {
	var result []V
	for v := range i {
		if !filterFunc(v) {
			continue
		}
		result = append(result, v)
	}
	return result
}

// FilterAndCollectWithError is like FilterAndCollect, but stops and returns the
// values collected so far if filterFunc returns an error.
func (i ComparableIterator[V]) FilterAndCollectWithError(filterFunc func(value V) (bool, error)) ([]V, error) {
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

// FilterAndCollectParallel filters values concurrently and returns the matching
// values. The result order is unspecified; parallelOptions control the worker
// group.
func (i ComparableIterator[V]) FilterAndCollectParallel(filterFunc func(value V) bool, parallelOptions ...ParallelOption) []V {
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

	_ = wg.Wait() //nolint:errcheck
	return result
}

// FilterAndCollectWithErrorParallel filters values concurrently. It waits for all workers,
// then returns the collected values and the first error, if any. The result
// order is unspecified.
func (i ComparableIterator[V]) FilterAndCollectWithErrorParallel(filterFunc func(value V) (bool, error), parallelOptions ...ParallelOption) ([]V, error) {
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

// Find returns a pointer to the first value matching matchFunc, or nil if none
// matches.
func (i ComparableIterator[V]) Find(matchFunc func(v V) bool) *V {
	for v := range i {
		if matchFunc(v) {
			return &v
		}
	}

	return nil
}

// FindOr returns the first value matching matchFunc, or defaultValue if none
// matches.
func (i ComparableIterator[V]) FindOr(matchFunc func(v V) bool, defaultValue V) V {
	if v := i.Find(matchFunc); v != nil {
		return *v
	}

	return defaultValue
}

// FindOrNone returns the first value matching matchFunc, or the zero value of V
// if none matches.
func (i ComparableIterator[V]) FindOrNone(matchFunc func(v V) bool) V {
	if v := i.Find(matchFunc); v != nil {
		return *v
	}

	var none V
	return none
}

// First returns a pointer to the first value, or nil if the iterator is empty.
func (i ComparableIterator[V]) First() *V {
	for v := range i {
		return &v
	}

	return nil
}

// FirstOr returns the first value, or defaultValue if the iterator is empty.
func (i ComparableIterator[V]) FirstOr(defaultValue V) V {
	if v := i.First(); v != nil {
		return *v
	}

	return defaultValue
}

// FirstOrNone returns the first value, or the zero value of V if the iterator is
// empty.
func (i ComparableIterator[V]) FirstOrNone() V {
	if v := i.First(); v != nil {
		return *v
	}

	var none V
	return none
}

// ForEach calls callback for each value in order. It stops and returns the first
// error from callback.
func (i ComparableIterator[V]) ForEach(callback func(v V) error) error {
	for v := range i {
		if err := callback(v); err != nil {
			return err
		}
	}

	return nil
}

// ForEachParallel calls callback for values concurrently and returns the first
// error from the worker group. parallelOptions control concurrency; callback
// calls may run out of order.
func (i ComparableIterator[V]) ForEachParallel(callback func(v V) error, parallelOptions ...ParallelOption) error {
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

// Get returns a pointer to the value at zero-based position n, or nil if n is
// negative or outside the iterator.
func (i ComparableIterator[V]) Get[N constraints.Signed | constraints.Unsigned](n N) *V {
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

// GetOr returns the value at zero-based position n, or defaultValue if n is
// outside the iterator.
func (i ComparableIterator[V]) GetOr[N constraints.Signed | constraints.Unsigned](n N, defaultValue V) V {
	if v := i.Get(n); v != nil {
		return *v
	}

	return defaultValue
}

// GetOrNone returns the value at zero-based position n, or the zero value of V if
// n is outside the iterator.
func (i ComparableIterator[V]) GetOrNone[N constraints.Signed | constraints.Unsigned](n N) V {
	if v := i.Get(n); v != nil {
		return *v
	}

	var none V
	return none
}

// Indexed returns a MapIterator that yields each value with its zero-based index.
func (i ComparableIterator[V]) Indexed() MapIterator[int, V] {
	return func(yield func(int, V) bool) {
		index := 0
		for v := range i {
			if !yield(index, v) {
				return
			}
			index++
		}
	}
}

// Last returns a pointer to the last value, or nil if the iterator is empty.
func (i ComparableIterator[V]) Last() *V {
	var last *V
	for v := range i {
		last = &v
	}

	return last
}

// LastOr returns the last value, or defaultValue if the iterator is empty.
func (i ComparableIterator[V]) LastOr(defaultValue V) V {
	if v := i.Last(); v != nil {
		return *v
	}

	return defaultValue
}

// LastOrNone returns the last value, or the zero value of V if the iterator is
// empty.
func (i ComparableIterator[V]) LastOrNone() V {
	if v := i.Last(); v != nil {
		return *v
	}

	var none V
	return none
}

// Limit returns an iterator over at most n values. It yields nothing if n is not
// positive. The returned iterator retains its remaining limit across traversals.
func (i ComparableIterator[V]) Limit[N constraints.Signed | constraints.Unsigned](n N) Iterator[V] {
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

// Map returns an iterator that applies mapFunc to each value as it is requested.
func (i ComparableIterator[V]) Map[V2 any](mapFunc func(value V) V2) Iterator[V2] {
	return func(yield func(V2) bool) {
		for v := range i {
			if !yield(mapFunc(v)) {
				return
			}
		}
	}
}

// MapAndCollect applies mapFunc to every value and returns the results in
// iteration order.
func (i ComparableIterator[V]) MapAndCollect[V2 any](mapFunc func(value V) V2) []V2 {
	var result []V2
	for v := range i {
		result = append(result, mapFunc(v))
	}
	return result
}

// MapAndCollectParallel applies mapFunc concurrently and returns the results in
// unspecified order. parallelOptions control the worker group.
func (i ComparableIterator[V]) MapAndCollectParallel[V2 any](mapFunc func(value V) V2, parallelOptions ...ParallelOption) []V2 {
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

	_ = wg.Wait() //nolint:errcheck
	return result
}

// MapAndCollectWithError applies mapFunc in order and returns the results
// collected before the first error.
func (i ComparableIterator[V]) MapAndCollectWithError[V2 any](mapFunc func(value V) (V2, error)) ([]V2, error) {
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

// MapAndCollectWithErrorParallel applies mapFunc concurrently and returns
// collected results and the first error, if any. The result order is unspecified.
func (i ComparableIterator[V]) MapAndCollectWithErrorParallel[V2 any](mapFunc func(value V) (V2, error), parallelOptions ...ParallelOption) ([]V2, error) {
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

// Max returns a pointer to the greatest value according to compareFunc, or nil if
// the iterator is empty. For ordered values, [OrderedIterator.Max] avoids a
// comparison function.
func (i ComparableIterator[V]) Max(compareFunc func(a, b V) int) *V {
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

// MaxOr returns the greatest value according to compareFunc, or defaultValue if
// the iterator is empty.
//
// For ordered values, [OrderedIterator.MaxOr] avoids a comparison function.
func (i ComparableIterator[V]) MaxOr(compareFunc func(a, b V) int, defaultValue V) V {
	if max := i.Max(compareFunc); max != nil {
		return *max
	}
	return defaultValue
}

// MaxOrNone returns the greatest value according to compareFunc, or the zero
// value of V if the iterator is empty.
//
// For ordered values, [OrderedIterator.MaxOrNone] avoids a comparison function.
func (i ComparableIterator[V]) MaxOrNone(compareFunc func(a, b V) int) V {
	if max := i.Max(compareFunc); max != nil {
		return *max
	}

	var none V
	return none
}

// Min returns a pointer to the least value according to compareFunc, or nil if
// the iterator is empty. For ordered values, [OrderedIterator.Min] avoids a
// comparison function.
func (i ComparableIterator[V]) Min(compareFunc func(a, b V) int) *V {
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

// MinMax returns pointers to the least and greatest values according to
// compareFunc. Both are nil if the iterator is empty. For ordered values,
// [OrderedIterator.MinMax] avoids a comparison function.
func (i ComparableIterator[V]) MinMax(compareFunc func(a, b V) int) (min *V, max *V) {
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

// MinMaxOr returns the least and greatest values according to compareFunc, or
// defaultValue twice if the iterator is empty.
//
// For ordered values, [OrderedIterator.MinMaxOr] avoids a comparison function.
func (i ComparableIterator[V]) MinMaxOr(compareFunc func(a, b V) int, defaultValue V) (min V, max V) {
	if minP, maxP := i.MinMax(compareFunc); minP != nil && maxP != nil {
		return *minP, *maxP
	}

	return defaultValue, defaultValue
}

// MinMaxOrNone returns the least and greatest values according to compareFunc, or
// two zero values if the iterator is empty.
//
// For ordered values, [OrderedIterator.MinMaxOrNone] avoids a comparison
// function.
func (i ComparableIterator[V]) MinMaxOrNone(compareFunc func(a, b V) int) (min V, max V) {
	if minP, maxP := i.MinMax(compareFunc); minP != nil && maxP != nil {
		return *minP, *maxP
	}

	var none V
	return none, none
}

// MinOr returns the least value according to compareFunc, or defaultValue if the
// iterator is empty.
//
// For ordered values, [OrderedIterator.MinOr] avoids a comparison function.
func (i ComparableIterator[V]) MinOr(compareFunc func(a, b V) int, defaultValue V) V {
	if min := i.Min(compareFunc); min != nil {
		return *min
	}

	return defaultValue
}

// MinOrNone returns the least value according to compareFunc, or the zero value
// of V if the iterator is empty.
//
// For ordered values, [OrderedIterator.MinOrNone] avoids a comparison function.
func (i ComparableIterator[V]) MinOrNone(compareFunc func(a, b V) int) V {
	if min := i.Min(compareFunc); min != nil {
		return *min
	}

	var none V
	return none
}

// Pull returns next and stop functions for pulling values from the iterator. Call
// stop if next is not called until exhaustion.
func (i ComparableIterator[V]) Pull() (next func() (V, bool), stop func()) {
	return iter.Pull(iter.Seq[V](i))
}

// Reduce returns startValue for an empty iterator. Otherwise it calls accumulator
// for each value and returns the last result; accumulator does not receive the
// preceding result.
func (i ComparableIterator[V]) Reduce(startValue V, accumulator func(v V) V) V {
	curr := startValue
	for v := range i {
		curr = accumulator(v)
	}
	return curr
}

// Reverse returns an iterator over the values in reverse order.
//
// This reads all values before returning the new iterator and may be slow for
// large inputs.
func (i ComparableIterator[V]) Reverse() Iterator[V] {
	s := i.Collect()
	slices.Reverse(s)
	return ToIterator(s)
}

// Skip returns an iterator that discards the first n values. A nonpositive n
// discards nothing. The returned iterator retains its remaining skip count across
// traversals.
func (i ComparableIterator[V]) Skip[N constraints.Signed | constraints.Unsigned](n N) Iterator[V] {
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

// SkipWhile returns an iterator that discards the leading values matching
// matchFunc, then yields all remaining values.
func (i ComparableIterator[V]) SkipWhile(matchFunc func(v V) bool) Iterator[V] {
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

// Sorted returns an iterator over the values sorted by compareFunc. The sort is
// not guaranteed to be stable. For ordered values, [OrderedIterator.Sorted] avoids
// a comparison function.
//
// This reads all values before returning the new iterator and may be slow for
// large inputs.
func (i ComparableIterator[V]) Sorted(compareFunc func(a, b V) int) Iterator[V] {
	return ToIterator(i.CollectSorted(compareFunc))
}

// SortedStable returns an iterator over the values sorted by compareFunc,
// preserving the order of values that compare equal. For ordered values,
// [OrderedIterator.SortedStable] avoids a comparison function.
//
// This reads all values before returning the new iterator and may be slow for
// large inputs.
func (i ComparableIterator[V]) SortedStable(compareFunc func(a, b V) int) Iterator[V] {
	return ToIterator(i.CollectSortedStable(compareFunc))
}

// TakeWhile returns an iterator over the leading values matching matchFunc. It
// stops at the first nonmatching value.
func (i ComparableIterator[V]) TakeWhile(matchFunc func(v V) bool) Iterator[V] {
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

// ToSeq returns the iterator as an iter.Seq.
func (i ComparableIterator[V]) ToSeq() iter.Seq[V] {
	return iter.Seq[V](i)
}
