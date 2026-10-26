package abc

// Go-only validation for the truth-table bi-decomposition port.

import (
	"errors"
	"math/bits"
	"slices"
	"testing"
)

func bdcCheckFunction(t testing.TB, manager *BdcManager, function, care *TruthTable) {
	t.Helper()
	for a := 0; a < 1<<uint(function.Variables()); a++ {
		if care != nil {
			c, e := care.Get(uint64(a))
			if e != nil {
				t.Fatal(e)
			}
			if !c {
				continue
			}
		}
		want, e := function.Get(uint64(a))
		if e != nil {
			t.Fatal(e)
		}
		got, e := manager.Evaluate(uint64(a))
		if e != nil {
			t.Fatal(e)
		}
		if got != want {
			t.Fatalf("assignment %d: got %v want %v", a, got, want)
		}
	}
}

func TestBdcManagerExhaustiveThreeVariables(t *testing.T) {
	manager, err := NewBdcManager(BdcParameters{MaxVariables: 6, MaxWords: 64, MaxNodes: 256})
	if err != nil {
		t.Fatal(err)
	}
	for word := 0; word < 256; word++ {
		function := bdcTestTruth(t, 3, uint64(word))
		n, err := manager.Decompose(function, nil, nil, BdcDecomposeOptions{MaxNewNodes: 64})
		if err != nil {
			t.Fatalf("%02x: %v", word, err)
		}
		if n < 0 || n > 64 {
			t.Fatalf("bad node count %d", n)
		}
		bdcCheckFunction(t, manager, function, nil)
		root, err := manager.Root()
		if err != nil || !root.Valid() {
			t.Fatalf("root %v %v", root, err)
		}
		truth, err := manager.NodeTruth(int(root.ID()))
		if err != nil {
			t.Fatal(err)
		}
		if root.Complemented() {
			truth = truth.Not()
		}
		if !truth.Equal(function) {
			t.Fatalf("truth mismatch for %02x", word)
		}
		if _, err = manager.Format(); err != nil {
			t.Fatal(err)
		}
	}
	if manager.Statistics().Calls != 256 {
		t.Fatalf("stats %+v", manager.Statistics())
	}
}

func TestBdcCareDivisorsAccessAndCopies(t *testing.T) {
	m, err := NewBdcManager(BdcParameters{MaxVariables: 6, MaxNodes: 128})
	if err != nil {
		t.Fatal(err)
	}
	f := bdcTestTruth(t, 3, 0x96)
	care := bdcTestTruth(t, 3, 0x77)
	div := bdcTestTruth(t, 3, 0x3c)
	if _, err = m.Decompose(f, care, []*TruthTable{div}, BdcDecomposeOptions{}); err != nil {
		t.Fatal(err)
	}
	bdcCheckFunction(t, m, f, care)
	for id := 0; id < m.NodeCount(); id++ {
		n, e := m.Node(id)
		if e != nil {
			t.Fatal(e)
		}
		if n.Type == BdcNodeAND {
			if int(n.Fanin0.ID()) >= id || int(n.Fanin1.ID()) >= id {
				t.Fatalf("non-topological node %+v", n)
			}
		}
	}
	root, _ := m.Root()
	if err = m.SetCopyInt(int(root.ID()), 42); err != nil {
		t.Fatal(err)
	}
	query := root.Not()
	wantCopy := 42
	if query.Complemented() {
		wantCopy ^= 1
	}
	got, err := m.ObjectCopy(query)
	if err != nil || got != wantCopy {
		t.Fatalf("copy %d %v", got, err)
	}
	if err = m.SetCopyToken(int(root.ID()), 0x12345678); err != nil {
		t.Fatal(err)
	}
	token, err := m.ObjectCopyToken(root.NotIf(true))
	if err != nil || token != 0x12345678 {
		t.Fatalf("copy token %x %v", token, err)
	}
	words := div.Words()
	words[0] = 0
	if div.Words()[0] == 0 {
		t.Fatal("divisor aliases caller")
	}
	nodeTruth, err := m.NodeTruth(0)
	if err != nil {
		t.Fatal(err)
	}
	_ = nodeTruth.Set(0, false)
	again, _ := m.NodeTruth(0)
	if !again.IsConstant(true) {
		t.Fatal("node truth aliases manager")
	}
}

