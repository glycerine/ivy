package goivy

import (
	"fmt"
)

// Find performs path-compressing find on a SortOrVar.
func TypeInferFind(x SortOrVar) SortOrVar {
	sv, ok := x.(*SortVar)
	if !ok || sv.Instance == nil {
		return x
	}
	sv.Instance = TypeInferFind(sv.Instance)
	return sv.Instance
}

// OccursIn checks if s1 occurs in s2.
func OccursIn(s1, s2 SortOrVar) bool {
	if s1 == nil || s2 == nil {
		return false
	}
	s1 = TypeInferFind(s1)
	s2 = TypeInferFind(s2)
	if s1 == nil || s2 == nil {
		return false
	}
	if sortOrVarEqual(s1, s2) {
		return true
	}
	if sw, ok := s2.(*SortWrapper); ok {
		if fs, ok := sw.Sort.(*FunctionSort); ok && fs != nil {
			for _, sub := range fs.Sorts {
				if sub == nil {
					continue
				}
				if OccursIn(s1, Wrap(sub)) {
					return true
				}
			}
		}
	}
	if fsv, ok := s2.(*FunctionSortVar); ok {
		for _, sub := range fsv.Sorts {
			if sub == nil {
				continue
			}
			if OccursIn(s1, sub) {
				return true
			}
		}
	}
	return false
}

// Unify unifies two SortOrVar values.
func TypeInferUnify(s1, s2 SortOrVar) error {
	if s1 == nil || s2 == nil {
		return nil
	}
	s1 = TypeInferFind(s1)
	s2 = TypeInferFind(s2)
	if s1 == nil || s2 == nil {
		return nil
	}

	// Same or TopSort
	if sortOrVarEqual(s1, s2) {
		return nil
	}
	if typeinferIsTopSort(s1) || typeinferIsTopSort(s2) {
		return nil
	}

	// s1 is SortVar
	if sv1, ok := s1.(*SortVar); ok {
		if sv1.Instance != nil {
			panic("Find should have resolved this")
		}
		if OccursIn(s1, s2) {
			return &SortError{Msg: fmt.Sprintf("Recursive unification: %s, %s", s1, s2)}
		}
		sv1.Instance = s2
		return nil
	}

	// s2 is SortVar
	if _, ok := s2.(*SortVar); ok {
		return TypeInferUnify(s2, s1)
	}

	// Handle FunctionSortVar ↔ FunctionSortVar
	fsv1, isFSV1 := s1.(*FunctionSortVar)
	fsv2, isFSV2 := s2.(*FunctionSortVar)
	if isFSV1 && isFSV2 && fsv1.Arity() == fsv2.Arity() {
		for i := range fsv1.Sorts {
			if err := TypeInferUnify(fsv1.Sorts[i], fsv2.Sorts[i]); err != nil {
				return err
			}
		}
		return nil
	}

	// Handle FunctionSortVar ↔ SortWrapper(FunctionSort)
	if isFSV1 {
		if sw2, ok := s2.(*SortWrapper); ok {
			if fs2, ok := sw2.Sort.(*FunctionSort); ok && fs2 != nil && fsv1.Arity() == fs2.Arity() {
				for i := range fsv1.Sorts {
					if err := TypeInferUnify(fsv1.Sorts[i], Wrap(fs2.Sorts[i])); err != nil {
						return err
					}
				}
				return nil
			}
		}
	}
	if isFSV2 {
		if sw1, ok := s1.(*SortWrapper); ok {
			if fs1, ok := sw1.Sort.(*FunctionSort); ok && fs1 != nil && fs1.Arity() == fsv2.Arity() {
				for i := range fsv2.Sorts {
					if err := TypeInferUnify(Wrap(fs1.Sorts[i]), fsv2.Sorts[i]); err != nil {
						return err
					}
				}
				return nil
			}
		}
	}

	// Both are concrete sorts (SortWrapper ↔ SortWrapper)
	sw1, ok1 := s1.(*SortWrapper)
	sw2, ok2 := s2.(*SortWrapper)
	if !ok1 || !ok2 {
		return &SortError{Msg: fmt.Sprintf("Cannot unify sorts: %s, %s", s1, s2)}
	}

	fs1, isFS1 := sw1.Sort.(*FunctionSort)
	fs2, isFS2 := sw2.Sort.(*FunctionSort)
	if isFS1 && isFS2 && fs1 != nil && fs2 != nil && fs1.Arity() == fs2.Arity() {
		for i := range fs1.Sorts {
			if err := TypeInferUnify(Wrap(fs1.Sorts[i]), Wrap(fs2.Sorts[i])); err != nil {
				return err
			}
		}
		return nil
	}

	return &SortError{Msg: fmt.Sprintf("Cannot unify sorts: %s, %s", s1, s2)}
}

