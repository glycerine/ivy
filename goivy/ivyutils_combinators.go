package goivy

// Combinator functions corresponding to Python's ivy_utils.py lines 15-78.
// Python uses higher-order generator factories; Go uses concrete generic functions.

// Unique returns a new slice with duplicates removed, preserving order.
// Corresponds to Python's unique(gen).
func Unique[T comparable](gen []T) []T {
	memo := make(map[T]struct{})
	var result []T
	for _, x := range gen {
		if _, seen := memo[x]; !seen {
			memo[x] = struct{}{}
			result = append(result, x)
		}
	}
	return result
}

// GenList applies fn to each element of l and collects all results (flattening).
// Corresponds to Python's gen_list(gen, l, *args).
func GenList[T any, U any](fn func(T) []U, l []T) []U {
	var result []U
	for _, x := range l {
		result = append(result, fn(x)...)
	}
	return result
}

// GenListList applies fn to each element of each inner list (double flattening).
// Corresponds to Python's gen_list_list(gen, l, *args).
func GenListList[T any, U any](fn func(T) []U, l [][]T) []U {
	var result []U
	for _, inner := range l {
		for _, x := range inner {
			result = append(result, fn(x)...)
		}
	}
	return result
}

// Pair is a generic pair type used by Unique2 and other functions.
type Pair[X any, Y any] struct {
	First  X
	Second Y
}

// Unique2 given pairs (x,y), yields one x for each unique y value.
// Corresponds to Python's unique2(gen).
func Unique2[X any, Y comparable](pairs []Pair[X, Y]) []X {
	memo := make(map[Y]struct{})
	var result []X
	for _, p := range pairs {
		if _, seen := memo[p.Second]; !seen {
			memo[p.Second] = struct{}{}
			result = append(result, p.First)
		}
	}
	return result
}

// AnyIn returns true if any element in items is in the given set.
// Corresponds to Python's any_in(func) applied.
func AnyIn[T comparable](aSet map[T]struct{}, items []T) bool {
	for _, x := range items {
		if _, ok := aSet[x]; ok {
			return true
		}
	}
	return false
}

// GenToSet applies fn and returns results as a set (map).
// Corresponds to Python's gen_to_set(gen).
func GenToSet[T comparable](items []T) map[T]struct{} {
	result := make(map[T]struct{}, len(items))
	for _, x := range items {
		result[x] = struct{}{}
	}
	return result
}

// Filter2 filters pairs where the second element matches the given value,
// returning a set of first elements.
// Corresponds to Python's filter2(func).
func Filter2[X comparable, Y comparable](pairs []Pair[X, Y], value2 Y) map[X]struct{} {
	result := make(map[X]struct{})
	for _, p := range pairs {
		if p.Second == value2 {
			result[p.First] = struct{}{}
		}
	}
	return result
}
