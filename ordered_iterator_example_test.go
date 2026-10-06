package itertools_test

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/ertaquo/itertools"
)

func ExampleOrderedIterator() {
	i := itertools.OrderedIterator[int](func(yield func(int) bool) {
		for _, v := range []int{3, 1, 2} {
			if !yield(v) {
				return
			}
		}
	})
	for v := range i {
		fmt.Println(v)
	}
	// Output:
	// 3
	// 1
	// 2
}

func ExampleToOrderedIterator() {
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	fmt.Println(i.Collect())
	fmt.Println(i.CollectSorted())
	// Output:
	// [3 1 2]
	// [1 2 3]
}

func ExampleSeqToOrderedIterator() {
	i := itertools.SeqToOrderedIterator(slices.Values([]string{"pear", "apple", "orange"}))
	fmt.Println(i.CollectSorted())
	// Output: [apple orange pear]
}

func ExampleOrderedIterator_CollectSorted() {
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	fmt.Println(i.CollectSorted())
	// Output: [1 2 3]
}

func ExampleOrderedIterator_CollectSortedAs() {
	type Numbers []int
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	values := i.CollectSortedAs[Numbers]()
	fmt.Printf("%T %v\n", values, values)
	// Output: itertools_test.Numbers [1 2 3]
}

func ExampleOrderedIterator_CollectSortedStable() {
	i := itertools.ToOrderedIterator([]int{3, 1, 2, 1})
	fmt.Println(i.CollectSortedStable())
	// Output: [1 1 2 3]
}

func ExampleOrderedIterator_CollectSortedStableAs() {
	type Numbers []int
	i := itertools.ToOrderedIterator([]int{3, 1, 2, 1})
	values := i.CollectSortedStableAs[Numbers]()
	fmt.Printf("%T %v\n", values, values)
	// Output: itertools_test.Numbers [1 1 2 3]
}

func ExampleOrderedIterator_CollectSortedFunc() {
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	values := i.CollectSortedFunc(func(a, b int) int {
		return cmp.Compare(b, a)
	})
	fmt.Println(values)
	// Output: [3 2 1]
}

func ExampleOrderedIterator_CollectSortedFuncAs() {
	type Numbers []int
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	values := i.CollectSortedFuncAs[Numbers](func(a, b int) int {
		return cmp.Compare(b, a)
	})
	fmt.Printf("%T %v\n", values, values)
	// Output: itertools_test.Numbers [3 2 1]
}

func ExampleOrderedIterator_CollectSortedStableFunc() {
	i := itertools.ToOrderedIterator([]string{"pear", "fig", "plum", "kiwi"})
	values := i.CollectSortedStableFunc(func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	})
	fmt.Println(values)
	// Output: [fig pear plum kiwi]
}

func ExampleOrderedIterator_CollectSortedStableFuncAs() {
	type Words []string
	i := itertools.ToOrderedIterator([]string{"pear", "fig", "plum", "kiwi"})
	values := i.CollectSortedStableFuncAs[Words](func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	})
	fmt.Printf("%T %v\n", values, values)
	// Output: itertools_test.Words [fig pear plum kiwi]
}

func ExampleOrderedIterator_Concat() {
	a := itertools.ToOrderedIterator([]int{1, 2})
	b := itertools.ToOrderedIterator([]int{3, 4})
	fmt.Println(a.Concat(b).Collect())
	// Output: [1 2 3 4]
}

func ExampleOrderedIterator_Equal() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	fmt.Println(i.Equal(itertools.ToOrderedIterator([]int{1, 2, 3})))
	fmt.Println(i.Equal(itertools.ToOrderedIterator([]int{3, 2, 1})))
	// Output:
	// true
	// false
}

func ExampleOrderedIterator_EqualFunc() {
	a := itertools.ToOrderedIterator([]string{"GO", "Hello"})
	b := itertools.ToOrderedIterator([]string{"go", "hello"})
	fmt.Println(a.EqualFunc(b, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	}))
	// Output: true
}

