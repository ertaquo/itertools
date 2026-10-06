package itertools_test

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/ertaquo/itertools"
)

func Example() {
	numbers := itertools.ToIterator([]int{1, 2, 3, 4, 5})
	evenSquares := numbers.Filter(func(n int) bool {
		return n%2 == 0
	}).Map(func(n int) int {
		return n * n
	})
	fmt.Println(evenSquares.Collect())
	// Output:
	// [4 16]
}

func ExampleIterator() {
	numbers := itertools.Iterator[int](func(yield func(int) bool) {
		for n := 1; n <= 3; n++ {
			if !yield(n) {
				return
			}
		}
	})
	for n := range numbers {
		fmt.Println(n)
	}
	// Output:
	// 1
	// 2
	// 3
}

func ExampleToIterator() {
	names := []string{"Alice", "Bob", "Vera"}
	for name := range itertools.ToIterator(names) {
		fmt.Println(name)
	}
	// Output:
	// Alice
	// Bob
	// Vera
}

func ExampleSeqToIterator() {
	seq := slices.Values([]string{"Alice", "Bob", "Vera"})
	fmt.Println(itertools.SeqToIterator(seq).Collect())
	// Output:
	// [Alice Bob Vera]
}

func ExampleIterator_All() {
	numbers := itertools.ToIterator([]int{2, 4, 6})
	fmt.Println(numbers.All(func(n int) bool { return n%2 == 0 }))
	fmt.Println(numbers.All(func(n int) bool { return n > 3 }))
	// Output:
	// true
	// false
}

func ExampleIterator_Any() {
	numbers := itertools.ToIterator([]int{1, 2, 3})
	fmt.Println(numbers.Any(func(n int) bool { return n%2 == 0 }))
	fmt.Println(numbers.Any(func(n int) bool { return n < 0 }))
	// Output:
	// true
	// false
}

func ExampleIterator_Collect() {
	numbers := itertools.ToIterator([]int{1, 2, 3})
	fmt.Println(numbers.Collect())
	// Output:
	// [1 2 3]
}

func ExampleIterator_CollectAs() {
	type Names []string
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"}).CollectAs[Names]()
	fmt.Printf("%T: %v\n", names, names)
	// Output:
	// itertools_test.Names: [Alice Bob Vera]
}

func ExampleIterator_CollectSorted() {
	names := itertools.ToIterator([]string{"Vera", "Alice", "Bob"})
	fmt.Println(names.CollectSorted(strings.Compare))
	// Output:
	// [Alice Bob Vera]
}

func ExampleIterator_CollectSortedAs() {
	type Names []string
	names := itertools.ToIterator([]string{"Vera", "Alice", "Bob"}).CollectSortedAs[Names](strings.Compare)
	fmt.Printf("%T: %v\n", names, names)
	// Output:
	// itertools_test.Names: [Alice Bob Vera]
}

func ExampleIterator_CollectSortedStable() {
	names := itertools.ToIterator([]string{"Bob", "Vera", "Amy", "Alice"})
	compare := func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	}
	fmt.Println(names.CollectSortedStable(compare))
	// Output:
	// [Bob Amy Vera Alice]
}

func ExampleIterator_CollectSortedStableAs() {
	type Names []string
	names := itertools.ToIterator([]string{"Bob", "Vera", "Amy", "Alice"})
	compare := func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	}
	sorted := names.CollectSortedStableAs[Names](compare)
	fmt.Printf("%T: %v\n", sorted, sorted)
	// Output:
	// itertools_test.Names: [Bob Amy Vera Alice]
}

func ExampleIterator_SortedStable() {
	names := itertools.ToIterator([]string{"Bob", "Vera", "Amy", "Alice"})
	compare := func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	}
	fmt.Println(names.SortedStable(compare).Collect())
	// Output:
	// [Bob Amy Vera Alice]
}

func ExampleIterator_Contains() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	fmt.Println(names.Contains("Bob"))
	fmt.Println(names.Contains("Bill"))
	// Output:
	// true
	// false
}

func ExampleIterator_ContainsFunc() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	fmt.Println(names.ContainsFunc(func(name string) bool {
		return strings.HasPrefix(name, "V")
	}))
	// Output:
	// true
}

func ExampleIterator_Count() {
	fmt.Println(itertools.ToIterator([]string{"Alice", "Bob", "Vera"}).Count())
	// Output:
	// 3
}

func ExampleIterator_Equal() {
	names := itertools.ToIterator([]string{"Alice", "Bob"})
	same := itertools.ToIterator([]string{"Alice", "Bob"})
	different := itertools.ToIterator([]string{"Bob", "Alice"})
	fmt.Println(names.Equal(same, strings.Compare))
	fmt.Println(names.Equal(different, strings.Compare))
	// Output:
	// true
	// false
}

