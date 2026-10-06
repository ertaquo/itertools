package itertools_test

import (
	"cmp"
	"errors"
	"maps"
	"math"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ertaquo/itertools"
)

func TestComparableIteratorConversion(t *testing.T) {
	type numbers []int
	assert.Equal(t, []int{1, 2}, itertools.ToComparableIterator(numbers{1, 2}).Collect(), "ToComparableIterator")
	var seq = slices.Values([]int{3, 4})
	assert.Equal(t, []int{3, 4}, itertools.SeqToComparableIterator(seq).Collect(), "SeqToComparableIterator")
	assert.Equal(t, []int{5, 6}, slices.Collect(itertools.ToComparableIterator([]int{5, 6}).ToSeq()), "ToSeq")
}

func TestComparableIteratorPredicates(t *testing.T) {
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
			i := itertools.ToComparableIterator(tt.values)
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

func TestComparableIteratorCollect(t *testing.T) {
	type numbers []int
	i := itertools.ToComparableIterator([]int{3, 1, 2})
	assert.Equal(t, []int{3, 1, 2}, i.Collect(), "Collect")
	assert.EqualValues(t, []int{3, 1, 2}, i.CollectAs[numbers](), "CollectAs")
	if got := itertools.ToComparableIterator([]int(nil)).Collect(); len(got) != 0 {
		t.Errorf("Collect(empty) = %v, want empty", got)
	}
	assert.Equal(t, []int{1, 2, 3}, i.CollectSorted(cmp.Compare[int]), "CollectSorted")
	assert.EqualValues(t, []int{1, 2, 3}, i.CollectSortedAs[numbers](cmp.Compare[int]), "CollectSortedAs")
	type item struct{ key, order int }
	items := itertools.ToComparableIterator([]item{{2, 0}, {1, 1}, {2, 2}})
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

func TestComparableIteratorContains(t *testing.T) {
	i := itertools.ToComparableIterator([]int{1, 2})
	if !i.Contains(2) || i.Contains(3) {
		t.Error("Contains gave wrong membership")
	}
}

func TestComparableIteratorCountEqual(t *testing.T) {
	if got := itertools.ToComparableIterator([]int{1, 2, 3}).Count(); got != 3 {
		t.Errorf("Count = %d, want 3", got)
	}
	if got := itertools.ToComparableIterator([]int(nil)).Count(); got != 0 {
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
		if got := itertools.ToComparableIterator(tt.a).Equal(itertools.ToComparableIterator(tt.b)); got != tt.want {
			t.Errorf("Equal(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
	if !itertools.ToComparableIterator([]int{1, 2}).Equal(itertools.ToIterator([]int{1, 2})) {
		t.Error("Equal with Iterator returned false for equal values")
	}
}

func TestComparableIteratorEqualFunc(t *testing.T) {
	foldCompare := func(a, b string) int {
		return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
	}
	for _, tt := range []struct {
		name        string
		left, right []string
		want        bool
	}{
		{"empty", nil, nil, true},
		{"equal ignoring case", []string{"Go", "Lang"}, []string{"go", "LANG"}, true},
		{"different value", []string{"Go"}, []string{"Rust"}, false},
		{"shorter right", []string{"Go", "Lang"}, []string{"go"}, false},
		{"shorter left", []string{"Go"}, []string{"go", "Lang"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			left := itertools.ToComparableIterator(tt.left)
			if got := left.EqualFunc(itertools.ToIterator(tt.right), foldCompare); got != tt.want {
				t.Errorf("EqualFunc(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.want)
			}
		})
	}
	if !itertools.ToComparableIterator([]string{"Go"}).EqualFunc(
		itertools.ToComparableIterator([]string{"go"}), foldCompare,
	) {
		t.Error("EqualFunc with ComparableIterator returned false")
	}
}

func TestComparableIteratorFilter(t *testing.T) {
	i := itertools.ToComparableIterator([]int{1, 2, 3, 4})
	even := func(v int) bool { return v%2 == 0 }
	assert.Equal(t, []int{2, 4}, i.Filter(even).Collect(), "Filter")
	assert.Equal(t, []int{2, 4}, i.FilterAndCollect(even), "FilterAndCollect")
	assert.Equal(t, []int{2, 4}, slices.Sorted(slices.Values(i.FilterAndCollectParallel(even, itertools.WithLimit(2)))), "FilterAndCollectParallel")
	stopped := 0
	for range i.Filter(even) {
		stopped++
		break
	}
	if stopped != 1 {
		t.Errorf("Filter early stop yielded %d values", stopped)
	}
}

func TestComparableIteratorFilterNone(t *testing.T) {
	for _, tt := range []struct {
		name         string
		values, want []int
	}{
		{"nil", nil, nil},
		{"empty", []int{}, nil},
		{"single zero", []int{0}, nil},
		{"single nonzero", []int{-1}, []int{-1}},
		{"all zero", []int{0, 0, 0}, nil},
		{"no zero", []int{3, -1, 3, 2}, []int{3, -1, 3, 2}},
		{"mixed", []int{0, 3, 0, -1, 3, 0, 2, 0}, []int{3, -1, 3, 2}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := itertools.ToComparableIterator(tt.values).FilterNone().Collect()
			if !slices.Equal(got, tt.want) {
				t.Errorf("FilterNone(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}

	t.Run("strings", func(t *testing.T) {
		values := []string{"", "Go", "", " ", "Go", ""}
		want := []string{"Go", " ", "Go"}
		if got := itertools.ToComparableIterator(values).FilterNone().Collect(); !slices.Equal(got, want) {
			t.Errorf("FilterNone(%q) = %q, want %q", values, got, want)
		}
	})

	t.Run("floats", func(t *testing.T) {
		values := []float64{0, -1.5, math.Copysign(0, -1), math.NaN(), math.Inf(1), 2.5, 0}
		want := []float64{-1.5, math.NaN(), math.Inf(1), 2.5}
		equal := func(a, b float64) bool { return a == b || math.IsNaN(a) && math.IsNaN(b) }
		if got := itertools.ToComparableIterator(values).FilterNone().Collect(); !slices.EqualFunc(got, want, equal) {
			t.Errorf("FilterNone(%v) = %v, want %v", values, got, want)
		}
	})
}

func TestComparableIteratorFilterNoneRange(t *testing.T) {
	visited := 0
	i := itertools.ComparableIterator[int](func(yield func(int) bool) {
		for _, v := range []int{0, 0, 5, 0, 7} {
			visited++
			if !yield(v) {
				return
			}
		}
	}).FilterNone()
	if visited != 0 {
		t.Fatalf("FilterNone visited %d values before iteration, want 0", visited)
	}

	var got []int
	for v := range i {
		got = append(got, v)
		break
	}
	if want := []int{5}; !slices.Equal(got, want) {
		t.Errorf("FilterNone after break = %v, want %v", got, want)
	}
	if visited != 3 {
		t.Errorf("FilterNone visited %d values after break, want 3", visited)
	}

	if got := i.Collect(); !slices.Equal(got, []int{5, 7}) {
		t.Errorf("FilterNone after restart = %v, want [5 7]", got)
	}
}

func TestComparableIteratorFilterWithError(t *testing.T) {
	bad := errors.New("bad value")
	f := func(v int) (bool, error) {
		if v == 3 {
			return false, bad
		}
		return v%2 == 0, nil
	}
	i := itertools.ToComparableIterator([]int{1, 2, 3, 4})
	got, err := i.FilterAndCollectWithError(f)
	assert.Equal(t, []int{2}, got, "FilterAndCollectWithError partial")
	if !errors.Is(err, bad) {
		t.Errorf("FilterAndCollectWithError error = %v", err)
	}
	got, err = i.FilterAndCollectWithErrorParallel(f, itertools.WithLimit(1))
	if !errors.Is(err, bad) {
		t.Errorf("FilterAndCollectWithErrorParallel error = %v", err)
	}
	if !slices.Contains(got, 2) || slices.Contains(got, 1) {
		t.Errorf("FilterAndCollectWithErrorParallel = %v", got)
	}
	got, err = i.FilterAndCollectWithErrorParallel(func(v int) (bool, error) { return v%2 == 0, nil })
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(got)
	assert.Equal(t, []int{2, 4}, got, "FilterAndCollectWithErrorParallel")
}

func TestComparableIteratorFind(t *testing.T) {
	i := itertools.ToComparableIterator([]int{1, 2, 3})
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

func TestComparableIteratorFirstLastGet(t *testing.T) {
	i := itertools.ToComparableIterator([]int{4, 5, 6})
	empty := itertools.ToComparableIterator([]int(nil))
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

func TestComparableIteratorIndexed(t *testing.T) {
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
			i := itertools.ToComparableIterator(tt.values).Indexed()
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

func TestComparableIteratorIndexedRange(t *testing.T) {
	visited := 0
	i := itertools.ComparableIterator[int](func(yield func(int) bool) {
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

func TestComparableIteratorForEach(t *testing.T) {
	bad := errors.New("stop")
	i := itertools.ToComparableIterator([]int{1, 2, 3})
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
	err = i.ForEachParallel(func(v int) error { mu.Lock(); got = append(got, v); mu.Unlock(); return nil }, itertools.WithLimit(2))
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(got)
	assert.Equal(t, []int{1, 2, 3}, got, "ForEachParallel")
	if err := i.ForEachParallel(func(v int) error {
		if v == 2 {
			return bad
		}
		return nil
	}); !errors.Is(err, bad) {
		t.Errorf("ForEachParallel error = %v", err)
	}
}

func TestComparableIteratorConcatLimit(t *testing.T) {
	i := itertools.ToComparableIterator([]int{1, 2}).Concat(itertools.ToComparableIterator([]int{3, 4}))
	assert.Equal(t, []int{1, 2, 3, 4}, i.Collect(), "Concat")
	assert.Equal(t, []int{1, 2, 3, 4}, itertools.ToComparableIterator([]int{1, 2}).Concat(itertools.ToIterator([]int{3, 4})).Collect(), "Concat Iterator")
	assert.Equal(t, []int{1, 2}, i.Limit(2).Collect(), "Limit")
	assert.Empty(t, i.Limit(0).Collect(), "Limit(0)")
	assert.Empty(t, i.Limit(-1).Collect(), "Limit(-1)")
	assert.Equal(t, []int{1, 2, 3, 4}, i.Limit(10).Collect(), "Limit(large)")
}

func TestComparableIteratorMap(t *testing.T) {
	i := itertools.ToComparableIterator([]int{1, 2, 3})
	double := func(v int) int { return v * 2 }
	assert.Equal(t, []int{2, 4, 6}, i.Map(double).Collect(), "Map")
	assert.Equal(t, []int{2, 4, 6}, i.MapAndCollect(double), "MapAndCollect")
	got := i.MapAndCollectParallel(double, itertools.WithLimit(2))
	slices.Sort(got)
	assert.Equal(t, []int{2, 4, 6}, got, "MapAndCollectParallel")
}

func TestComparableIteratorMapWithError(t *testing.T) {
	bad := errors.New("bad value")
	f := func(v int) (int, error) {
		if v == 3 {
			return 0, bad
		}
		return v * 2, nil
	}
	i := itertools.ToComparableIterator([]int{1, 2, 3})
	got, err := i.MapAndCollectWithError(f)
	assert.Equal(t, []int{2, 4}, got, "MapAndCollectWithError partial")
	if !errors.Is(err, bad) {
		t.Errorf("MapAndCollectWithError error = %v", err)
	}
	_, err = i.MapAndCollectWithErrorParallel(f, itertools.WithLimit(1))
	if !errors.Is(err, bad) {
		t.Errorf("MapAndCollectWithErrorParallel error = %v", err)
	}
	got, err = i.MapAndCollectWithErrorParallel(func(v int) (int, error) { return v * 2, nil })
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(got)
	assert.Equal(t, []int{2, 4, 6}, got, "MapAndCollectWithErrorParallel")
}

func TestComparableIteratorExtrema(t *testing.T) {
	i := itertools.ToComparableIterator([]int{3, 1, 4, 2})
	empty := itertools.ToComparableIterator([]int(nil))
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

func TestComparableIteratorPull(t *testing.T) {
	next, stop := itertools.ToComparableIterator([]int{1, 2}).Pull()
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

func TestComparableIteratorReduce(t *testing.T) {
	// Reduce currently passes each element to the accumulator, without a running value.
	if got := itertools.ToComparableIterator([]int{1, 2, 3}).Reduce(9, func(v int) int { return v * 2 }); got != 6 {
		t.Errorf("Reduce = %d, want 6", got)
	}
	if got := itertools.ToComparableIterator([]int(nil)).Reduce(9, func(v int) int { return v * 2 }); got != 9 {
		t.Errorf("Reduce(empty) = %d, want 9", got)
	}
}

func TestComparableIteratorReduceWithError(t *testing.T) {
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
			i := itertools.ComparableIterator[int](func(yield func(int) bool) {
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

func TestComparableIteratorReverseSkip(t *testing.T) {
	i := itertools.ToComparableIterator([]int{1, 2, 3, 4})
	assert.Equal(t, []int{4, 3, 2, 1}, i.Reverse().Collect(), "Reverse")
	assert.Equal(t, []int{3, 4}, i.Skip(2).Collect(), "Skip")
	assert.Equal(t, []int{1, 2, 3, 4}, i.Skip(0).Collect(), "Skip(0)")
	assert.Empty(t, i.Skip(10).Collect(), "Skip(large)")
	assert.Equal(t, []int{3, 1}, itertools.ToComparableIterator([]int{1, 2, 3, 1}).SkipWhile(func(v int) bool { return v < 3 }).Collect(), "SkipWhile")
}

func TestComparableIteratorShuffle(t *testing.T) {
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
			shuffled := itertools.ToComparableIterator(values).Shuffle()
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

func TestComparableIteratorShuffleRange(t *testing.T) {
	values := []int{1, 2, 3, 4}
	visited := 0
	i := itertools.ComparableIterator[int](func(yield func(int) bool) {
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

func TestComparableIteratorSorted(t *testing.T) {
	i := itertools.ToComparableIterator([]int{3, 1, 2})
	assert.Equal(t, []int{1, 2, 3}, i.Sorted(cmp.Compare[int]).Collect(), "Sorted")
	type item struct{ key, order int }
	items := itertools.ToComparableIterator([]item{{2, 0}, {1, 1}, {2, 2}})
	got := items.SortedStable(func(a, b item) int { return cmp.Compare(a.key, b.key) }).Collect()
	want := []item{{1, 1}, {2, 0}, {2, 2}}
	if !slices.Equal(got, want) {
		t.Errorf("SortedStable = %v, want %v", got, want)
	}
}

func TestComparableIteratorTakeWhileUnique(t *testing.T) {
	assert.Equal(t, []int{1, 2}, itertools.ToComparableIterator([]int{1, 2, 3, 1}).TakeWhile(func(v int) bool { return v < 3 }).Collect(), "TakeWhile")
	assert.Equal(t, []int{2, 1, 3}, itertools.ToComparableIterator([]int{2, 1, 2, 3, 1}).Unique().Collect(), "Unique")
}

func TestComparableIteratorGroupBy(t *testing.T) {
	for _, tt := range []struct {
		name   string
		values []int
		want   map[bool][]int
	}{
		{"nil", nil, nil},
		{"empty", []int{}, nil},
		{"single", []int{3}, map[bool][]int{false: {3}}},
		{"one group", []int{4, 2, 4}, map[bool][]int{true: {4, 2, 4}}},
		{"mixed", []int{3, 2, 3, 0, -1, 4}, map[bool][]int{false: {3, 3, -1}, true: {2, 0, 4}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var called []int
			groups := itertools.ToComparableIterator(tt.values).GroupBy(func(v int) bool {
				called = append(called, v)
				return v%2 == 0
			})
			got := make(map[bool][]int)
			for key, group := range groups {
				if _, ok := got[key]; ok {
					t.Errorf("GroupBy yielded key %v more than once", key)
				}
				got[key] = group
			}
			if !maps.EqualFunc(got, tt.want, slices.Equal[[]int]) {
				t.Errorf("GroupBy(%v) = %v, want %v", tt.values, got, tt.want)
			}
			if !slices.Equal(called, tt.values) {
				t.Errorf("GroupBy key function values = %v, want %v", called, tt.values)
			}
		})
	}
}

func TestComparableIteratorGroupByRange(t *testing.T) {
	values := []int{3, 2, 3, 0, -1}
	visited := 0
	i := itertools.ComparableIterator[int](func(yield func(int) bool) {
		for _, v := range values {
			visited++
			if !yield(v) {
				return
			}
		}
	})
	var called []int
	groups := i.GroupBy(func(v int) bool {
		called = append(called, v)
		return v%2 == 0
	})
	if visited != len(values) || !slices.Equal(called, values) {
		t.Fatalf("GroupBy before iteration visited %d values and made key calls %v, want %d, %v", visited, called, len(values), values)
	}

	want := map[bool][]int{false: {3, 3, -1}, true: {2, 0}}
	yielded := 0
	for key, group := range groups {
		yielded++
		if visited != len(values) || !slices.Equal(called, values) {
			t.Errorf("GroupBy before yielding visited %d values and made key calls %v, want %d, %v", visited, called, len(values), values)
		}
		if !slices.Equal(group, want[key]) {
			t.Errorf("GroupBy group %v = %v, want %v", key, group, want[key])
		}
		break
	}
	if yielded != 1 {
		t.Errorf("GroupBy after break yielded %d groups, want 1", yielded)
	}
}

func TestComparableIteratorUniqueFunc(t *testing.T) {
	i := itertools.ToComparableIterator([]string{"Go", "Rust", "go", "RUST", "C"})
	got := i.UniqueFunc(func(a, b string) int {
		return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
	}).Collect()
	want := []string{"Go", "Rust", "C"}
	if !slices.Equal(got, want) {
		t.Errorf("UniqueFunc = %v, want %v", got, want)
	}
}

func TestComparableIteratorToIterator(t *testing.T) {
	i := itertools.ToComparableIterator([]int{1, 2, 3})
	var converted = i.ToIterator()
	assert.Equal(t, []int{1, 2, 3}, converted.Collect(), "ToIterator")
}

func TestComparableIteratorZip(t *testing.T) {
	values := itertools.ToComparableIterator([]int{10, 20})
	keys := itertools.ToComparableIterator([]string{"a", "b", "c"})
	got := make(map[string]int)
	for k, v := range values.Zip(keys) {
		got[k] = v
	}
	want := map[string]int{"a": 10, "b": 20}
	if len(got) != len(want) || got["a"] != 10 || got["b"] != 20 {
		t.Errorf("Zip = %v, want %v", got, want)
	}
	if n := itertools.ToComparableIterator([]int(nil)).Zip(keys).Count(); n != 0 {
		t.Errorf("Zip(empty) count = %d", n)
	}
	var pairs []struct {
		key   string
		value int
	}
	for k, v := range values.Zip(itertools.ToIterator([]string{"x"})) {
		pairs = append(pairs, struct {
			key   string
			value int
		}{k, v})
	}
	if want := []struct {
		key   string
		value int
	}{{"x", 10}}; !slices.Equal(pairs, want) {
		t.Errorf("Zip with Iterator keys = %v, want %v", pairs, want)
	}
}
