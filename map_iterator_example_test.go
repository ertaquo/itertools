package itertools_test

import (
	"cmp"
	"errors"
	"fmt"
	"iter"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/ertaquo/itertools"
)

func ExampleMapIterator() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	for fruit, count := range stock {
		fmt.Println(fruit, count)
	}
	// Unordered output:
	// apple 3
	// pear 5
}

func ExampleToMapIterator() {
	stock := map[string]int{"apple": 3, "pear": 5}
	for fruit, count := range itertools.ToMapIterator(stock) {
		fmt.Println(fruit, count)
	}
	// Unordered output:
	// apple 3
	// pear 5
}

func ExampleSeq2ToMapIterator() {
	seq := slices.All([]string{"apple", "pear"})
	for index, fruit := range itertools.Seq2ToMapIterator(seq) {
		fmt.Println(index, fruit)
	}
	// Output:
	// 0 apple
	// 1 pear
}

func ExampleMapIterator_All() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(stock.All(func(fruit string, count int) bool { return count > 0 }))
	fmt.Println(stock.All(func(fruit string, count int) bool { return count > 3 }))
	// Output:
	// true
	// false
}

func ExampleMapIterator_Any() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(stock.Any(func(fruit string, count int) bool { return count > 3 }))
	fmt.Println(stock.Any(func(fruit string, count int) bool { return count > 5 }))
	// Output:
	// true
	// false
}

func ExampleMapIterator_Collect() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(stock.Collect())
	// Output:
	// map[apple:3 pear:5]
}

func ExampleMapIterator_CollectAs() {
	type inventory map[string]int
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	var result = stock.CollectAs[inventory]()
	fmt.Println(result)
	// Output:
	// map[apple:3 pear:5]
}

func ExampleMapIterator_CollectKeys() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	for _, fruit := range stock.CollectKeys() {
		fmt.Println(fruit)
	}
	// Unordered output:
	// apple
	// pear
}

func ExampleMapIterator_CollectKeysAndValues() {
	fruits := itertools.Seq2ToMapIterator(slices.All([]string{"apple", "pear"}))
	keys, values := fruits.CollectKeysAndValues()
	fmt.Println(keys)
	fmt.Println(values)
	// Output:
	// [0 1]
	// [apple pear]
}

func ExampleMapIterator_CollectKeysAndValuesAs() {
	type indexes []int
	type names []string
	fruits := itertools.Seq2ToMapIterator(slices.All([]string{"apple", "pear"}))
	keys, values := fruits.CollectKeysAndValuesAs[indexes, names]()
	fmt.Println(keys)
	fmt.Println(values)
	// Output:
	// [0 1]
	// [apple pear]
}

func ExampleMapIterator_CollectKeysAs() {
	type names []string
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	for _, fruit := range stock.CollectKeysAs[names]() {
		fmt.Println(fruit)
	}
	// Unordered output:
	// apple
	// pear
}

func ExampleMapIterator_CollectValues() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	for _, count := range stock.CollectValues() {
		fmt.Println(count)
	}
	// Unordered output:
	// 3
	// 5
}

func ExampleMapIterator_CollectValuesAs() {
	type counts []int
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	for _, count := range stock.CollectValuesAs[counts]() {
		fmt.Println(count)
	}
	// Unordered output:
	// 3
	// 5
}

func ExampleMapIterator_ContainsFunc() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(stock.ContainsFunc(func(fruit string, count int) bool {
		return fruit == "pear" && count >= 5
	}))
	// Output:
	// true
}

func ExampleMapIterator_ContainsKey() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(stock.ContainsKey("apple"))
	fmt.Println(stock.ContainsKey("plum"))
	// Output:
	// true
	// false
}

func ExampleMapIterator_ContainsValue() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(stock.ContainsValue(5))
	fmt.Println(stock.ContainsValue(7))
	// Output:
	// true
	// false
}

func ExampleMapIterator_Count() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(stock.Count())
	// Output:
	// 2
}

func ExampleMapIterator_Equal() {
	fruits := itertools.Seq2ToMapIterator(slices.All([]string{"apple", "pear"}))
	same := itertools.Seq2ToMapIterator(slices.All([]string{"apple", "pear"}))
	reversed := itertools.Seq2ToMapIterator(slices.All([]string{"pear", "apple"}))
	fmt.Println(fruits.Equal(same, strings.Compare))
	fmt.Println(fruits.Equal(reversed, strings.Compare))
	// Output:
	// true
	// false
}

func ExampleMapIterator_Filter() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5, "plum": 7})
	filtered := stock.Filter(func(fruit string, count int) bool { return count >= 5 })
	for fruit, count := range filtered {
		fmt.Println(fruit, count)
	}
	// Unordered output:
	// pear 5
	// plum 7
}

