package itertools_test

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/ertaquo/itertools"
)

func ExampleComparableIterator() {
	values := itertools.ComparableIterator[string](func(yield func(string) bool) {
		for _, value := range []string{"red", "green", "blue"} {
			if !yield(value) {
				return
			}
		}
	})
	for value := range values {
		fmt.Println(value)
	}
	// Output:
	// red
	// green
	// blue
}

func ExampleToComparableIterator() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	for value := range values {
		fmt.Println(value)
	}
	// Output:
	// red
	// green
	// blue
}

func ExampleSeqToComparableIterator() {
	values := itertools.SeqToComparableIterator(slices.Values([]string{"red", "green", "red"}))
	fmt.Println(values.Unique().Collect())
	// Output: [red green]
}

func ExampleComparableIterator_Concat() {
	first := itertools.ToComparableIterator([]string{"red", "green"})
	second := itertools.ToComparableIterator([]string{"blue"})
	fmt.Println(first.Concat(second).Collect())
	// Output: [red green blue]
}

func ExampleComparableIterator_Contains() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(values.Contains("green"))
	fmt.Println(values.Contains("yellow"))
	// Output:
	// true
	// false
}

func ExampleComparableIterator_Equal() {
	values := itertools.ToComparableIterator([]string{"red", "green"})
	fmt.Println(values.Equal(itertools.ToComparableIterator([]string{"red", "green"})))
	fmt.Println(values.Equal(itertools.ToComparableIterator([]string{"green", "red"})))
	// Output:
	// true
	// false
}

func ExampleComparableIterator_EqualFunc() {
	lower := itertools.ToComparableIterator([]string{"red", "green"})
	upper := itertools.ToComparableIterator([]string{"RED", "GREEN"})
	equal := lower.EqualFunc(upper, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	fmt.Println(equal)
	// Output: true
}

func ExampleComparableIterator_GroupBy() {
	words := itertools.ToComparableIterator([]string{"Go", "Rust", "C", "Java", "Go"})
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

func ExampleComparableIterator_Unique() {
	values := itertools.ToComparableIterator([]string{"red", "green", "red", "blue", "green"})
	fmt.Println(values.Unique().Collect())
	// Output: [red green blue]
}

func ExampleComparableIterator_UniqueFunc() {
	values := itertools.ToComparableIterator([]string{"red", "RED", "green", "Green"})
	unique := values.UniqueFunc(func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	fmt.Println(unique.Collect())
	// Output: [red green]
}

func ExampleComparableIterator_ToIterator() {
	values := itertools.ToComparableIterator([]string{"red", "green"})
	var iterator = values.ToIterator()
	fmt.Println(iterator.Collect())
	// Output: [red green]
}

func ExampleComparableIterator_Zip() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	keys := itertools.ToIterator([]int{1, 2})
	for key, value := range values.Zip(keys) {
		fmt.Println(key, value)
	}
	// Output:
	// 1 red
	// 2 green
}

func ExampleComparableIterator_All() {
	values := itertools.ToComparableIterator([]int{2, 4, 6})
	fmt.Println(values.All(func(value int) bool { return value%2 == 0 }))
	fmt.Println(values.All(func(value int) bool { return value > 3 }))
	// Output:
	// true
	// false
}

func ExampleComparableIterator_Any() {
	values := itertools.ToComparableIterator([]int{1, 2, 3})
	fmt.Println(values.Any(func(value int) bool { return value%2 == 0 }))
	fmt.Println(values.Any(func(value int) bool { return value > 3 }))
	// Output:
	// true
	// false
}

func ExampleComparableIterator_Collect() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(values.Collect())
	// Output: [red green blue]
}

func ExampleComparableIterator_CollectAs() {
	type Colors []string
	values := itertools.ToComparableIterator([]string{"red", "green"})
	colors := values.CollectAs[Colors]()
	fmt.Printf("%T: %v\n", colors, colors)
	// Output: itertools_test.Colors: [red green]
}

func ExampleComparableIterator_CollectSorted() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(values.CollectSorted(strings.Compare))
	// Output: [blue green red]
}

func ExampleComparableIterator_CollectSortedAs() {
	type Colors []string
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	colors := values.CollectSortedAs[Colors](strings.Compare)
	fmt.Printf("%T: %v\n", colors, colors)
	// Output: itertools_test.Colors: [blue green red]
}

func ExampleComparableIterator_CollectSortedStable() {
	values := itertools.ToComparableIterator([]string{"pear", "fig", "plum", "kiwi"})
	sorted := values.CollectSortedStable(func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	})
	fmt.Println(sorted)
	// Output: [fig pear plum kiwi]
}

