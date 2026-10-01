package itertools_test

import (
	"cmp"
	"errors"
	"maps"
	"slices"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ertaquo/itertools"
)

type mapPair struct {
	key   string
	value int
}

func mapSeq(pairs ...mapPair) itertools.MapIterator[string, int] {
	return itertools.Seq2ToMapIterator(func(yield func(string, int) bool) {
		for _, p := range pairs {
			if !yield(p.key, p.value) {
				return
			}
		}
	})
}

func collectMapPairs(i itertools.MapIterator[string, int]) []mapPair {
	var pairs []mapPair
	for k, v := range i {
		pairs = append(pairs, mapPair{k, v})
	}
	return pairs
}

func TestMapIteratorConversion(t *testing.T) {
	type namedMap map[string]int
	assert.Equal(t, map[string]int{"a": 1, "b": 2}, itertools.ToMapIterator(namedMap{"a": 1, "b": 2}).Collect(), "ToMapIterator")
	i := mapSeq(mapPair{"a", 1}, mapPair{"b", 2})
	assert.Equal(t, []mapPair{{"a", 1}, {"b", 2}}, collectMapPairs(i), "Seq2ToMapIterator")
	assert.Equal(t, []mapPair{{"a", 1}, {"b", 2}}, collectMapPairs(itertools.Seq2ToMapIterator(i.ToSeq2())), "ToSeq2")
}

func TestMapIteratorPredicates(t *testing.T) {
	match := func(k string, v int) bool { return k == "b" && v == 2 }
	for _, tt := range []struct {
		name     string
		pairs    []mapPair
		all, any bool
	}{
		{"empty", nil, true, false},
		{"all", []mapPair{{"b", 2}}, true, true},
		{"mixed", []mapPair{{"a", 1}, {"b", 2}}, false, true},
		{"none", []mapPair{{"a", 1}}, false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			i := mapSeq(tt.pairs...)
			if got := i.All(match); got != tt.all {
				t.Errorf("All = %v, want %v", got, tt.all)
			}
			if got := i.Any(match); got != tt.any {
				t.Errorf("Any = %v, want %v", got, tt.any)
			}
			if got := i.ContainsFunc(match); got != tt.any {
				t.Errorf("ContainsFunc = %v, want %v", got, tt.any)
			}
		})
	}
}

func TestMapIteratorCollect(t *testing.T) {
	type namedMap map[string]int
	type keys []string
	type values []int
	i := mapSeq(mapPair{"b", 2}, mapPair{"a", 1})
	want := map[string]int{"a": 1, "b": 2}
	assert.Equal(t, want, i.Collect(), "Collect")
	assert.EqualValues(t, want, i.CollectAs[namedMap](), "CollectAs")
	if got := mapSeq().Collect(); len(got) != 0 {
		t.Errorf("Collect(empty) = %v", got)
	}
	if got := i.CollectKeys(); !slices.Equal(got, []string{"b", "a"}) {
		t.Errorf("CollectKeys = %v", got)
	}
	if got := i.CollectKeysAs[keys](); !slices.Equal(got, []string{"b", "a"}) {
		t.Errorf("CollectKeysAs = %v", got)
	}
	assert.Equal(t, []int{2, 1}, i.CollectValues(), "CollectValues")
	assert.EqualValues(t, []int{2, 1}, i.CollectValuesAs[values](), "CollectValuesAs")
	k, v := i.CollectKeysAndValues()
	if !slices.Equal(k, []string{"b", "a"}) || !slices.Equal(v, []int{2, 1}) {
		t.Errorf("CollectKeysAndValues = %v, %v", k, v)
	}
	k2, v2 := i.CollectKeysAndValuesAs[keys, values]()
	if !slices.Equal(k2, k) || !slices.Equal(v2, v) {
		t.Errorf("CollectKeysAndValuesAs = %v, %v", k2, v2)
	}
	duplicates := mapSeq(mapPair{"a", 1}, mapPair{"a", 2})
	assert.Equal(t, map[string]int{"a": 2}, duplicates.Collect(), "Collect duplicate key")
	if got := duplicates.Count(); got != 2 {
		t.Errorf("Count duplicate key = %d, want 2", got)
	}
}