func ExampleOrderedIterator_Max() {
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	fmt.Println(*i.Max())
	fmt.Println(itertools.ToOrderedIterator([]int{}).Max())
	// Output:
	// 3
	// <nil>
}

func ExampleOrderedIterator_MaxOr() {
	fmt.Println(itertools.ToOrderedIterator([]int{3, 1, 2}).MaxOr(-1))
	fmt.Println(itertools.ToOrderedIterator([]int{}).MaxOr(-1))
	// Output:
	// 3
	// -1
}

func ExampleOrderedIterator_MaxOrNone() {
	fmt.Println(itertools.ToOrderedIterator([]int{3, 1, 2}).MaxOrNone())
	fmt.Println(itertools.ToOrderedIterator([]int{}).MaxOrNone())
	// Output:
	// 3
	// 0
}

func ExampleOrderedIterator_MaxFunc() {
	i := itertools.ToOrderedIterator([]string{"pear", "fig", "orange"})
	longest := i.MaxFunc(func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	})
	fmt.Println(*longest)
	// Output: orange
}

func ExampleOrderedIterator_MaxFuncOr() {
	byLength := func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	}
	fmt.Println(itertools.ToOrderedIterator([]string{"pear", "fig", "orange"}).MaxFuncOr(byLength, "none"))
	fmt.Println(itertools.ToOrderedIterator([]string{}).MaxFuncOr(byLength, "none"))
	// Output:
	// orange
	// none
}

func ExampleOrderedIterator_MaxFuncOrNone() {
	byLength := func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	}
	fmt.Printf("%q\n", itertools.ToOrderedIterator([]string{"pear", "fig", "orange"}).MaxFuncOrNone(byLength))
	fmt.Printf("%q\n", itertools.ToOrderedIterator([]string{}).MaxFuncOrNone(byLength))
	// Output:
	// "orange"
	// ""
}

func ExampleOrderedIterator_Min() {
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	fmt.Println(*i.Min())
	fmt.Println(itertools.ToOrderedIterator([]int{}).Min())
	// Output:
	// 1
	// <nil>
}

func ExampleOrderedIterator_MinOr() {
	fmt.Println(itertools.ToOrderedIterator([]int{3, 1, 2}).MinOr(-1))
	fmt.Println(itertools.ToOrderedIterator([]int{}).MinOr(-1))
	// Output:
	// 1
	// -1
}

func ExampleOrderedIterator_MinOrNone() {
	fmt.Println(itertools.ToOrderedIterator([]int{3, 1, 2}).MinOrNone())
	fmt.Println(itertools.ToOrderedIterator([]int{}).MinOrNone())
	// Output:
	// 1
	// 0
}

func ExampleOrderedIterator_MinFunc() {
	i := itertools.ToOrderedIterator([]string{"pear", "fig", "orange"})
	shortest := i.MinFunc(func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	})
	fmt.Println(*shortest)
	// Output: fig
}

func ExampleOrderedIterator_MinFuncOr() {
	byLength := func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	}
	fmt.Println(itertools.ToOrderedIterator([]string{"pear", "fig", "orange"}).MinFuncOr(byLength, "none"))
	fmt.Println(itertools.ToOrderedIterator([]string{}).MinFuncOr(byLength, "none"))
	// Output:
	// fig
	// none
}

func ExampleOrderedIterator_MinFuncOrNone() {
	byLength := func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	}
	fmt.Printf("%q\n", itertools.ToOrderedIterator([]string{"pear", "fig", "orange"}).MinFuncOrNone(byLength))
	fmt.Printf("%q\n", itertools.ToOrderedIterator([]string{}).MinFuncOrNone(byLength))
	// Output:
	// "fig"
	// ""
}

func ExampleOrderedIterator_MinMax() {
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	min, max := i.MinMax()
	fmt.Println(*min, *max)
	fmt.Println(itertools.ToOrderedIterator([]int{}).MinMax())
	// Output:
	// 1 3
	// <nil> <nil>
}