func ExampleComparableIterator_CollectSortedStableAs() {
	type Fruits []string
	values := itertools.ToComparableIterator([]string{"pear", "fig", "plum", "kiwi"})
	fruits := values.CollectSortedStableAs[Fruits](func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	})
	fmt.Printf("%T: %v\n", fruits, fruits)
	// Output: itertools_test.Fruits: [fig pear plum kiwi]
}

func ExampleComparableIterator_ContainsFunc() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(values.ContainsFunc(func(value string) bool {
		return strings.HasPrefix(value, "gr")
	}))
	// Output: true
}

func ExampleComparableIterator_Count() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(values.Count())
	// Output: 3
}

func ExampleComparableIterator_FilterNone() {
	values := itertools.ToComparableIterator([]int{0, 1, 0, 2, 3, 0})
	fmt.Println(values.FilterNone().Collect())
	// Output: [1 2 3]
}

func ExampleComparableIterator_Filter() {
	values := itertools.ToComparableIterator([]int{1, 2, 3, 4, 5})
	even := values.Filter(func(value int) bool { return value%2 == 0 })
	fmt.Println(even.Collect())
	// Output: [2 4]
}

func ExampleComparableIterator_FilterAndCollect() {
	values := itertools.ToComparableIterator([]int{1, 2, 3, 4, 5})
	fmt.Println(values.FilterAndCollect(func(value int) bool { return value%2 == 0 }))
	// Output: [2 4]
}

func ExampleComparableIterator_FilterAndCollectWithError() {
	values := itertools.ToComparableIterator([]int{1, 2, -1, 4})
	even, err := values.FilterAndCollectWithError(func(value int) (bool, error) {
		if value < 0 {
			return false, fmt.Errorf("negative value: %d", value)
		}
		return value%2 == 0, nil
	})
	fmt.Println(even)
	fmt.Println(err)
	// Output:
	// [2]
	// negative value: -1
}

func ExampleComparableIterator_FilterAndCollectParallel() {
	values := itertools.ToComparableIterator([]int{1, 2, 3, 4, 5, 6})
	even := values.FilterAndCollectParallel(func(value int) bool { return value%2 == 0 })
	for _, value := range even {
		fmt.Println(value)
	}
	// Unordered output:
	// 2
	// 4
	// 6
}

func ExampleComparableIterator_FilterAndCollectWithErrorParallel() {
	values := itertools.ToComparableIterator([]int{2, -1, 4})
	even, err := values.FilterAndCollectWithErrorParallel(func(value int) (bool, error) {
		if value < 0 {
			return false, fmt.Errorf("negative value: %d", value)
		}
		return value%2 == 0, nil
	})
	for _, value := range even {
		fmt.Println(value)
	}
	fmt.Println(err)
	// Unordered output:
	// 2
	// 4
	// negative value: -1
}

func ExampleComparableIterator_Find() {
	values := itertools.ToComparableIterator([]int{1, 2, 3, 4})
	found := values.Find(func(value int) bool { return value%2 == 0 })
	fmt.Println(*found)
	fmt.Println(values.Find(func(value int) bool { return value > 4 }))
	// Output:
	// 2
	// <nil>
}

func ExampleComparableIterator_FindOr() {
	values := itertools.ToComparableIterator([]int{1, 2, 3})
	fmt.Println(values.FindOr(func(value int) bool { return value%2 == 0 }, -1))
	fmt.Println(values.FindOr(func(value int) bool { return value > 3 }, -1))
	// Output:
	// 2
	// -1
}

func ExampleComparableIterator_FindOrNone() {
	values := itertools.ToComparableIterator([]int{1, 2, 3})
	fmt.Println(values.FindOrNone(func(value int) bool { return value%2 == 0 }))
	fmt.Println(values.FindOrNone(func(value int) bool { return value > 3 }))
	// Output:
	// 2
	// 0
}

func ExampleComparableIterator_First() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(*values.First())
	fmt.Println(itertools.ToComparableIterator([]string{}).First())
	// Output:
	// red
	// <nil>
}

func ExampleComparableIterator_FirstOr() {
	fmt.Println(itertools.ToComparableIterator([]string{"red", "green"}).FirstOr("none"))
	fmt.Println(itertools.ToComparableIterator([]string{}).FirstOr("none"))
	// Output:
	// red
	// none
}