func ExampleIterator_Filter() {
	numbers := itertools.ToIterator([]int{1, 2, 3, 4})
	even := numbers.Filter(func(n int) bool {
		return n%2 == 0
	})
	fmt.Println(even.Collect())
	// Output:
	// [2 4]
}

func ExampleIterator_FilterAndCollect() {
	numbers := itertools.ToIterator([]int{1, 2, 3, 4})
	even := numbers.FilterAndCollect(func(n int) bool {
		return n%2 == 0
	})
	fmt.Println(even)
	// Output:
	// [2 4]
}

func ExampleIterator_FilterAndCollectWithError() {
	numbers := itertools.ToIterator([]int{1, 2, 3, -1, 4})
	even, err := numbers.FilterAndCollectWithError(func(n int) (bool, error) {
		if n < 0 {
			return false, fmt.Errorf("negative number: %d", n)
		}
		return n%2 == 0, nil
	})
	fmt.Println(even)
	fmt.Println(err)
	// Output:
	// [2]
	// negative number: -1
}

func ExampleIterator_FilterAndCollectParallel() {
	numbers := itertools.ToIterator([]int{1, 2, 3, 4})
	even := numbers.FilterAndCollectParallel(func(n int) bool {
		return n%2 == 0
	}, itertools.WithLimit(2))
	for _, n := range even {
		fmt.Println(n)
	}
	// Unordered output:
	// 2
	// 4
}

func ExampleIterator_FilterAndCollectWithErrorParallel() {
	numbers := itertools.ToIterator([]int{1, 2, -1, 4})
	even, err := numbers.FilterAndCollectWithErrorParallel(func(n int) (bool, error) {
		if n < 0 {
			return false, fmt.Errorf("negative number: %d", n)
		}
		return n%2 == 0, nil
	}, itertools.WithLimit(2))
	for _, n := range even {
		fmt.Println(n)
	}
	fmt.Println(err)
	// Unordered output:
	// 2
	// 4
	// negative number: -1
}

func ExampleIterator_Find() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	if name := names.Find(func(name string) bool { return len(name) == 3 }); name != nil {
		fmt.Println(*name)
	}
	fmt.Println(names.Find(func(name string) bool { return name == "Bill" }) == nil)
	// Output:
	// Bob
	// true
}

func ExampleIterator_FindOr() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	fmt.Println(names.FindOr(func(name string) bool { return len(name) == 3 }, "unknown"))
	fmt.Printf("%q\n", names.FindOr(func(name string) bool { return name == "Bill" }, "unknown"))
	// Output:
	// Bob
	// "unknown"
}

func ExampleIterator_FindOrNone() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	fmt.Println(names.FindOrNone(func(name string) bool { return len(name) == 3 }))
	fmt.Printf("%q\n", names.FindOrNone(func(name string) bool { return name == "Bill" }))
	// Output:
	// Bob
	// ""
}

func ExampleIterator_First() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	if name := names.First(); name != nil {
		fmt.Println(*name)
	}
	fmt.Println(itertools.ToIterator([]string{}).First() == nil)
	// Output:
	// Alice
	// true
}

func ExampleIterator_FirstOr() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	fmt.Println(names.FirstOr("unknown"))
	fmt.Printf("%q\n", itertools.ToIterator([]string{}).FirstOr("unknown"))
	// Output:
	// Alice
	// "unknown"
}

func ExampleIterator_FirstOrNone() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	fmt.Println(names.FirstOrNone())
	fmt.Printf("%q\n", itertools.ToIterator([]string{}).FirstOrNone())
	// Output:
	// Alice
	// ""
}

func ExampleIterator_Indexed() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	for index, name := range names.Indexed() {
		fmt.Println(index, name)
	}
	// Output:
	// 0 Alice
	// 1 Bob
	// 2 Vera
}

func ExampleIterator_Last() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	if name := names.Last(); name != nil {
		fmt.Println(*name)
	}
	fmt.Println(itertools.ToIterator([]string{}).Last() == nil)
	// Output:
	// Vera
	// true
}

func ExampleIterator_LastOr() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	fmt.Println(names.LastOr("unknown"))
	fmt.Printf("%q\n", itertools.ToIterator([]string{}).LastOr("unknown"))
	// Output:
	// Vera
	// "unknown"
}

func ExampleIterator_LastOrNone() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	fmt.Println(names.LastOrNone())
	fmt.Printf("%q\n", itertools.ToIterator([]string{}).LastOrNone())
	// Output:
	// Vera
	// ""
}

func ExampleIterator_ForEach() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	err := names.ForEach(func(name string) error {
		if name == "Vera" {
			return fmt.Errorf("stopped at %s", name)
		}
		fmt.Println(name)
		return nil
	})
	fmt.Println(err)
	// Output:
	// Alice
	// Bob
	// stopped at Vera
}

