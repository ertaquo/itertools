package itertools_test

import (
	"cmp"
	"errors"
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ertaquo/itertools"
)

func TestIteratorConversion(t *testing.T) {
	type numbers []int
	assert.Equal(t, []int{1, 2}, itertools.ToIterator(numbers{1, 2}).Collect(), "ToIterator")
	var seq = slices.Values([]int{3, 4})
	assert.Equal(t, []int{3, 4}, itertools.SeqToIterator(seq).Collect(), "SeqToIterator")
	assert.Equal(t, []int{5, 6}, slices.Collect(itertools.ToIterator([]int{5, 6}).ToSeq()), "ToSeq")
}

func TestIteratorPredicates(t *testing.T) {
	even := func(v int) bool { return v%2 == 0 }
	for _, tt := range []struct {
		name     string
		values   []int
		all, any bool
	}{
		{"empty", nil, true, false},
		{"all", []int{2, 4}, true, true},
		{"mixed", []int{1, 2}, false, true},
		{"none", []int{1, 3}, false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			i := itertools.ToIterator(tt.values)
			if got := i.All(even); got != tt.all {
				t.Errorf("All = %v, want %v", got, tt.all)
			}
			if got := i.Any(even); got != tt.any {
				t.Errorf("Any = %v, want %v", got, tt.any)
			}
			if got := i.ContainsFunc(even); got != tt.any {
				t.Errorf("ContainsFunc = %v, want %v", got, tt.any)
			}
		})
	}
}

func TestIteratorCollect(t *testing.T) {
	type numbers []int
	i := itertools.ToIterator([]int{3, 1, 2})
	assert.Equal(t, []int{3, 1, 2}, i.Collect(), "Collect")
	assert.EqualValues(t, []int{3, 1, 2}, i.CollectAs[numbers](), "CollectAs")
	if got := itertools.ToIterator([]int(nil)).Collect(); len(got) != 0 {
		t.Errorf("Collect(empty) = %v, want empty", got)
	}
	assert.Equal(t, []int{1, 2, 3}, i.CollectSorted(cmp.Compare[int]), "CollectSorted")
	assert.EqualValues(t, []int{1, 2, 3}, i.CollectSortedAs[numbers](cmp.Compare[int]), "CollectSortedAs")
	type item struct{ key, order int }
	items := itertools.ToIterator([]item{{2, 0}, {1, 1}, {2, 2}})
	less := func(a, b item) int { return cmp.Compare(a.key, b.key) }
	want := []item{{1, 1}, {2, 0}, {2, 2}}
	if got := items.CollectSortedStable(less); !slices.Equal(got, want) {
		t.Errorf("CollectSortedStable = %v, want %v", got, want)
	}
	type namedItems []item
	if got := items.CollectSortedStableAs[namedItems](less); !slices.Equal(got, want) {
		t.Errorf("CollectSortedStableAs = %v, want %v", got, want)
	}
}

func TestIteratorContains(t *testing.T) {
	i := itertools.ToIterator([]int{1, 2})
	if !i.Contains(2) || i.Contains(3) {
		t.Error("Contains gave wrong membership")
	}
	defer func() {
		if recover() == nil {
			t.Error("Contains(noncomparable) did not panic")
		}
	}()
	itertools.ToIterator([][]int{{1}}).Contains([]int{1})
}