func ExampleComparableIterator_FirstOrNone() {
	fmt.Println(itertools.ToComparableIterator([]int{2, 4, 6}).FirstOrNone())
	fmt.Println(itertools.ToComparableIterator([]int{}).FirstOrNone())
	// Output:
	// 2
	// 0
}

func ExampleComparableIterator_ForEach() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	err := values.ForEach(func(value string) error {
		fmt.Println(value)
		return nil
	})
	fmt.Println(err)
	// Output:
	// red
	// green
	// blue
	// <nil>
}

func ExampleComparableIterator_ForEachParallel() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	err := values.ForEachParallel(func(value string) error {
		fmt.Println(strings.ToUpper(value))
		return nil
	})
	fmt.Println(err)
	// Unordered output:
	// RED
	// GREEN
	// BLUE
	// <nil>
}

func ExampleComparableIterator_Get() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(*values.Get(1))
	fmt.Println(values.Get(3))
	// Output:
	// green
	// <nil>
}

func ExampleComparableIterator_GetOr() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(values.GetOr(1, "none"))
	fmt.Println(values.GetOr(3, "none"))
	// Output:
	// green
	// none
}

func ExampleComparableIterator_GetOrNone() {
	values := itertools.ToComparableIterator([]int{2, 4, 6})
	fmt.Println(values.GetOrNone(1))
	fmt.Println(values.GetOrNone(3))
	// Output:
	// 4
	// 0
}

func ExampleComparableIterator_Indexed() {
	names := itertools.ToComparableIterator([]string{"Alice", "Bob", "Vera"})
	for index, name := range names.Indexed() {
		fmt.Println(index, name)
	}
	// Output:
	// 0 Alice
	// 1 Bob
	// 2 Vera
}

func ExampleComparableIterator_Last() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(*values.Last())
	fmt.Println(itertools.ToComparableIterator([]string{}).Last())
	// Output:
	// blue
	// <nil>
}

func ExampleComparableIterator_LastOr() {
	fmt.Println(itertools.ToComparableIterator([]string{"red", "green"}).LastOr("none"))
	fmt.Println(itertools.ToComparableIterator([]string{}).LastOr("none"))
	// Output:
	// green
	// none
}

func ExampleComparableIterator_LastOrNone() {
	fmt.Println(itertools.ToComparableIterator([]int{2, 4, 6}).LastOrNone())
	fmt.Println(itertools.ToComparableIterator([]int{}).LastOrNone())
	// Output:
	// 6
	// 0
}

func ExampleComparableIterator_Limit() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(values.Limit(2).Collect())
	// Output: [red green]
}

func ExampleComparableIterator_Map() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	lengths := values.Map(func(value string) int { return len(value) })
	fmt.Println(lengths.Collect())
	// Output: [3 5 4]
}

func ExampleComparableIterator_MapAndCollect() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(values.MapAndCollect(strings.ToUpper))
	// Output: [RED GREEN BLUE]
}

func ExampleComparableIterator_MapAndCollectParallel() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	upper := values.MapAndCollectParallel(strings.ToUpper)
	for _, value := range upper {
		fmt.Println(value)
	}
	// Unordered output:
	// RED
	// GREEN
	// BLUE
}

func ExampleComparableIterator_MapAndCollectWithError() {
	values := itertools.ToComparableIterator([]string{"10", "20", "oops", "30"})
	numbers, err := values.MapAndCollectWithError(strconv.Atoi)
	fmt.Println(numbers)
	fmt.Println(err)
	// Output:
	// [10 20]
	// strconv.Atoi: parsing "oops": invalid syntax
}

func ExampleComparableIterator_MapAndCollectWithErrorParallel() {
	values := itertools.ToComparableIterator([]string{"10", "oops", "20"})
	numbers, err := values.MapAndCollectWithErrorParallel(strconv.Atoi)
	for _, value := range numbers {
		fmt.Println(value)
	}
	fmt.Println(err)
	// Unordered output:
	// 10
	// 20
	// strconv.Atoi: parsing "oops": invalid syntax
}

func ExampleComparableIterator_Max() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	longest := values.Max(func(a, b string) int { return cmp.Compare(len(a), len(b)) })
	fmt.Println(*longest)
	// Output: green
}

func ExampleComparableIterator_MaxOr() {
	values := itertools.ToComparableIterator([]int{3, 1, 4})
	fmt.Println(values.MaxOr(cmp.Compare[int], -1))
	fmt.Println(itertools.ToComparableIterator([]int{}).MaxOr(cmp.Compare[int], -1))
	// Output:
	// 4
	// -1
}

