package abc

// MiniSat integer sorting and synchronized float permutation sorting from
// src/aig/gia/giaSort.c.

type GiaSortOptions struct {
	MaxEntries int
	Stop       func() bool
}

func checkGiaSortOptions(options GiaSortOptions, entries int) (GiaSortOptions, error) {
	if options.MaxEntries < 0 || entries < 0 {
		return options, ErrDimension
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 1 << 20
	}
	if entries > options.MaxEntries {
		return options, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return options, ErrCancelled
	}
	return options, nil
}

func giaSelectionSortInts(values []int, stop func() bool) error {
	for i := 0; i+1 < len(values); i++ {
		if stop != nil && stop() {
			return ErrCancelled
		}
		best := i
		for j := i + 1; j < len(values); j++ {
			if values[j] < values[best] {
				best = j
			}
		}
		values[i], values[best] = values[best], values[i]
	}
	return nil
}

func giaMiniSatSortInts(values []int, stop func() bool) error {
	if stop != nil && stop() {
		return ErrCancelled
	}
	if len(values) <= 15 {
		return giaSelectionSortInts(values, stop)
	}
	pivot := values[len(values)/2]
	i, j := -1, len(values)
	for {
		if stop != nil && stop() {
			return ErrCancelled
		}
		for {
			i++
			if values[i] >= pivot {
				break
			}
		}
		for {
			j--
			if pivot >= values[j] {
				break
			}
		}
		if i >= j {
			break
		}
		values[i], values[j] = values[j], values[i]
	}
	if err := giaMiniSatSortInts(values[:i], stop); err != nil {
		return err
	}
	return giaMiniSatSortInts(values[i:], stop)
}

// GiaSortInts applies the exact inlined-comparison MiniSat sort. Sorting is
// atomic on cancellation: the caller's slice changes only after completion.
func GiaSortInts(values []int, options GiaSortOptions) error {
	options, err := checkGiaSortOptions(options, len(values))
	if err != nil {
		return err
	}
	work := append([]int(nil), values...)
	if err = giaMiniSatSortInts(work, options.Stop); err != nil {
		return err
	}
	if options.Stop != nil && options.Stop() {
		return ErrCancelled
	}
	copy(values, work)
	return nil
}

func giaSelectionSortFloats(values []float32, permutation []int, stop func() bool) error {
	for i := 0; i+1 < len(values); i++ {
		if stop != nil && stop() {
			return ErrCancelled
		}
		best := i
		for j := i + 1; j < len(values); j++ {
			if values[j] < values[best] {
				best = j
			}
		}
		values[i], values[best] = values[best], values[i]
		permutation[i], permutation[best] = permutation[best], permutation[i]
	}
	return nil
}

func giaMiniSatSortFloats(values []float32, permutation []int, stop func() bool) error {
	if stop != nil && stop() {
		return ErrCancelled
	}
	if len(values) <= 15 {
		return giaSelectionSortFloats(values, permutation, stop)
	}
	pivot := values[len(values)/2]
	i, j := -1, len(values)
	for {
		if stop != nil && stop() {
			return ErrCancelled
		}
		for {
			i++
			if !(values[i] < pivot) {
				break
			}
		}
		for {
			j--
			if !(pivot < values[j]) {
				break
			}
		}
		if i >= j {
			break
		}
		values[i], values[j] = values[j], values[i]
		permutation[i], permutation[j] = permutation[j], permutation[i]
	}
	if err := giaMiniSatSortFloats(values[:i], permutation[:i], stop); err != nil {
		return err
	}
	return giaMiniSatSortFloats(values[i:], permutation[i:], stop)
}

// GiaSortFloats sorts values in place and applies every swap to permutation.
// A nil permutation creates [0,n); a non-nil permutation must have exactly n
// entries and is updated in place. The returned slice is the updated owned or
// caller-provided permutation. Float comparisons deliberately use C's strict
// less-than behavior, so NaNs, signed zeros and equal values retain the exact
// source partition behavior rather than receiving an invented total order.
func GiaSortFloats(values []float32, permutation []int, options GiaSortOptions) ([]int, error) {
	options, err := checkGiaSortOptions(options, len(values))
	if err != nil {
		return nil, err
	}
	provided := permutation != nil
	if provided && len(permutation) != len(values) {
		return nil, ErrDimension
	}
	workValues := append([]float32(nil), values...)
	workPermutation := append([]int(nil), permutation...)
	if !provided {
		workPermutation = make([]int, len(values))
		for i := range workPermutation {
			workPermutation[i] = i
		}
	}
	if err = giaMiniSatSortFloats(workValues, workPermutation, options.Stop); err != nil {
		return nil, err
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	copy(values, workValues)
	if provided {
		copy(permutation, workPermutation)
		return permutation, nil
	}
	return workPermutation, nil
}