func TestMapIteratorContains(t *testing.T) {
	i := mapSeq(mapPair{"a", 1}, mapPair{"b", 2})
	if !i.ContainsKey("b") || i.ContainsKey("c") {
		t.Error("ContainsKey gave wrong membership")
	}
	if !i.ContainsValue(2) || i.ContainsValue(3) {
		t.Error("ContainsValue gave wrong membership")
	}
	defer func() {
		if recover() == nil {
			t.Error("ContainsValue(noncomparable) did not panic")
		}
	}()
	itertools.Seq2ToMapIterator(func(yield func(string, []int) bool) { yield("a", []int{1}) }).ContainsValue([]int{1})
}

func TestMapIteratorCountEqual(t *testing.T) {
	if got := mapSeq(mapPair{"a", 1}, mapPair{"b", 2}).Count(); got != 2 {
		t.Errorf("Count = %d", got)
	}
	if got := mapSeq().Count(); got != 0 {
		t.Errorf("Count(empty) = %d", got)
	}
	for _, tt := range []struct {
		name        string
		left, right []mapPair
		want        bool
	}{
		{"empty", nil, nil, true},
		{"equal", []mapPair{{"a", 1}, {"b", 2}}, []mapPair{{"a", 1}, {"b", 2}}, true},
		{"different key", []mapPair{{"a", 1}}, []mapPair{{"b", 1}}, false},
		{"different value", []mapPair{{"a", 1}}, []mapPair{{"a", 2}}, false},
		{"different order", []mapPair{{"a", 1}, {"b", 2}}, []mapPair{{"b", 2}, {"a", 1}}, false},
		{"shorter left", []mapPair{{"a", 1}}, []mapPair{{"a", 1}, {"b", 2}}, false},
		{"shorter right", []mapPair{{"a", 1}, {"b", 2}}, []mapPair{{"a", 1}}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapSeq(tt.left...).Equal(mapSeq(tt.right...), cmp.Compare[int]); got != tt.want {
				t.Errorf("Equal(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.want)
			}
		})
	}
}

func TestMapIteratorFilter(t *testing.T) {
	i := mapSeq(mapPair{"a", 1}, mapPair{"b", 2}, mapPair{"c", 3})
	even := func(_ string, v int) bool { return v%2 == 0 }
	want := map[string]int{"b": 2}
	assert.Equal(t, []mapPair{{"b", 2}}, collectMapPairs(i.Filter(even)), "Filter")
	assert.Equal(t, want, i.FilterAndCollect(even), "FilterAndCollect")
	assert.Equal(t, want, i.FilterAndCollectParallel(even, itertools.WithLimit(2)), "FilterAndCollectParallel")
	assert.Equal(t, []mapPair{{"b", 2}}, collectMapPairs(i.FilterKeys(func(k string) bool { return k == "b" })), "FilterKeys")
	assert.Equal(t, []mapPair{{"b", 2}}, collectMapPairs(i.FilterValues(func(v int) bool { return v%2 == 0 })), "FilterValues")
	visited := 0
	source := itertools.Seq2ToMapIterator(func(yield func(string, int) bool) {
		for _, p := range []mapPair{{"a", 1}, {"b", 2}, {"c", 3}} {
			visited++
			if !yield(p.key, p.value) {
				return
			}
		}
	})
	for range source.Filter(even) {
		break
	}
	if visited != 2 {
		t.Errorf("Filter read %d source pairs after first match, want 2", visited)
	}
}