func ExampleComparableIterator_MaxOrNone() {
	values := itertools.ToComparableIterator([]int{3, 1, 4})
	fmt.Println(values.MaxOrNone(cmp.Compare[int]))
	fmt.Println(itertools.ToComparableIterator([]int{}).MaxOrNone(cmp.Compare[int]))
	// Output:
	// 4
	// 0
}

func ExampleComparableIterator_Min() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	shortest := values.Min(func(a, b string) int { return cmp.Compare(len(a), len(b)) })
	fmt.Println(*shortest)
	// Output: red
}

func ExampleComparableIterator_MinMax() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	shortest, longest := values.MinMax(func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	})
	fmt.Println(*shortest, *longest)
	// Output: red green
}

func ExampleComparableIterator_MinMaxOr() {
	values := itertools.ToComparableIterator([]int{3, 1, 4})
	fmt.Println(values.MinMaxOr(cmp.Compare[int], -1))
	fmt.Println(itertools.ToComparableIterator([]int{}).MinMaxOr(cmp.Compare[int], -1))
	// Output:
	// 1 4
	// -1 -1
}

func ExampleComparableIterator_MinMaxOrNone() {
	values := itertools.ToComparableIterator([]int{3, 1, 4})
	fmt.Println(values.MinMaxOrNone(cmp.Compare[int]))
	fmt.Println(itertools.ToComparableIterator([]int{}).MinMaxOrNone(cmp.Compare[int]))
	// Output:
	// 1 4
	// 0 0
}

func ExampleComparableIterator_MinOr() {
	values := itertools.ToComparableIterator([]int{3, 1, 4})
	fmt.Println(values.MinOr(cmp.Compare[int], -1))
	fmt.Println(itertools.ToComparableIterator([]int{}).MinOr(cmp.Compare[int], -1))
	// Output:
	// 1
	// -1
}

func ExampleComparableIterator_MinOrNone() {
	values := itertools.ToComparableIterator([]int{3, 1, 4})
	fmt.Println(values.MinOrNone(cmp.Compare[int]))
	fmt.Println(itertools.ToComparableIterator([]int{}).MinOrNone(cmp.Compare[int]))
	// Output:
	// 1
	// 0
}

func ExampleComparableIterator_Pull() {
	values := itertools.ToComparableIterator([]string{"red", "green"})
	next, stop := values.Pull()
	defer stop()
	fmt.Println(next())
	fmt.Println(next())
	value, ok := next()
	fmt.Printf("%q %t\n", value, ok)
	// Output:
	// red true
	// green true
	// "" false
}

func ExampleComparableIterator_Reduce() {
	values := itertools.ToComparableIterator([]int{1, 2, 3, 4})
	// Keep the running total in the callback; Reduce passes each input value.
	total := 0
	result := values.Reduce(0, func(value int) int {
		total += value
		return total
	})
	fmt.Println(result)
	// Output: 10
}

func ExampleComparableIterator_ReduceWithError() {
	numbers := itertools.ToComparableIterator([]int{1, 2, 3, -1, 4})
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

func ExampleComparableIterator_Reverse() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(values.Reverse().Collect())
	// Output: [blue green red]
}

func ExampleComparableIterator_Skip() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(values.Skip(1).Collect())
	// Output: [green blue]
}

func ExampleComparableIterator_SkipWhile() {
	values := itertools.ToComparableIterator([]int{1, 2, 3, 1, 4})
	remaining := values.SkipWhile(func(value int) bool { return value < 3 })
	fmt.Println(remaining.Collect())
	// Output: [3 1 4]
}

func ExampleComparableIterator_Sorted() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(values.Sorted(strings.Compare).Collect())
	// Output: [blue green red]
}

func ExampleComparableIterator_SortedStable() {
	values := itertools.ToComparableIterator([]string{"pear", "fig", "plum", "kiwi"})
	sorted := values.SortedStable(func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	})
	fmt.Println(sorted.Collect())
	// Output: [fig pear plum kiwi]
}

func ExampleComparableIterator_TakeWhile() {
	values := itertools.ToComparableIterator([]int{1, 2, 3, 1, 4})
	leading := values.TakeWhile(func(value int) bool { return value < 3 })
	fmt.Println(leading.Collect())
	// Output: [1 2]
}

func ExampleComparableIterator_ToSeq() {
	values := itertools.ToComparableIterator([]string{"red", "green", "blue"})
	fmt.Println(slices.Collect(values.ToSeq()))
	// Output: [red green blue]
}