func ExampleIterator_ForEachParallel() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	err := names.ForEachParallel(func(name string) error {
		fmt.Println(name)
		return nil
	}, itertools.WithLimit(2))
	fmt.Println(err)
	// Unordered output:
	// Alice
	// Bob
	// Vera
	// <nil>
}

func ExampleIterator_Get() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	if name := names.Get(1); name != nil {
		fmt.Println(*name)
	}
	fmt.Println(names.Get(10) == nil)
	// Output:
	// Bob
	// true
}

func ExampleIterator_GetOr() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	fmt.Println(names.GetOr(1, "unknown"))
	fmt.Printf("%q\n", names.GetOr(10, "unknown"))
	// Output:
	// Bob
	// "unknown"
}

func ExampleIterator_GetOrNone() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	fmt.Println(names.GetOrNone(1))
	fmt.Printf("%q\n", names.GetOrNone(10))
	// Output:
	// Bob
	// ""
}

func ExampleIterator_Concat() {
	first := itertools.ToIterator([]string{"Alice", "Bob"})
	second := itertools.ToIterator([]string{"Vera"})
	fmt.Println(first.Concat(second).Collect())
	// Output:
	// [Alice Bob Vera]
}

func ExampleIterator_Limit() {
	numbers := itertools.ToIterator([]int{1, 2, 3, 4, 5})
	fmt.Println(numbers.Limit(3).Collect())
	// Output:
	// [1 2 3]
}

func ExampleIterator_Map() {
	numbers := itertools.ToIterator([]int{1, 2, 3})
	labels := numbers.Map(func(n int) string {
		return fmt.Sprintf("item-%d", n)
	})
	fmt.Println(labels.Collect())
	// Output:
	// [item-1 item-2 item-3]
}

func ExampleIterator_MapAndCollect() {
	numbers := itertools.ToIterator([]int{1, 2, 3})
	labels := numbers.MapAndCollect(func(n int) string {
		return fmt.Sprintf("item-%d", n)
	})
	fmt.Println(labels)
	// Output:
	// [item-1 item-2 item-3]
}

func ExampleIterator_MapAndCollectParallel() {
	numbers := itertools.ToIterator([]int{1, 2, 3})
	labels := numbers.MapAndCollectParallel(func(n int) string {
		return fmt.Sprintf("item-%d", n)
	}, itertools.WithLimit(2))
	for _, label := range labels {
		fmt.Println(label)
	}
	// Unordered output:
	// item-1
	// item-2
	// item-3
}

func ExampleIterator_MapAndCollectWithError() {
	text := itertools.ToIterator([]string{"10", "20", "oops", "30"})
	numbers, err := text.MapAndCollectWithError(func(s string) (int, error) {
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", s)
		}
		return n, nil
	})
	fmt.Println(numbers)
	fmt.Println(err)
	// Output:
	// [10 20]
	// invalid number: oops
}

func ExampleIterator_MapAndCollectWithErrorParallel() {
	text := itertools.ToIterator([]string{"10", "oops", "20"})
	numbers, err := text.MapAndCollectWithErrorParallel(func(s string) (int, error) {
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", s)
		}
		return n, nil
	}, itertools.WithLimit(2))
	for _, n := range numbers {
		fmt.Println(n)
	}
	fmt.Println(err)
	// Unordered output:
	// 10
	// 20
	// invalid number: oops
}

func ExampleIterator_Max() {
	numbers := itertools.ToIterator([]int{3, 1, 4, 2})
	if n := numbers.Max(cmp.Compare[int]); n != nil {
		fmt.Println(*n)
	}
	fmt.Println(itertools.ToIterator([]int{}).Max(cmp.Compare[int]) == nil)
	// Output:
	// 4
	// true
}

func ExampleIterator_MaxOr() {
	numbers := itertools.ToIterator([]int{3, 1, 4, 2})
	fmt.Println(numbers.MaxOr(cmp.Compare[int], -1))
	fmt.Println(itertools.ToIterator([]int{}).MaxOr(cmp.Compare[int], -1))
	// Output:
	// 4
	// -1
}

func ExampleIterator_MaxOrNone() {
	numbers := itertools.ToIterator([]int{3, 1, 4, 2})
	fmt.Println(numbers.MaxOrNone(cmp.Compare[int]))
	fmt.Println(itertools.ToIterator([]int{}).MaxOrNone(cmp.Compare[int]))
	// Output:
	// 4
	// 0
}

func ExampleIterator_Min() {
	numbers := itertools.ToIterator([]int{3, 1, 4, 2})
	if n := numbers.Min(cmp.Compare[int]); n != nil {
		fmt.Println(*n)
	}
	fmt.Println(itertools.ToIterator([]int{}).Min(cmp.Compare[int]) == nil)
	// Output:
	// 1
	// true
}