func ExampleOrderedIterator_MinMaxOr() {
	fmt.Println(itertools.ToOrderedIterator([]int{3, 1, 2}).MinMaxOr(-1))
	fmt.Println(itertools.ToOrderedIterator([]int{}).MinMaxOr(-1))
	// Output:
	// 1 3
	// -1 -1
}

func ExampleOrderedIterator_MinMaxOrNone() {
	fmt.Println(itertools.ToOrderedIterator([]int{3, 1, 2}).MinMaxOrNone())
	fmt.Println(itertools.ToOrderedIterator([]int{}).MinMaxOrNone())
	// Output:
	// 1 3
	// 0 0
}

func ExampleOrderedIterator_MinMaxFunc() {
	i := itertools.ToOrderedIterator([]string{"pear", "fig", "orange"})
	shortest, longest := i.MinMaxFunc(func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	})
	fmt.Println(*shortest, *longest)
	// Output: fig orange
}

func ExampleOrderedIterator_MinMaxFuncOr() {
	byLength := func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	}
	fmt.Println(itertools.ToOrderedIterator([]string{"pear", "fig", "orange"}).MinMaxFuncOr(byLength, "none"))
	fmt.Println(itertools.ToOrderedIterator([]string{}).MinMaxFuncOr(byLength, "none"))
	// Output:
	// fig orange
	// none none
}

func ExampleOrderedIterator_MinMaxFuncOrNone() {
	byLength := func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	}
	shortest, longest := itertools.ToOrderedIterator([]string{"pear", "fig", "orange"}).MinMaxFuncOrNone(byLength)
	fmt.Printf("%q %q\n", shortest, longest)
	shortest, longest = itertools.ToOrderedIterator([]string{}).MinMaxFuncOrNone(byLength)
	fmt.Printf("%q %q\n", shortest, longest)
	// Output:
	// "fig" "orange"
	// "" ""
}

func ExampleOrderedIterator_Sorted() {
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	for v := range i.Sorted() {
		fmt.Println(v)
	}
	// Output:
	// 1
	// 2
	// 3
}

func ExampleOrderedIterator_SortedStable() {
	i := itertools.ToOrderedIterator([]int{3, 1, 2, 1})
	fmt.Println(i.SortedStable().Collect())
	// Output: [1 1 2 3]
}

func ExampleOrderedIterator_SortedFunc() {
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	sorted := i.SortedFunc(func(a, b int) int {
		return cmp.Compare(b, a)
	})
	fmt.Println(sorted.Collect())
	// Output: [3 2 1]
}

func ExampleOrderedIterator_SortedStableFunc() {
	i := itertools.ToOrderedIterator([]string{"pear", "fig", "plum", "kiwi"})
	sorted := i.SortedStableFunc(func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	})
	fmt.Println(sorted.Collect())
	// Output: [fig pear plum kiwi]
}

func ExampleOrderedIterator_ToComparableIterator() {
	fmt.Println(itertools.
		ToOrderedIterator([]int{1, 2, 1, 3}).
		ToComparableIterator().
		Unique().
		Collect(),
	)
	// Output: [1 2 3]
}

func ExampleOrderedIterator_Contains() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	fmt.Println(i.Contains(2))
	fmt.Println(i.Contains(4))
	// Output:
	// true
	// false
}

func ExampleOrderedIterator_GroupBy() {
	words := itertools.ToOrderedIterator([]string{"Go", "Rust", "C", "Java", "Go"})
	groups := words.GroupBy(func(word string) int {
		return len(word)
	})
	for length, group := range groups {
		fmt.Println(length, group)
	}
	// Unordered output:
	// 1 [C]
	// 2 [Go Go]
	// 4 [Rust Java]
}

func ExampleOrderedIterator_Unique() {
	i := itertools.ToOrderedIterator([]int{1, 2, 1, 3, 2})
	fmt.Println(i.Unique().Collect())
	// Output: [1 2 3]
}

