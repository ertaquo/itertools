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

type MapIterator[K comparable, V any] func(yield func(K, V) bool)

func ToMapIterator[Map ~map[K]V, K comparable, V any](m Map) MapIterator[K, V] {
	return MapIterator[K, V](maps.All(m))
}

func Seq2ToMapIterator[K comparable, V any](s iter.Seq2[K, V]) MapIterator[K, V] {
	return MapIterator[K, V](s)
}

func (i MapIterator[K, V]) All(matchFunc func(k K, v V) bool) bool {
	for k, v := range i {
		if !matchFunc(k, v) {
			return false
		}
	}

	return true
}

func (i MapIterator[K, V]) Any(matchFunc func(k K, v V) bool) bool {
	return i.ContainsFunc(matchFunc)
}

func (i MapIterator[K, V]) Collect() map[K]V {
	return maps.Collect(iter.Seq2[K, V](i))
}

func (i MapIterator[K, V]) CollectAs[Map ~map[K]V]() Map {
	return maps.Collect(iter.Seq2[K, V](i))
}

func (i MapIterator[K, V]) CollectKeys() []K {
	var keys []K
	for k, _ := range i {
		keys = append(keys, k)
	}
	return keys
}

func (i MapIterator[K, V]) CollectKeysAndValues() (keys []K, values []V) {
	for k, v := range i {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}

func (i MapIterator[K, V]) CollectKeysAndValuesAs[Keys ~[]K, Values ~[]V]() (keys Keys, values Values) {
	return i.CollectKeysAndValues()
}

func (i MapIterator[K, V]) CollectKeysAs[Slice ~[]K]() Slice {
	return i.CollectKeys()
}

func (i MapIterator[K, V]) CollectValues() []V {
	var values []V
	for _, v := range i {
		values = append(values, v)
	}
	return values
}

func (i MapIterator[K, V]) CollectValuesAs[Slice ~[]V]() Slice {
	return i.CollectValues()
}

func (i MapIterator[K, V]) ContainsFunc(matchFunc func(k K, v V) bool) bool {
	for k, v := range i {
		if matchFunc(k, v) {
			return true
		}
	}

	return false
}

func (i MapIterator[K, V]) ContainsKey(key K) bool {
	for k, _ := range i {
		if k == key {
			return true
		}
	}

	return false
}

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

func (i MapIterator[K, V]) Count() int {
	var count int
	for range i {
		count++
	}
	return count
}

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

func (i MapIterator[K, V]) FilterAndCollect(filterFunc func(key K, value V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range i {
		if filterFunc(k, v) {
			result[k] = v
		}
	}
	return result
}

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

func (i MapIterator[K, V]) Find(matchFunc func(k K, v V) bool) (*K, *V) {
	for k, v := range i {
		if matchFunc(k, v) {
			return &k, &v
		}
	}
	return nil, nil
}

func (i MapIterator[K, V]) FindKey(matchFunc func(k K, v V) bool) *K {
	for k, v := range i {
		if matchFunc(k, v) {
			return &k
		}
	}
	return nil
}

func (i MapIterator[K, V]) FindKeyOr(matchFunc func(k K, v V) bool, defaultKey K) K {
	if k := i.FindKey(matchFunc); k != nil {
		return *k
	}

	return defaultKey
}

func (i MapIterator[K, V]) FindKeyOrNone(matchFunc func(k K, v V) bool) K {
	if k := i.FindKey(matchFunc); k != nil {
		return *k
	}

	var none K
	return none
}

func (i MapIterator[K, V]) FindOr(matchFunc func(k K, v V) bool, defaultKey K, defaultValue V) (K, V) {
	if k, v := i.Find(matchFunc); v != nil {
		return *k, *v
	}

	return defaultKey, defaultValue
}

func (i MapIterator[K, V]) FindOrNone(matchFunc func(k K, v V) bool) (K, V) {
	if k, v := i.Find(matchFunc); v != nil {
		return *k, *v
	}

	var noneK K
	var noneV V
	return noneK, noneV
}

func (i MapIterator[K, V]) FindValue(matchFunc func(k K, v V) bool) *V {
	for k, v := range i {
		if matchFunc(k, v) {
			return &v
		}
	}
	return nil
}

func (i MapIterator[K, V]) FindValueOr(matchFunc func(k K, v V) bool, defaultValue V) V {
	if v := i.FindValue(matchFunc); v != nil {
		return *v
	}

	return defaultValue
}

func (i MapIterator[K, V]) FindValueOrNone(matchFunc func(k K, v V) bool) V {
	if v := i.FindValue(matchFunc); v != nil {
		return *v
	}

	var none V
	return none
}

func (i MapIterator[K, V]) ForEach(callback func(k K, v V) error) error {
	for k, v := range i {
		if err := callback(k, v); err != nil {
			return err
		}
	}

	return nil
}

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

func (i MapIterator[K, V]) Get(key K) *V {
	for k, v := range i {
		if k == key {
			return &v
		}
	}

	return nil
}

func (i MapIterator[K, V]) GetOr(key K, defaultValue V) V {
	if v := i.Get(key); v != nil {
		return *v
	}

	return defaultValue
}

func (i MapIterator[K, V]) GetOrNone(key K) V {
	if v := i.Get(key); v != nil {
		return *v
	}

	var none V
	return none
}

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

func (i MapIterator[K, V]) Keys() ComparableIterator[K] {
	return func(yield func(K) bool) {
		for k := range i {
			if !yield(k) {
				return
			}
		}
	}
}

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

func (i MapIterator[K, V]) Map[K2 comparable, V2 any](mapFunc func(key K, value V) (K2, V2)) MapIterator[K2, V2] {
	return func(yield func(K2, V2) bool) {
		for k, v := range i {
			if !yield(mapFunc(k, v)) {
				return
			}
		}
	}
}

func (i MapIterator[K, V]) MapAndCollect[K2 comparable, V2 any](mapFunc func(key K, value V) (K2, V2)) map[K2]V2 {
	result := make(map[K2]V2)
	for k, v := range i {
		k2, v2 := mapFunc(k, v)
		result[k2] = v2
	}
	return result
}

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

func (i MapIterator[K, V]) MapKeys[K2 comparable](mapFunc func(key K) K2) MapIterator[K2, V] {
	return func(yield func(K2, V) bool) {
		for k, v := range i {
			if !yield(mapFunc(k), v) {
				return
			}
		}
	}
}

func (i MapIterator[K, V]) MapValues[V2 any](mapFunc func(value V) V2) MapIterator[K, V2] {
	return func(yield func(K, V2) bool) {
		for k, v := range i {
			if !yield(k, mapFunc(v)) {
				return
			}
		}
	}
}

func (i MapIterator[K, V]) Pull() (next func() (K, V, bool), stop func()) {
	return iter.Pull2(iter.Seq2[K, V](i))
}

func (i MapIterator[K, V]) Reduce(startValue V, accumulator func(k K, v V) V) V {
	curr := startValue
	for k, v := range i {
		curr = accumulator(k, v)
	}
	return curr
}

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

func (i MapIterator[K, V]) ToSeq2() iter.Seq2[K, V] {
	return iter.Seq2[K, V](i)
}

func (i MapIterator[K, V]) Values() Iterator[V] {
	return func(yield func(V) bool) {
		for _, v := range i {
			if !yield(v) {
				return
			}
		}
	}
}
