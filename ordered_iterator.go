package itertools

import (
	"cmp"
	"iter"
	"slices"
	"sync"

	"golang.org/x/exp/constraints"
	"golang.org/x/sync/errgroup"
)

// OrderedIterator yields ordered values of type V. It has the same signature
// as [iter.Seq] and can be used directly in a for range loop.
type OrderedIterator[V cmp.Ordered] func(yield func(V) bool)

// ToOrderedIterator returns an iterator over the elements of s in slice order.
func ToOrderedIterator[Slice ~[]V, V cmp.Ordered](s Slice) OrderedIterator[V] {
	return OrderedIterator[V](slices.Values(s))
}

// SeqToOrderedIterator converts s to an OrderedIterator without collecting its
// values.
func SeqToOrderedIterator[V cmp.Ordered](s iter.Seq[V]) OrderedIterator[V] {
	return OrderedIterator[V](s)
}

// CollectSorted collects the values, sorts them in ascending order, and returns
// the slice.
func (i OrderedIterator[V]) CollectSorted() []V {
	return slices.SortedFunc(iter.Seq[V](i), cmp.Compare)
}

// CollectSortedAs is like CollectSorted but returns a slice of type Slice.
func (i OrderedIterator[V]) CollectSortedAs[Slice ~[]V]() Slice {
	return slices.SortedFunc(iter.Seq[V](i), cmp.Compare)
}

// CollectSortedStable collects the values and sorts them in ascending order while
// preserving the order of equal values.
func (i OrderedIterator[V]) CollectSortedStable() []V {
	return slices.SortedStableFunc(iter.Seq[V](i), cmp.Compare)
}

// CollectSortedStableAs is like CollectSortedStable but returns a slice of type
// Slice.
func (i OrderedIterator[V]) CollectSortedStableAs[Slice ~[]V]() Slice {
	return slices.SortedStableFunc(iter.Seq[V](i), cmp.Compare)
}

// CollectSortedFunc collects the values, sorts them using compareFunc, and
// returns the slice. The sort is not guaranteed to be stable. Use CollectSorted
// for natural order; it avoids a comparison function.
func (i OrderedIterator[V]) CollectSortedFunc(compareFunc func(a, b V) int) []V {
	return slices.SortedFunc(iter.Seq[V](i), compareFunc)
}

// CollectSortedFuncAs is like CollectSortedFunc but returns a slice of type
// Slice.
//
// For natural order, [OrderedIterator.CollectSortedAs] avoids a comparison
// function.
func (i OrderedIterator[V]) CollectSortedFuncAs[Slice ~[]V](compareFunc func(a, b V) int) Slice {
	return slices.SortedFunc(iter.Seq[V](i), compareFunc)
}

// CollectSortedStableFunc collects the values and sorts them using compareFunc
// while preserving the order of values that compare equal. Use
// CollectSortedStable for natural order; it avoids a comparison function.
func (i OrderedIterator[V]) CollectSortedStableFunc(compareFunc func(a, b V) int) []V {
	return slices.SortedStableFunc(iter.Seq[V](i), compareFunc)
}

// CollectSortedStableFuncAs is like CollectSortedStableFunc but returns a slice
// of type Slice.
//
// For natural order, [OrderedIterator.CollectSortedStableAs] avoids a comparison
// function.
func (i OrderedIterator[V]) CollectSortedStableFuncAs[Slice ~[]V](compareFunc func(a, b V) int) Slice {
	return slices.SortedStableFunc(iter.Seq[V](i), compareFunc)
}