func TestMapIteratorFilterWithError(t *testing.T) {
	bad := errors.New("bad value")
	f := func(_ string, v int) (bool, error) {
		if v == 3 {
			return false, bad
		}
		return v%2 == 0, nil
	}
	i := mapSeq(mapPair{"a", 1}, mapPair{"b", 2}, mapPair{"c", 3})
	got, err := i.FilterAndCollectWithError(f)
	assert.Equal(t, map[string]int{"b": 2}, got, "FilterAndCollectWithError partial")
	if !errors.Is(err, bad) {
		t.Errorf("FilterAndCollectWithError error = %v", err)
	}
	got, err = i.FilterAndCollectWithErrorParallel(f, itertools.WithLimit(1))
	if !errors.Is(err, bad) {
		t.Errorf("FilterAndCollectWithErrorParallel error = %v", err)
	}
	if !maps.Equal(got, map[string]int{"b": 2}) {
		t.Errorf("FilterAndCollectWithErrorParallel partial = %v", got)
	}
	got, err = i.FilterAndCollectWithErrorParallel(func(_ string, v int) (bool, error) { return v%2 == 0, nil })
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, map[string]int{"b": 2}, got, "FilterAndCollectWithErrorParallel")
}

func TestMapIteratorFind(t *testing.T) {
	i := mapSeq(mapPair{"a", 1}, mapPair{"b", 2})
	match := func(k string, v int) bool { return k == "b" && v == 2 }
	missing := func(k string, v int) bool { return k == "z" }
	k, v := i.Find(match)
	if k == nil || v == nil || *k != "b" || *v != 2 {
		t.Errorf("Find = %v, %v", k, v)
	}
	k, v = i.Find(missing)
	if k != nil || v != nil {
		t.Errorf("Find(missing) = %v, %v", k, v)
	}
	if got := i.FindKey(match); got == nil || *got != "b" {
		t.Errorf("FindKey = %v", got)
	}
	if got := i.FindKey(missing); got != nil {
		t.Errorf("FindKey(missing) = %v", got)
	}
	if got := i.FindKeyOr(match, "z"); got != "b" {
		t.Errorf("FindKeyOr = %q", got)
	}
	if got := i.FindKeyOr(missing, "z"); got != "z" {
		t.Errorf("FindKeyOr(missing) = %q", got)
	}
	if got := i.FindKeyOrNone(match); got != "b" {
		t.Errorf("FindKeyOrNone = %q", got)
	}
	if got := i.FindKeyOrNone(missing); got != "" {
		t.Errorf("FindKeyOrNone(missing) = %q", got)
	}
	a, b := i.FindOr(match, "z", 9)
	if a != "b" || b != 2 {
		t.Errorf("FindOr = %q, %d", a, b)
	}
	a, b = i.FindOr(missing, "z", 9)
	if a != "z" || b != 9 {
		t.Errorf("FindOr(missing) = %q, %d", a, b)
	}
	a, b = i.FindOrNone(match)
	if a != "b" || b != 2 {
		t.Errorf("FindOrNone = %q, %d", a, b)
	}
	a, b = i.FindOrNone(missing)
	if a != "" || b != 0 {
		t.Errorf("FindOrNone(missing) = %q, %d", a, b)
	}
	if got := i.FindValue(match); got == nil || *got != 2 {
		t.Errorf("FindValue = %v", got)
	}
	if got := i.FindValue(missing); got != nil {
		t.Errorf("FindValue(missing) = %v", got)
	}
	if got := i.FindValueOr(match, 9); got != 2 {
		t.Errorf("FindValueOr = %d", got)
	}
	if got := i.FindValueOr(missing, 9); got != 9 {
		t.Errorf("FindValueOr(missing) = %d", got)
	}
	if got := i.FindValueOrNone(match); got != 2 {
		t.Errorf("FindValueOrNone = %d", got)
	}
	if got := i.FindValueOrNone(missing); got != 0 {
		t.Errorf("FindValueOrNone(missing) = %d", got)
	}
}