func TestBdcAtomicErrorsLimitsCancellation(t *testing.T) {
	if m, e := NewBdcManager(BdcParameters{MaxVariables: 1}); m != nil || !errors.Is(e, ErrDimension) {
		t.Fatalf("constructor %v %v", m, e)
	}
	m, err := NewBdcManager(BdcParameters{MaxVariables: 6, MaxNodes: 128})
	if err != nil {
		t.Fatal(err)
	}
	f := bdcTestTruth(t, 3, 0x69)
	if _, err = m.Decompose(f, nil, nil, BdcDecomposeOptions{MaxNewNodes: 64}); err != nil {
		t.Fatal(err)
	}
	before, _ := m.Format()
	count := m.NodeCount()
	stats := m.Statistics()
	if _, err = m.Decompose(f, nil, nil, BdcDecomposeOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel %v", err)
	}
	if got, _ := m.Format(); got != before || m.NodeCount() != count || m.Statistics() != stats {
		t.Fatal("cancellation mutated manager")
	}
	hard := bdcTestTruth(t, 3, 0x96)
	if _, err = m.Decompose(hard, nil, nil, BdcDecomposeOptions{MaxNewNodes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("limit %v", err)
	}
	if got, _ := m.Format(); got != before || m.NodeCount() != count {
		t.Fatal("limit failure mutated manager")
	}
	wrong := bdcTestTruth(t, 2, 3)
	if _, err = m.Decompose(f, wrong, nil, BdcDecomposeOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("dimension %v", err)
	}
	if _, err = m.Node(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = m.Evaluate(8); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestBdcSourceChoicesZeroCareSupportAndCounts(t *testing.T) {
	m, err := NewBdcManager(BdcParameters{MaxVariables: 6, MaxNodes: 128})
	if err != nil {
		t.Fatal(err)
	}
	function := bdcTestTruth(t, 2, 0xe)
	zeroCare := bdcTestTruth(t, 2, 0)
	if nodes, err := m.Decompose(function, zeroCare, nil, BdcDecomposeOptions{}); err != nil || nodes != 0 {
		t.Fatalf("zero care nodes=%d err=%v", nodes, err)
	}
	root, _ := m.Root()
	if root != BdcLit(1) {
		t.Fatalf("zero care root %v", root)
	}
	if _, err = m.Decompose(function, nil, nil, BdcDecomposeOptions{}); err != nil {
		t.Fatal(err)
	}
	stats := m.Statistics()
	if stats.OrChoices+stats.AndChoices+stats.Muxes == 0 {
		t.Fatalf("missing choice stats %+v", stats)
	}
	variable := bdcTestTruth(t, 2, 0xa)
	divisor := bdcTestTruth(t, 2, 0xc)
	if _, err = m.Decompose(variable, nil, []*TruthTable{divisor}, BdcDecomposeOptions{}); err != nil {
		t.Fatal(err)
	}
	if m.AndCount() != 1 || m.GateCount() != 0 {
		t.Fatalf("source count=%d gates=%d", m.AndCount(), m.GateCount())
	}
	isf := bdcISFState{on: bdcTestTruth(t, 3, 0xaa), off: bdcTestTruth(t, 3, 0x55)}
	if err = bdcMinimizeISF(&isf); err != nil {
		t.Fatal(err)
	}
	if isf.support != 1 {
		t.Fatalf("support %03b", isf.support)
	}
}

func bdcTestBuild(t testing.TB, variables int) *bdcBuild {
	t.Helper()
	m, err := NewBdcManager(BdcParameters{MaxVariables: 6, MaxNodes: 256})
	if err != nil {
		t.Fatal(err)
	}
	table, err := NewBdcTable(BdcTableOptions{Variables: variables, MaxFunctions: 256})
	if err != nil {
		t.Fatal(err)
	}
	b := &bdcBuild{manager: m, variables: variables, maxNew: 128, table: table, andHash: make(map[[4]int]BdcLit)}
	one, err := NewConstantTruthTable(variables, true, 64)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = b.addBase(BdcNodeConst1, one); err != nil {
		t.Fatal(err)
	}
	for v := 0; v < variables; v++ {
		x, e := NewVariableTruthTable(variables, v, 64)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = b.addBase(BdcNodePI, x); e != nil {
			t.Fatal(e)
		}
	}
	return b
}

func TestBdcPinnedDecompositionDecisions(t *testing.T) {
	b := bdcTestBuild(t, 2)
	strong := bdcISFState{on: bdcTestTruth(t, 2, 0xe), off: bdcTestTruth(t, 2, 0x1), support: 3}
	cost, left, right, err := b.decomposeOR(strong)
	if err != nil {
		t.Fatal(err)
	}
	if cost != 3000 || left.unique != 1 || right.unique != 2 {
		t.Fatalf("strong OR cost=%d left=%b right=%b", cost, left.unique, right.unique)
	}
	b = bdcTestBuild(t, 3)
	weak := bdcISFState{on: bdcTestTruth(t, 3, 0x06), off: bdcTestTruth(t, 3, 0x29), support: 7}
	cost, left, right, err = b.decomposeOR(weak)
	if err != nil {
		t.Fatal(err)
	}
	if cost != 125 || left.unique != 4 || right.on != nil || right.off != nil {
		t.Fatalf("weak OR cost=%d left=%b right=%+v", cost, left.unique, right)
	}
	parent := bdcISFState{on: bdcTestTruth(t, 3, 0xaa), off: bdcTestTruth(t, 3, 0x55), support: 1}
	right = bdcISFState{}
	absorbed, err := b.updateRight(parent, bdcISFState{unique: 1}, &right, bdcMakeLit(1, false), bdcDecompositionOR)
	if err != nil || !absorbed {
		t.Fatalf("OR absorption=%v err=%v", absorbed, err)
	}
	x0, _ := NewVariableTruthTable(3, 0, 64)
	x1, _ := NewVariableTruthTable(3, 1, 64)
	x2, _ := NewVariableTruthTable(3, 2, 64)
	and01, _ := x0.And(x1, false, false)
	function, _ := and01.Or(x2)
	m, err := NewBdcManager(BdcParameters{MaxVariables: 6, MaxNodes: 256})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.Decompose(function, nil, []*TruthTable{and01}, BdcDecomposeOptions{}); err != nil {
		t.Fatal(err)
	}
	if m.Statistics().Reused == 0 || m.Statistics().OrChoices == 0 {
		t.Fatalf("reuse stats %+v", m.Statistics())
	}
	xor, _ := x0.Xor(x2, false)
	m2, err := NewBdcManager(BdcParameters{MaxVariables: 6, MaxNodes: 256})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m2.Decompose(xor, nil, nil, BdcDecomposeOptions{}); err != nil {
		t.Fatal(err)
	}
	firstGate := 1 + 3
	if firstGate >= len(m2.nodes) || int(m2.nodes[firstGate].Fanin0.ID()) != 1 && int(m2.nodes[firstGate].Fanin1.ID()) != 1 {
		t.Fatalf("MUX did not select first support variable: %+v", m2.nodes)
	}
}

func TestBdcDeterministicModelCampaign(t *testing.T) {
	var state uint64 = 0x9e3779b97f4a7c15
	for run := 0; run < 5000; run++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		variables := 2 + int(state%4)
		mask := uint64(1) << uint(1<<uint(variables))
		logical := state
		if variables < 6 {
			logical &= mask - 1
		}
		f := bdcTestTruth(t, variables, logical)
		m, e := NewBdcManager(BdcParameters{MaxVariables: 6, MaxNodes: 512})
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Decompose(f, nil, nil, BdcDecomposeOptions{MaxNewNodes: 256}); e != nil {
			t.Fatalf("run %d variables=%d logical=%x: %v", run, variables, logical, e)
		}
		bdcCheckFunction(t, m, f, nil)
	}
}

func TestBdcSPFDPrimitivesDatasetAndOverlap(t *testing.T) {
	if got := BdcSPFDAdjacencyCost(0); got != 0 {
		t.Fatalf("zero cost %d", got)
	}
	if got := BdcSPFDAdjacencyCost(truthVariablePattern(3)); got != 64 {
		t.Fatalf("variable cost %d", got)
	}
	for v := 0; v < 6; v++ {
		a, e := BdcSPFDCofactor(truthVariablePattern(v), v, false)
		if e != nil || a != 0 {
			t.Fatalf("cof0 %d %x %v", v, a, e)
		}
		b, e := BdcSPFDCofactor(truthVariablePattern(v), v, true)
		if e != nil || b != ^uint64(0) {
			t.Fatalf("cof1 %d %x %v", v, b, e)
		}
	}
	if BdcSPFDSeparationCount(0, 0) != 0 {
		t.Fatal("separation")
	}
	if _, e := BdcSPFDHash(1, 0); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	for _, truth := range []uint64{0, 1, 0x123456789abcdef0, ^uint64(0)} {
		primes := [8]uint32{12582917, 25165843, 50331653, 100663319, 201326611, 402653189, 805306457, 1610612741}
		var value uint32
		for i := 0; i < 8; i++ {
			value ^= primes[i] * uint32(byte(truth>>uint(8*i)))
		}
		got, e := BdcSPFDHash(truth, 12289)
		if e != nil || got != int(value%12289) {
			t.Fatalf("hash %x=%d want %d err=%v", truth, got, value%12289, e)
		}
	}
	selection, e := SelectBdcSPFDDivisor([]uint64{truthVariablePattern(0), truthVariablePattern(1)}, []int{1, 2}, ^truthVariablePattern(0), truthVariablePattern(0))
	if e != nil || selection.Index != 0 {
		t.Fatalf("selection %+v %v", selection, e)
	}
	truths := []uint64{0, 1, ^uint64(0), 0x123456789abcdef0}
	weights := []int32{-1, 0, 7, 19}
	data, e := MarshalBdcSPFDDataset(truths, weights, 1024)
	if e != nil {
		t.Fatal(e)
	}
	rt, rw, e := ParseBdcSPFDDataset(data, 10)
	if e != nil || !slices.Equal(rt, truths) || !slices.Equal(rw, weights) {
		t.Fatalf("roundtrip %v %v %v", rt, rw, e)
	}
	data[0] ^= 1
	if rt[0] != 0 {
		t.Fatal("parser aliases input")
	}
	entries := []BdcSPFDEntry{{Fanin0: -1, Fanin1: 0, Truth: truthVariablePattern(0)}, {Fanin0: -1, Fanin1: 1, Truth: truthVariablePattern(1)}, {Fanin0: 0, Fanin1: 1, Truth: truthVariablePattern(0) & truthVariablePattern(1)}, {Fanin0: 0, Fanin1: 1, XOR: true, Truth: truthVariablePattern(0) ^ truthVariablePattern(1)}, {Fanin0: 2, Fanin1: 3, Truth: (truthVariablePattern(0) & truthVariablePattern(1)) & (truthVariablePattern(0) ^ truthVariablePattern(1))}}
	if e = ValidateBdcSPFDEntries(entries); e != nil {
		t.Fatal(e)
	}
	over, e := BdcSPFDOverlap(entries, 2, 4)
	if e != nil || over != 1 {
		t.Fatalf("overlap %v %v", over, e)
	}
	not, e := BdcSPFDOverlap(entries, 2, 3)
	if e != nil || not != 0 {
		t.Fatalf("shared overlap %v %v", not, e)
	}
	form, e := FormatBdcSPFD(entries, 3)
	if e != nil || form != "(a+b)" {
		t.Fatalf("format %q %v", form, e)
	}
}

func TestBdcSPFDIndependentModels(t *testing.T) {
	var state uint64 = 0xd1b54a32d192ed03
	for i := 0; i < 5000; i++ {
		state ^= state << 13
		state ^= state >> 7
		state ^= state << 17
		truth := state
		want := 0
		for v := 0; v < 6; v++ {
			step := 1 << uint(v)
			for a := 0; a < 64; a++ {
				if a&step == 0 && bits.OnesCount64(((truth>>uint(a))^(truth>>uint(a+step)))&1) == 1 {
					want += 2
				}
			}
		}
		if got := BdcSPFDAdjacencyCost(truth); got != want {
			t.Fatalf("%x: %d != %d", truth, got, want)
		}
		entries := []BdcSPFDEntry{{Fanin0: -1, Fanin1: 0}, {Fanin0: -1, Fanin1: 1}, {Fanin0: -1, Fanin1: 2}}
		for n := 0; n < 6; n++ {
			limit := len(entries)
			a := int(state >> uint(n*5) % uint64(limit))
			c := int(state >> uint(n*7+3) % uint64(limit))
			entries = append(entries, BdcSPFDEntry{Fanin0: a, Fanin1: c})
		}
		first := 3 + int(state%3)
		second := 6 + int((state>>9)%3)
		var ancestors func(int, map[int]bool)
		ancestors = func(index int, set map[int]bool) {
			e := entries[index]
			if e.Fanin0 < 0 {
				return
			}
			if set[index] {
				return
			}
			set[index] = true
			ancestors(e.Fanin0, set)
			ancestors(e.Fanin1, set)
		}
		aSet, cSet := map[int]bool{}, map[int]bool{}
		ancestors(first, aSet)
		ancestors(second, cSet)
		expected := 0
		for id := range aSet {
			if cSet[id] {
				expected++
			}
		}
		got, err := BdcSPFDOverlap(entries, first, second)
		if err != nil || got != expected {
			t.Fatalf("overlap run=%d got=%d want=%d err=%v", i, got, expected, err)
		}
	}
}

func TestBdcSPFDBoundedSearch(t *testing.T) {
	a, b := truthVariablePattern(0), truthVariablePattern(1)
	for _, test := range []struct {
		truth uint64
		gates int
	}{
		{a, 0}, {^a, 0}, {a & b, 1}, {a ^ b, 1}, {^(a & b), 1},
	} {
		result, err := SearchBdcSPFD(test.truth, BdcSPFDSearchOptions{Variables: 2, Candidates: 32, MaxGates: 3, MaxGenerated: 10000})
		if err != nil || result.Root < 0 || result.GateCount != test.gates {
			t.Fatalf("%x: %+v %v", test.truth, result, err)
		}
		if err = ValidateBdcSPFDEntries(result.Entries); err != nil {
			t.Fatal(err)
		}
		got := result.Entries[result.Root].Truth
		if result.Complemented {
			got = ^got
		}
		if got != test.truth {
			t.Fatalf("got %x want %x", got, test.truth)
		}
	}
	if result, err := SearchBdcSPFD(a&b, BdcSPFDSearchOptions{Variables: 2, Candidates: 4, MaxGates: 2, Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) || result.Entries != nil {
		t.Fatalf("cancel %+v %v", result, err)
	}
	if result, err := SearchBdcSPFD((a&b)|truthVariablePattern(2), BdcSPFDSearchOptions{Variables: 3, Candidates: 16, MaxGates: 3, MaxGenerated: 1}); !errors.Is(err, ErrLimit) || result.Entries != nil {
		t.Fatalf("limit %+v %v", result, err)
	}
	levelOne, err := SearchBdcSPFD(0x0123456789abcdef, BdcSPFDSearchOptions{Variables: 3, Candidates: 1, MaxGates: 1, MaxGenerated: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(levelOne.Entries) != 18 || levelOne.Generated != 15 {
		t.Fatalf("level one was pruned: %+v", levelOne)
	}
	calls := 0
	if result, err := SearchBdcSPFD(0x0123456789abcdef, BdcSPFDSearchOptions{Variables: 6, Candidates: 200, MaxGates: 5, MaxGenerated: 1 << 20, Stop: func() bool { calls++; return calls >= 3 }}); !errors.Is(err, ErrCancelled) || result.Entries != nil || calls < 3 {
		t.Fatalf("inner cancel calls=%d result=%+v err=%v", calls, result, err)
	}
}

func FuzzBdcSPFDSearchAndDataset(f *testing.F) {
	f.Add(uint64(0x0123456789abcdef), uint8(3), uint8(12), uint8(2))
	f.Add(uint64(0x6996966996696996), uint8(4), uint8(20), uint8(3))
	f.Fuzz(func(t *testing.T, target uint64, rawVariables, rawCandidates, rawGates uint8) {
		variables := 2 + int(rawVariables%4)
		candidates := 1 + int(rawCandidates%24)
		gates := int(rawGates % 4)
		result, err := SearchBdcSPFD(target, BdcSPFDSearchOptions{Variables: variables, Candidates: candidates, MaxGates: gates, MaxGenerated: 50000})
		if err != nil && !errors.Is(err, ErrLimit) {
			t.Fatal(err)
		}
		if err != nil {
			return
		}
		if err = ValidateBdcSPFDEntries(result.Entries); err != nil {
			t.Fatal(err)
		}
		if result.Root >= 0 {
			got := result.Entries[result.Root].Truth
			if result.Complemented {
				got = ^got
			}
			if got != target {
				t.Fatalf("search got %x want %x", got, target)
			}
		}
		truths := make([]uint64, len(result.Entries))
		weights := make([]int32, len(result.Entries))
		for i, e := range result.Entries {
			truths[i] = e.Truth
			weights[i] = int32(i % 17)
		}
		data, err := MarshalBdcSPFDDataset(truths, weights, 1<<20)
		if err != nil {
			t.Fatal(err)
		}
		gotTruths, gotWeights, err := ParseBdcSPFDDataset(data, len(truths)+1)
		if err != nil || !slices.Equal(gotTruths, truths) || !slices.Equal(gotWeights, weights) {
			t.Fatalf("dataset mismatch %v", err)
		}
	})
}

func FuzzBdcDecomposition(f *testing.F) {
	f.Add(uint8(3), uint64(0x96), uint64(0xff))
	f.Add(uint8(2), uint64(6), uint64(15))
	f.Fuzz(func(t *testing.T, n uint8, logical, careBits uint64) {
		variables := 2 + int(n%4)
		width := 1 << uint(variables)
		if width < 64 {
			mask := uint64(1) << uint(width)
			logical &= mask - 1
			careBits &= mask - 1
		}
		fn := bdcTestTruth(t, variables, logical)
		care := bdcTestTruth(t, variables, careBits)
		m, e := NewBdcManager(BdcParameters{MaxVariables: 6, MaxNodes: 512})
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Decompose(fn, care, nil, BdcDecomposeOptions{MaxNewNodes: 256}); e != nil {
			t.Fatal(e)
		}
		bdcCheckFunction(t, m, fn, care)
	})
}
