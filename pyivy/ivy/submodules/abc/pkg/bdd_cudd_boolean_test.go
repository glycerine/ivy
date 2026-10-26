package abc

// Go-only semantic tests for the CUDD Boolean-operation port.

import (
	"errors"
	"testing"
)

type cuddBooleanTruthKey struct {
	variable int
	truth    uint64
}

func cuddBooleanBuildTruth(t *testing.T, manager *ExtraBDDManager, truth uint64, variables int) ExtraBDDLit {
	t.Helper()
	memo := make(map[cuddBooleanTruthKey]ExtraBDDLit)
	var build func(int, uint64) ExtraBDDLit
	build = func(variable int, word uint64) ExtraBDDLit {
		if variable == variables {
			if word&1 != 0 {
				return ExtraBDDTrue
			}
			return ExtraBDDFalse
		}
		key := cuddBooleanTruthKey{variable: variable, truth: word}
		if result, ok := memo[key]; ok {
			return result
		}
		remaining := variables - variable
		entries := 1 << remaining
		var lowWord, highWord uint64
		for index := 0; index < entries; index++ {
			if word>>index&1 == 0 {
				continue
			}
			if index&1 == 0 {
				lowWord |= uint64(1) << (index >> 1)
			} else {
				highWord |= uint64(1) << (index >> 1)
			}
		}
		high := build(variable+1, highWord)
		low := build(variable+1, lowWord)
		result, err := manager.uniqueCreate(variable, high, low, nil)
		if err != nil {
			t.Fatalf("build truth %x at variable %d: %v", truth, variable, err)
		}
		memo[key] = result
		return result
	}
	return build(0, truth)
}

func cuddBooleanAssignment(index, variables int) []bool {
	assignment := make([]bool, variables)
	for variable := range assignment {
		assignment[variable] = index>>variable&1 != 0
	}
	return assignment
}

func cuddBooleanTruthValue(truth uint64, assignment []bool) bool {
	index := 0
	for variable, value := range assignment {
		if value {
			index |= 1 << variable
		}
	}
	return truth>>index&1 != 0
}

func cuddBooleanAssert(t *testing.T, manager *ExtraBDDManager, root ExtraBDDLit, expected func([]bool) bool) {
	t.Helper()
	for index := 0; index < 1<<manager.variables; index++ {
		assignment := cuddBooleanAssignment(index, manager.variables)
		got, err := manager.Evaluate(root, assignment)
		if err != nil {
			t.Fatalf("evaluate %d: %v", index, err)
		}
		if want := expected(assignment); got != want {
			t.Fatalf("assignment %v: got %t want %t", assignment, got, want)
		}
	}
}

func TestCUDDBooleanOperations(t *testing.T) {
	manager, err := NewExtraBDDManager(4, 1<<16)
	if err != nil {
		t.Fatal(err)
	}
	fTruth, gTruth := uint64(0x6a3c), uint64(0xc5a7)
	f := cuddBooleanBuildTruth(t, manager, fTruth, 4)
	g := cuddBooleanBuildTruth(t, manager, gTruth, 4)
	x := make([]ExtraBDDLit, 4)
	for variable := range x {
		x[variable], err = manager.Variable(variable)
		if err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name string
		make func() (ExtraBDDLit, error)
		want func(bool, bool) bool
	}{
		{"ite", func() (ExtraBDDLit, error) { return manager.Ite(x[0], f, g, ExtraBDDOptions{}) }, func(fv, gv bool) bool { return fv }},
		{"intersect", func() (ExtraBDDLit, error) { return manager.Intersect(f, g, ExtraBDDOptions{}) }, func(fv, gv bool) bool { return fv && gv }},
		{"nand", func() (ExtraBDDLit, error) { return manager.Nand(f, g, ExtraBDDOptions{}) }, func(fv, gv bool) bool { return !(fv && gv) }},
		{"nor", func() (ExtraBDDLit, error) { return manager.Nor(f, g, ExtraBDDOptions{}) }, func(fv, gv bool) bool { return !(fv || gv) }},
		{"xor", func() (ExtraBDDLit, error) { return manager.Xor(f, g, ExtraBDDOptions{}) }, func(fv, gv bool) bool { return fv != gv }},
		{"xnor", func() (ExtraBDDLit, error) { return manager.Xnor(f, g, ExtraBDDOptions{}) }, func(fv, gv bool) bool { return fv == gv }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, makeErr := test.make()
			if makeErr != nil {
				t.Fatal(makeErr)
			}
			cuddBooleanAssert(t, manager, root, func(assignment []bool) bool {
				fv, gv := cuddBooleanTruthValue(fTruth, assignment), cuddBooleanTruthValue(gTruth, assignment)
				if test.name == "ite" {
					if assignment[0] {
						return fv
					}
					return gv
				}
				return test.want(fv, gv)
			})
		})
	}
	constant, isConstant, err := manager.IteConstant(f, ExtraBDDTrue, ExtraBDDTrue, ExtraBDDOptions{})
	if err != nil || !isConstant || constant != ExtraBDDTrue {
		t.Fatalf("constant ITE = %d,%t,%v", constant, isConstant, err)
	}
	leq, err := manager.Leq(f, g, ExtraBDDOptions{})
	wantLeq := true
	for index := 0; index < 16; index++ {
		assignment := cuddBooleanAssignment(index, 4)
		if cuddBooleanTruthValue(fTruth, assignment) && !cuddBooleanTruthValue(gTruth, assignment) {
			wantLeq = false
		}
	}
	if err != nil || leq != wantLeq {
		t.Fatalf("Leq = %t,%v want %t", leq, err, wantLeq)
	}
}