func TestIteratorCountEqual(t *testing.T) {
	if got := itertools.ToIterator([]int{1, 2, 3}).Count(); got != 3 {
		t.Errorf("Count = %d, want 3", got)
	}
	if got := itertools.ToIterator([]int(nil)).Count(); got != 0 {
		t.Errorf("Count(empty) = %d", got)
	}
	for _, tt := range []struct {
		a, b []int
		want bool
	}{
		{nil, nil, true},
		{[]int{1, 2}, []int{1, 2}, true},
		{[]int{1, 2}, []int{2, 1}, false},
		{[]int{1}, []int{1, 2}, false},
		{[]int{1, 2}, []int{1}, false},
	} {
		if got := itertools.ToIterator(tt.a).Equal(itertools.ToIterator(tt.b), cmp.Compare[int]); got != tt.want {
			t.Errorf("Equal(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestIteratorFilter(t *testing.T) {
	i := itertools.ToIterator([]int{1, 2, 3, 4})
	even := func(v int) bool { return v%2 == 0 }
	assert.Equal(t, []int{2, 4}, i.Filter(even).Collect(), "Filter")
	assert.Equal(t, []int{2, 4}, i.FilterAndCollect(even), "FilterAndCollect")
	assert.Equal(t, []int{2, 4}, slices.Sorted(slices.Values(i.FilterAndCollect(even, itertools.WithParallelLimit(2)))), "FilterAndCollect (parallel)")
	stopped := 0
	for range i.Filter(even) {
		stopped++
		break
	}
	if stopped != 1 {
		t.Errorf("Filter early stop yielded %d values", stopped)
	}
}

func TestIteratorFilterWithError(t *testing.T) {
	bad := errors.New("bad value")
	f := func(v int) (bool, error) {
		if v == 3 {
			return false, bad
		}
		return v%2 == 0, nil
	}
	i := itertools.ToIterator([]int{1, 2, 3, 4})
	got, err := i.FilterAndCollectWithError(f)
	assert.Equal(t, []int{2}, got, "FilterAndCollectWithError partial")
	if !errors.Is(err, bad) {
		t.Errorf("FilterAndCollectWithError error = %v", err)
	}
	got, err = i.FilterAndCollectWithError(f, itertools.WithParallelLimit(1))
	if !errors.Is(err, bad) {
		t.Errorf("FilterAndCollectWithError (parallel) error = %v", err)
	}
	if !slices.Contains(got, 2) || slices.Contains(got, 1) {
		t.Errorf("FilterAndCollectWithError (parallel) = %v", got)
	}
	got, err = i.FilterAndCollectWithError(func(v int) (bool, error) { return v%2 == 0, nil }, itertools.WithParallel())
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(got)
	assert.Equal(t, []int{2, 4}, got, "FilterAndCollectWithError (parallel)")
}

func TestIteratorFilterMap(t *testing.T) {
	for _, tt := range []struct {
		name   string
		values []string
		want   []int
	}{
		{"nil", nil, nil},
		{"empty", []string{}, nil},
		{"single match", []string{"0"}, []int{0}},
		{"single rejected", []string{"oops"}, nil},
		{"all match", []string{"3", "-1", "3", "0"}, []int{3, -1, 3, 0}},
		{"none match", []string{"oops", "bad"}, nil},
		{"mixed", []string{"bad", "3", "oops", "-1", "3", "0", "bad"}, []int{3, -1, 3, 0}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var called []string
			i := itertools.ToIterator(tt.values).FilterMap(func(value string) (int, bool) {
				called = append(called, value)
				n, err := strconv.Atoi(value)
				if err != nil {
					return -1, false
				}
				return n, true
			})
			for pass := range 2 {
				called = nil
				if got := i.Collect(); !slices.Equal(got, tt.want) {
					t.Errorf("FilterMap(%q), pass %d = %v, want %v", tt.values, pass, got, tt.want)
				}
				if !slices.Equal(called, tt.values) {
					t.Errorf("FilterMap callback values = %q, want %q", called, tt.values)
				}
			}
		})
	}
}

func TestIteratorFilterMapRange(t *testing.T) {
	var visited, called []string
	i := itertools.Iterator[string](func(yield func(string) bool) {
		for _, value := range []string{"bad", "2", "oops", "4", "6"} {
			visited = append(visited, value)
			if !yield(value) {
				return
			}
		}
	}).FilterMap(func(value string) (int, bool) {
		called = append(called, value)
		n, err := strconv.Atoi(value)
		return n, err == nil
	})
	if len(visited) != 0 || len(called) != 0 {
		t.Fatalf("FilterMap before iteration visited %q and made callback calls %q, want neither", visited, called)
	}

	var got []int
	for value := range i {
		got = append(got, value)
		break
	}
	if want := []int{2}; !slices.Equal(got, want) {
		t.Errorf("FilterMap after break = %v, want %v", got, want)
	}
	want := []string{"bad", "2"}
	if !slices.Equal(visited, want) {
		t.Errorf("FilterMap source values after break = %q, want %q", visited, want)
	}
	if !slices.Equal(called, want) {
		t.Errorf("FilterMap callback values after break = %q, want %q", called, want)
	}
}

func TestIteratorFind(t *testing.T) {
	i := itertools.ToIterator([]int{1, 2, 3})
	even := func(v int) bool { return v%2 == 0 }
	missing := func(v int) bool { return v == 9 }
	{
		got := i.Find(even)
		if assert.NotNil(t, got, "Find") {
			assert.Equal(t, 2, *got, "Find")
		}
	}
	if got := i.Find(missing); got != nil {
		t.Errorf("Find(missing) = %v", got)
	}
	if got := i.FindOr(even, 9); got != 2 {
		t.Errorf("FindOr = %d", got)
	}
	if got := i.FindOr(missing, 9); got != 9 {
		t.Errorf("FindOr(missing) = %d", got)
	}
	if got := i.FindOrNone(missing); got != 0 {
		t.Errorf("FindOrNone(missing) = %d", got)
	}
}

func TestIteratorFirstLastGet(t *testing.T) {
	i := itertools.ToIterator([]int{4, 5, 6})
	empty := itertools.ToIterator([]int(nil))
	{
		got := i.First()
		if assert.NotNil(t, got, "First") {
			assert.Equal(t, 4, *got, "First")
		}
	}
	{
		got := i.Last()
		if assert.NotNil(t, got, "Last") {
			assert.Equal(t, 6, *got, "Last")
		}
	}
	{
		got := i.Get(1)
		if assert.NotNil(t, got, "Get") {
			assert.Equal(t, 5, *got, "Get")
		}
	}
	if empty.First() != nil || empty.Last() != nil || i.Get(-1) != nil || i.Get(3) != nil {
		t.Error("missing element did not return nil")
	}
	if got := i.FirstOr(9); got != 4 {
		t.Errorf("FirstOr = %d", got)
	}
	if got := empty.FirstOr(9); got != 9 {
		t.Errorf("FirstOr(empty) = %d", got)
	}
	if got := empty.FirstOrNone(); got != 0 {
		t.Errorf("FirstOrNone(empty) = %d", got)
	}
	if got := i.FirstOrNone(); got != 4 {
		t.Errorf("FirstOrNone = %d", got)
	}
	if got := i.LastOr(9); got != 6 {
		t.Errorf("LastOr = %d", got)
	}
	if got := empty.LastOr(9); got != 9 {
		t.Errorf("LastOr(empty) = %d", got)
	}
	if got := empty.LastOrNone(); got != 0 {
		t.Errorf("LastOrNone(empty) = %d", got)
	}
	if got := i.LastOrNone(); got != 6 {
		t.Errorf("LastOrNone = %d", got)
	}
	if got := i.GetOr(1, 9); got != 5 {
		t.Errorf("GetOr = %d", got)
	}
	if got := i.GetOr(3, 9); got != 9 {
		t.Errorf("GetOr(missing) = %d", got)
	}
	if got := i.GetOrNone(2); got != 6 {
		t.Errorf("GetOrNone = %d", got)
	}
	if got := i.GetOrNone(3); got != 0 {
		t.Errorf("GetOrNone(missing) = %d", got)
	}
}

func TestIteratorIndexed(t *testing.T) {
	type pair struct {
		index, value int
	}
	for _, tt := range []struct {
		name   string
		values []int
		want   []pair
	}{
		{"nil", nil, nil},
		{"empty", []int{}, nil},
		{"single zero", []int{0}, []pair{{0, 0}}},
		{"single nonzero", []int{5}, []pair{{0, 5}}},
		{"mixed", []int{3, -1, 3, 0, 2}, []pair{{0, 3}, {1, -1}, {2, 3}, {3, 0}, {4, 2}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			i := itertools.ToIterator(tt.values).Indexed()
			for pass := range 2 {
				var got []pair
				for index, value := range i {
					got = append(got, pair{index, value})
				}
				if !slices.Equal(got, tt.want) {
					t.Errorf("Indexed(%v), pass %d = %v, want %v", tt.values, pass, got, tt.want)
				}
			}
		})
	}
}

func TestIteratorIndexedRange(t *testing.T) {
	visited := 0
	i := itertools.Iterator[int](func(yield func(int) bool) {
		for _, v := range []int{10, 20, 30} {
			visited++
			if !yield(v) {
				return
			}
		}
	}).Indexed()
	if visited != 0 {
		t.Fatalf("Indexed visited %d values before iteration, want 0", visited)
	}

	var got [][2]int
	for index, value := range i {
		got = append(got, [2]int{index, value})
		if len(got) == 2 {
			break
		}
	}
	if want := [][2]int{{0, 10}, {1, 20}}; !slices.Equal(got, want) {
		t.Errorf("Indexed after break = %v, want %v", got, want)
	}
	if visited != 2 {
		t.Errorf("Indexed visited %d values after break, want 2", visited)
	}

	got = nil
	for index, value := range i {
		got = append(got, [2]int{index, value})
	}
	if want := [][2]int{{0, 10}, {1, 20}, {2, 30}}; !slices.Equal(got, want) {
		t.Errorf("Indexed after restart = %v, want %v", got, want)
	}
}

func TestIteratorForEach(t *testing.T) {
	bad := errors.New("stop")
	i := itertools.ToIterator([]int{1, 2, 3})
	var got []int
	err := i.ForEach(func(v int) error {
		got = append(got, v)
		if v == 2 {
			return bad
		}
		return nil
	})
	assert.Equal(t, []int{1, 2}, got, "ForEach visits")
	if !errors.Is(err, bad) {
		t.Errorf("ForEach error = %v", err)
	}
	if err := i.ForEach(func(int) error { return nil }); err != nil {
		t.Errorf("ForEach success = %v", err)
	}
	var mu sync.Mutex
	got = nil
	err = i.ForEach(func(v int) error { mu.Lock(); got = append(got, v); mu.Unlock(); return nil }, itertools.WithParallelLimit(2))
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(got)
	assert.Equal(t, []int{1, 2, 3}, got, "ForEach (parallel)")
	if err := i.ForEach(func(v int) error {
		if v == 2 {
			return bad
		}
		return nil
	}, itertools.WithParallel()); !errors.Is(err, bad) {
		t.Errorf("ForEach (parallel) error = %v", err)
	}
}

func TestIteratorConcatLimit(t *testing.T) {
	i := itertools.ToIterator([]int{1, 2}).Concat(itertools.ToIterator([]int{3, 4}))
	assert.Equal(t, []int{1, 2, 3, 4}, i.Collect(), "Concat")
	assert.Equal(t, []int{1, 2}, i.Limit(2).Collect(), "Limit")
	assert.Empty(t, i.Limit(0).Collect(), "Limit(0)")
	assert.Empty(t, i.Limit(-1).Collect(), "Limit(-1)")
	assert.Equal(t, []int{1, 2, 3, 4}, i.Limit(10).Collect(), "Limit(large)")
}

func TestIteratorMap(t *testing.T) {
	i := itertools.ToIterator([]int{1, 2, 3})
	double := func(v int) int { return v * 2 }
	assert.Equal(t, []int{2, 4, 6}, i.Map(double).Collect(), "Map")
	assert.Equal(t, []int{2, 4, 6}, i.MapAndCollect(double), "MapAndCollect")
	got := i.MapAndCollect(double, itertools.WithParallelLimit(2))
	slices.Sort(got)
	assert.Equal(t, []int{2, 4, 6}, got, "MapAndCollect (parallel)")
}

func TestIteratorMapWithError(t *testing.T) {
	bad := errors.New("bad value")
	f := func(v int) (int, error) {
		if v == 3 {
			return 0, bad
		}
		return v * 2, nil
	}
	i := itertools.ToIterator([]int{1, 2, 3})
	got, err := i.MapAndCollectWithError(f)
	assert.Equal(t, []int{2, 4}, got, "MapAndCollectWithError partial")
	if !errors.Is(err, bad) {
		t.Errorf("MapAndCollectWithError error = %v", err)
	}
	_, err = i.MapAndCollectWithError(f, itertools.WithParallelLimit(1))
	if !errors.Is(err, bad) {
		t.Errorf("MapAndCollectWithError (parallel) error = %v", err)
	}
	got, err = i.MapAndCollectWithError(func(v int) (int, error) { return v * 2, nil }, itertools.WithParallel())
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(got)
	assert.Equal(t, []int{2, 4, 6}, got, "MapAndCollectWithError (parallel)")
}

func TestIteratorExtrema(t *testing.T) {
	i := itertools.ToIterator([]int{3, 1, 4, 2})
	empty := itertools.ToIterator([]int(nil))
	{
		got := i.Min(cmp.Compare[int])
		if assert.NotNil(t, got, "Min") {
			assert.Equal(t, 1, *got, "Min")
		}
	}
	{
		got := i.Max(cmp.Compare[int])
		if assert.NotNil(t, got, "Max") {
			assert.Equal(t, 4, *got, "Max")
		}
	}
	min, max := i.MinMax(cmp.Compare[int])
	{
		got := min
		if assert.NotNil(t, got, "MinMax min") {
			assert.Equal(t, 1, *got, "MinMax min")
		}
	}
	{
		got := max
		if assert.NotNil(t, got, "MinMax max") {
			assert.Equal(t, 4, *got, "MinMax max")
		}
	}
	if empty.Min(cmp.Compare[int]) != nil || empty.Max(cmp.Compare[int]) != nil {
		t.Error("empty extrema should be nil")
	}
	min, max = empty.MinMax(cmp.Compare[int])
	if min != nil || max != nil {
		t.Errorf("MinMax(empty) = %v, %v", min, max)
	}
	if got := i.MinOr(cmp.Compare[int], 9); got != 1 {
		t.Errorf("MinOr = %d", got)
	}
	if got := empty.MinOr(cmp.Compare[int], 9); got != 9 {
		t.Errorf("MinOr(empty) = %d", got)
	}
	if got := i.MinOrNone(cmp.Compare[int]); got != 1 {
		t.Errorf("MinOrNone = %d", got)
	}
	if got := empty.MinOrNone(cmp.Compare[int]); got != 0 {
		t.Errorf("MinOrNone(empty) = %d", got)
	}
	if got := i.MaxOr(cmp.Compare[int], 9); got != 4 {
		t.Errorf("MaxOr = %d", got)
	}
	if got := empty.MaxOr(cmp.Compare[int], 9); got != 9 {
		t.Errorf("MaxOr(empty) = %d", got)
	}
	if got := i.MaxOrNone(cmp.Compare[int]); got != 4 {
		t.Errorf("MaxOrNone = %d", got)
	}
	if got := empty.MaxOrNone(cmp.Compare[int]); got != 0 {
		t.Errorf("MaxOrNone(empty) = %d", got)
	}
	a, b := i.MinMaxOr(cmp.Compare[int], 9)
	if a != 1 || b != 4 {
		t.Errorf("MinMaxOr = %d, %d", a, b)
	}
	a, b = empty.MinMaxOr(cmp.Compare[int], 9)
	if a != 9 || b != 9 {
		t.Errorf("MinMaxOr(empty) = %d, %d", a, b)
	}
	a, b = i.MinMaxOrNone(cmp.Compare[int])
	if a != 1 || b != 4 {
		t.Errorf("MinMaxOrNone = %d, %d", a, b)
	}
	a, b = empty.MinMaxOrNone(cmp.Compare[int])
	if a != 0 || b != 0 {
		t.Errorf("MinMaxOrNone(empty) = %d, %d", a, b)
	}
}

func TestIteratorPull(t *testing.T) {
	next, stop := itertools.ToIterator([]int{1, 2}).Pull()
	defer stop()
	if v, ok := next(); !ok || v != 1 {
		t.Errorf("first Pull = %d, %v", v, ok)
	}
	if v, ok := next(); !ok || v != 2 {
		t.Errorf("second Pull = %d, %v", v, ok)
	}
	if v, ok := next(); ok || v != 0 {
		t.Errorf("exhausted Pull = %d, %v", v, ok)
	}
}

func TestIteratorReduce(t *testing.T) {
	// Reduce currently passes each element to the accumulator, without a running value.
	if got := itertools.ToIterator([]int{1, 2, 3}).Reduce(9, func(v int) int { return v * 2 }); got != 6 {
		t.Errorf("Reduce = %d, want 6", got)
	}
	if got := itertools.ToIterator([]int(nil)).Reduce(9, func(v int) int { return v * 2 }); got != 9 {
		t.Errorf("Reduce(empty) = %d, want 9", got)
	}
}

func TestIteratorReduceWithError(t *testing.T) {
	bad := errors.New("bad value")
	for _, tt := range []struct {
		name       string
		values     []int
		start      int
		failAt     int
		want       int
		wantErr    error
		wantValues []int
	}{
		{"nil", nil, 9, 0, 9, nil, nil},
		{"empty", []int{}, 9, 0, 9, nil, nil},
		{"zero start", nil, 0, 0, 0, nil, nil},
		{"single", []int{5}, 9, 0, 10, nil, []int{5}},
		{"multiple", []int{3, -1, 3, 0, 2}, 9, 0, 4, nil, []int{3, -1, 3, 0, 2}},
		{"zero result", []int{3, 0}, 9, 0, 0, nil, []int{3, 0}},
		{"error first", []int{1, 2, 3}, 9, 1, 2, bad, []int{1}},
		{"error middle", []int{1, 2, 3}, 9, 2, 4, bad, []int{1, 2}},
		{"error last", []int{1, 2, 3}, 9, 3, 6, bad, []int{1, 2, 3}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var visited, called []int
			i := itertools.Iterator[int](func(yield func(int) bool) {
				for _, v := range tt.values {
					visited = append(visited, v)
					if !yield(v) {
						return
					}
				}
			})
			got, err := i.ReduceWithError(tt.start, func(v int) (int, error) {
				called = append(called, v)
				if len(called) == tt.failAt {
					return v * 2, bad
				}
				return v * 2, nil
			})
			if got != tt.want || err != tt.wantErr {
				t.Errorf("ReduceWithError(%v, %d) = %d, %v, want %d, %v", tt.values, tt.start, got, err, tt.want, tt.wantErr)
			}
			if !slices.Equal(called, tt.wantValues) {
				t.Errorf("ReduceWithError accumulator values = %v, want %v", called, tt.wantValues)
			}
			if !slices.Equal(visited, tt.wantValues) {
				t.Errorf("ReduceWithError source values = %v, want %v", visited, tt.wantValues)
			}
		})
	}
}

func TestIteratorRepeat(t *testing.T) {
	for _, tt := range []struct {
		name   string
		values []int
		n      int
		want   []int
	}{
		{"nil", nil, 3, nil},
		{"empty", []int{}, 3, nil},
		{"negative", []int{1, 2}, -1, nil},
		{"zero", []int{1, 2}, 0, nil},
		{"once", []int{1, 2}, 1, []int{1, 2}},
		{"twice", []int{1, 2}, 2, []int{1, 2, 1, 2}},
		{"single", []int{5}, 3, []int{5, 5, 5}},
		{"duplicates", []int{0, -1, -1}, 3, []int{0, -1, -1, 0, -1, -1, 0, -1, -1}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var got []int
			for v := range itertools.ToIterator(tt.values).Repeat(tt.n) {
				got = append(got, v)
				if len(got) > len(tt.want) {
					break
				}
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("Repeat(%v, %d) = %v, want %v", tt.values, tt.n, got, tt.want)
			}
		})
	}

	t.Run("unsigned", func(t *testing.T) {
		got := itertools.ToIterator([]int{1, 2}).Repeat(uint8(2)).Collect()
		want := []int{1, 2, 1, 2}
		if !slices.Equal(got, want) {
			t.Errorf("Repeat with unsigned count = %v, want %v", got, want)
		}
	})
}