// Concat returns an iterator over the values of i followed by those of another.
func (i OrderedIterator[V]) Concat[A Iterator[V] | ComparableIterator[V] | OrderedIterator[V]](another A) OrderedIterator[V] {
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

// Equal reports whether two iterators yield equal values in the same order. It
// stops at the first difference.
func (i OrderedIterator[V]) Equal[A Iterator[V] | ComparableIterator[V] | OrderedIterator[V]](another A) bool {
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
func (i OrderedIterator[V]) EqualFunc[A Iterator[V] | ComparableIterator[V] | OrderedIterator[V]](another A, compareFunc func(a, b V) int) bool {
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
func (i OrderedIterator[V]) FilterNone() OrderedIterator[V] {
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

// Max returns a pointer to the greatest value, or nil if the iterator is empty.
func (i OrderedIterator[V]) Max() *V {
	return i.MaxFunc(cmp.Compare)
}

// MaxOr returns the greatest value, or defaultValue if the iterator is empty.
func (i OrderedIterator[V]) MaxOr(defaultValue V) V {
	if max := i.Max(); max != nil {
		return *max
	}
	return defaultValue
}

// MaxOrNone returns the greatest value, or the zero value of V if the iterator is
// empty.
func (i OrderedIterator[V]) MaxOrNone() V {
	if max := i.Max(); max != nil {
		return *max
	}

	var none V
	return none
}

// MaxFunc returns a pointer to the greatest value according to compareFunc, or
// nil if the iterator is empty. Use Max for natural order; it avoids a comparison
// function.
func (i OrderedIterator[V]) MaxFunc(compareFunc func(a, b V) int) *V {
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

// MaxFuncOr returns the greatest value according to compareFunc, or defaultValue
// if the iterator is empty.
//
// For natural order, [OrderedIterator.MaxOr] avoids a comparison function.
func (i OrderedIterator[V]) MaxFuncOr(compareFunc func(a, b V) int, defaultValue V) V {
	if max := i.MaxFunc(compareFunc); max != nil {
		return *max
	}
	return defaultValue
}

// MaxFuncOrNone returns the greatest value according to compareFunc, or the zero
// value of V if the iterator is empty.
//
// For natural order, [OrderedIterator.MaxOrNone] avoids a comparison function.
func (i OrderedIterator[V]) MaxFuncOrNone(compareFunc func(a, b V) int) V {
	if max := i.MaxFunc(compareFunc); max != nil {
		return *max
	}

	var none V
	return none
}

// Min returns a pointer to the least value, or nil if the iterator is empty.
func (i OrderedIterator[V]) Min() *V {
	return i.MinFunc(cmp.Compare)
}

// MinOr returns the least value, or defaultValue if the iterator is empty.
func (i OrderedIterator[V]) MinOr(defaultValue V) V {
	if min := i.Min(); min != nil {
		return *min
	}

	return defaultValue
}

// MinOrNone returns the least value, or the zero value of V if the iterator is
// empty.
func (i OrderedIterator[V]) MinOrNone() V {
	if min := i.Min(); min != nil {
		return *min
	}

	var none V
	return none
}

// MinFunc returns a pointer to the least value according to compareFunc, or nil
// if the iterator is empty. Use Min for natural order; it avoids a comparison
// function.
func (i OrderedIterator[V]) MinFunc(compareFunc func(a, b V) int) *V {
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

// MinFuncOr returns the least value according to compareFunc, or defaultValue if
// the iterator is empty.
//
// For natural order, [OrderedIterator.MinOr] avoids a comparison function.
func (i OrderedIterator[V]) MinFuncOr(compareFunc func(a, b V) int, defaultValue V) V {
	if min := i.MinFunc(compareFunc); min != nil {
		return *min
	}

	return defaultValue
}

// MinFuncOrNone returns the least value according to compareFunc, or the zero
// value of V if the iterator is empty.
//
// For natural order, [OrderedIterator.MinOrNone] avoids a comparison function.
func (i OrderedIterator[V]) MinFuncOrNone(compareFunc func(a, b V) int) V {
	if min := i.MinFunc(compareFunc); min != nil {
		return *min
	}

	var none V
	return none
}

// MinMax returns pointers to the least and greatest values. Both are nil if the
// iterator is empty.
func (i OrderedIterator[V]) MinMax() (min *V, max *V) {
	return i.MinMaxFunc(cmp.Compare)
}

// MinMaxOr returns the least and greatest values, or defaultValue twice if the
// iterator is empty.
func (i OrderedIterator[V]) MinMaxOr(defaultValue V) (min V, max V) {
	if minP, maxP := i.MinMax(); minP != nil && maxP != nil {
		return *minP, *maxP
	}

	return defaultValue, defaultValue
}

// MinMaxOrNone returns the least and greatest values, or two zero values if the
// iterator is empty.
func (i OrderedIterator[V]) MinMaxOrNone() (min V, max V) {
	if minP, maxP := i.MinMax(); minP != nil && maxP != nil {
		return *minP, *maxP
	}

	var none V
	return none, none
}

// MinMaxFunc returns pointers to the least and greatest values according to
// compareFunc. Both are nil if the iterator is empty. Use MinMax for natural
// order; it avoids a comparison function.
func (i OrderedIterator[V]) MinMaxFunc(compareFunc func(a, b V) int) (min *V, max *V) {
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

// MinMaxFuncOr returns the least and greatest values according to compareFunc, or
// defaultValue twice if the iterator is empty.
//
// For natural order, [OrderedIterator.MinMaxOr] avoids a comparison function.
func (i OrderedIterator[V]) MinMaxFuncOr(compareFunc func(a, b V) int, defaultValue V) (min V, max V) {
	if minP, maxP := i.MinMaxFunc(compareFunc); minP != nil && maxP != nil {
		return *minP, *maxP
	}

	return defaultValue, defaultValue
}

// MinMaxFuncOrNone returns the least and greatest values according to
// compareFunc, or two zero values if the iterator is empty.
//
// For natural order, [OrderedIterator.MinMaxOrNone] avoids a comparison function.
func (i OrderedIterator[V]) MinMaxFuncOrNone(compareFunc func(a, b V) int) (min V, max V) {
	if minP, maxP := i.MinMaxFunc(compareFunc); minP != nil && maxP != nil {
		return *minP, *maxP
	}

	var none V
	return none, none
}

// Sorted returns an iterator over the values sorted in ascending order.
//
// This reads all values before returning the new iterator and may be slow for
// large inputs.
func (i OrderedIterator[V]) Sorted() OrderedIterator[V] {
	return ToOrderedIterator(i.CollectSorted())
}

// SortedStable returns an iterator over the values sorted in ascending order,
// preserving the order of equal values.
//
// This reads all values before returning the new iterator and may be slow for
// large inputs.
func (i OrderedIterator[V]) SortedStable() OrderedIterator[V] {
	return ToOrderedIterator(i.CollectSortedStable())
}

// SortedFunc returns an iterator over the values sorted by compareFunc. The sort
// is not guaranteed to be stable. Use Sorted for natural order; it avoids a
// comparison function.
//
// This reads all values before returning the new iterator and may be slow for
// large inputs.
func (i OrderedIterator[V]) SortedFunc(compareFunc func(a, b V) int) OrderedIterator[V] {
	return ToOrderedIterator(i.CollectSortedFunc(compareFunc))
}

// SortedStableFunc returns an iterator over the values sorted by compareFunc,
// preserving the order of values that compare equal. Use SortedStable for natural
// order; it avoids a comparison function.
//
// This reads all values before returning the new iterator and may be slow for
// large inputs.
func (i OrderedIterator[V]) SortedStableFunc(compareFunc func(a, b V) int) OrderedIterator[V] {
	return ToOrderedIterator(i.CollectSortedStableFunc(compareFunc))
}

// ToComparableIterator returns i as a ComparableIterator without collecting its
// values.
func (i OrderedIterator[V]) ToComparableIterator() ComparableIterator[V] {
	return ComparableIterator[V](i)
}

// --- From ComparableIterator[V] ---

// Contains reports whether the iterator contains value. It stops at the first
// match.
func (i OrderedIterator[V]) Contains(value V) bool {
	for v := range i {
		if v == value {
			return true
		}
	}

	return false
}

// Unique returns an iterator that yields only the first occurrence of each value.
// It uses a map to track seen values and retains that state across traversals.
func (i OrderedIterator[V]) Unique() ComparableIterator[V] {
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
func (i OrderedIterator[V]) UniqueFunc(compareFunc func(a, b V) int) Iterator[V] {
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
func (i OrderedIterator[V]) ToIterator() Iterator[V] {
	return Iterator[V](i)
}

// Zip returns an iterator pairing keys with the values of i until either side is
// exhausted. It collects all values of i before reading keys and consumes those
// values across traversals.
//
// Collecting the values first may be slow or use substantial memory for large
// inputs.
func (i OrderedIterator[V]) Zip[K comparable, Keys Iterator[K] | ComparableIterator[K]](keys Keys) MapIterator[K, V] {
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
func (i OrderedIterator[V]) All(matchFunc func(v V) bool) bool {
	for v := range i {
		if !matchFunc(v) {
			return false
		}
	}

	return true
}

// Any reports whether matchFunc returns true for any value. It stops at the first
// match.
func (i OrderedIterator[V]) Any(matchFunc func(v V) bool) bool {
	return i.ContainsFunc(matchFunc)
}

// Collect returns the values in iteration order as a slice. The result is nil if
// the iterator is empty.
func (i OrderedIterator[V]) Collect() []V {
	return slices.Collect(iter.Seq[V](i))
}

// CollectAs returns the values in iteration order as a slice of type Slice. The
// result is nil if the iterator is empty.
func (i OrderedIterator[V]) CollectAs[Slice ~[]V]() Slice {
	return slices.Collect(iter.Seq[V](i))
}

// ContainsFunc reports whether matchFunc returns true for any value. It stops at
// the first match.
func (i OrderedIterator[V]) ContainsFunc(matchFunc func(v V) bool) bool {
	for v := range i {
		if matchFunc(v) {
			return true
		}
	}

	return false
}

// Count returns the number of values yielded by the iterator.
func (i OrderedIterator[V]) Count() int {
	var count int
	for range i {
		count++
	}
	return count
}

// Filter returns an iterator over values for which filterFunc returns true. It
// evaluates filterFunc as values are requested.
func (i OrderedIterator[V]) Filter(filterFunc func(value V) bool) Iterator[V] {
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
func (i OrderedIterator[V]) FilterAndCollect(filterFunc func(value V) bool) []V {
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
func (i OrderedIterator[V]) FilterAndCollectWithError(filterFunc func(value V) (bool, error)) ([]V, error) {
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
func (i OrderedIterator[V]) FilterAndCollectParallel(filterFunc func(value V) bool, parallelOptions ...ParallelOption) []V {
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
func (i OrderedIterator[V]) FilterAndCollectWithErrorParallel(filterFunc func(value V) (bool, error), parallelOptions ...ParallelOption) ([]V, error) {
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
func (i OrderedIterator[V]) Find(matchFunc func(v V) bool) *V {
	for v := range i {
		if matchFunc(v) {
			return &v
		}
	}

	return nil
}

// FindOr returns the first value matching matchFunc, or defaultValue if none
// matches.
func (i OrderedIterator[V]) FindOr(matchFunc func(v V) bool, defaultValue V) V {
	if v := i.Find(matchFunc); v != nil {
		return *v
	}

	return defaultValue
}

// FindOrNone returns the first value matching matchFunc, or the zero value of V
// if none matches.
func (i OrderedIterator[V]) FindOrNone(matchFunc func(v V) bool) V {
	if v := i.Find(matchFunc); v != nil {
		return *v
	}

	var none V
	return none
}

// First returns a pointer to the first value, or nil if the iterator is empty.
func (i OrderedIterator[V]) First() *V {
	for v := range i {
		return &v
	}

	return nil
}

// FirstOr returns the first value, or defaultValue if the iterator is empty.
func (i OrderedIterator[V]) FirstOr(defaultValue V) V {
	if v := i.First(); v != nil {
		return *v
	}

	return defaultValue
}

// FirstOrNone returns the first value, or the zero value of V if the iterator is
// empty.
func (i OrderedIterator[V]) FirstOrNone() V {
	if v := i.First(); v != nil {
		return *v
	}

	var none V
	return none
}

// ForEach calls callback for each value in order. It stops and returns the first
// error from callback.
func (i OrderedIterator[V]) ForEach(callback func(v V) error) error {
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
func (i OrderedIterator[V]) ForEachParallel(callback func(v V) error, parallelOptions ...ParallelOption) error {
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
func (i OrderedIterator[V]) Get[N constraints.Signed | constraints.Unsigned](n N) *V {
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
func (i OrderedIterator[V]) GetOr[N constraints.Signed | constraints.Unsigned](n N, defaultValue V) V {
	if v := i.Get(n); v != nil {
		return *v
	}

	return defaultValue
}

// GetOrNone returns the value at zero-based position n, or the zero value of V if
// n is outside the iterator.
func (i OrderedIterator[V]) GetOrNone[N constraints.Signed | constraints.Unsigned](n N) V {
	if v := i.Get(n); v != nil {
		return *v
	}

	var none V
	return none
}

// Indexed returns a MapIterator that yields each value with its zero-based index.
func (i OrderedIterator[V]) Indexed() MapIterator[int, V] {
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
func (i OrderedIterator[V]) Last() *V {
	var last *V
	for v := range i {
		last = &v
	}

	return last
}

// LastOr returns the last value, or defaultValue if the iterator is empty.
func (i OrderedIterator[V]) LastOr(defaultValue V) V {
	if v := i.Last(); v != nil {
		return *v
	}

	return defaultValue
}

// LastOrNone returns the last value, or the zero value of V if the iterator is
// empty.
func (i OrderedIterator[V]) LastOrNone() V {
	if v := i.Last(); v != nil {
		return *v
	}

	var none V
	return none
}

// Limit returns an iterator over at most n values. It yields nothing if n is not
// positive. The returned iterator retains its remaining limit across traversals.
func (i OrderedIterator[V]) Limit[N constraints.Signed | constraints.Unsigned](n N) Iterator[V] {
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
func (i OrderedIterator[V]) Map[V2 any](mapFunc func(value V) V2) Iterator[V2] {
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
func (i OrderedIterator[V]) MapAndCollect[V2 any](mapFunc func(value V) V2) []V2 {
	var result []V2
	for v := range i {
		result = append(result, mapFunc(v))
	}
	return result
}

// MapAndCollectParallel applies mapFunc concurrently and returns the results in
// unspecified order. parallelOptions control the worker group.
func (i OrderedIterator[V]) MapAndCollectParallel[V2 any](mapFunc func(value V) V2, parallelOptions ...ParallelOption) []V2 {
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
func (i OrderedIterator[V]) MapAndCollectWithError[V2 any](mapFunc func(value V) (V2, error)) ([]V2, error) {
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
func (i OrderedIterator[V]) MapAndCollectWithErrorParallel[V2 any](mapFunc func(value V) (V2, error), parallelOptions ...ParallelOption) ([]V2, error) {
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

// Pull returns next and stop functions for pulling values from the iterator. Call
// stop if next is not called until exhaustion.
func (i OrderedIterator[V]) Pull() (next func() (V, bool), stop func()) {
	return iter.Pull(iter.Seq[V](i))
}

// Reduce returns startValue for an empty iterator. Otherwise it calls accumulator
// for each value and returns the last result; accumulator does not receive the
// preceding result.
func (i OrderedIterator[V]) Reduce(startValue V, accumulator func(v V) V) V {
	curr := startValue
	for v := range i {
		curr = accumulator(v)
	}
	return curr
}

// ReduceWithError returns startValue for an empty iterator. Otherwise it calls
// accumulator for each value and returns the last result; accumulator does not
// receive the preceding result. If accumulator returns an error, ReduceWithError
// stops and returns the last result and that error.
func (i OrderedIterator[V]) ReduceWithError(startValue V, accumulator func(v V) (V, error)) (V, error) {
	curr := startValue
	for v := range i {
		var err error
		curr, err = accumulator(v)
		if err != nil {
			return curr, err
		}
	}
	return curr, nil
}

// Reverse returns an iterator over the values in reverse order.
//
// This reads all values before returning the new iterator and may be slow for
// large inputs.
func (i OrderedIterator[V]) Reverse() Iterator[V] {
	s := i.Collect()
	slices.Reverse(s)
	return ToIterator(s)
}

// Skip returns an iterator that discards the first n values. A nonpositive n
// discards nothing. The returned iterator retains its remaining skip count across
// traversals.
func (i OrderedIterator[V]) Skip[N constraints.Signed | constraints.Unsigned](n N) Iterator[V] {
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
func (i OrderedIterator[V]) SkipWhile(matchFunc func(v V) bool) Iterator[V] {
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

// TakeWhile returns an iterator over the leading values matching matchFunc. It
// stops at the first nonmatching value.
func (i OrderedIterator[V]) TakeWhile(matchFunc func(v V) bool) Iterator[V] {
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
func (i OrderedIterator[V]) ToSeq() iter.Seq[V] {
	return iter.Seq[V](i)
}
