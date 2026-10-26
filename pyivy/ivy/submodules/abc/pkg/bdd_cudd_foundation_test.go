package abc

// Stand-alone validation for the concrete CUDD foundation.

import (
	"math"
	"slices"
	"strings"
	"sync"
	"testing"
)

func newTestCuddFoundation(t *testing.T, variables int) *CuddFoundation {
	t.Helper()
	foundation, err := NewCuddFoundation(CuddFoundationOptions{Variables: variables, BDDCapacity: 512, ADDCapacity: 512, CacheSlots: 32, MaxOperations: 1 << 20, MaxBytes: 4 << 20})
	if err != nil {
		t.Fatal(err)
	}
	return foundation
}

func TestCuddFoundationLifecycleAndAPI(t *testing.T) {
	foundation := newTestCuddFoundation(t, 4)
	bdd, err := foundation.BDDManager()
	if err != nil {
		t.Fatal(err)
	}
	add, err := foundation.ADDManager()
	if err != nil {
		t.Fatal(err)
	}
	a, _ := foundation.BDDVariable(0)
	b, _ := foundation.BDDVariable(1)
	conjunction, err := bdd.And(a, b.Not(), ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 16; assignment++ {
		values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
		got, evalErr := bdd.Evaluate(conjunction, values)
		if evalErr != nil || got != (values[0] && !values[1]) {
			t.Fatalf("assignment=%d got=%v err=%v", assignment, got, evalErr)
		}
	}
	x, _ := foundation.ADDVariable(0)
	y, _ := foundation.ADDVariable(1)
	sum, err := add.Apply(CuddADDPlus, x, y, ExtraBDDOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < 4; assignment++ {
		got, evalErr := add.Evaluate(sum, assignment)
		want := float64((assignment & 1) + (assignment >> 1 & 1))
		if evalErr != nil || got != want {
			t.Fatalf("ADD assignment=%d got=%g want=%g err=%v", assignment, got, want, evalErr)
		}
	}
	if err = foundation.RefBDD(conjunction); err != nil {
		t.Fatal(err)
	}
	if err = foundation.RefBDD(conjunction.Not()); err != nil {
		t.Fatal(err)
	}
	count, err := foundation.ReferenceCount(CuddCacheBDD, uint32(conjunction))
	if err != nil || count != 2 {
		t.Fatalf("BDD refs=%d err=%v", count, err)
	}
	if err = foundation.DerefBDD(conjunction.Not()); err != nil {
		t.Fatal(err)
	}
	if err = foundation.RefADD(sum); err != nil {
		t.Fatal(err)
	}
	key := CuddCacheKey{Domain: CuddCacheBDD, Operation: 17, Arity: 2, Operands: [3]uint32{uint32(a), uint32(b)}}
	if err = foundation.CacheInsert(key, uint32(conjunction)); err != nil {
		t.Fatal(err)
	}
	result, hit, err := foundation.CacheLookup(key)
	if err != nil || !hit || result != uint32(conjunction) {
		t.Fatalf("cache result=%d hit=%v err=%v", result, hit, err)
	}
	if err = foundation.SetPermutation([]int{2, 0, 3, 1}); err != nil {
		t.Fatal(err)
	}
	permutation, inverse, err := foundation.Permutation()
	if err != nil || !slices.Equal(permutation, []int{2, 0, 3, 1}) || !slices.Equal(inverse, []int{1, 3, 0, 2}) {
		t.Fatalf("perm=%v inv=%v err=%v", permutation, inverse, err)
	}
	metadata, _ := foundation.Variable(1)
	metadata.Role, metadata.Bound, metadata.PairIndex, metadata.Group = CuddVariablePI, true, 3, CuddGroupHard
	if err = foundation.SetVariable(1, metadata); err != nil {
		t.Fatal(err)
	}
	if err = foundation.SetEpsilon(1e-9); err != nil {
		t.Fatal(err)
	}
	background, err := foundation.SetBackground(3.5)
	if err != nil {
		t.Fatal(err)
	}
	value, err := add.Evaluate(background, 0)
	if err != nil || value != 3.5 {
		t.Fatalf("background=%g err=%v", value, err)
	}
	snapshot, err := foundation.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Permutation[0] = 99
	again, _, _ := foundation.Permutation()
	if again[0] != 2 {
		t.Fatal("snapshot aliases manager permutation")
	}
	if err = foundation.Validate(); err != nil {
		t.Fatal(err)
	}
	statistics, err := foundation.Statistics()
	if err != nil {
		t.Fatal(err)
	}
	if statistics.Variables != 4 || statistics.BDDNodes < 6 || statistics.ADDNodes < 7 || statistics.BDDExternalRefs != 1 || statistics.ADDExternalRefs != 1 || statistics.CacheUsed != 1 || statistics.CacheHits != 1 || !strings.Contains(statistics.Format(), "CUDD stats:") {
		t.Fatalf("statistics=%+v", statistics)
	}
	if err = foundation.DerefBDD(conjunction); err != nil {
		t.Fatal(err)
	}
	if err = foundation.DerefADD(sum); err != nil {
		t.Fatal(err)
	}
	zero, err := foundation.CheckZeroReferences()
	if err != nil || zero != 0 {
		t.Fatalf("zero refs=%d err=%v", zero, err)
	}
	closed, err := foundation.Close()
	if err != nil || closed.Variables != 4 {
		t.Fatalf("close=%+v err=%v", closed, err)
	}
	if _, err = foundation.Statistics(); err == nil {
		t.Fatal("closed foundation accepted statistics")
	}
}

func TestCuddFoundationValidationAndFailures(t *testing.T) {
	if CuddVariablePI != 0 || CuddVariablePresentState != 1 || CuddVariableNextState != 2 || CuddVariableOrdinary != CuddVariablePI {
		t.Fatalf("variable role values PI=%d PS=%d NS=%d ordinary=%d", CuddVariablePI, CuddVariablePresentState, CuddVariableNextState, CuddVariableOrdinary)
	}
	if _, err := NewCuddFoundation(CuddFoundationOptions{Variables: 2, BDDCapacity: 2}); err == nil {
		t.Fatal("accepted undersized BDD")
	}
	if _, err := NewCuddFoundation(CuddFoundationOptions{Variables: 2, MaxBytes: 1}); err != ErrLimit {
		t.Fatalf("small bytes err=%v", err)
	}
	cancel := true
	cancelled, err := NewCuddFoundation(CuddFoundationOptions{Variables: 2, BDDCapacity: 32, ADDCapacity: 32, CacheSlots: 4, MaxBytes: 1 << 20, Stop: func() bool { return cancel }})
	if err != nil {
		t.Fatal(err)
	}
	before := append([]int(nil), cancelled.permutation...)
	if err = cancelled.SetPermutation([]int{1, 0}); err != ErrCancelled {
		t.Fatalf("cancel err=%v", err)
	}
	if !slices.Equal(before, cancelled.permutation) {
		t.Fatal("cancelled permutation mutated")
	}
	cancel = false
	if err = cancelled.SetPermutation([]int{1, 0}); err != nil {
		t.Fatal(err)
	}
	if err = cancelled.SetPermutation([]int{0, 0}); err != ErrDimension {
		t.Fatalf("duplicate permutation err=%v", err)
	}
	metadata, _ := cancelled.Variable(0)
	metadata.PairIndex = 0
	if err = cancelled.SetVariable(0, metadata); err != ErrDimension {
		t.Fatalf("self pair err=%v", err)
	}
	if err = cancelled.SetEpsilon(math.NaN()); err != ErrDimension {
		t.Fatalf("NaN epsilon err=%v", err)
	}
	if err = cancelled.RefBDD(ExtraBDDLit(999)); err != ErrInvalidLiteral {
		t.Fatalf("invalid BDD err=%v", err)
	}
	if err = cancelled.DerefADD(0); err != ErrInvariant {
		t.Fatalf("zero deref err=%v", err)
	}
	cancelled.addReferences[0] = math.MaxUint32
	if err = cancelled.RefADD(0); err != ErrLimit {
		t.Fatalf("saturated ref err=%v", err)
	}
	delete(cancelled.addReferences, 0)
	badKey := CuddCacheKey{Domain: CuddCacheDomain(9), Arity: 1}
	if err = cancelled.CacheInsert(badKey, 0); err != ErrDimension {
		t.Fatalf("bad cache err=%v", err)
	}
	nonCanonicalKey := CuddCacheKey{Domain: CuddCacheADD, Arity: 1, Operands: [3]uint32{0, 1}}
	if err = cancelled.CacheInsert(nonCanonicalKey, 0); err != ErrDimension {
		t.Fatalf("non-canonical cache key err=%v", err)
	}
	if err = cancelled.Validate(); err != nil {
		t.Fatal(err)
	}
	// Corrupt a unique-table head and prove the deep checker catches it.
	original := cancelled.bdd.uniqueHeads[0]
	cancelled.bdd.uniqueHeads[0] = uint32(len(cancelled.bdd.nodes) + 1)
	if err = cancelled.Validate(); err != ErrInvariant {
		t.Fatalf("corrupt unique err=%v", err)
	}
	cancelled.bdd.uniqueHeads[0] = original
	if err = cancelled.Validate(); err != nil {
		t.Fatal(err)
	}
	limited, err := NewCuddFoundation(CuddFoundationOptions{Variables: 1, BDDCapacity: 32, ADDCapacity: 32, CacheSlots: 4, MaxOperations: 1, MaxBytes: 1 << 20})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = limited.BDDVariable(0); err != nil {
		t.Fatal(err)
	}
	if _, err = limited.BDDVariable(0); err != ErrLimit {
		t.Fatalf("operation limit err=%v", err)
	}
	// Deep validation must reject surplus canonical-table entries as well as
	// missing entries.
	cancelled.add.terminals[math.Float64bits(17)] = 0
	if err = cancelled.Validate(); err != ErrInvariant {
		t.Fatalf("surplus ADD terminal err=%v", err)
	}
	delete(cancelled.add.terminals, math.Float64bits(17))
	if err = cancelled.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestCuddFoundationCacheCollisionAndFlush(t *testing.T) {
	foundation, err := NewCuddFoundation(CuddFoundationOptions{Variables: 2, BDDCapacity: 64, ADDCapacity: 64, CacheSlots: 1, MaxOperations: 1000, MaxBytes: 1 << 20})
	if err != nil {
		t.Fatal(err)
	}
	a, _ := foundation.BDDVariable(0)
	b, _ := foundation.BDDVariable(1)
	first := CuddCacheKey{Domain: CuddCacheBDD, Operation: 1, Arity: 1, Operands: [3]uint32{uint32(a)}}
	second := CuddCacheKey{Domain: CuddCacheBDD, Operation: 2, Arity: 1, Operands: [3]uint32{uint32(b)}}
	if err = foundation.CacheInsert(first, uint32(a)); err != nil {
		t.Fatal(err)
	}
	if err = foundation.CacheInsert(second, uint32(b)); err != nil {
		t.Fatal(err)
	}
	if _, hit, lookupErr := foundation.CacheLookup(first); lookupErr != nil || hit {
		t.Fatalf("evicted hit=%v err=%v", hit, lookupErr)
	}
	if result, hit, lookupErr := foundation.CacheLookup(second); lookupErr != nil || !hit || result != uint32(b) {
		t.Fatalf("second=%d hit=%v err=%v", result, hit, lookupErr)
	}
	if err = foundation.CacheFlush(); err != nil {
		t.Fatal(err)
	}
	statistics, _ := foundation.Statistics()
	if statistics.CacheUsed != 0 || statistics.CacheLookups != 2 || statistics.CacheHits != 1 {
		t.Fatalf("stats=%+v", statistics)
	}
}

func TestCuddFoundationBuildInventory(t *testing.T) {
	sources := CuddFoundationBuildSources()
	if len(sources) != 61 || sources[0] != "cuddAPI.c" || sources[len(sources)-1] != "cuddZddUtil.c" {
		t.Fatalf("inventory len=%d first=%q last=%q", len(sources), sources[0], sources[len(sources)-1])
	}
	seen := make(map[string]bool, len(sources))
	for _, source := range sources {
		if seen[source] || !strings.HasPrefix(source, "cudd") || !strings.HasSuffix(source, ".c") {
			t.Fatalf("bad source %q", source)
		}
		seen[source] = true
	}
	sources[0] = "changed"
	if CuddFoundationBuildSources()[0] != "cuddAPI.c" {
		t.Fatal("inventory aliases package storage")
	}
}

func TestCuddFoundationIndependentStateModel(t *testing.T) {
	foundation, err := NewCuddFoundation(CuddFoundationOptions{Variables: 3, BDDCapacity: 128, ADDCapacity: 128, CacheSlots: 16, MaxOperations: 100000, MaxBytes: 2 << 20})
	if err != nil {
		t.Fatal(err)
	}
	bdd := make([]ExtraBDDLit, 3)
	add := make([]CuddADDEdge, 3)
	for index := 0; index < 3; index++ {
		bdd[index], _ = foundation.BDDVariable(index)
		add[index], _ = foundation.ADDVariable(index)
	}
	modelBDD := map[uint32]uint32{}
	modelADD := map[uint32]uint32{}
	modelCache := make([]CuddCacheRecord, 16)
	state := uint64(0x9e3779b97f4a7c15)
	next := func() uint64 { state ^= state << 7; state ^= state >> 9; state ^= state << 8; return state }
	for step := 0; step < 5000; step++ {
		value := next()
		domain := CuddCacheDomain(value & 1)
		index := int(value>>1) % 3
		operand := uint32(add[index])
		model := modelADD
		if domain == CuddCacheBDD {
			operand, model = uint32(bdd[index]), modelBDD
		}
		switch (value >> 5) % 5 {
		case 0:
			if err = foundation.ref(domain, operand, true); err != nil {
				t.Fatal(err)
			}
			key := operand
			if domain == CuddCacheBDD {
				key = uint32(ExtraBDDLit(operand).ID())
			}
			model[key]++
		case 1:
			key := operand
			if domain == CuddCacheBDD {
				key = uint32(ExtraBDDLit(operand).ID())
			}
			if model[key] != 0 {
				if err = foundation.ref(domain, operand, false); err != nil {
					t.Fatal(err)
				}
				model[key]--
				if model[key] == 0 {
					delete(model, key)
				}
			}
		case 2:
			key := CuddCacheKey{Domain: domain, Operation: uint16(value >> 16), Arity: 1, Operands: [3]uint32{operand}}
			if err = foundation.CacheInsert(key, operand); err != nil {
				t.Fatal(err)
			}
			modelCache[int(cuddCacheHash(key))&15] = CuddCacheRecord{Key: key, Result: operand, Valid: true}
		case 3:
			key := CuddCacheKey{Domain: domain, Operation: uint16(value >> 16), Arity: 1, Operands: [3]uint32{operand}}
			result, hit, lookupErr := foundation.CacheLookup(key)
			if lookupErr != nil {
				t.Fatal(lookupErr)
			}
			record := modelCache[int(cuddCacheHash(key))&15]
			wantHit := record.Valid && record.Key == key
			if hit != wantHit || hit && result != record.Result {
				t.Fatalf("step=%d cache hit=%v/%v result=%d/%d", step, hit, wantHit, result, record.Result)
			}
		case 4:
			permutation := []int{0, 1, 2}
			if value&0x100 != 0 {
				permutation = []int{2, 0, 1}
			}
			if err = foundation.SetPermutation(permutation); err != nil {
				t.Fatal(err)
			}
		}
		if err = foundation.Validate(); err != nil {
			t.Fatalf("step=%d: %v", step, err)
		}
	}
	for node, want := range modelBDD {
		if got := foundation.bddReferences[node]; got != want {
			t.Fatalf("BDD ref %d=%d want %d", node, got, want)
		}
	}
	for node, want := range modelADD {
		if got := foundation.addReferences[node]; got != want {
			t.Fatalf("ADD ref %d=%d want %d", node, got, want)
		}
	}
}

func TestCuddFoundationIndependentManagersRace(t *testing.T) {
	var wait sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wait.Add(1)
		go func(worker int) {
			defer wait.Done()
			foundation, err := NewCuddFoundation(CuddFoundationOptions{Variables: 3, BDDCapacity: 128, ADDCapacity: 128, CacheSlots: 16, MaxOperations: 10000, MaxBytes: 2 << 20})
			if err != nil {
				t.Errorf("worker %d: %v", worker, err)
				return
			}
			for index := 0; index < 200; index++ {
				literal, _ := foundation.BDDVariable(index % 3)
				if err = foundation.RefBDD(literal); err != nil {
					t.Errorf("worker %d ref: %v", worker, err)
					return
				}
				if err = foundation.DerefBDD(literal); err != nil {
					t.Errorf("worker %d deref: %v", worker, err)
					return
				}
			}
			if err = foundation.Validate(); err != nil {
				t.Errorf("worker %d validate: %v", worker, err)
			}
		}(worker)
	}
	wait.Wait()
}

func FuzzCuddFoundationState(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
	f.Add([]byte{255, 0, 17, 33, 65, 129})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 192 {
			data = data[:192]
		}
		foundation, err := NewCuddFoundation(CuddFoundationOptions{Variables: 3, BDDCapacity: 256, ADDCapacity: 256, CacheSlots: 16, MaxOperations: 10000, MaxBytes: 4 << 20})
		if err != nil {
			t.Fatal(err)
		}
		bdd := make([]ExtraBDDLit, 3)
		add := make([]CuddADDEdge, 3)
		for index := 0; index < 3; index++ {
			bdd[index], _ = foundation.BDDVariable(index)
			add[index], _ = foundation.ADDVariable(index)
		}
		bddManager, _ := foundation.BDDManager()
		addManager, _ := foundation.ADDManager()
		for offset, value := range data {
			index := int(value) % 3
			switch value % 10 {
			case 0:
				if root, buildErr := bddManager.And(bdd[index], bdd[(index+1)%3], ExtraBDDOptions{MaxOperations: 1000, MaxBytes: 1 << 20}); buildErr == nil {
					bdd[index] = root
				}
			case 1:
				if root, buildErr := addManager.Apply(CuddADDPlus, add[index], add[(index+1)%3], ExtraBDDOptions{MaxOperations: 1000, MaxBytes: 1 << 20}); buildErr == nil {
					add[index] = root
				}
			case 2:
				_ = foundation.RefBDD(bdd[index])
			case 3:
				if count, _ := foundation.ReferenceCount(CuddCacheBDD, uint32(bdd[index])); count > 0 {
					_ = foundation.DerefBDD(bdd[index])
				}
			case 4:
				_ = foundation.RefADD(add[index])
			case 5:
				if count, _ := foundation.ReferenceCount(CuddCacheADD, uint32(add[index])); count > 0 {
					_ = foundation.DerefADD(add[index])
				}
			case 6:
				key := CuddCacheKey{Domain: CuddCacheBDD, Operation: uint16(offset), Arity: 1, Operands: [3]uint32{uint32(bdd[index])}}
				_ = foundation.CacheInsert(key, uint32(bdd[index]))
			case 7:
				_ = foundation.SetEpsilon(float64(value) / 255)
			case 8:
				metadata, _ := foundation.Variable(index)
				metadata.Bound = !metadata.Bound
				_ = foundation.SetVariable(index, metadata)
			case 9:
				if value&0x10 == 0 {
					_ = foundation.SetPermutation([]int{0, 1, 2})
				} else {
					_ = foundation.SetPermutation([]int{1, 2, 0})
				}
			}
			if err = foundation.Validate(); err != nil {
				t.Fatalf("offset=%d value=%d: %v", offset, value, err)
			}
		}
		statistics, err := foundation.Statistics()
		if err != nil || statistics.BDDNodes > 256 || statistics.ADDNodes > 256 {
			t.Fatalf("stats=%+v err=%v", statistics, err)
		}
	})
}