func TestIteratorRepeatRange(t *testing.T) {
	for _, tt := range []struct {
		name        string
		n, stop     int
		want        []int
		wantVisited int
	}{
		{"negative", -1, 1, nil, 0},
		{"zero", 0, 1, nil, 0},
		{"break in source", 3, 1, []int{1}, 1},
		{"break in replay", 3, 4, []int{1, 2, 3, 1}, 3},
		{"complete", 2, 7, []int{1, 2, 3, 1, 2, 3}, 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			visited := 0
			i := itertools.Iterator[int](func(yield func(int) bool) {
				for _, v := range []int{1, 2, 3} {
					visited++
					if !yield(v) {
						return
					}
				}
			}).Repeat(tt.n)
			if visited != 0 {
				t.Fatalf("Repeat visited %d values before iteration, want 0", visited)
			}
			var got []int
			for v := range i {
				got = append(got, v)
				if len(got) == tt.stop {
					break
				}
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("Repeat(%d) after iteration = %v, want %v", tt.n, got, tt.want)
			}
			if visited != tt.wantVisited {
				t.Errorf("Repeat(%d) visited %d source values, want %d", tt.n, visited, tt.wantVisited)
			}
		})
	}
}

func TestIteratorReverseSkip(t *testing.T) {
	i := itertools.ToIterator([]int{1, 2, 3, 4})
	assert.Equal(t, []int{4, 3, 2, 1}, i.Reverse().Collect(), "Reverse")
	assert.Equal(t, []int{3, 4}, i.Skip(2).Collect(), "Skip")
	assert.Equal(t, []int{1, 2, 3, 4}, i.Skip(0).Collect(), "Skip(0)")
	assert.Empty(t, i.Skip(10).Collect(), "Skip(large)")
	assert.Equal(t, []int{3, 1}, itertools.ToIterator([]int{1, 2, 3, 1}).SkipWhile(func(v int) bool { return v < 3 }).Collect(), "SkipWhile")
}