func ExampleIterator_MinOr() {
	numbers := itertools.ToIterator([]int{3, 1, 4, 2})
	fmt.Println(numbers.MinOr(cmp.Compare[int], -1))
	fmt.Println(itertools.ToIterator([]int{}).MinOr(cmp.Compare[int], -1))
	// Output:
	// 1
	// -1
}

func ExampleIterator_MinOrNone() {
	numbers := itertools.ToIterator([]int{3, 1, 4, 2})
	fmt.Println(numbers.MinOrNone(cmp.Compare[int]))
	fmt.Println(itertools.ToIterator([]int{}).MinOrNone(cmp.Compare[int]))
	// Output:
	// 1
	// 0
}

func ExampleIterator_MinMax() {
	numbers := itertools.ToIterator([]int{3, 1, 4, 2})
	min, max := numbers.MinMax(cmp.Compare[int])
	if min != nil && max != nil {
		fmt.Println(*min, *max)
	}
	// Output:
	// 1 4
}

func ExampleIterator_MinMaxOr() {
	numbers := itertools.ToIterator([]int{3, 1, 4, 2})
	fmt.Println(numbers.MinMaxOr(cmp.Compare[int], -1))
	fmt.Println(itertools.ToIterator([]int{}).MinMaxOr(cmp.Compare[int], -1))
	// Output:
	// 1 4
	// -1 -1
}

func ExampleIterator_MinMaxOrNone() {
	numbers := itertools.ToIterator([]int{3, 1, 4, 2})
	fmt.Println(numbers.MinMaxOrNone(cmp.Compare[int]))
	fmt.Println(itertools.ToIterator([]int{}).MinMaxOrNone(cmp.Compare[int]))
	// Output:
	// 1 4
	// 0 0
}

func ExampleIterator_Pull() {
	numbers := itertools.ToIterator([]int{1, 2})
	next, stop := numbers.Pull()
	defer stop()
	fmt.Println(next())
	fmt.Println(next())
	fmt.Println(next())
	// Output:
	// 1 true
	// 2 true
	// 0 false
}

func ExampleIterator_Reduce() {
	numbers := itertools.ToIterator([]int{1, 2, 3})
	sum := 0
	total := numbers.Reduce(0, func(n int) int {
		// The callback receives each value; keep the running sum in the closure.
		sum += n
		return sum
	})
	fmt.Println(total)
	// Output:
	// 6
}

func ExampleIterator_ReduceWithError() {
	numbers := itertools.ToIterator([]int{1, 2, 3, -1, 4})
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

func ExampleIterator_Reverse() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	fmt.Println(names.Reverse().Collect())
	// Output:
	// [Vera Bob Alice]
}

func ExampleIterator_Shuffle() {
	names := itertools.ToIterator([]string{"Alice", "Bob", "Vera"})
	for name := range names.Shuffle() {
		fmt.Println(name)
	}
	// Unordered output:
	// Vera
	// Alice
	// Bob
}

func ExampleIterator_Skip() {
	numbers := itertools.ToIterator([]int{1, 2, 3, 4, 5})
	fmt.Println(numbers.Skip(2).Collect())
	// Output:
	// [3 4 5]
}

func ExampleIterator_SkipWhile() {
	numbers := itertools.ToIterator([]int{1, 2, 3, 1, 4})
	fmt.Println(numbers.SkipWhile(func(n int) bool {
		return n < 3
	}).Collect())
	// Output:
	// [3 1 4]
}

func ExampleIterator_Sorted() {
	names := itertools.ToIterator([]string{"Vera", "Alice", "Bob"})
	for name := range names.Sorted(strings.Compare) {
		fmt.Println(name)
	}
	// Output:
	// Alice
	// Bob
	// Vera
}

func ExampleIterator_TakeWhile() {
	numbers := itertools.ToIterator([]int{1, 2, 3, 1, 4})
	fmt.Println(numbers.TakeWhile(func(n int) bool {
		return n < 3
	}).Collect())
	// Output:
	// [1 2]
}

func ExampleIterator_ToSeq() {
	numbers := itertools.ToIterator([]int{1, 2, 3})
	fmt.Println(slices.Collect(numbers.ToSeq()))
	// Output:
	// [1 2 3]
}

func ExampleIterator_Unique() {
	names := itertools.ToIterator([]string{"Bob", "bob", "Alice", "BOB"})
	unique := names.Unique(func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	fmt.Println(unique.Collect())
	// Output:
	// [Bob Alice]
}

func ExampleIterator_Zip() {
	values := itertools.ToIterator([]int{10, 20, 30})
	keys := itertools.ToIterator([]string{"Alice", "Bob"})
	for name, score := range values.Zip(keys) {
		fmt.Println(name, score)
	}
	// Output:
	// Alice 10
	// Bob 20
}
