package itertools

import (
	"iter"
	"maps"
	"reflect"
	"slices"
	"sync"

	"golang.org/x/exp/constraints"
	"golang.org/x/sync/errgroup"
)

// MapIterator yields key-value pairs. It has the same signature as [iter.Seq2]
// and can be used directly in a for range loop.
type MapIterator[K comparable, V any] func(yield func(K, V) bool)

// ToMapIterator returns an iterator over the entries of m. Like map iteration,
// the order is unspecified.
func ToMapIterator[Map ~map[K]V, K comparable, V any](m Map) MapIterator[K, V] {
	return MapIterator[K, V](maps.All(m))
}

// Seq2ToMapIterator converts s to a MapIterator without collecting its pairs.
func Seq2ToMapIterator[K comparable, V any](s iter.Seq2[K, V]) MapIterator[K, V] {
	return MapIterator[K, V](s)
}

// All reports whether matchFunc returns true for every pair. It returns true for
// an empty iterator and stops at the first nonmatching pair.
func (i MapIterator[K, V]) All(matchFunc func(k K, v V) bool) bool {
	for k, v := range i {
		if !matchFunc(k, v) {
			return false
		}
	}

	return true
}

// Any reports whether matchFunc returns true for any pair. It stops at the first
// match.
func (i MapIterator[K, V]) Any(matchFunc func(k K, v V) bool) bool {
	return i.ContainsFunc(matchFunc)
}

// Collect collects the pairs into a new map. A later occurrence of a key replaces
// an earlier value.
func (i MapIterator[K, V]) Collect() map[K]V {
	return maps.Collect(iter.Seq2[K, V](i))
}

// CollectAs is like Collect but returns a map of type Map.
func (i MapIterator[K, V]) CollectAs[Map ~map[K]V]() Map {
	return maps.Collect(iter.Seq2[K, V](i))
}

// CollectKeys returns the keys in iteration order, including repeated keys.
//
// For lazy traversal, use [MapIterator.Keys] instead.
func (i MapIterator[K, V]) CollectKeys() []K {
	var keys []K
	for k, _ := range i {
		keys = append(keys, k)
	}
	return keys
}

// CollectKeysAndValues returns aligned slices of keys and values in iteration
// order, including repeated keys.
func (i MapIterator[K, V]) CollectKeysAndValues() (keys []K, values []V) {
	for k, v := range i {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}

// CollectKeysAndValuesAs is like CollectKeysAndValues but returns slices of types
// Keys and Values.
func (i MapIterator[K, V]) CollectKeysAndValuesAs[Keys ~[]K, Values ~[]V]() (keys Keys, values Values) {
	return i.CollectKeysAndValues()
}

// CollectKeysAs is like CollectKeys but returns a slice of type Slice.
//
// For lazy traversal, use [MapIterator.Keys] instead.
func (i MapIterator[K, V]) CollectKeysAs[Slice ~[]K]() Slice {
	return i.CollectKeys()
}

// CollectValues returns the values in iteration order, including values for
// repeated keys.
//
// For lazy traversal, use [MapIterator.Values] instead.
func (i MapIterator[K, V]) CollectValues() []V {
	var values []V
	for _, v := range i {
		values = append(values, v)
	}
	return values
}

// CollectValuesAs is like CollectValues but returns a slice of type Slice.
//
// For lazy traversal, use [MapIterator.Values] instead.
func (i MapIterator[K, V]) CollectValuesAs[Slice ~[]V]() Slice {
	return i.CollectValues()
}

// ContainsFunc reports whether matchFunc returns true for any pair. It stops at
// the first match.
func (i MapIterator[K, V]) ContainsFunc(matchFunc func(k K, v V) bool) bool {
	for k, v := range i {
		if matchFunc(k, v) {
			return true
		}
	}

	return false
}

// ContainsKey reports whether the iterator yields key. It stops at the first
// match.
func (i MapIterator[K, V]) ContainsKey(key K) bool {
	for k, _ := range i {
		if k == key {
			return true
		}
	}

	return false
}

// ContainsValue reports whether the iterator yields value. It stops at the first
// match and panics if value is not comparable.
func (i MapIterator[K, V]) ContainsValue(value V) bool {
	rv := reflect.ValueOf(value)
	if !rv.Type().Comparable() {
		panic("iterator value type " + rv.Type().String() + " is not comparable")
	}

	for _, v := range i {
		if reflect.ValueOf(v).Equal(rv) {
			return true
		}
	}

	return false
}

// Count returns the number of pairs yielded, including repeated keys.
func (i MapIterator[K, V]) Count() int {
	var count int
	for range i {
		count++
	}
	return count
}

// Equal reports whether two iterators yield the same keys and compareFunc-equal
// values in the same order. It stops at the first difference.
func (i MapIterator[K, V]) Equal(another MapIterator[K, V], compareFunc func(a, b V) int) bool {
	p1, p1stop := iter.Pull2(iter.Seq2[K, V](i))
	p2, p2stop := iter.Pull2(iter.Seq2[K, V](another))
	defer p1stop()
	defer p2stop()

	for {
		k1, v1, ok1 := p1()
		k2, v2, ok2 := p2()

		if !ok1 && !ok2 {
			return true
		}

		if !ok1 || !ok2 {
			return false
		}

		if k1 != k2 {
			return false
		}

		if c := compareFunc(v1, v2); c != 0 {
			return false
		}
	}
}

// Filter returns an iterator over pairs for which filterFunc returns true. It
// evaluates filterFunc as pairs are requested.
func (i MapIterator[K, V]) Filter(filterFunc func(key K, value V) bool) MapIterator[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range i {
			if !filterFunc(k, v) {
				continue
			}

			if !yield(k, v) {
				return
			}
		}
	}
}