func TestIteratorShuffle(t *testing.T) {
	for _, tt := range []struct {
		name   string
		values []int
	}{
		{"nil", nil},
		{"empty", []int{}},
		{"single", []int{5}},
		{"all equal", []int{2, 2, 2}},
		{"distinct", []int{5, 1, 4, 2, 3}},
		{"duplicates", []int{-1, 2, 0, 2, -1}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			values := slices.Clone(tt.values)
			shuffled := itertools.ToIterator(values).Shuffle()
			if !slices.Equal(values, tt.values) {
				t.Errorf("Shuffle modified source = %v, want %v", values, tt.values)
			}
			got := shuffled.Collect()
			want := slices.Clone(tt.values)
			slices.Sort(got)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Errorf("Shuffle(%v) sorted = %v, want %v", tt.values, got, want)
			}
		})
	}
}

func TestIteratorShuffleRange(t *testing.T) {
	values := []int{1, 2, 3, 4}
	visited := 0
	i := itertools.Iterator[int](func(yield func(int) bool) {
		for _, v := range values {
			visited++
			if !yield(v) {
				return
			}
		}
	}).Shuffle()
	if visited != len(values) {
		t.Fatalf("Shuffle before iteration visited %d values, want %d", visited, len(values))
	}
	want := i.Collect()
	if len(want) != len(values) {
		t.Fatalf("Shuffle yielded %d values, want %d", len(want), len(values))
	}

	var got []int
	for v := range i {
		got = append(got, v)
		if len(got) == 2 {
			break
		}
	}
	if !slices.Equal(got, want[:2]) {
		t.Errorf("Shuffle after break = %v, want %v", got, want[:2])
	}
	if got := i.Collect(); !slices.Equal(got, want) {
		t.Errorf("Shuffle after restart = %v, want %v", got, want)
	}
	if visited != len(values) {
		t.Errorf("Shuffle after iteration visited %d source values, want %d", visited, len(values))
	}
}