func ExampleOrderedIterator_UniqueFunc() {
	i := itertools.ToOrderedIterator([]string{"Go", "go", "Rust", "GO"})
	unique := i.UniqueFunc(func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	fmt.Println(unique.Collect())
	// Output: [Go Rust]
}

func ExampleOrderedIterator_ToIterator() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	var values = i.ToIterator()
	fmt.Println(values.Collect())
	// Output: [1 2 3]
}

func ExampleOrderedIterator_Zip() {
	values := itertools.ToOrderedIterator([]int{10, 20, 30})
	keys := itertools.ToIterator([]string{"a", "b"})
	for key, value := range values.Zip(keys) {
		fmt.Printf("%s: %d\n", key, value)
	}
	// Output:
	// a: 10
	// b: 20
}

func ExampleOrderedIterator_All() {
	i := itertools.ToOrderedIterator([]int{2, 4, 6})
	fmt.Println(i.All(func(v int) bool { return v%2 == 0 }))
	fmt.Println(i.All(func(v int) bool { return v > 3 }))
	// Output:
	// true
	// false
}

func ExampleOrderedIterator_Any() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	fmt.Println(i.Any(func(v int) bool { return v%2 == 0 }))
	fmt.Println(i.Any(func(v int) bool { return v > 3 }))
	// Output:
	// true
	// false
}

func ExampleOrderedIterator_Collect() {
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	fmt.Println(i.Collect())
	// Output: [3 1 2]
}

func ExampleOrderedIterator_CollectAs() {
	type Numbers []int
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	values := i.CollectAs[Numbers]()
	fmt.Printf("%T %v\n", values, values)
	// Output: itertools_test.Numbers [3 1 2]
}

func ExampleOrderedIterator_ContainsFunc() {
	i := itertools.ToOrderedIterator([]string{"pear", "apple", "orange"})
	fmt.Println(i.ContainsFunc(func(v string) bool {
		return strings.HasPrefix(v, "a")
	}))
	// Output: true
}

func ExampleOrderedIterator_Count() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	fmt.Println(i.Count())
	// Output: 3
}

func ExampleOrderedIterator_FilterNone() {
	values := itertools.ToOrderedIterator([]int{0, 1, 0, 2, 3, 0})
	fmt.Println(values.FilterNone().Collect())
	// Output: [1 2 3]
}

func ExampleOrderedIterator_Filter() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3, 4})
	even := i.Filter(func(v int) bool { return v%2 == 0 })
	fmt.Println(even.Collect())
	// Output: [2 4]
}

func ExampleOrderedIterator_FilterAndCollect() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3, 4})
	fmt.Println(i.FilterAndCollect(func(v int) bool { return v%2 == 0 }))
	// Output: [2 4]
}

func ExampleOrderedIterator_FilterAndCollectWithError() {
	i := itertools.ToOrderedIterator([]int{1, 2, -3, 4})
	values, err := i.FilterAndCollectWithError(func(v int) (bool, error) {
		if v < 0 {
			return false, errors.New("negative value")
		}
		return v%2 == 0, nil
	})
	fmt.Println(values, err)
	// Output: [2] negative value
}

func ExampleOrderedIterator_FilterAndCollectParallel() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3, 4})
	values := i.FilterAndCollectParallel(func(v int) bool { return v%2 == 0 })
	for _, v := range values {
		fmt.Println(v)
	}
	// Unordered output:
	// 2
	// 4
}

func ExampleOrderedIterator_FilterAndCollectWithErrorParallel() {
	i := itertools.ToOrderedIterator([]int{1, 2, -3, 4})
	values, err := i.FilterAndCollectWithErrorParallel(func(v int) (bool, error) {
		if v < 0 {
			return false, errors.New("negative value")
		}
		return v%2 == 0, nil
	})
	for _, v := range values {
		fmt.Println(v)
	}
	fmt.Println(err)
	// Unordered output:
	// 2
	// 4
	// negative value
}

