package itertools

import (
	"cmp"
	"iter"
	"slices"
	"sync"

	"golang.org/x/exp/constraints"
	"golang.org/x/sync/errgroup"
)

type OrderedIterator[V cmp.Ordered] func(yield func(V) bool)

func ToOrderedIterator[Slice ~[]V, V cmp.Ordered](s Slice) OrderedIterator[V] {
	return OrderedIterator[V](slices.Values(s))
}

func SeqToOrderedIterator[V cmp.Ordered](s iter.Seq[V]) OrderedIterator[V] {
	return OrderedIterator[V](s)
}

func (i OrderedIterator[V]) CollectSorted() []V {
	return slices.SortedFunc(iter.Seq[V](i), cmp.Compare)
}

func (i OrderedIterator[V]) CollectSortedAs[Slice ~[]V]() Slice {
	return slices.SortedFunc(iter.Seq[V](i), cmp.Compare)
}

func (i OrderedIterator[V]) CollectSortedStable() []V {
	return slices.SortedStableFunc(iter.Seq[V](i), cmp.Compare)
}

func (i OrderedIterator[V]) CollectSortedStableAs[Slice ~[]V]() Slice {
	return slices.SortedStableFunc(iter.Seq[V](i), cmp.Compare)
}

func (i OrderedIterator[V]) CollectSortedFunc(compareFunc func(a, b V) int) []V {
	return slices.SortedFunc(iter.Seq[V](i), compareFunc)
}

func (i OrderedIterator[V]) CollectSortedFuncAs[Slice ~[]V](compareFunc func(a, b V) int) Slice {
	return slices.SortedFunc(iter.Seq[V](i), compareFunc)
}

func (i OrderedIterator[V]) CollectSortedStableFunc(compareFunc func(a, b V) int) []V {
	return slices.SortedStableFunc(iter.Seq[V](i), compareFunc)
}

func (i OrderedIterator[V]) CollectSortedStableFuncAs[Slice ~[]V](compareFunc func(a, b V) int) Slice {
	return slices.SortedStableFunc(iter.Seq[V](i), compareFunc)
}

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

func (i OrderedIterator[V]) Max() *V {
	return i.MaxFunc(cmp.Compare)
}

func (i OrderedIterator[V]) MaxOr(defaultValue V) V {
	if max := i.Max(); max != nil {
		return *max
	}
	return defaultValue
}

func (i OrderedIterator[V]) MaxOrNone() V {
	if max := i.Max(); max != nil {
		return *max
	}

	var none V
	return none
}

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

func (i OrderedIterator[V]) MaxFuncOr(compareFunc func(a, b V) int, defaultValue V) V {
	if max := i.MaxFunc(compareFunc); max != nil {
		return *max
	}
	return defaultValue
}

func (i OrderedIterator[V]) MaxFuncOrNone(compareFunc func(a, b V) int) V {
	if max := i.MaxFunc(compareFunc); max != nil {
		return *max
	}

	var none V
	return none
}

func (i OrderedIterator[V]) Min() *V {
	return i.MinFunc(cmp.Compare)
}

func (i OrderedIterator[V]) MinOr(defaultValue V) V {
	if min := i.Min(); min != nil {
		return *min
	}

	return defaultValue
}

func (i OrderedIterator[V]) MinOrNone() V {
	if min := i.Min(); min != nil {
		return *min
	}

	var none V
	return none
}

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

func (i OrderedIterator[V]) MinFuncOr(compareFunc func(a, b V) int, defaultValue V) V {
	if min := i.MinFunc(compareFunc); min != nil {
		return *min
	}

	return defaultValue
}

func (i OrderedIterator[V]) MinFuncOrNone(compareFunc func(a, b V) int) V {
	if min := i.MinFunc(compareFunc); min != nil {
		return *min
	}

	var none V
	return none
}

func (i OrderedIterator[V]) MinMax() (min *V, max *V) {
	return i.MinMaxFunc(cmp.Compare)
}

func (i OrderedIterator[V]) MinMaxOr(defaultValue V) (min V, max V) {
	if minP, maxP := i.MinMax(); minP != nil && maxP != nil {
		return *minP, *maxP
	}

	return defaultValue, defaultValue
}

func (i OrderedIterator[V]) MinMaxOrNone() (min V, max V) {
	if minP, maxP := i.MinMax(); minP != nil && maxP != nil {
		return *minP, *maxP
	}

	var none V
	return none, none
}

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

func (i OrderedIterator[V]) MinMaxFuncOr(compareFunc func(a, b V) int, defaultValue V) (min V, max V) {
	if minP, maxP := i.MinMaxFunc(compareFunc); minP != nil && maxP != nil {
		return *minP, *maxP
	}

	return defaultValue, defaultValue
}

func (i OrderedIterator[V]) MinMaxFuncOrNone(compareFunc func(a, b V) int) (min V, max V) {
	if minP, maxP := i.MinMaxFunc(compareFunc); minP != nil && maxP != nil {
		return *minP, *maxP
	}

	var none V
	return none, none
}

func (i OrderedIterator[V]) Sorted() OrderedIterator[V] {
	return ToOrderedIterator(i.CollectSorted())
}