// FilterAndCollect returns a map of pairs for which filterFunc returns true. A
// later matching occurrence of a key replaces an earlier value.
func (i MapIterator[K, V]) FilterAndCollect(filterFunc func(key K, value V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range i {
		if filterFunc(k, v) {
			result[k] = v
		}
	}
	return result
}

// FilterAndCollectWithError is like FilterAndCollect, but stops and returns the
// map collected so far if filterFunc returns an error.
func (i MapIterator[K, V]) FilterAndCollectWithError(filterFunc func(key K, value V) (bool, error)) (map[K]V, error) {
	result := make(map[K]V)
	for k, v := range i {
		ok, err := filterFunc(k, v)
		if err != nil {
			return result, err
		}
		if ok {
			result[k] = v
		}
	}
	return result, nil
}

// FilterAndCollectParallel filters pairs concurrently and returns a map of
// matches. parallelOptions control the worker group. If a key repeats, the
// retained value depends on worker completion order.
func (i MapIterator[K, V]) FilterAndCollectParallel(filterFunc func(key K, value V) bool, parallelOptions ...ParallelOption) map[K]V {
	var wg errgroup.Group
	result := make(map[K]V)
	var resultMutex sync.Mutex

	for _, option := range parallelOptions {
		option(&wg)
	}
	for k, v := range i {
		wg.Go(func() error {
			if filterFunc(k, v) {
				resultMutex.Lock()
				result[k] = v
				resultMutex.Unlock()
			}
			return nil
		})
	}

	_ = wg.Wait()
	return result
}

// FilterAndCollectWithErrorParallel filters pairs concurrently and returns
// collected matches and the first error, if any. If a key repeats, the retained
// value depends on worker completion order.
func (i MapIterator[K, V]) FilterAndCollectWithErrorParallel(filterFunc func(key K, value V) (bool, error), parallelOptions ...ParallelOption) (map[K]V, error) {
	var wg errgroup.Group
	result := make(map[K]V)
	var resultMutex sync.Mutex

	for _, option := range parallelOptions {
		option(&wg)
	}
	for k, v := range i {
		wg.Go(func() error {
			ok, err := filterFunc(k, v)
			if err != nil {
				return err
			}
			if ok {
				resultMutex.Lock()
				result[k] = v
				resultMutex.Unlock()
			}
			return nil
		})
	}

	err := wg.Wait()
	return result, err
}

// FilterKeys returns an iterator over pairs whose key satisfies filterFunc.
func (i MapIterator[K, V]) FilterKeys(filterFunc func(key K) bool) MapIterator[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range i {
			if !filterFunc(k) {
				continue
			}

			if !yield(k, v) {
				return
			}
		}
	}
}

// FilterValues returns an iterator over pairs whose value satisfies filterFunc.
func (i MapIterator[K, V]) FilterValues(filterFunc func(value V) bool) MapIterator[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range i {
			if !filterFunc(v) {
				continue
			}

			if !yield(k, v) {
				return
			}
		}
	}
}

// FilterMap returns an iterator over pairs for which filterFunc returns true, applying
// the transformation to each matching pair. It evaluates filterFunc as pairs are
// requested.
func (i MapIterator[K, V]) FilterMap[K2 comparable, V2 any](filterFunc func(key K, value V) (K2, V2, bool)) MapIterator[K2, V2] {
	return func(yield func(K2, V2) bool) {
		for k, v := range i {
			k2, v2, ok := filterFunc(k, v)
			if !ok {
				continue
			}

			if !yield(k2, v2) {
				return
			}
		}
	}
}