func TestMapIteratorForEach(t *testing.T) {
	bad := errors.New("stop")
	i := mapSeq(mapPair{"a", 1}, mapPair{"b", 2}, mapPair{"c", 3})
	var got []mapPair
	err := i.ForEach(func(k string, v int) error {
		got = append(got, mapPair{k, v})
		if k == "b" {
			return bad
		}
		return nil
	})
	assert.Equal(t, []mapPair{{"a", 1}, {"b", 2}}, got, "ForEach visits")
	if !errors.Is(err, bad) {
		t.Errorf("ForEach error = %v", err)
	}
	if err := i.ForEach(func(string, int) error { return nil }); err != nil {
		t.Errorf("ForEach success = %v", err)
	}
	var mu sync.Mutex
	got = nil
	err = i.ForEachParallel(func(k string, v int) error {
		mu.Lock()
		got = append(got, mapPair{k, v})
		mu.Unlock()
		return nil
	}, itertools.WithLimit(2))
	if err != nil {
		t.Fatal(err)
	}
	slices.SortFunc(got, func(a, b mapPair) int { return cmp.Compare(a.key, b.key) })
	assert.Equal(t, []mapPair{{"a", 1}, {"b", 2}, {"c", 3}}, got, "ForEachParallel")
	if err := i.ForEachParallel(func(k string, _ int) error {
		if k == "b" {
			return bad
		}
		return nil
	}); !errors.Is(err, bad) {
		t.Errorf("ForEachParallel error = %v", err)
	}
}

func TestMapIteratorGet(t *testing.T) {
	i := mapSeq(mapPair{"a", 1}, mapPair{"b", 2})
	if got := i.Get("b"); got == nil || *got != 2 {
		t.Errorf("Get = %v", got)
	}
	if got := i.Get("z"); got != nil {
		t.Errorf("Get(missing) = %v", got)
	}
	if got := i.GetOr("b", 9); got != 2 {
		t.Errorf("GetOr = %d", got)
	}
	if got := i.GetOr("z", 9); got != 9 {
		t.Errorf("GetOr(missing) = %d", got)
	}
	if got := i.GetOrNone("b"); got != 2 {
		t.Errorf("GetOrNone = %d", got)
	}
	if got := i.GetOrNone("z"); got != 0 {
		t.Errorf("GetOrNone(missing) = %d", got)
	}
}

func TestMapIteratorConcatKeysLimit(t *testing.T) {
	i := mapSeq(mapPair{"a", 1}, mapPair{"b", 2}).Concat(mapSeq(mapPair{"c", 3}))
	assert.Equal(t, []mapPair{{"a", 1}, {"b", 2}, {"c", 3}}, collectMapPairs(i), "Concat")
	if got := i.Keys().Collect(); !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Errorf("Keys = %v", got)
	}
	assert.Equal(t, []mapPair{{"a", 1}, {"b", 2}}, collectMapPairs(i.Limit(2)), "Limit")
	assert.Empty(t, collectMapPairs(i.Limit(0)), "Limit(0)")
	assert.Empty(t, collectMapPairs(i.Limit(-1)), "Limit(-1)")
	assert.Equal(t, []mapPair{{"a", 1}, {"b", 2}, {"c", 3}}, collectMapPairs(i.Limit(10)), "Limit(large)")
}

func TestMapIteratorMap(t *testing.T) {
	i := mapSeq(mapPair{"a", 1}, mapPair{"b", 2})
	f := func(k string, v int) (string, int) { return k + "!", v * 2 }
	wantPairs := []mapPair{{"a!", 2}, {"b!", 4}}
	wantMap := map[string]int{"a!": 2, "b!": 4}
	assert.Equal(t, wantPairs, collectMapPairs(i.Map(f)), "Map")
	assert.Equal(t, wantMap, i.MapAndCollect(f), "MapAndCollect")
	assert.Equal(t, wantMap, i.MapAndCollectParallel(f, itertools.WithLimit(2)), "MapAndCollectParallel")
	assert.Equal(t, []mapPair{{"a!", 1}, {"b!", 2}}, collectMapPairs(i.MapKeys(func(k string) string { return k + "!" })), "MapKeys")
	assert.Equal(t, []mapPair{{"a", 2}, {"b", 4}}, collectMapPairs(i.MapValues(func(v int) int { return v * 2 })), "MapValues")
}