func ExampleMapIterator_FilterAndCollect() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5, "plum": 7})
	result := stock.FilterAndCollect(func(fruit string, count int) bool { return count >= 5 })
	fmt.Println(result)
	// Output:
	// map[pear:5 plum:7]
}

func ExampleMapIterator_FilterAndCollectWithError() {
	counts := itertools.Seq2ToMapIterator(slices.All([]int{3, -1, 5}))
	result, err := counts.FilterAndCollectWithError(func(index, count int) (bool, error) {
		if count < 0 {
			return false, errors.New("negative stock")
		}
		return count > 0, nil
	})
	fmt.Println(result)
	fmt.Println(err)
	// Output:
	// map[0:3]
	// negative stock
}

func ExampleMapIterator_FilterAndCollectParallel() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5, "plum": 7})
	result := stock.FilterAndCollectParallel(func(fruit string, count int) bool {
		return count >= 5
	}, itertools.WithLimit(2))
	for fruit, count := range result {
		fmt.Println(fruit, count)
	}
	// Unordered output:
	// pear 5
	// plum 7
}

func ExampleMapIterator_FilterAndCollectWithErrorParallel() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": -1, "plum": 7})
	result, err := stock.FilterAndCollectWithErrorParallel(func(fruit string, count int) (bool, error) {
		if count < 0 {
			return false, errors.New("negative stock")
		}
		return count > 0, nil
	}, itertools.WithLimit(2))
	for fruit, count := range result {
		fmt.Println(fruit, count)
	}
	fmt.Println(err)
	// Unordered output:
	// apple 3
	// plum 7
	// negative stock
}

func ExampleMapIterator_FilterKeys() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5, "plum": 7})
	for fruit, count := range stock.FilterKeys(func(fruit string) bool {
		return strings.HasPrefix(fruit, "p")
	}) {
		fmt.Println(fruit, count)
	}
	// Unordered output:
	// pear 5
	// plum 7
}

func ExampleMapIterator_FilterValues() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5, "plum": 7})
	for fruit, count := range stock.FilterValues(func(count int) bool { return count >= 5 }) {
		fmt.Println(fruit, count)
	}
	// Unordered output:
	// pear 5
	// plum 7
}

func ExampleMapIterator_FilterMap() {
	stock := itertools.ToMapIterator(map[string]string{"apple": "3", "pear": "oops", "plum": "5"})
	counts := stock.FilterMap(func(fruit, text string) (string, int, bool) {
		count, err := strconv.Atoi(text)
		return strings.ToUpper(fruit), count, err == nil
	})
	for fruit, count := range counts {
		fmt.Println(fruit, count)
	}
	// Unordered output:
	// APPLE 3
	// PLUM 5
}

func ExampleMapIterator_Find() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fruit, count := stock.Find(func(fruit string, count int) bool { return count > 3 })
	fmt.Println(*fruit, *count)
	fruit, count = stock.Find(func(fruit string, count int) bool { return count > 5 })
	fmt.Println(fruit, count)
	// Output:
	// pear 5
	// <nil> <nil>
}

func ExampleMapIterator_FindKey() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fruit := stock.FindKey(func(fruit string, count int) bool { return count > 3 })
	fmt.Println(*fruit)
	fmt.Println(stock.FindKey(func(fruit string, count int) bool { return count > 5 }))
	// Output:
	// pear
	// <nil>
}

func ExampleMapIterator_FindKeyOr() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(stock.FindKeyOr(func(fruit string, count int) bool { return count > 3 }, "none"))
	fmt.Println(stock.FindKeyOr(func(fruit string, count int) bool { return count > 5 }, "none"))
	// Output:
	// pear
	// none
}

func ExampleMapIterator_FindKeyOrNone() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Printf("%q\n", stock.FindKeyOrNone(func(fruit string, count int) bool { return count > 3 }))
	fmt.Printf("%q\n", stock.FindKeyOrNone(func(fruit string, count int) bool { return count > 5 }))
	// Output:
	// "pear"
	// ""
}

func ExampleMapIterator_FindOr() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(stock.FindOr(func(fruit string, count int) bool { return count > 3 }, "none", -1))
	fmt.Println(stock.FindOr(func(fruit string, count int) bool { return count > 5 }, "none", -1))
	// Output:
	// pear 5
	// none -1
}

func ExampleMapIterator_FindOrNone() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fruit, count := stock.FindOrNone(func(fruit string, count int) bool { return count > 3 })
	fmt.Printf("%q %d\n", fruit, count)
	fruit, count = stock.FindOrNone(func(fruit string, count int) bool { return count > 5 })
	fmt.Printf("%q %d\n", fruit, count)
	// Output:
	// "pear" 5
	// "" 0
}