func ExampleOrderedIterator_Find() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3, 4})
	fmt.Println(*i.Find(func(v int) bool { return v%2 == 0 }))
	fmt.Println(i.Find(func(v int) bool { return v > 4 }))
	// Output:
	// 2
	// <nil>
}

func ExampleOrderedIterator_FindOr() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	fmt.Println(i.FindOr(func(v int) bool { return v%2 == 0 }, -1))
	fmt.Println(i.FindOr(func(v int) bool { return v > 3 }, -1))
	// Output:
	// 2
	// -1
}

func ExampleOrderedIterator_FindOrNone() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	fmt.Println(i.FindOrNone(func(v int) bool { return v%2 == 0 }))
	fmt.Println(i.FindOrNone(func(v int) bool { return v > 3 }))
	// Output:
	// 2
	// 0
}

func ExampleOrderedIterator_First() {
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	fmt.Println(*i.First())
	fmt.Println(itertools.ToOrderedIterator([]int{}).First())
	// Output:
	// 3
	// <nil>
}

func ExampleOrderedIterator_FirstOr() {
	fmt.Println(itertools.ToOrderedIterator([]int{3, 1, 2}).FirstOr(-1))
	fmt.Println(itertools.ToOrderedIterator([]int{}).FirstOr(-1))
	// Output:
	// 3
	// -1
}

func ExampleOrderedIterator_FirstOrNone() {
	fmt.Println(itertools.ToOrderedIterator([]int{3, 1, 2}).FirstOrNone())
	fmt.Println(itertools.ToOrderedIterator([]int{}).FirstOrNone())
	// Output:
	// 3
	// 0
}

func ExampleOrderedIterator_ForEach() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	err := i.ForEach(func(v int) error {
		fmt.Println(v)
		return nil
	})
	fmt.Println(err)
	// Output:
	// 1
	// 2
	// 3
	// <nil>
}

func ExampleOrderedIterator_ForEachParallel() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	err := i.ForEachParallel(func(v int) error {
		fmt.Println(v)
		return nil
	})
	fmt.Println(err)
	// Unordered output:
	// 1
	// 2
	// 3
	// <nil>
}

func ExampleOrderedIterator_Get() {
	i := itertools.ToOrderedIterator([]int{10, 20, 30})
	fmt.Println(*i.Get(1))
	fmt.Println(i.Get(3))
	// Output:
	// 20
	// <nil>
}

func ExampleOrderedIterator_GetOr() {
	i := itertools.ToOrderedIterator([]int{10, 20, 30})
	fmt.Println(i.GetOr(1, -1))
	fmt.Println(i.GetOr(3, -1))
	// Output:
	// 20
	// -1
}

func ExampleOrderedIterator_GetOrNone() {
	i := itertools.ToOrderedIterator([]int{10, 20, 30})
	fmt.Println(i.GetOrNone(1))
	fmt.Println(i.GetOrNone(3))
	// Output:
	// 20
	// 0
}

func ExampleOrderedIterator_Indexed() {
	names := itertools.ToOrderedIterator([]string{"Alice", "Bob", "Vera"})
	for index, name := range names.Indexed() {
		fmt.Println(index, name)
	}
	// Output:
	// 0 Alice
	// 1 Bob
	// 2 Vera
}

func ExampleOrderedIterator_Last() {
	i := itertools.ToOrderedIterator([]int{3, 1, 2})
	fmt.Println(*i.Last())
	fmt.Println(itertools.ToOrderedIterator([]int{}).Last())
	// Output:
	// 2
	// <nil>
}

func ExampleOrderedIterator_LastOr() {
	fmt.Println(itertools.ToOrderedIterator([]int{3, 1, 2}).LastOr(-1))
	fmt.Println(itertools.ToOrderedIterator([]int{}).LastOr(-1))
	// Output:
	// 2
	// -1
}

func ExampleOrderedIterator_LastOrNone() {
	fmt.Println(itertools.ToOrderedIterator([]int{3, 1, 2}).LastOrNone())
	fmt.Println(itertools.ToOrderedIterator([]int{}).LastOrNone())
	// Output:
	// 2
	// 0
}