func (i OrderedIterator[V]) SortedStable() OrderedIterator[V] {
	return ToOrderedIterator(i.CollectSortedStable())
}

func (i OrderedIterator[V]) SortedFunc(compareFunc func(a, b V) int) OrderedIterator[V] {
	return ToOrderedIterator(i.CollectSortedFunc(compareFunc))
}

func (i OrderedIterator[V]) SortedStableFunc(compareFunc func(a, b V) int) OrderedIterator[V] {
	return ToOrderedIterator(i.CollectSortedStableFunc(compareFunc))
}

func (i OrderedIterator[V]) ToComparableIterator() ComparableIterator[V] {
	return ComparableIterator[V](i)
}

// --- From ComparableIterator[V] ---

func (i OrderedIterator[V]) Contains(value V) bool {
	for v := range i {
		if v == value {
			return true
		}
	}

	return false
}

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

func (i OrderedIterator[V]) ToIterator() Iterator[V] {
	return Iterator[V](i)
}

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

func (i OrderedIterator[V]) All(matchFunc func(v V) bool) bool {
	for v := range i {
		if !matchFunc(v) {
			return false
		}
	}

	return true
}

func (i OrderedIterator[V]) Any(matchFunc func(v V) bool) bool {
	return i.ContainsFunc(matchFunc)
}

func (i OrderedIterator[V]) Collect() []V {
	return slices.Collect(iter.Seq[V](i))
}

func (i OrderedIterator[V]) CollectAs[Slice ~[]V]() Slice {
	return slices.Collect(iter.Seq[V](i))
}

func (i OrderedIterator[V]) ContainsFunc(matchFunc func(v V) bool) bool {
	for v := range i {
		if matchFunc(v) {
			return true
		}
	}

	return false
}

func (i OrderedIterator[V]) Count() int {
	var count int
	for range i {
		count++
	}
	return count
}

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

	wg.Wait()
	return result
}

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

func (i OrderedIterator[V]) Find(matchFunc func(v V) bool) *V {
	for v := range i {
		if matchFunc(v) {
			return &v
		}
	}

	return nil
}

func (i OrderedIterator[V]) FindOr(matchFunc func(v V) bool, defaultValue V) V {
	if v := i.Find(matchFunc); v != nil {
		return *v
	}

	return defaultValue
}

func (i OrderedIterator[V]) FindOrNone(matchFunc func(v V) bool) V {
	if v := i.Find(matchFunc); v != nil {
		return *v
	}

	var none V
	return none
}

func (i OrderedIterator[V]) First() *V {
	for v := range i {
		return &v
	}

	return nil
}

func (i OrderedIterator[V]) FirstOr(defaultValue V) V {
	if v := i.First(); v != nil {
		return *v
	}

	return defaultValue
}

func (i OrderedIterator[V]) FirstOrNone() V {
	if v := i.First(); v != nil {
		return *v
	}

	var none V
	return none
}

func (i OrderedIterator[V]) ForEach(callback func(v V) error) error {
	for v := range i {
		if err := callback(v); err != nil {
			return err
		}
	}

	return nil
}

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

func (i OrderedIterator[V]) GetOr[N constraints.Signed | constraints.Unsigned](n N, defaultValue V) V {
	if v := i.Get(n); v != nil {
		return *v
	}

	return defaultValue
}

func (i OrderedIterator[V]) GetOrNone[N constraints.Signed | constraints.Unsigned](n N) V {
	if v := i.Get(n); v != nil {
		return *v
	}

	var none V
	return none
}

func (i OrderedIterator[V]) Last() *V {
	var last *V
	for v := range i {
		last = &v
	}

	return last
}

func (i OrderedIterator[V]) LastOr(defaultValue V) V {
	if v := i.Last(); v != nil {
		return *v
	}

	return defaultValue
}

func (i OrderedIterator[V]) LastOrNone() V {
	if v := i.Last(); v != nil {
		return *v
	}

	var none V
	return none
}

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

func (i OrderedIterator[V]) Map[V2 any](mapFunc func(value V) V2) Iterator[V2] {
	return func(yield func(V2) bool) {
		for v := range i {
			if !yield(mapFunc(v)) {
				return
			}
		}
	}
}

func (i OrderedIterator[V]) MapAndCollect[V2 any](mapFunc func(value V) V2) []V2 {
	var result []V2
	for v := range i {
		result = append(result, mapFunc(v))
	}
	return result
}

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

	wg.Wait()
	return result
}

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

func (i OrderedIterator[V]) Pull() (next func() (V, bool), stop func()) {
	return iter.Pull(iter.Seq[V](i))
}

func (i OrderedIterator[V]) Reduce(startValue V, accumulator func(v V) V) V {
	curr := startValue
	for v := range i {
		curr = accumulator(v)
	}
	return curr
}

func (i OrderedIterator[V]) Reverse() Iterator[V] {
	s := i.Collect()
	slices.Reverse(s)
	return ToIterator(s)
}

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

func (i OrderedIterator[V]) ToSeq() iter.Seq[V] {
	return iter.Seq[V](i)
}
