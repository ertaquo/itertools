package itertools_test

import (
	"cmp"
	"errors"
	"math"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ertaquo/itertools"
)

func TestOrderedIteratorConversion(t *testing.T) {
	type numbers []int
	assert.Equal(t, []int{1, 2}, itertools.ToOrderedIterator(numbers{1, 2}).Collect(), "ToOrderedIterator")
	var seq = slices.Values([]int{3, 4})
	assert.Equal(t, []int{3, 4}, itertools.SeqToOrderedIterator(seq).Collect(), "SeqToOrderedIterator")
	assert.Equal(t, []int{5, 6}, slices.Collect(itertools.ToOrderedIterator([]int{5, 6}).ToSeq()), "ToSeq")
}

func TestOrderedIteratorPredicates(t *testing.T) {
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
			i := itertools.ToOrderedIterator(tt.values)
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

func TestOrderedIteratorCollect(t *testing.T) {
	type numbers []int
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	assert.Equal(t, []int{3, 1, 2}, i.Collect(), "Collect")
	assert.EqualValues(t, []int{3, 1, 2}, i.CollectAs[numbers](), "CollectAs")
	if got := itertools.ToOrderedIterator([]int(nil)).Collect(); len(got) != 0 {
		t.Errorf("Collect(empty) = %v, want empty", got)
	}
	assert.Equal(t, []int{1, 2, 3}, i.CollectSorted(), "CollectSorted")
	assert.EqualValues(t, []int{1, 2, 3}, i.CollectSortedAs[numbers](), "CollectSortedAs")
}

func TestOrderedIteratorCollectSortedVariants(t *testing.T) {
	type numbers []int
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	assert.Equal(t, []int{1, 2, 3}, i.CollectSortedStable(), "CollectSortedStable")
	assert.EqualValues(t, []int{1, 2, 3}, i.CollectSortedStableAs[numbers](), "CollectSortedStableAs")
	reverse := func(a, b int) int { return cmp.Compare(b, a) }
	assert.Equal(t, []int{3, 2, 1}, i.CollectSortedFunc(reverse), "CollectSortedFunc")
	assert.EqualValues(t, []int{3, 2, 1}, i.CollectSortedFuncAs[numbers](reverse), "CollectSortedFuncAs")
	words := itertools.ToOrderedIterator([]string{"b-first", "a-only", "b-second"})
	byInitial := func(a, b string) int { return cmp.Compare(a[0], b[0]) }
	want := []string{"a-only", "b-first", "b-second"}
	if got := words.CollectSortedStableFunc(byInitial); !slices.Equal(got, want) {
		t.Errorf("CollectSortedStableFunc = %v, want %v", got, want)
	}
	type namedWords []string
	if got := words.CollectSortedStableFuncAs[namedWords](byInitial); !slices.Equal(got, want) {
		t.Errorf("CollectSortedStableFuncAs = %v, want %v", got, want)
	}
}

func TestOrderedIteratorContains(t *testing.T) {
	i := itertools.ToOrderedIterator([]int{1, 2})
	if !i.Contains(2) || i.Contains(3) {
		t.Error("Contains gave wrong membership")
	}
}

func TestOrderedIteratorCountEqual(t *testing.T) {
	if got := itertools.ToOrderedIterator([]int{1, 2, 3}).Count(); got != 3 {
		t.Errorf("Count = %d, want 3", got)
	}
	if got := itertools.ToOrderedIterator([]int(nil)).Count(); got != 0 {
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
		if got := itertools.ToOrderedIterator(tt.a).Equal(itertools.ToOrderedIterator(tt.b)); got != tt.want {
			t.Errorf("Equal(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
	left := itertools.ToOrderedIterator([]int{1, 2})
	if !left.Equal(itertools.ToIterator([]int{1, 2})) || !left.Equal(itertools.ToComparableIterator([]int{1, 2})) {
		t.Error("Equal with Iterator or ComparableIterator returned false")
	}
}

func TestOrderedIteratorEqualFunc(t *testing.T) {
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
			left := itertools.ToOrderedIterator(tt.left)
			if got := left.EqualFunc(itertools.ToIterator(tt.right), foldCompare); got != tt.want {
				t.Errorf("EqualFunc(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.want)
			}
		})
	}
	left := itertools.ToOrderedIterator([]string{"Go"})
	if !left.EqualFunc(itertools.ToComparableIterator([]string{"go"}), foldCompare) ||
		!left.EqualFunc(itertools.ToOrderedIterator([]string{"go"}), foldCompare) {
		t.Error("EqualFunc with ComparableIterator or OrderedIterator returned false")
	}
}

func TestOrderedIteratorFilter(t *testing.T) {
	i := itertools.ToOrderedIterator([]int{1, 2, 3, 4})
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

func TestOrderedIteratorFilterNone(t *testing.T) {
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
			got := itertools.ToOrderedIterator(tt.values).FilterNone().Collect()
			if !slices.Equal(got, tt.want) {
				t.Errorf("FilterNone(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}

	t.Run("strings", func(t *testing.T) {
		values := []string{"", "Go", "", " ", "Go", ""}
		want := []string{"Go", " ", "Go"}
		if got := itertools.ToOrderedIterator(values).FilterNone().Collect(); !slices.Equal(got, want) {
			t.Errorf("FilterNone(%q) = %q, want %q", values, got, want)
		}
	})

	t.Run("floats", func(t *testing.T) {
		values := []float64{0, -1.5, math.Copysign(0, -1), math.NaN(), math.Inf(1), 2.5, 0}
		want := []float64{-1.5, math.NaN(), math.Inf(1), 2.5}
		equal := func(a, b float64) bool { return a == b || math.IsNaN(a) && math.IsNaN(b) }
		if got := itertools.ToOrderedIterator(values).FilterNone().Collect(); !slices.EqualFunc(got, want, equal) {
			t.Errorf("FilterNone(%v) = %v, want %v", values, got, want)
		}
	})
}

func TestOrderedIteratorFilterNoneRange(t *testing.T) {
	visited := 0
	i := itertools.OrderedIterator[int](func(yield func(int) bool) {
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

func TestOrderedIteratorFilterWithError(t *testing.T) {
	bad := errors.New("bad value")
	f := func(v int) (bool, error) {
		if v == 3 {
			return false, bad
		}
		return v%2 == 0, nil
	}
	i := itertools.ToOrderedIterator([]int{1, 2, 3, 4})
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

func TestOrderedIteratorFind(t *testing.T) {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
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

func TestOrderedIteratorFirstLastGet(t *testing.T) {
	i := itertools.ToOrderedIterator([]int{4, 5, 6})
	empty := itertools.ToOrderedIterator([]int(nil))
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

func TestOrderedIteratorIndexed(t *testing.T) {
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
			i := itertools.ToOrderedIterator(tt.values).Indexed()
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

func TestOrderedIteratorIndexedRange(t *testing.T) {
	visited := 0
	i := itertools.OrderedIterator[int](func(yield func(int) bool) {
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

func TestOrderedIteratorForEach(t *testing.T) {
	bad := errors.New("stop")
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
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

func TestOrderedIteratorConcatLimit(t *testing.T) {
	i := itertools.ToOrderedIterator([]int{1, 2}).Concat(itertools.ToOrderedIterator([]int{3, 4}))
	assert.Equal(t, []int{1, 2, 3, 4}, i.Collect(), "Concat")
	assert.Equal(t, []int{1, 2, 3, 4}, itertools.ToOrderedIterator([]int{1, 2}).Concat(itertools.ToIterator([]int{3, 4})).Collect(), "Concat Iterator")
	assert.Equal(t, []int{1, 2, 3, 4}, itertools.ToOrderedIterator([]int{1, 2}).Concat(itertools.ToComparableIterator([]int{3, 4})).Collect(), "Concat ComparableIterator")
	assert.Equal(t, []int{1, 2}, i.Limit(2).Collect(), "Limit")
	assert.Empty(t, i.Limit(0).Collect(), "Limit(0)")
	assert.Empty(t, i.Limit(-1).Collect(), "Limit(-1)")
	assert.Equal(t, []int{1, 2, 3, 4}, i.Limit(10).Collect(), "Limit(large)")
}

func TestOrderedIteratorMap(t *testing.T) {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	double := func(v int) int { return v * 2 }
	assert.Equal(t, []int{2, 4, 6}, i.Map(double).Collect(), "Map")
	assert.Equal(t, []int{2, 4, 6}, i.MapAndCollect(double), "MapAndCollect")
	got := i.MapAndCollectParallel(double, itertools.WithLimit(2))
	slices.Sort(got)
	assert.Equal(t, []int{2, 4, 6}, got, "MapAndCollectParallel")
}

func TestOrderedIteratorMapWithError(t *testing.T) {
	bad := errors.New("bad value")
	f := func(v int) (int, error) {
		if v == 3 {
			return 0, bad
		}
		return v * 2, nil
	}
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
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

func TestOrderedIteratorExtrema(t *testing.T) {
	i := itertools.ToOrderedIterator([]int{3, 1, 4, 2})
	empty := itertools.ToOrderedIterator([]int(nil))
	{
		got := i.Min()
		if assert.NotNil(t, got, "Min") {
			assert.Equal(t, 1, *got, "Min")
		}
	}
	{
		got := i.Max()
		if assert.NotNil(t, got, "Max") {
			assert.Equal(t, 4, *got, "Max")
		}
	}
	min, max := i.MinMax()
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
	if empty.Min() != nil || empty.Max() != nil {
		t.Error("empty extrema should be nil")
	}
	min, max = empty.MinMax()
	if min != nil || max != nil {
		t.Errorf("MinMax(empty) = %v, %v", min, max)
	}
	if got := i.MinOr(9); got != 1 {
		t.Errorf("MinOr = %d", got)
	}
	if got := empty.MinOr(9); got != 9 {
		t.Errorf("MinOr(empty) = %d", got)
	}
	if got := i.MinOrNone(); got != 1 {
		t.Errorf("MinOrNone = %d", got)
	}
	if got := empty.MinOrNone(); got != 0 {
		t.Errorf("MinOrNone(empty) = %d", got)
	}
	if got := i.MaxOr(9); got != 4 {
		t.Errorf("MaxOr = %d", got)
	}
	if got := empty.MaxOr(9); got != 9 {
		t.Errorf("MaxOr(empty) = %d", got)
	}
	if got := i.MaxOrNone(); got != 4 {
		t.Errorf("MaxOrNone = %d", got)
	}
	if got := empty.MaxOrNone(); got != 0 {
		t.Errorf("MaxOrNone(empty) = %d", got)
	}
	a, b := i.MinMaxOr(9)
	if a != 1 || b != 4 {
		t.Errorf("MinMaxOr = %d, %d", a, b)
	}
	a, b = empty.MinMaxOr(9)
	if a != 9 || b != 9 {
		t.Errorf("MinMaxOr(empty) = %d, %d", a, b)
	}
	a, b = i.MinMaxOrNone()
	if a != 1 || b != 4 {
		t.Errorf("MinMaxOrNone = %d, %d", a, b)
	}
	a, b = empty.MinMaxOrNone()
	if a != 0 || b != 0 {
		t.Errorf("MinMaxOrNone(empty) = %d, %d", a, b)
	}
}

func TestOrderedIteratorExtremaFunc(t *testing.T) {
	i := itertools.ToOrderedIterator([]int{3, 1, 4, 2})
	empty := itertools.ToOrderedIterator([]int(nil))
	reverse := func(a, b int) int { return cmp.Compare(b, a) }
	{
		got := i.MinFunc(reverse)
		if assert.NotNil(t, got, "MinFunc") {
			assert.Equal(t, 4, *got, "MinFunc")
		}
	}
	{
		got := i.MaxFunc(reverse)
		if assert.NotNil(t, got, "MaxFunc") {
			assert.Equal(t, 1, *got, "MaxFunc")
		}
	}
	min, max := i.MinMaxFunc(reverse)
	{
		got := min
		if assert.NotNil(t, got, "MinMaxFunc min") {
			assert.Equal(t, 4, *got, "MinMaxFunc min")
		}
	}
	{
		got := max
		if assert.NotNil(t, got, "MinMaxFunc max") {
			assert.Equal(t, 1, *got, "MinMaxFunc max")
		}
	}
	if empty.MinFunc(reverse) != nil || empty.MaxFunc(reverse) != nil {
		t.Error("empty comparator extrema should be nil")
	}
	min, max = empty.MinMaxFunc(reverse)
	if min != nil || max != nil {
		t.Errorf("MinMaxFunc(empty) = %v, %v", min, max)
	}
	for _, tt := range []struct {
		name      string
		got, want int
	}{
		{"MinFuncOr", i.MinFuncOr(reverse, 9), 4},
		{"MinFuncOr(empty)", empty.MinFuncOr(reverse, 9), 9},
		{"MinFuncOrNone", i.MinFuncOrNone(reverse), 4},
		{"MinFuncOrNone(empty)", empty.MinFuncOrNone(reverse), 0},
		{"MaxFuncOr", i.MaxFuncOr(reverse, 9), 1},
		{"MaxFuncOr(empty)", empty.MaxFuncOr(reverse, 9), 9},
		{"MaxFuncOrNone", i.MaxFuncOrNone(reverse), 1},
		{"MaxFuncOrNone(empty)", empty.MaxFuncOrNone(reverse), 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %d, want %d", tt.got, tt.want)
			}
		})
	}
	for _, tt := range []struct {
		name             string
		get              func() (int, int)
		wantMin, wantMax int
	}{
		{"MinMaxFuncOr", func() (int, int) { return i.MinMaxFuncOr(reverse, 9) }, 4, 1},
		{"MinMaxFuncOr(empty)", func() (int, int) { return empty.MinMaxFuncOr(reverse, 9) }, 9, 9},
		{"MinMaxFuncOrNone", func() (int, int) { return i.MinMaxFuncOrNone(reverse) }, 4, 1},
		{"MinMaxFuncOrNone(empty)", func() (int, int) { return empty.MinMaxFuncOrNone(reverse) }, 0, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			min, max := tt.get()
			if min != tt.wantMin || max != tt.wantMax {
				t.Errorf("got %d, %d; want %d, %d", min, max, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestOrderedIteratorSortedVariants(t *testing.T) {
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	assert.Equal(t, []int{1, 2, 3}, i.Sorted().Collect(), "Sorted")
	assert.Equal(t, []int{1, 2, 3}, i.SortedStable().Collect(), "SortedStable")
	reverse := func(a, b int) int { return cmp.Compare(b, a) }
	assert.Equal(t, []int{3, 2, 1}, i.SortedFunc(reverse).Collect(), "SortedFunc")
	words := itertools.ToOrderedIterator([]string{"b-first", "a-only", "b-second"})
	byInitial := func(a, b string) int { return cmp.Compare(a[0], b[0]) }
	want := []string{"a-only", "b-first", "b-second"}
	if got := words.SortedStableFunc(byInitial).Collect(); !slices.Equal(got, want) {
		t.Errorf("SortedStableFunc = %v, want %v", got, want)
	}
}

func TestOrderedIteratorConversions(t *testing.T) {
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	var comparable = i.ToComparableIterator()
	var plain = i.ToIterator()
	assert.Equal(t, []int{3, 1, 2}, comparable.Collect(), "ToComparableIterator")
	assert.Equal(t, []int{3, 1, 2}, plain.Collect(), "ToIterator")
}

func TestOrderedIteratorPull(t *testing.T) {
	next, stop := itertools.ToOrderedIterator([]int{1, 2}).Pull()
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

func TestOrderedIteratorReduce(t *testing.T) {
	// Reduce currently passes each element to the accumulator, without a running value.
	if got := itertools.ToOrderedIterator([]int{1, 2, 3}).Reduce(9, func(v int) int { return v * 2 }); got != 6 {
		t.Errorf("Reduce = %d, want 6", got)
	}
	if got := itertools.ToOrderedIterator([]int(nil)).Reduce(9, func(v int) int { return v * 2 }); got != 9 {
		t.Errorf("Reduce(empty) = %d, want 9", got)
	}
}

func TestOrderedIteratorReduceWithError(t *testing.T) {
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
			i := itertools.OrderedIterator[int](func(yield func(int) bool) {
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

func TestOrderedIteratorReverseSkip(t *testing.T) {
	i := itertools.ToOrderedIterator([]int{1, 2, 3, 4})
	assert.Equal(t, []int{4, 3, 2, 1}, i.Reverse().Collect(), "Reverse")
	assert.Equal(t, []int{3, 4}, i.Skip(2).Collect(), "Skip")
	assert.Equal(t, []int{1, 2, 3, 4}, i.Skip(0).Collect(), "Skip(0)")
	assert.Empty(t, i.Skip(10).Collect(), "Skip(large)")
	assert.Equal(t, []int{3, 1}, itertools.ToOrderedIterator([]int{1, 2, 3, 1}).SkipWhile(func(v int) bool { return v < 3 }).Collect(), "SkipWhile")
}

func TestOrderedIteratorTakeWhileUnique(t *testing.T) {
	assert.Equal(t, []int{1, 2}, itertools.ToOrderedIterator([]int{1, 2, 3, 1}).TakeWhile(func(v int) bool { return v < 3 }).Collect(), "TakeWhile")
	assert.Equal(t, []int{2, 1, 3}, itertools.ToOrderedIterator([]int{2, 1, 2, 3, 1}).Unique().Collect(), "Unique")
}

func TestOrderedIteratorUniqueFunc(t *testing.T) {
	i := itertools.ToOrderedIterator([]string{"Go", "Rust", "go", "RUST", "C"})
	got := i.UniqueFunc(func(a, b string) int {
		return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
	}).Collect()
	want := []string{"Go", "Rust", "C"}
	if !slices.Equal(got, want) {
		t.Errorf("UniqueFunc = %v, want %v", got, want)
	}
}

func TestOrderedIteratorZip(t *testing.T) {
	values := itertools.ToOrderedIterator([]int{10, 20})
	keys := itertools.ToIterator([]string{"a", "b", "c"})
	got := make(map[string]int)
	for k, v := range values.Zip(keys) {
		got[k] = v
	}
	want := map[string]int{"a": 10, "b": 20}
	if len(got) != len(want) || got["a"] != 10 || got["b"] != 20 {
		t.Errorf("Zip = %v, want %v", got, want)
	}
	if n := itertools.ToOrderedIterator([]int(nil)).Zip(keys).Count(); n != 0 {
		t.Errorf("Zip(empty) count = %d", n)
	}
	var pairs []struct {
		key   string
		value int
	}
	for k, v := range values.Zip(itertools.ToComparableIterator([]string{"x"})) {
		pairs = append(pairs, struct {
			key   string
			value int
		}{k, v})
	}
	if want := []struct {
		key   string
		value int
	}{{"x", 10}}; !slices.Equal(pairs, want) {
		t.Errorf("Zip with ComparableIterator keys = %v, want %v", pairs, want)
	}
}