func ExampleOrderedIterator_Limit() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3, 4})
	fmt.Println(i.Limit(2).Collect())
	// Output: [1 2]
}

func ExampleOrderedIterator_Map() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	labels := i.Map(func(v int) string { return fmt.Sprintf("item-%d", v) })
	fmt.Println(labels.Collect())
	// Output: [item-1 item-2 item-3]
}

func ExampleOrderedIterator_MapAndCollect() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	labels := i.MapAndCollect(func(v int) string { return fmt.Sprintf("item-%d", v) })
	fmt.Println(labels)
	// Output: [item-1 item-2 item-3]
}

func ExampleOrderedIterator_MapAndCollectParallel() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	labels := i.MapAndCollectParallel(func(v int) string { return fmt.Sprintf("item-%d", v) })
	for _, label := range labels {
		fmt.Println(label)
	}
	// Unordered output:
	// item-1
	// item-2
	// item-3
}

func ExampleOrderedIterator_MapAndCollectWithError() {
	i := itertools.ToOrderedIterator([]int{1, 2, -3, 4})
	labels, err := i.MapAndCollectWithError(func(v int) (string, error) {
		if v < 0 {
			return "", errors.New("negative value")
		}
		return fmt.Sprintf("item-%d", v), nil
	})
	fmt.Println(labels, err)
	// Output: [item-1 item-2] negative value
}

func ExampleOrderedIterator_MapAndCollectWithErrorParallel() {
	i := itertools.ToOrderedIterator([]int{1, 2, -3, 4})
	labels, err := i.MapAndCollectWithErrorParallel(func(v int) (string, error) {
		if v < 0 {
			return "", errors.New("negative value")
		}
		return fmt.Sprintf("item-%d", v), nil
	})
	for _, label := range labels {
		fmt.Println(label)
	}
	fmt.Println(err)
	// Unordered output:
	// item-1
	// item-2
	// item-4
	// negative value
}

func ExampleOrderedIterator_Pull() {
	i := itertools.ToOrderedIterator([]int{10, 20})
	next, stop := i.Pull()
	defer stop()
	fmt.Println(next())
	fmt.Println(next())
	fmt.Println(next())
	// Output:
	// 10 true
	// 20 true
	// 0 false
}

func ExampleOrderedIterator_Reduce() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	// The callback receives each value, so keep the running total in a closure.
	total := 0
	result := i.Reduce(0, func(v int) int {
		total += v
		return total
	})
	fmt.Println(result)
	// Output: 6
}

func ExampleOrderedIterator_ReduceWithError() {
	numbers := itertools.ToOrderedIterator([]int{1, 2, 3, -1, 4})
	sum := 0
	total, err := numbers.ReduceWithError(0, func(n int) (int, error) {
		if n < 0 {
			return sum, fmt.Errorf("negative number: %d", n)
		}
		// The callback receives each value; keep the running sum in the closure.
		sum += n
		return sum, nil
	})
	fmt.Println(total)
	fmt.Println(err)
	// Output:
	// 6
	// negative number: -1
}

func ExampleOrderedIterator_Reverse() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	fmt.Println(i.Reverse().Collect())
	// Output: [3 2 1]
}

func ExampleOrderedIterator_Skip() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3, 4})
	fmt.Println(i.Skip(2).Collect())
	// Output: [3 4]
}

func ExampleOrderedIterator_SkipWhile() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3, 1, 4})
	fmt.Println(i.SkipWhile(func(v int) bool { return v < 3 }).Collect())
	// Output: [3 1 4]
}

func ExampleOrderedIterator_TakeWhile() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3, 1, 4})
	fmt.Println(i.TakeWhile(func(v int) bool { return v < 3 }).Collect())
	// Output: [1 2]
}

func ExampleOrderedIterator_ToSeq() {
	i := itertools.ToOrderedIterator([]int{1, 2, 3})
	fmt.Println(slices.Collect(i.ToSeq()))
	// Output: [1 2 3]
}