func TestCUDDBooleanAbstractionCofactorAndCompose(t *testing.T) {
	manager, err := NewExtraBDDManager(4, 1<<17)
	if err != nil {
		t.Fatal(err)
	}
	fTruth, gTruth := uint64(0x936c), uint64(0x2db7)
	f := cuddBooleanBuildTruth(t, manager, fTruth, 4)
	g := cuddBooleanBuildTruth(t, manager, gTruth, 4)
	quantified := []int{0, 2}
	checkQuantified := func(assignment []bool, combine func(bool, bool) bool, value func([]bool) bool, identity bool) bool {
		result := identity
		copyAssignment := append([]bool(nil), assignment...)
		for mask := 0; mask < 4; mask++ {
			copyAssignment[0] = mask&1 != 0
			copyAssignment[2] = mask&2 != 0
			result = combine(result, value(copyAssignment))
		}
		return result
	}
	exist, err := manager.ExistAbstract(f, quantified, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddBooleanAssert(t, manager, exist, func(a []bool) bool {
		return checkQuantified(a, func(x, y bool) bool { return x || y }, func(x []bool) bool { return cuddBooleanTruthValue(fTruth, x) }, false)
	})
	universal, err := manager.UnivAbstract(f, quantified, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddBooleanAssert(t, manager, universal, func(a []bool) bool {
		return checkQuantified(a, func(x, y bool) bool { return x && y }, func(x []bool) bool { return cuddBooleanTruthValue(fTruth, x) }, true)
	})
	andAbstract, err := manager.AndAbstract(f, g, quantified, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddBooleanAssert(t, manager, andAbstract, func(a []bool) bool {
		return checkQuantified(a, func(x, y bool) bool { return x || y }, func(x []bool) bool {
			return cuddBooleanTruthValue(fTruth, x) && cuddBooleanTruthValue(gTruth, x)
		}, false)
	})
	xorAbstract, err := manager.XorExistAbstract(f, g, quantified, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddBooleanAssert(t, manager, xorAbstract, func(a []bool) bool {
		return checkQuantified(a, func(x, y bool) bool { return x || y }, func(x []bool) bool {
			return cuddBooleanTruthValue(fTruth, x) != cuddBooleanTruthValue(gTruth, x)
		}, false)
	})

	difference, err := manager.BooleanDifference(f, 1, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddBooleanAssert(t, manager, difference, func(a []bool) bool {
		copyAssignment := append([]bool(nil), a...)
		copyAssignment[1] = false
		low := cuddBooleanTruthValue(fTruth, copyAssignment)
		copyAssignment[1] = true
		return low != cuddBooleanTruthValue(fTruth, copyAssignment)
	})
	dependent, err := manager.VariableDependent(f, 1, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantDependent := true
	for index := 0; index < 16; index++ {
		a := cuddBooleanAssignment(index, 4)
		a[1] = false
		low := cuddBooleanTruthValue(fTruth, a)
		a[1] = true
		if low && cuddBooleanTruthValue(fTruth, a) {
			wantDependent = false
		}
	}
	if dependent != wantDependent {
		t.Fatalf("dependent = %t want %t", dependent, wantDependent)
	}

	x0, _ := manager.Variable(0)
	x2, _ := manager.Variable(2)
	cube, err := manager.And(x0, x2.Not(), ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if ok, checkErr := manager.CheckCube(cube); checkErr != nil || !ok {
		t.Fatalf("cube check = %t,%v", ok, checkErr)
	}
	cofactor, err := manager.Cofactor(f, cube, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddBooleanAssert(t, manager, cofactor, func(a []bool) bool {
		copyAssignment := append([]bool(nil), a...)
		copyAssignment[0], copyAssignment[2] = true, false
		return cuddBooleanTruthValue(fTruth, copyAssignment)
	})
	if ok, checkErr := manager.CheckCube(f); checkErr != nil || ok {
		t.Fatalf("noncube check = %t,%v", ok, checkErr)
	}

	composed, err := manager.Compose(f, g, 2, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddBooleanAssert(t, manager, composed, func(a []bool) bool {
		copyAssignment := append([]bool(nil), a...)
		copyAssignment[2] = cuddBooleanTruthValue(gTruth, a)
		return cuddBooleanTruthValue(fTruth, copyAssignment)
	})
	permutation := []int{2, 0, 3, 1}
	permuted, err := manager.Permute(f, permutation, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddBooleanAssert(t, manager, permuted, func(a []bool) bool {
		original := make([]bool, 4)
		for variable, image := range permutation {
			original[variable] = a[image]
		}
		return cuddBooleanTruthValue(fTruth, original)
	})
	swapped, err := manager.SwapVariables(f, []int{0, 1}, []int{2, 3}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddBooleanAssert(t, manager, swapped, func(a []bool) bool {
		return cuddBooleanTruthValue(fTruth, []bool{a[2], a[3], a[0], a[1]})
	})
	adjacent, err := manager.AdjacentPermuteX(f, []int{1, 2, 0}, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cuddBooleanAssert(t, manager, adjacent, func(a []bool) bool {
		return cuddBooleanTruthValue(fTruth, []bool{a[0], a[2], a[1], a[3]})
	})
	unchanged, err := manager.AdjacentPermuteX(f, []int{1, 2}, ExtraBDDOptions{})
	if err != nil || unchanged != f {
		t.Fatalf("short adjacent group = %d,%v want unchanged %d", unchanged, err, f)
	}
}

func TestCUDDBooleanAllFourVariableFunctions(t *testing.T) {
	manager, err := NewExtraBDDManager(4, 1<<19)
	if err != nil {
		t.Fatal(err)
	}
	permutation := []int{1, 0, 2, 3}
	for truth := uint64(0); truth < 1<<16; truth++ {
		root := cuddBooleanBuildTruth(t, manager, truth, 4)
		difference, diffErr := manager.BooleanDifference(root, 0, ExtraBDDOptions{})
		exist, existErr := manager.ExistAbstract(root, []int{0}, ExtraBDDOptions{})
		universal, univErr := manager.UnivAbstract(root, []int{0}, ExtraBDDOptions{})
		permuted, permErr := manager.Permute(root, permutation, ExtraBDDOptions{})
		if diffErr != nil || existErr != nil || univErr != nil || permErr != nil {
			t.Fatalf("truth %04x: diff=%v exist=%v univ=%v perm=%v", truth, diffErr, existErr, univErr, permErr)
		}
		for index := 0; index < 16; index++ {
			a := cuddBooleanAssignment(index, 4)
			lowIndex, highIndex := index&^1, index|1
			low, high := truth>>lowIndex&1 != 0, truth>>highIndex&1 != 0
			checks := []struct {
				root ExtraBDDLit
				want bool
			}{
				{difference, low != high},
				{exist, low || high},
				{universal, low && high},
				{permuted, truth>>((index&^3)|((index&1)<<1)|((index&2)>>1))&1 != 0},
			}
			for _, check := range checks {
				got, evalErr := manager.Evaluate(check.root, a)
				if evalErr != nil || got != check.want {
					t.Fatalf("truth %04x assignment %d root %d = %t,%v want %t", truth, index, check.root, got, evalErr, check.want)
				}
			}
		}
	}
}

func TestCUDDBooleanErrorsAndRollback(t *testing.T) {
	manager, err := NewExtraBDDManager(4, 1024)
	if err != nil {
		t.Fatal(err)
	}
	f := cuddBooleanBuildTruth(t, manager, 0x6a3c, 4)
	g := cuddBooleanBuildTruth(t, manager, 0xc5a7, 4)
	beforeNodes, beforeLookups, beforeMisses := len(manager.nodes), manager.cacheLookups, manager.cacheMisses
	if result, limitErr := manager.Ite(f, g, g.Not(), ExtraBDDOptions{MaxOperations: 1}); result != 0 || !errors.Is(limitErr, ErrLimit) {
		t.Fatalf("limit = %d,%v", result, limitErr)
	}
	if len(manager.nodes) != beforeNodes || manager.cacheLookups != beforeLookups || manager.cacheMisses != beforeMisses {
		t.Fatal("failed operation changed manager")
	}
	limited, limitCreateErr := NewExtraBDDManager(2, 64)
	if limitCreateErr != nil {
		t.Fatal(limitCreateErr)
	}
	x0, _ := limited.Variable(0)
	x1, _ := limited.Variable(1)
	limitedNodes := len(limited.nodes)
	if result, limitErr := limited.AndLimit(x0, x1, 0, ExtraBDDOptions{}); result != 0 || !errors.Is(limitErr, ErrLimit) {
		t.Fatalf("AND node limit = %d,%v", result, limitErr)
	}
	if len(limited.nodes) != limitedNodes {
		t.Fatal("AND node-limit failure was not atomic")
	}
	if result, limitErr := limited.AndAbstractLimit(x0, x1, nil, 0, ExtraBDDOptions{}); result != 0 || !errors.Is(limitErr, ErrLimit) {
		t.Fatalf("AND-abstract node limit = %d,%v", result, limitErr)
	}
	if len(limited.nodes) != limitedNodes {
		t.Fatal("AND-abstract node-limit failure was not atomic")
	}
	if result, cancelErr := manager.ExistAbstract(f, []int{0, 1}, ExtraBDDOptions{Stop: func() bool { return true }}); result != 0 || !errors.Is(cancelErr, ErrCancelled) {
		t.Fatalf("cancel = %d,%v", result, cancelErr)
	}
	invalid := ExtraBDDLit(len(manager.nodes)+1) << 1
	if _, invalidErr := manager.Ite(invalid, f, g, ExtraBDDOptions{}); !errors.Is(invalidErr, ErrInvalidLiteral) {
		t.Fatalf("invalid literal: %v", invalidErr)
	}
	badCalls := []func() error{
		func() error { _, callErr := manager.Permute(f, []int{0, 0, 2, 3}, ExtraBDDOptions{}); return callErr },
		func() error {
			_, callErr := manager.VectorCompose(f, []ExtraBDDLit{f}, ExtraBDDOptions{})
			return callErr
		},
		func() error {
			_, callErr := manager.SwapVariables(f, []int{0}, []int{0}, ExtraBDDOptions{})
			return callErr
		},
		func() error {
			_, callErr := manager.Restrict(f, []ExtraBDDAssignment{{Variable: 0}, {Variable: 0, Value: true}}, ExtraBDDOptions{})
			return callErr
		},
		func() error { _, callErr := manager.Cofactor(f, ExtraBDDFalse, ExtraBDDOptions{}); return callErr },
	}
	for index, call := range badCalls {
		if call() == nil {
			t.Fatalf("bad call %d succeeded", index)
		}
	}
	if _, nilErr := (*ExtraBDDManager)(nil).Permute(f, nil, ExtraBDDOptions{}); !errors.Is(nilErr, ErrDimension) {
		t.Fatalf("nil manager: %v", nilErr)
	}
}

func FuzzCUDDBooleanOperations(f *testing.F) {
	f.Add(uint16(0x6a3c), uint16(0xc5a7), uint8(0), uint8(5))
	f.Add(uint16(0xffff), uint16(0x0000), uint8(7), uint8(3))
	f.Add(uint16(0xaaaa), uint16(0xcccc), uint8(10), uint8(0))
	f.Fuzz(func(t *testing.T, fTruth16, gTruth16 uint16, operation, mask uint8) {
		manager, err := NewExtraBDDManager(4, 4096)
		if err != nil {
			t.Fatal(err)
		}
		fTruth, gTruth := uint64(fTruth16), uint64(gTruth16)
		first := cuddBooleanBuildTruth(t, manager, fTruth, 4)
		second := cuddBooleanBuildTruth(t, manager, gTruth, 4)
		quantified := make([]int, 0, 4)
		for variable := 0; variable < 4; variable++ {
			if mask>>variable&1 != 0 {
				quantified = append(quantified, variable)
			}
		}
		var root ExtraBDDLit
		var makeErr error
		op := operation % 10
		switch op {
		case 0:
			root, makeErr = manager.Ite(first, second, second.Not(), ExtraBDDOptions{})
		case 1:
			root, makeErr = manager.Xor(first, second, ExtraBDDOptions{})
		case 2:
			root, makeErr = manager.ExistAbstract(first, quantified, ExtraBDDOptions{})
		case 3:
			root, makeErr = manager.UnivAbstract(first, quantified, ExtraBDDOptions{})
		case 4:
			root, makeErr = manager.AndAbstract(first, second, quantified, ExtraBDDOptions{})
		case 5:
			root, makeErr = manager.XorExistAbstract(first, second, quantified, ExtraBDDOptions{})
		case 6:
			root, makeErr = manager.BooleanDifference(first, int(mask&3), ExtraBDDOptions{})
		case 7:
			root, makeErr = manager.Compose(first, second, int(mask&3), ExtraBDDOptions{})
		case 8:
			root, makeErr = manager.Permute(first, []int{1, 0, 3, 2}, ExtraBDDOptions{})
		case 9:
			root, makeErr = manager.Restrict(first, []ExtraBDDAssignment{{Variable: int(mask & 3), Value: mask&4 != 0}}, ExtraBDDOptions{})
		}
		if makeErr != nil {
			t.Fatal(makeErr)
		}
		for index := 0; index < 16; index++ {
			assignment := cuddBooleanAssignment(index, 4)
			fv, gv := cuddBooleanTruthValue(fTruth, assignment), cuddBooleanTruthValue(gTruth, assignment)
			want := false
			switch op {
			case 0:
				if fv {
					want = gv
				} else {
					want = !gv
				}
			case 1:
				want = fv != gv
			case 2, 3, 4, 5:
				want = op == 3
				count := 1 << len(quantified)
				for values := 0; values < count; values++ {
					candidate := append([]bool(nil), assignment...)
					for offset, variable := range quantified {
						candidate[variable] = values>>offset&1 != 0
					}
					value := cuddBooleanTruthValue(fTruth, candidate)
					if op == 4 {
						value = value && cuddBooleanTruthValue(gTruth, candidate)
					} else if op == 5 {
						value = value != cuddBooleanTruthValue(gTruth, candidate)
					}
					if op == 3 {
						want = want && value
					} else {
						want = want || value
					}
				}
			case 6:
				variable := int(mask & 3)
				candidate := append([]bool(nil), assignment...)
				candidate[variable] = false
				low := cuddBooleanTruthValue(fTruth, candidate)
				candidate[variable] = true
				want = low != cuddBooleanTruthValue(fTruth, candidate)
			case 7:
				candidate := append([]bool(nil), assignment...)
				candidate[int(mask&3)] = gv
				want = cuddBooleanTruthValue(fTruth, candidate)
			case 8:
				want = cuddBooleanTruthValue(fTruth, []bool{assignment[1], assignment[0], assignment[3], assignment[2]})
			case 9:
				candidate := append([]bool(nil), assignment...)
				candidate[int(mask&3)] = mask&4 != 0
				want = cuddBooleanTruthValue(fTruth, candidate)
			}
			got, evalErr := manager.Evaluate(root, assignment)
			if evalErr != nil || got != want {
				t.Fatalf("op %d f=%04x g=%04x mask=%x assignment=%v: got %t,%v want %t", op, fTruth, gTruth, mask, assignment, got, evalErr, want)
			}
		}
	})
}
