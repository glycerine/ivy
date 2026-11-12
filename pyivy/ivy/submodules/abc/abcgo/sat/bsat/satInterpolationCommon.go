package bsat

import "fmt"

// Variable types follow ABC: A-local=1, B-local=0, shared=-(position+1).
func interpolationVariables(cnf *Sto_Man_t, requested []int32) (types, shared []int32, err error) {
	if cnf == nil || cnf.nVars < 0 {
		return nil, nil, fmt.Errorf("invalid interpolation CNF")
	}
	inA, inB := make([]bool, cnf.nVars), make([]bool, cnf.nVars)
	for c := cnf.pHead; c != nil; c = c.pNext {
		if c.fRoot == 0 {
			continue
		}
		for _, l := range c.pLits {
			v := lit_var(l)
			if l < 0 || v >= cnf.nVars {
				return nil, nil, fmt.Errorf("invalid interpolation literal")
			}
			if c.fA != 0 {
				inA[v] = true
			} else {
				inB[v] = true
			}
		}
	}
	types = make([]int32, cnf.nVars)
	for v := range types {
		if inA[v] {
			types[v] = 1
		}
	}
	if len(requested) == 0 {
		for v := int32(0); v < cnf.nVars; v++ {
			if inA[v] && inB[v] {
				shared = append(shared, v)
			}
		}
	} else {
		shared = append([]int32(nil), requested...)
	}
	seen := map[int32]bool{}
	for position, v := range shared {
		if v < 0 || v >= cnf.nVars || seen[v] {
			return nil, nil, fmt.Errorf("invalid or duplicate shared variable %d", v)
		}
		seen[v] = true
		types[v] = -int32(position) - 1
	}
	return types, shared, nil
}