func TestMapIteratorMapWithError(t *testing.T) {
	bad := errors.New("bad value")
	f := func(k string, v int) (string, int, error) {
		if k == "c" {
			return "", 0, bad
		}
		return k + "!", v * 2, nil
	}
	i := mapSeq(mapPair{"a", 1}, mapPair{"b", 2}, mapPair{"c", 3})
	want := map[string]int{"a!": 2, "b!": 4}
	got, err := i.MapAndCollectWithError(f)
	assert.Equal(t, want, got, "MapAndCollectWithError partial")
	if !errors.Is(err, bad) {
		t.Errorf("MapAndCollectWithError error = %v", err)
	}
	got, err = i.MapAndCollectWithErrorParallel(f, itertools.WithLimit(1))
	assert.Equal(t, want, got, "MapAndCollectWithErrorParallel partial")
	if !errors.Is(err, bad) {
		t.Errorf("MapAndCollectWithErrorParallel error = %v", err)
	}
	got, err = i.MapAndCollectWithErrorParallel(func(k string, v int) (string, int, error) { return k + "!", v * 2, nil })
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, map[string]int{"a!": 2, "b!": 4, "c!": 6}, got, "MapAndCollectWithErrorParallel")
}

func TestMapIteratorPull(t *testing.T) {
	next, stop := mapSeq(mapPair{"a", 1}, mapPair{"b", 2}).Pull()
	defer stop()
	if k, v, ok := next(); !ok || k != "a" || v != 1 {
		t.Errorf("first Pull = %q, %d, %v", k, v, ok)
	}
	if k, v, ok := next(); !ok || k != "b" || v != 2 {
		t.Errorf("second Pull = %q, %d, %v", k, v, ok)
	}
	if k, v, ok := next(); ok || k != "" || v != 0 {
		t.Errorf("exhausted Pull = %q, %d, %v", k, v, ok)
	}
}

func TestMapIteratorReduce(t *testing.T) {
	// Reduce currently passes each pair to the accumulator without a running value.
	f := func(k string, v int) int { return len(k) + v }
	if got := mapSeq(mapPair{"a", 1}, mapPair{"bb", 2}).Reduce(9, f); got != 4 {
		t.Errorf("Reduce = %d, want 4", got)
	}
	if got := mapSeq().Reduce(9, f); got != 9 {
		t.Errorf("Reduce(empty) = %d, want 9", got)
	}
}

func TestMapIteratorReverseSkip(t *testing.T) {
	i := mapSeq(mapPair{"a", 1}, mapPair{"b", 2}, mapPair{"c", 3})
	assert.Equal(t, []mapPair{{"c", 3}, {"b", 2}, {"a", 1}}, collectMapPairs(i.Reverse()), "Reverse")
	assert.Equal(t, []mapPair{{"c", 3}}, collectMapPairs(i.Skip(2)), "Skip")
	assert.Equal(t, []mapPair{{"a", 1}, {"b", 2}, {"c", 3}}, collectMapPairs(i.Skip(0)), "Skip(0)")
	assert.Empty(t, collectMapPairs(i.Skip(10)), "Skip(large)")
	assert.Equal(t, []mapPair{{"b", 2}, {"c", 3}}, collectMapPairs(i.SkipWhile(func(_ string, v int) bool { return v < 2 })), "SkipWhile")
}

func TestMapIteratorSorted(t *testing.T) {
	i := mapSeq(mapPair{"b-first", 1}, mapPair{"a", 2}, mapPair{"b-second", 3})
	assert.Equal(t, []mapPair{{"a", 2}, {"b-first", 1}, {"b-second", 3}}, collectMapPairs(i.Sorted(cmp.Compare[string])), "Sorted")
	byInitial := func(a, b string) int { return cmp.Compare(a[0], b[0]) }
	assert.Equal(t, []mapPair{{"a", 2}, {"b-first", 1}, {"b-second", 3}}, collectMapPairs(i.SortedStable(byInitial)), "SortedStable")
}

func TestMapIteratorTakeWhile(t *testing.T) {
	i := mapSeq(mapPair{"a", 1}, mapPair{"b", 2}, mapPair{"c", 3})
	assert.Equal(t, []mapPair{{"a", 1}, {"b", 2}}, collectMapPairs(i.TakeWhile(func(_ string, v int) bool { return v < 3 })), "TakeWhile")
}

func TestMapIteratorValues(t *testing.T) {
	i := mapSeq(mapPair{"a", 1}, mapPair{"b", 2})
	assert.Equal(t, []int{1, 2}, i.Values().Collect(), "Values")
}