func ExampleMapIterator_FindValue() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	count := stock.FindValue(func(fruit string, count int) bool { return fruit == "pear" })
	fmt.Println(*count)
	fmt.Println(stock.FindValue(func(fruit string, count int) bool { return fruit == "plum" }))
	// Output:
	// 5
	// <nil>
}

func ExampleMapIterator_FindValueOr() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(stock.FindValueOr(func(fruit string, count int) bool { return fruit == "pear" }, -1))
	fmt.Println(stock.FindValueOr(func(fruit string, count int) bool { return fruit == "plum" }, -1))
	// Output:
	// 5
	// -1
}

func ExampleMapIterator_FindValueOrNone() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(stock.FindValueOrNone(func(fruit string, count int) bool { return fruit == "pear" }))
	fmt.Println(stock.FindValueOrNone(func(fruit string, count int) bool { return fruit == "plum" }))
	// Output:
	// 5
	// 0
}

func ExampleMapIterator_ForEach() {
	fruits := itertools.Seq2ToMapIterator(slices.All([]string{"apple", "pear"}))
	err := fruits.ForEach(func(index int, fruit string) error {
		fmt.Println(index, fruit)
		return nil
	})
	fmt.Println(err)
	// Output:
	// 0 apple
	// 1 pear
	// <nil>
}

func ExampleMapIterator_ForEachParallel() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	err := stock.ForEachParallel(func(fruit string, count int) error {
		fmt.Println(fruit, count)
		return nil
	}, itertools.WithLimit(2))
	fmt.Println(err)
	// Unordered output:
	// apple 3
	// pear 5
	// <nil>
}

func ExampleMapIterator_Get() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(*stock.Get("apple"))
	fmt.Println(stock.Get("plum"))
	// Output:
	// 3
	// <nil>
}

func ExampleMapIterator_GetOr() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(stock.GetOr("apple", -1))
	fmt.Println(stock.GetOr("plum", -1))
	// Output:
	// 3
	// -1
}

func ExampleMapIterator_GetOrNone() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(stock.GetOrNone("apple"))
	fmt.Println(stock.GetOrNone("plum"))
	// Output:
	// 3
	// 0
}

func ExampleMapIterator_Concat() {
	first := itertools.Seq2ToMapIterator(slices.All([]string{"apple", "pear"}))
	second := itertools.Seq2ToMapIterator(slices.All([]string{"plum"}))
	for index, fruit := range first.Concat(second) {
		fmt.Println(index, fruit)
	}
	// Output:
	// 0 apple
	// 1 pear
	// 0 plum
}

func ExampleMapIterator_Keys() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	for fruit := range stock.Keys() {
		fmt.Println(fruit)
	}
	// Unordered output:
	// apple
	// pear
}

func ExampleMapIterator_Limit() {
	fruits := itertools.Seq2ToMapIterator(slices.All([]string{"apple", "pear", "plum"}))
	for index, fruit := range fruits.Limit(2) {
		fmt.Println(index, fruit)
	}
	// Output:
	// 0 apple
	// 1 pear
}

func ExampleMapIterator_Map() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	mapped := stock.Map(func(fruit string, count int) (string, string) {
		return strings.ToUpper(fruit), strconv.Itoa(count)
	})
	for fruit, count := range mapped {
		fmt.Printf("%s: %q\n", fruit, count)
	}
	// Unordered output:
	// APPLE: "3"
	// PEAR: "5"
}

func ExampleMapIterator_MapAndCollect() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	result := stock.MapAndCollect(func(fruit string, count int) (string, string) {
		return strings.ToUpper(fruit), strconv.Itoa(count)
	})
	fmt.Println(result)
	// Output:
	// map[APPLE:3 PEAR:5]
}

func ExampleMapIterator_MapAndCollectParallel() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	result := stock.MapAndCollectParallel(func(fruit string, count int) (string, string) {
		return strings.ToUpper(fruit), strconv.Itoa(count)
	}, itertools.WithLimit(2))
	for fruit, count := range result {
		fmt.Printf("%s: %q\n", fruit, count)
	}
	// Unordered output:
	// APPLE: "3"
	// PEAR: "5"
}

func ExampleMapIterator_MapAndCollectWithError() {
	counts := itertools.Seq2ToMapIterator(slices.All([]string{"3", "invalid", "5"}))
	result, err := counts.MapAndCollectWithError(func(index int, text string) (int, int, error) {
		count, err := strconv.Atoi(text)
		return index, count, err
	})
	fmt.Println(result)
	fmt.Println(err)
	// Output:
	// map[0:3]
	// strconv.Atoi: parsing "invalid": invalid syntax
}