// Find returns pointers to the first key and value matching matchFunc, or two nil
// pointers if none matches.
func (i MapIterator[K, V]) Find(matchFunc func(k K, v V) bool) (*K, *V) {
	for k, v := range i {
		if matchFunc(k, v) {
			return &k, &v
		}
	}
	return nil, nil
}

// FindKey returns a pointer to the first key whose pair matches matchFunc, or nil
// if none matches.
func (i MapIterator[K, V]) FindKey(matchFunc func(k K, v V) bool) *K {
	for k, v := range i {
		if matchFunc(k, v) {
			return &k
		}
	}
	return nil
}

// FindKeyOr returns the first key whose pair matches matchFunc, or defaultKey if
// none matches.
func (i MapIterator[K, V]) FindKeyOr(matchFunc func(k K, v V) bool, defaultKey K) K {
	if k := i.FindKey(matchFunc); k != nil {
		return *k
	}

	return defaultKey
}

// FindKeyOrNone returns the first key whose pair matches matchFunc, or the zero
// value of K if none matches.
func (i MapIterator[K, V]) FindKeyOrNone(matchFunc func(k K, v V) bool) K {
	if k := i.FindKey(matchFunc); k != nil {
		return *k
	}

	var none K
	return none
}

// FindOr returns the first pair matching matchFunc, or defaultKey and
// defaultValue if none matches.
func (i MapIterator[K, V]) FindOr(matchFunc func(k K, v V) bool, defaultKey K, defaultValue V) (K, V) {
	if k, v := i.Find(matchFunc); v != nil {
		return *k, *v
	}

	return defaultKey, defaultValue
}

// FindOrNone returns the first pair matching matchFunc, or the zero values of K
// and V if none matches.
func (i MapIterator[K, V]) FindOrNone(matchFunc func(k K, v V) bool) (K, V) {
	if k, v := i.Find(matchFunc); v != nil {
		return *k, *v
	}

	var noneK K
	var noneV V
	return noneK, noneV
}

// FindValue returns a pointer to the first value whose pair matches matchFunc, or
// nil if none matches.
func (i MapIterator[K, V]) FindValue(matchFunc func(k K, v V) bool) *V {
	for k, v := range i {
		if matchFunc(k, v) {
			return &v
		}
	}
	return nil
}

// FindValueOr returns the first value whose pair matches matchFunc, or
// defaultValue if none matches.
func (i MapIterator[K, V]) FindValueOr(matchFunc func(k K, v V) bool, defaultValue V) V {
	if v := i.FindValue(matchFunc); v != nil {
		return *v
	}

	return defaultValue
}

// FindValueOrNone returns the first value whose pair matches matchFunc, or the
// zero value of V if none matches.
func (i MapIterator[K, V]) FindValueOrNone(matchFunc func(k K, v V) bool) V {
	if v := i.FindValue(matchFunc); v != nil {
		return *v
	}

	var none V
	return none
}

// ForEach calls callback for each pair in order. It stops and returns the first
// error from callback.
func (i MapIterator[K, V]) ForEach(callback func(k K, v V) error) error {
	for k, v := range i {
		if err := callback(k, v); err != nil {
			return err
		}
	}

	return nil
}

// ForEachParallel calls callback for pairs concurrently and returns the first
// error from the worker group. parallelOptions control concurrency; callback
// calls may run out of order.
func (i MapIterator[K, V]) ForEachParallel(callback func(k K, v V) error, parallelOptions ...ParallelOption) error {
	var wg errgroup.Group

	for _, option := range parallelOptions {
		option(&wg)
	}

	for k, v := range i {
		wg.Go(func() error {
			return callback(k, v)
		})
	}

	return wg.Wait()
}

// Get returns a pointer to the value of the first pair with key, or nil if no
// pair has that key.
func (i MapIterator[K, V]) Get(key K) *V {
	for k, v := range i {
		if k == key {
			return &v
		}
	}

	return nil
}

// GetOr returns the value of the first pair with key, or defaultValue if no pair
// has that key.
func (i MapIterator[K, V]) GetOr(key K, defaultValue V) V {
	if v := i.Get(key); v != nil {
		return *v
	}

	return defaultValue
}

