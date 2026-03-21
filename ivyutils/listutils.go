package ivyutils

// List utility functions corresponding to Python's ivy_utils.py lines 82-147.

// Concat flattens a slice of slices into one slice.
// Corresponds to Python: concat(args) = functools.reduce(operator.add, args, [])
func Concat[T any](args [][]T) []T {
	var result []T
	for _, a := range args {
		result = append(result, a...)
	}
	return result
}

// UnzipAppend unzips a list of N-tuples (represented as [][]T)
// and concatenates each position.
// Corresponds to Python: unzip_append(tups) = (concat(x) for x in zip(*tups))
func UnzipAppend[T any](tups [][]T) [][]T {
	if len(tups) == 0 {
		return nil
	}
	width := len(tups[0])
	result := make([][]T, width)
	for _, tup := range tups {
		for i := 0; i < width && i < len(tup); i++ {
			result[i] = append(result[i], tup[i])
		}
	}
	return result
}

// UnzipPairs unzips a list of pairs into two lists.
// Corresponds to Python: unzip_pairs(tups) = list(zip(*tups)) if tups else ([], [])
func UnzipPairs[A any, B any](pairs []Pair[A, B]) ([]A, []B) {
	if len(pairs) == 0 {
		return nil, nil
	}
	as := make([]A, len(pairs))
	bs := make([]B, len(pairs))
	for i, p := range pairs {
		as[i] = p.First
		bs[i] = p.Second
	}
	return as, bs
}

// Flatten flattens a slice of slices into one slice.
// Corresponds to Python's flatten(l) for one level of nesting.
// (Python's version is recursive for nested lists/tuples; Go's typed
// generics handle one level, which is sufficient for all call sites.)
func Flatten[T any](l [][]T) []T {
	return Concat(l)
}

// UnionOfList unions all sets (represented as maps) into one set.
// Corresponds to Python: union_of_list(sets)
func UnionOfList[T comparable](sets []map[T]struct{}) map[T]struct{} {
	res := make(map[T]struct{})
	for _, s := range sets {
		for k := range s {
			res[k] = struct{}{}
		}
	}
	return res
}

// UnionToList appends elements from fromList to toList, skipping duplicates.
// Mutates toList in-place.
// Corresponds to Python: union_to_list(to_list, from_list)
func UnionToList[T comparable](toList *[]T, fromList []T) {
	used := make(map[T]struct{}, len(*toList))
	for _, x := range *toList {
		used[x] = struct{}{}
	}
	for _, x := range fromList {
		if _, ok := used[x]; !ok {
			*toList = append(*toList, x)
			used[x] = struct{}{}
		}
	}
}

// ListUnion returns the union of two lists preserving order.
// Corresponds to Python: list_union(l1, l2)
func ListUnion[T comparable](l1, l2 []T) []T {
	res := make([]T, len(l1))
	copy(res, l1)
	UnionToList(&res, l2)
	return res
}

// ListDiff returns elements in l1 that are not in l2.
// Corresponds to Python: list_diff(l1, l2) = [s for s in l1 if s not in set(l2)]
func ListDiff[T comparable](l1, l2 []T) []T {
	sl2 := make(map[T]struct{}, len(l2))
	for _, x := range l2 {
		sl2[x] = struct{}{}
	}
	var result []T
	for _, x := range l1 {
		if _, ok := sl2[x]; !ok {
			result = append(result, x)
		}
	}
	return result
}

// DistinctUnorderedPairs yields all (l[i], l[j]) where i < j.
// Corresponds to Python: distinct_unordered_pairs(l)
func DistinctUnorderedPairs[T any](l []T) [][2]T {
	var result [][2]T
	for i := 0; i < len(l)-1; i++ {
		for j := i + 1; j < len(l); j++ {
			result = append(result, [2]T{l[i], l[j]})
		}
	}
	return result
}

// InverseMap inverts a map (swaps keys and values).
// Corresponds to Python: inverse_map(m) = dict((y,x) for x,y in m.items())
func InverseMap[K comparable, V comparable](m map[K]V) map[V]K {
	res := make(map[V]K, len(m))
	for k, v := range m {
		res[v] = k
	}
	return res
}

// ComposeMaps composes two maps as functions.
// A map is assumed to be the identity for keys not present.
// Corresponds to Python: compose_maps(m1, m2)
func ComposeMaps[K comparable](m1, m2 map[K]K) map[K]K {
	res := make(map[K]K, len(m2))
	for k, v := range m2 {
		res[k] = v
	}
	for k, v := range m1 {
		if v2, ok := m2[v]; ok {
			res[k] = v2
		} else {
			res[k] = v
		}
	}
	return res
}

// Partition groups elements by a key function.
// Corresponds to Python: partition(things, key) using collections.defaultdict(list)
func Partition[T any, K comparable](things []T, key func(T) K) map[K][]T {
	res := make(map[K][]T)
	for _, t := range things {
		k := key(t)
		res[k] = append(res[k], t)
	}
	return res
}

// SplitList splits l into two lists based on predicate map p.
// Returns (matching, non-matching).
// Corresponds to Python: split_list(l, p)
func SplitList[T comparable](l []T, p map[T]bool) ([]T, []T) {
	var yes, no []T
	for _, x := range l {
		if p[x] {
			yes = append(yes, x)
		} else {
			no = append(no, x)
		}
	}
	return yes, no
}