func ExampleMapIterator_MapAndCollectWithErrorParallel() {
	stock := itertools.ToMapIterator(map[string]string{"apple": "3", "pear": "invalid", "plum": "5"})
	result, err := stock.MapAndCollectWithErrorParallel(func(fruit, text string) (string, int, error) {
		count, err := strconv.Atoi(text)
		return fruit, count, err
	}, itertools.WithLimit(2))
	for fruit, count := range result {
		fmt.Println(fruit, count)
	}
	fmt.Println(err)
	// Unordered output:
	// apple 3
	// plum 5
	// strconv.Atoi: parsing "invalid": invalid syntax
}

func ExampleMapIterator_MapKeys() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	for fruit, count := range stock.MapKeys(strings.ToUpper) {
		fmt.Println(fruit, count)
	}
	// Unordered output:
	// APPLE 3
	// PEAR 5
}

func ExampleMapIterator_MapValues() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	for fruit, count := range stock.MapValues(strconv.Itoa) {
		fmt.Printf("%s: %q\n", fruit, count)
	}
	// Unordered output:
	// apple: "3"
	// pear: "5"
}

func ExampleMapIterator_Pull() {
	fruits := itertools.Seq2ToMapIterator(slices.All([]string{"apple", "pear"}))
	next, stop := fruits.Pull()
	defer stop()
	fmt.Println(next())
	fmt.Println(next())
	fmt.Println(next())
	// Output:
	// 0 apple true
	// 1 pear true
	// 0  false
}

func ExampleMapIterator_Reduce() {
	fruits := itertools.Seq2ToMapIterator(slices.All([]string{"apple", "pear", "plum"}))
	// The callback receives each pair; the last callback result is returned.
	last := fruits.Reduce("none", func(index int, fruit string) string {
		return fmt.Sprintf("%d: %s", index, fruit)
	})
	fmt.Println(last)
	empty := itertools.Seq2ToMapIterator(slices.All([]string{}))
	fmt.Println(empty.Reduce("none", func(index int, fruit string) string { return fruit }))
	// Output:
	// 2: plum
	// none
}

func ExampleMapIterator_Reverse() {
	fruits := itertools.Seq2ToMapIterator(slices.All([]string{"apple", "pear", "plum"}))
	for index, fruit := range fruits.Reverse() {
		fmt.Println(index, fruit)
	}
	// Output:
	// 2 plum
	// 1 pear
	// 0 apple
}

func ExampleMapIterator_Skip() {
	fruits := itertools.Seq2ToMapIterator(slices.All([]string{"apple", "pear", "plum"}))
	for index, fruit := range fruits.Skip(1) {
		fmt.Println(index, fruit)
	}
	// Output:
	// 1 pear
	// 2 plum
}

func ExampleMapIterator_SkipWhile() {
	counts := itertools.Seq2ToMapIterator(slices.All([]int{0, 0, 3, 0, 5}))
	for index, count := range counts.SkipWhile(func(index, count int) bool { return count == 0 }) {
		fmt.Println(index, count)
	}
	// Output:
	// 2 3
	// 3 0
	// 4 5
}

func ExampleMapIterator_Sorted() {
	stock := itertools.ToMapIterator(map[string]int{"pear": 5, "apple": 3, "plum": 7})
	for fruit, count := range stock.Sorted(cmp.Compare[string]) {
		fmt.Println(fruit, count)
	}
	// Output:
	// apple 3
	// pear 5
	// plum 7
}

func ExampleMapIterator_SortedStable() {
	var seq iter.Seq2[string, int] = func(yield func(string, int) bool) {
		for _, fruit := range []string{"pear", "fig", "plum"} {
			if !yield(fruit, len(fruit)) {
				return
			}
		}
	}
	stock := itertools.Seq2ToMapIterator(seq)
	for fruit, count := range stock.SortedStable(func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	}) {
		fmt.Println(fruit, count)
	}
	// Output:
	// fig 3
	// pear 4
	// plum 4
}

func ExampleMapIterator_TakeWhile() {
	counts := itertools.Seq2ToMapIterator(slices.All([]int{3, 5, 0, 7}))
	for index, count := range counts.TakeWhile(func(index, count int) bool { return count > 0 }) {
		fmt.Println(index, count)
	}
	// Output:
	// 0 3
	// 1 5
}

func ExampleMapIterator_ToSeq2() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	fmt.Println(maps.Collect(stock.ToSeq2()))
	// Output:
	// map[apple:3 pear:5]
}

func ExampleMapIterator_Values() {
	stock := itertools.ToMapIterator(map[string]int{"apple": 3, "pear": 5})
	for count := range stock.Values() {
		fmt.Println(count)
	}
	// Unordered output:
	// 3
	// 5
}