func TestIteratorSorted(t *testing.T) {
	i := itertools.ToIterator([]int{3, 1, 2})
	assert.Equal(t, []int{1, 2, 3}, i.Sorted(cmp.Compare[int]).Collect(), "Sorted")
	type item struct{ key, order int }
	items := itertools.ToIterator([]item{{2, 0}, {1, 1}, {2, 2}})
	got := items.SortedStable(func(a, b item) int { return cmp.Compare(a.key, b.key) }).Collect()
	want := []item{{1, 1}, {2, 0}, {2, 2}}
	if !slices.Equal(got, want) {
		t.Errorf("SortedStable = %v, want %v", got, want)
	}
}

func TestIteratorTakeWhileUnique(t *testing.T) {
	assert.Equal(t, []int{1, 2}, itertools.ToIterator([]int{1, 2, 3, 1}).TakeWhile(func(v int) bool { return v < 3 }).Collect(), "TakeWhile")
	assert.Equal(t, []int{2, 1, 3}, itertools.ToIterator([]int{2, 1, 2, 3, 1}).Unique(cmp.Compare[int]).Collect(), "Unique")
}

func TestIteratorZip(t *testing.T) {
	values := itertools.ToIterator([]int{10, 20})
	keys := itertools.ToIterator([]string{"a", "b", "c"})
	got := make(map[string]int)
	for k, v := range values.Zip(keys) {
		got[k] = v
	}
	want := map[string]int{"a": 10, "b": 20}
	if len(got) != len(want) || got["a"] != 10 || got["b"] != 20 {
		t.Errorf("Zip = %v, want %v", got, want)
	}
	if n := itertools.ToIterator([]int(nil)).Zip(keys).Count(); n != 0 {
		t.Errorf("Zip(empty) count = %d", n)
	}
}