// GetOrNone returns the value of the first pair with key, or the zero value of V
// if no pair has that key.
func (i MapIterator[K, V]) GetOrNone(key K) V {
	if v := i.Get(key); v != nil {
		return *v
	}

	var none V
	return none
}

// Concat returns an iterator over the pairs of i followed by those of another.
// Repeated keys are preserved.
func (i MapIterator[K, V]) Concat(another MapIterator[K, V]) MapIterator[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range i {
			if !yield(k, v) {
				return
			}
		}

		for k, v := range another {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Keys returns an iterator over the keys in pair order, including repeated keys.
func (i MapIterator[K, V]) Keys() ComparableIterator[K] {
	return func(yield func(K) bool) {
		for k := range i {
			if !yield(k) {
				return
			}
		}
	}
}

// Limit returns an iterator over at most n pairs. It yields nothing if n is not
// positive. The returned iterator retains its remaining limit across traversals.
func (i MapIterator[K, V]) Limit[N constraints.Signed | constraints.Unsigned](n N) MapIterator[K, V] {
	return func(yield func(K, V) bool) {
		if n <= 0 {
			return
		}

		for k, v := range i {
			if !yield(k, v) {
				return
			}

			n--
			if n <= 0 {
				return
			}
		}
	}
}

// Map returns an iterator that applies mapFunc to each pair as it is requested.
func (i MapIterator[K, V]) Map[K2 comparable, V2 any](mapFunc func(key K, value V) (K2, V2)) MapIterator[K2, V2] {
	return func(yield func(K2, V2) bool) {
		for k, v := range i {
			if !yield(mapFunc(k, v)) {
				return
			}
		}
	}
}

// MapAndCollect applies mapFunc to every pair and collects the results into a
// map. A later mapped key replaces an earlier value.
func (i MapIterator[K, V]) MapAndCollect[K2 comparable, V2 any](mapFunc func(key K, value V) (K2, V2)) map[K2]V2 {
	result := make(map[K2]V2)
	for k, v := range i {
		k2, v2 := mapFunc(k, v)
		result[k2] = v2
	}
	return result
}

// MapAndCollectParallel applies mapFunc concurrently and collects the results
// into a map. If mapped keys repeat, the retained value depends on worker
// completion order.
func (i MapIterator[K, V]) MapAndCollectParallel[K2 comparable, V2 any](mapFunc func(key K, value V) (K2, V2), parallelOptions ...ParallelOption) map[K2]V2 {
	var wg errgroup.Group
	result := make(map[K2]V2)
	var resultMutex sync.Mutex

	for _, option := range parallelOptions {
		option(&wg)
	}
	for k, v := range i {
		wg.Go(func() error {
			k2, v2 := mapFunc(k, v)
			resultMutex.Lock()
			result[k2] = v2
			resultMutex.Unlock()
			return nil
		})
	}

	_ = wg.Wait()
	return result
}

// MapAndCollectWithError applies mapFunc in order and returns the map collected
// before the first error.
func (i MapIterator[K, V]) MapAndCollectWithError[K2 comparable, V2 any](mapFunc func(key K, value V) (K2, V2, error)) (map[K2]V2, error) {
	result := make(map[K2]V2)
	for k, v := range i {
		k2, v2, err := mapFunc(k, v)
		if err != nil {
			return result, err
		}
		result[k2] = v2
	}
	return result, nil
}

// MapAndCollectWithErrorParallel applies mapFunc concurrently and returns
// collected results and the first error, if any. If mapped keys repeat, the
// retained value depends on worker completion order.
func (i MapIterator[K, V]) MapAndCollectWithErrorParallel[K2 comparable, V2 any](mapFunc func(key K, value V) (K2, V2, error), parallelOptions ...ParallelOption) (map[K2]V2, error) {
	var wg errgroup.Group
	result := make(map[K2]V2)
	var resultMutex sync.Mutex

	for _, option := range parallelOptions {
		option(&wg)
	}
	for k, v := range i {
		wg.Go(func() error {
			k2, v2, err := mapFunc(k, v)
			if err != nil {
				return err
			}
			resultMutex.Lock()
			result[k2] = v2
			resultMutex.Unlock()
			return nil
		})
	}

	err := wg.Wait()
	return result, err
}

// MapKeys returns an iterator that transforms each key with mapFunc and leaves
// its value unchanged.
func (i MapIterator[K, V]) MapKeys[K2 comparable](mapFunc func(key K) K2) MapIterator[K2, V] {
	return func(yield func(K2, V) bool) {
		for k, v := range i {
			if !yield(mapFunc(k), v) {
				return
			}
		}
	}
}

// MapValues returns an iterator that transforms each value with mapFunc and
// leaves its key unchanged.
func (i MapIterator[K, V]) MapValues[V2 any](mapFunc func(value V) V2) MapIterator[K, V2] {
	return func(yield func(K, V2) bool) {
		for k, v := range i {
			if !yield(k, mapFunc(v)) {
				return
			}
		}
	}
}

// Pull returns next and stop functions for pulling pairs from the iterator. Call
// stop if next is not called until exhaustion.
func (i MapIterator[K, V]) Pull() (next func() (K, V, bool), stop func()) {
	return iter.Pull2(iter.Seq2[K, V](i))
}

// Reduce returns startValue for an empty iterator. Otherwise it calls accumulator
// for each pair and returns the last result; accumulator does not receive the
// preceding result.
func (i MapIterator[K, V]) Reduce(startValue V, accumulator func(k K, v V) V) V {
	curr := startValue
	for k, v := range i {
		curr = accumulator(k, v)
	}
	return curr
}

// Reverse returns an iterator over the pairs in reverse order.
//
// This reads all pairs before returning the new iterator and may be slow for
// large inputs.
func (i MapIterator[K, V]) Reverse() MapIterator[K, V] {
	keys, values := i.CollectKeysAndValues()
	slices.Reverse(keys)
	slices.Reverse(values)
	return func(yield func(K, V) bool) {
		for i, k := range keys {
			if !yield(k, values[i]) {
				return
			}
		}
	}
}

// Skip returns an iterator that discards the first n pairs. A nonpositive n
// discards nothing. The returned iterator retains its remaining skip count across
// traversals.
func (i MapIterator[K, V]) Skip[N constraints.Signed | constraints.Unsigned](n N) MapIterator[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range i {
			if n > 0 {
				n--
				continue
			}

			if !yield(k, v) {
				return
			}
		}
	}
}

// SkipWhile returns an iterator that discards the leading pairs matching
// matchFunc, then yields all remaining pairs.
func (i MapIterator[K, V]) SkipWhile(matchFunc func(k K, v V) bool) MapIterator[K, V] {
	return func(yield func(K, V) bool) {
		wasSkipped := false

		for k, v := range i {
			if !wasSkipped && matchFunc(k, v) {
				continue
			}

			wasSkipped = true

			if !yield(k, v) {
				return
			}
		}
	}
}

// Sorted returns an iterator over pairs sorted by key using compareFunc. The sort
// is not guaranteed to be stable. If a key repeats, each occurrence yields the
// last value seen for that key.
//
// This reads all pairs before returning the new iterator and may be slow for
// large inputs.
func (i MapIterator[K, V]) Sorted(compareFunc func(a, b K) int) MapIterator[K, V] {
	var keys []K
	data := make(map[K]V)
	for k, v := range i {
		keys = append(keys, k)
		data[k] = v
	}
	slices.SortFunc(keys, compareFunc)
	return func(yield func(K, V) bool) {
		for _, k := range keys {
			if !yield(k, data[k]) {
				return
			}
		}
	}
}

// SortedStable returns an iterator over pairs sorted by key using compareFunc,
// preserving the order of keys that compare equal. If a key repeats, each
// occurrence yields the last value seen for that key.
//
// This reads all pairs before returning the new iterator and may be slow for
// large inputs.
func (i MapIterator[K, V]) SortedStable(compareFunc func(a, b K) int) MapIterator[K, V] {
	var keys []K
	data := make(map[K]V)
	for k, v := range i {
		keys = append(keys, k)
		data[k] = v
	}
	slices.SortStableFunc(keys, compareFunc)
	return func(yield func(K, V) bool) {
		for _, k := range keys {
			if !yield(k, data[k]) {
				return
			}
		}
	}
}

// TakeWhile returns an iterator over the leading pairs matching matchFunc. It
// stops at the first nonmatching pair.
func (i MapIterator[K, V]) TakeWhile(matchFunc func(k K, v V) bool) MapIterator[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range i {
			if !matchFunc(k, v) {
				return
			}

			if !yield(k, v) {
				return
			}
		}
	}
}

// ToSeq2 returns the iterator as an iter.Seq2.
func (i MapIterator[K, V]) ToSeq2() iter.Seq2[K, V] {
	return iter.Seq2[K, V](i)
}

// Values returns an iterator over the values in pair order, including values for
// repeated keys.
func (i MapIterator[K, V]) Values() Iterator[V] {
	return func(yield func(V) bool) {
		for _, v := range i {
			if !yield(v) {
				return
			}
		}
	}
}