// ConvertFromSortVars converts sort variables to TopSort.
// Mirrors Python type_inference.py convert_from_sortvars.
func ConvertFromSortVars(s SortOrVar) Sort {
	s = TypeInferFind(s)
	if _, ok := s.(*SortVar); ok {
		return NewTopSort()
	}
	if fsv, ok := s.(*FunctionSortVar); ok {
		sorts := make([]Sort, len(fsv.Sorts))
		for i, sub := range fsv.Sorts {
			sorts[i] = ConvertFromSortVars(sub)
		}
		result, err := NewFunctionSort(sorts...)
		if err != nil {
			return NewTopSort()
		}
		return result
	}
	if sw, ok := s.(*SortWrapper); ok {
		if sortIsNil(sw.Sort) {
			return NewTopSort()
		}
		if fs, ok := sw.Sort.(*FunctionSort); ok && fs != nil {
			sorts := make([]Sort, len(fs.Sorts))
			for i, sub := range fs.Sorts {
				if sub == nil {
					sorts[i] = NewTopSort()
					continue
				}
				sorts[i] = ConvertFromSortVars(Wrap(sub))
			}
			result, err := NewFunctionSort(sorts...)
			if err != nil {
				return NewTopSort()
			}
			return result
		}
		return sw.Sort
	}
	return NewTopSort()
}

// ConvertToSortVars converts TopSort occurrences to SortVar.
// Mirrors Python type_inference.py convert_to_sortvars.
// Uses FunctionSortVar to preserve SortVar linkage inside function sorts.
func ConvertToSortVars(s Sort) SortOrVar {
	if _, ok := s.(*TopSort); ok {
		return NewSortVar()
	}
	if fs, ok := s.(*FunctionSort); ok {
		hasSortVar := false
		sortVars := make([]SortOrVar, len(fs.Sorts))
		for i, sub := range fs.Sorts {
			sv := ConvertToSortVars(sub)
			sortVars[i] = sv
			if _, isSV := sv.(*SortVar); isSV {
				hasSortVar = true
			}
			if _, isFSV := sv.(*FunctionSortVar); isFSV {
				hasSortVar = true
			}
		}
		if hasSortVar {
			// Use FunctionSortVar to keep SortVar linkage intact.
			// This mirrors Python where FunctionSort(*[SortVar(), ...]) works.
			return NewFunctionSortVar(sortVars...)
		}
		// All elements are concrete — use plain FunctionSort.
		sorts := make([]Sort, len(sortVars))
		for i, sv := range sortVars {
			sorts[i] = Unwrap(sv)
		}
		result, _ := NewFunctionSort(sorts...)
		return Wrap(result)
	}
	return Wrap(s)
}

// InsertSortVars converts each named TopSort to a new SortVar using env.
// Mirrors Python type_inference.py insert_sortvars.
// Uses FunctionSortVar to preserve SortVar linkage inside function sorts.
func InsertSortVars(s Sort, env map[string]SortOrVar) SortOrVar {
	if ts, ok := s.(*TopSort); ok && ts.IsSortVariable() {
		key := ts.Name
		if sv, exists := env[key]; exists {
			return sv
		}
		sv := NewSortVar()
		env[key] = sv
		return sv
	}
	if fs, ok := s.(*FunctionSort); ok {
		hasSortVar := false
		sortVars := make([]SortOrVar, len(fs.Sorts))
		for i, sub := range fs.Sorts {
			sv := InsertSortVars(sub, env)
			sortVars[i] = sv
			if _, isSV := sv.(*SortVar); isSV {
				hasSortVar = true
			}
			if _, isFSV := sv.(*FunctionSortVar); isFSV {
				hasSortVar = true
			}
		}
		if hasSortVar {
			return NewFunctionSortVar(sortVars...)
		}
		sorts := make([]Sort, len(sortVars))
		for i, sv := range sortVars {
			sorts[i] = Unwrap(sv)
		}
		result, _ := NewFunctionSort(sorts...)
		return Wrap(result)
	}
	return Wrap(s)
}

func sortOrVarEqual(a, b SortOrVar) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	// Pointer equality for SortVar
	if av, ok := a.(*SortVar); ok {
		if bv, ok := b.(*SortVar); ok {
			return av == bv
		}
		return false
	}
	// Sort equality
	aw, aOk := a.(*SortWrapper)
	bw, bOk := b.(*SortWrapper)
	if aOk && bOk {
		if sortIsNil(aw.Sort) || sortIsNil(bw.Sort) {
			return sortIsNil(aw.Sort) && sortIsNil(bw.Sort)
		}
		return aw.Sort.Equal(bw.Sort)
	}
	return false
}

func typeinferIsTopSort(s SortOrVar) bool {
	if sw, ok := s.(*SortWrapper); ok {
		if sortIsNil(sw.Sort) {
			return false
		}
		_, isTop := sw.Sort.(*TopSort)
		return isTop
	}
	return false
}
