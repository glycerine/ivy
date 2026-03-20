package typeinfer

import (
	"fmt"

	"github.com/glycerine/goivy/logic"
)

// Find performs path-compressing find on a SortOrVar.
func Find(x SortOrVar) SortOrVar {
	sv, ok := x.(*SortVar)
	if !ok || sv.Instance == nil {
		return x
	}
	sv.Instance = Find(sv.Instance)
	return sv.Instance
}

// OccursIn checks if s1 occurs in s2.
func OccursIn(s1, s2 SortOrVar) bool {
	s1 = Find(s1)
	s2 = Find(s2)
	if sortOrVarEqual(s1, s2) {
		return true
	}
	if sw, ok := s2.(*SortWrapper); ok {
		if fs, ok := sw.Sort.(*logic.FunctionSort); ok {
			for _, sub := range fs.Sorts {
				if OccursIn(s1, Wrap(sub)) {
					return true
				}
			}
		}
	}
	if fsv, ok := s2.(*FunctionSortVar); ok {
		for _, sub := range fsv.Sorts {
			if OccursIn(s1, sub) {
				return true
			}
		}
	}
	return false
}

// Unify unifies two SortOrVar values.
func Unify(s1, s2 SortOrVar) error {
	s1 = Find(s1)
	s2 = Find(s2)

	// Same or TopSort
	if sortOrVarEqual(s1, s2) {
		return nil
	}
	if isTopSort(s1) || isTopSort(s2) {
		return nil
	}

	// s1 is SortVar
	if sv1, ok := s1.(*SortVar); ok {
		if sv1.Instance != nil {
			panic("Find should have resolved this")
		}
		if OccursIn(s1, s2) {
			return &logic.SortError{Msg: fmt.Sprintf("Recursive unification: %s, %s", s1, s2)}
		}
		sv1.Instance = s2
		return nil
	}

	// s2 is SortVar
	if _, ok := s2.(*SortVar); ok {
		return Unify(s2, s1)
	}

	// Handle FunctionSortVar ↔ FunctionSortVar
	fsv1, isFSV1 := s1.(*FunctionSortVar)
	fsv2, isFSV2 := s2.(*FunctionSortVar)
	if isFSV1 && isFSV2 && fsv1.Arity() == fsv2.Arity() {
		for i := range fsv1.Sorts {
			if err := Unify(fsv1.Sorts[i], fsv2.Sorts[i]); err != nil {
				return err
			}
		}
		return nil
	}

	// Handle FunctionSortVar ↔ SortWrapper(FunctionSort)
	if isFSV1 {
		if sw2, ok := s2.(*SortWrapper); ok {
			if fs2, ok := sw2.Sort.(*logic.FunctionSort); ok && fsv1.Arity() == fs2.Arity() {
				for i := range fsv1.Sorts {
					if err := Unify(fsv1.Sorts[i], Wrap(fs2.Sorts[i])); err != nil {
						return err
					}
				}
				return nil
			}
		}
	}
	if isFSV2 {
		if sw1, ok := s1.(*SortWrapper); ok {
			if fs1, ok := sw1.Sort.(*logic.FunctionSort); ok && fs1.Arity() == fsv2.Arity() {
				for i := range fsv2.Sorts {
					if err := Unify(Wrap(fs1.Sorts[i]), fsv2.Sorts[i]); err != nil {
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
		return &logic.SortError{Msg: fmt.Sprintf("Cannot unify sorts: %s, %s", s1, s2)}
	}

	fs1, isFS1 := sw1.Sort.(*logic.FunctionSort)
	fs2, isFS2 := sw2.Sort.(*logic.FunctionSort)
	if isFS1 && isFS2 && fs1.Arity() == fs2.Arity() {
		for i := range fs1.Sorts {
			if err := Unify(Wrap(fs1.Sorts[i]), Wrap(fs2.Sorts[i])); err != nil {
				return err
			}
		}
		return nil
	}

	return &logic.SortError{Msg: fmt.Sprintf("Cannot unify sorts: %s, %s", s1, s2)}
}

// ConvertFromSortVars converts sort variables to TopSort.
// Mirrors Python type_inference.py convert_from_sortvars.
func ConvertFromSortVars(s SortOrVar) logic.Sort {
	s = Find(s)
	if _, ok := s.(*SortVar); ok {
		return logic.NewTopSort()
	}
	if fsv, ok := s.(*FunctionSortVar); ok {
		sorts := make([]logic.Sort, len(fsv.Sorts))
		for i, sub := range fsv.Sorts {
			sorts[i] = ConvertFromSortVars(sub)
		}
		result, err := logic.NewFunctionSort(sorts...)
		if err != nil {
			return logic.NewTopSort()
		}
		return result
	}
	if sw, ok := s.(*SortWrapper); ok {
		if fs, ok := sw.Sort.(*logic.FunctionSort); ok {
			sorts := make([]logic.Sort, len(fs.Sorts))
			for i, sub := range fs.Sorts {
				sorts[i] = ConvertFromSortVars(Wrap(sub))
			}
			result, err := logic.NewFunctionSort(sorts...)
			if err != nil {
				return logic.NewTopSort()
			}
			return result
		}
		return sw.Sort
	}
	return logic.NewTopSort()
}

// ConvertToSortVars converts TopSort occurrences to SortVar.
// Mirrors Python type_inference.py convert_to_sortvars.
// Uses FunctionSortVar to preserve SortVar linkage inside function sorts.
func ConvertToSortVars(s logic.Sort) SortOrVar {
	if _, ok := s.(*logic.TopSort); ok {
		return NewSortVar()
	}
	if fs, ok := s.(*logic.FunctionSort); ok {
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
		sorts := make([]logic.Sort, len(sortVars))
		for i, sv := range sortVars {
			sorts[i] = Unwrap(sv)
		}
		result, _ := logic.NewFunctionSort(sorts...)
		return Wrap(result)
	}
	return Wrap(s)
}

// InsertSortVars converts each named TopSort to a new SortVar using env.
// Mirrors Python type_inference.py insert_sortvars.
// Uses FunctionSortVar to preserve SortVar linkage inside function sorts.
func InsertSortVars(s logic.Sort, env map[string]SortOrVar) SortOrVar {
	if ts, ok := s.(*logic.TopSort); ok && ts.IsSortVariable() {
		key := ts.Name
		if sv, exists := env[key]; exists {
			return sv
		}
		sv := NewSortVar()
		env[key] = sv
		return sv
	}
	if fs, ok := s.(*logic.FunctionSort); ok {
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
		sorts := make([]logic.Sort, len(sortVars))
		for i, sv := range sortVars {
			sorts[i] = Unwrap(sv)
		}
		result, _ := logic.NewFunctionSort(sorts...)
		return Wrap(result)
	}
	return Wrap(s)
}

func sortOrVarEqual(a, b SortOrVar) bool {
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
		return aw.Sort.Equal(bw.Sort)
	}
	return false
}

func isTopSort(s SortOrVar) bool {
	if sw, ok := s.(*SortWrapper); ok {
		_, isTop := sw.Sort.(*logic.TopSort)
		return isTop
	}
	return false
}
