package abc

// Go-only validation for src/aig/hop and its cudd2 adapter.

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func requireHopTruth(t testing.TB, manager *HopManager, edge HopEdge, variables int, want uint64) {
	t.Helper()
	truth, err := manager.Truth(edge, variables, false)
	assignmentMask := uint64(1)<<(1<<variables) - 1
	if err != nil || len(truth) != 1 || truth[0]&assignmentMask != want {
		t.Fatalf("truth %x want %x: %v", truth, want, err)
	}
}

func TestHopEdgeManagerOperationsTruthAndHash(t *testing.T) {
	edge, err := NewHopEdge(7, true)
	if err != nil || edge.ID() != 7 || !edge.Complemented() || edge.Regular() != 14 || edge.Not() != 14 {
		t.Fatal(edge, err)
	}
	manager, err := NewHopManager(HopLimits{MaxObjects: 256, MaxVars: 8})
	if err != nil {
		t.Fatal(err)
	}
	if manager.Const1() != 0 || manager.Const0() != 1 {
		t.Fatal(manager.Const1(), manager.Const0())
	}
	x0, _ := manager.IthVar(0)
	x1, _ := manager.IthVar(1)
	and, err := manager.And(x0, x1)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := manager.And(x1, x0)
	if again != and {
		t.Fatalf("strashing failed: %d != %d", and, again)
	}
	if same, _ := manager.And(x0, x0); same != x0 {
		t.Fatal(same)
	}
	if zero, _ := manager.And(x0, x0.Not()); zero != manager.Const0() {
		t.Fatal(zero)
	}
	or, _ := manager.Or(x0, x1)
	xor, _ := manager.Xor(x0, x1)
	mux, _ := manager.Mux(x0, x1, x1.Not())
	maj, _ := manager.Majority(x0, x1, manager.Const1())
	requireHopTruth(t, manager, and, 2, 0x8)
	requireHopTruth(t, manager, or, 2, 0xe)
	requireHopTruth(t, manager, xor, 2, 0x6)
	requireHopTruth(t, manager, mux, 2, 0x9)
	requireHopTruth(t, manager, maj, 2, 0xe)
	if muxType, err := manager.IsMuxType(mux.Regular()); err != nil || !muxType {
		t.Fatal(muxType, err)
	}
	control, whenTrue, whenFalse, recognized, err := manager.RecognizeMux(mux.Regular())
	if err != nil || !recognized {
		t.Fatal(recognized, err)
	}
	rebuiltMux, err := manager.Mux(control, whenTrue, whenFalse)
	if err != nil {
		t.Fatal(err)
	}
	requireHopTruth(t, manager, rebuiltMux, 2, 0x6)
	if _, _, recognized, err = manager.RecognizeXor(xor.Regular()); err != nil || !recognized {
		t.Fatal(recognized, err)
	}
	if _, err = manager.IthVar(2); err != nil {
		t.Fatal(err)
	}
	if prefixTruth, truthErr := manager.Truth(and, 2, false); truthErr != nil || prefixTruth[0] != 0x8888888888888888 {
		t.Fatal(prefixTruth, truthErr)
	}
	if err = manager.Check(); err != nil {
		t.Fatal(err)
	}
	object, _ := manager.Object(and)
	if object.Type != HopObjectAnd || object.Fanin0.ID() > object.Fanin1.ID() || object.Level != 1 {
		t.Fatal(object)
	}
	profile, err := manager.HashProfile()
	if err != nil {
		t.Fatal(err)
	}
	stats, _ := manager.Stats()
	if len(profile) != 16 || stats.HashEntries != stats.Ands+stats.Xors {
		t.Fatal(profile, stats)
	}
}

func TestHopDFSBalanceComposeTransferCleanupAndFormatting(t *testing.T) {
	manager, _ := NewHopManager(HopLimits{MaxObjects: 512, MaxVars: 8})
	x0, _ := manager.IthVar(0)
	x1, _ := manager.IthVar(1)
	x2, _ := manager.IthVar(2)
	left, _ := manager.And(x0, x1)
	root, _ := manager.And(left, x2)
	nodes, err := manager.DFS(root)
	if err != nil || len(nodes) != 2 || nodes[len(nodes)-1] != root.Regular() {
		t.Fatal(nodes, err)
	}
	if size, _ := manager.DagSize(root); size != len(nodes) {
		t.Fatal(size)
	}
	if fanouts, _ := manager.FanoutCount(root, x1); fanouts != 1 {
		t.Fatal(fanouts)
	}
	balanced, err := manager.Balance(root)
	if err != nil {
		t.Fatal(err)
	}
	originalTruth, _ := manager.Truth(root, 3, false)
	balancedTruth, _ := manager.Truth(balanced, 3, false)
	if !slices.Equal(originalTruth, balancedTruth) {
		t.Fatal(originalTruth, balancedTruth)
	}
	leaves, err := manager.CollectMulti(root)
	if err != nil || len(leaves) != 3 {
		t.Fatal(leaves, err)
	}
	refs, err := manager.References()
	if err != nil || refs[x1.ID()] == 0 {
		t.Fatal(refs, err)
	}
	if truth6, err := manager.Truth6(root, 3); err != nil || truth6 != 0x8080808080808080 {
		t.Fatal(truth6, err)
	}
	if words, err := manager.TruthWords32(root, 3, false); err != nil || !slices.Equal(words, []uint32{0x80808080}) {
		t.Fatal(words, err)
	}
	composed, err := manager.Compose(root, x2.Not(), 0)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		got, _ := manager.Evaluate(composed, inputs)
		want := (!inputs[2]) && inputs[1] && inputs[2]
		if got != want {
			t.Fatalf("compose assignment %d: %v want %v", assignment, got, want)
		}
	}
	permuted, err := manager.Permute(root, []int{2, 1, 0})
	if err != nil {
		t.Fatal(err)
	}
	permutedTruth, _ := manager.Truth(permuted, 3, false)
	if !slices.Equal(originalTruth, permutedTruth) {
		t.Fatal(originalTruth, permutedTruth)
	}
	remapped, err := manager.Remap(root, 0b101, 3)
	if err != nil {
		t.Fatal(err)
	}
	requireHopTruth(t, manager, remapped, 3, 0)
	zeroSupport, err := manager.Remap(root.Not(), 0, 3)
	if err != nil || zeroSupport != manager.Const1() {
		t.Fatal(zeroSupport, err)
	}
	beforeStats, _ := manager.Stats()
	if _, err = manager.Permute(root, []int{0, 1, 7}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	afterStats, _ := manager.Stats()
	if beforeStats.Created != afterStats.Created {
		t.Fatal(beforeStats, afterStats)
	}
	destination, _ := NewHopManager(HopLimits{MaxObjects: 512, MaxVars: 8})
	transferred, err := manager.TransferTo(destination, root, 3)
	if err != nil {
		t.Fatal(err)
	}
	transferTruth, _ := destination.Truth(transferred, 3, false)
	if !slices.Equal(originalTruth, transferTruth) {
		t.Fatal(originalTruth, transferTruth)
	}
	// The source manager may have more PIs than the transferred root.  The C
	// routine intentionally maps only this prefix; this also guards against a
	// destination PI slice overrun.
	extra, _ := manager.IthVar(3)
	if extra == 0 {
		t.Fatal("missing extra source PI")
	}
	prefixDestination, _ := NewHopManager(HopLimits{MaxObjects: 512, MaxVars: 8})
	prefixTransfer, err := manager.TransferTo(prefixDestination, left, 2)
	if err != nil {
		t.Fatal(err)
	}
	prefixTruth, err := prefixDestination.Truth(prefixTransfer, 2, false)
	if err != nil || prefixTruth[0] != 0x8888888888888888 {
		t.Fatal(prefixTruth, err)
	}
	if equation, _ := manager.Equation(root); equation == "" {
		t.Fatal("empty equation")
	}
	if verilog, _ := manager.Verilog(root.Not()); verilog == "" || verilog[0] != '~' {
		t.Fatal(verilog)
	}
	if verbose, err := manager.Verbose(root); err != nil || verbose == "" {
		t.Fatal(verbose, err)
	}
	if summary, err := manager.ManagerVerbose(); err != nil || summary == "" {
		t.Fatal(summary, err)
	}
	if blif, err := manager.BLIF(root, "hop"); err != nil || len(blif) < 20 || blif[:10] != ".model hop" {
		t.Fatal(blif, err)
	}
	for _, constantCase := range []struct {
		root HopEdge
		cube string
	}{{manager.Const1(), "1 1\n"}, {manager.Const0(), "0 1\n"}} {
		blif, blifErr := manager.BLIF(constantCase.root, "constant")
		if blifErr != nil || !strings.Contains(blif, ".names const1\n1\n") || !strings.Contains(blif, ".names const1 z\n"+constantCase.cube) {
			t.Fatalf("constant BLIF %d:\n%s\n%v", constantCase.root, blif, blifErr)
		}
	}
	if err = manager.CreateChoice(root, balanced); err != nil {
		t.Fatal(err)
	}
	if representative, _ := manager.Representative(balanced.Not()); representative != root.Not() {
		t.Fatal(representative)
	}
	if _, err = manager.CreatePO(root); err != nil {
		t.Fatal(err)
	}
	balancedManager, err := manager.Balanced(true)
	if err != nil {
		t.Fatal(err)
	}
	if len(balancedManager.pos) != 1 {
		t.Fatal(len(balancedManager.pos))
	}
	balancedOutput, _ := balancedManager.Object(balancedManager.pos[0])
	rebuiltTruth, _ := balancedManager.Truth(balancedOutput.Fanin0, 3, false)
	if !slices.Equal(rebuiltTruth, originalTruth) {
		t.Fatal(rebuiltTruth, originalTruth)
	}
	if err = balancedManager.Check(); err != nil {
		t.Fatal(err)
	}
	snapshot, err := manager.MemorySnapshot()
	if err != nil || len(snapshot) == 0 {
		t.Fatal(len(snapshot), err)
	}
	snapshot[0].Alive = false
	constant, _ := manager.Object(manager.Const1())
	if !constant.Alive {
		t.Fatal("memory snapshot aliases manager")
	}
	dangling, _ := manager.And(x0, x2.Not())
	if dangling == root {
		t.Fatal("unexpected sharing")
	}
	if removed, err := manager.Cleanup(); err != nil || removed == 0 {
		t.Fatal(removed, err)
	}
}

func TestHopUtilitiesCudd2LimitsCancellationAndLifecycle(t *testing.T) {
	if count, _ := HopBitWordCount(33); count != 2 {
		t.Fatal(count)
	}
	if count, _ := HopTruthWordCount(7); count != 4 {
		t.Fatal(count)
	}
	if HopBase2Log(9) != 4 || HopBase10Log(101) != 3 {
		t.Fatal(HopBase2Log(9), HopBase10Log(101))
	}
	wide, _ := NewHopManager(HopLimits{MaxObjects: 64, MaxVars: 31})
	for i := 0; i < 31; i++ {
		if _, wideErr := wide.IthVar(i); wideErr != nil {
			t.Fatal(wideErr)
		}
	}
	if _, wideErr := wide.Truth(wide.Const1(), 31, false); !errors.Is(wideErr, ErrDimension) {
		t.Fatalf("oversized truth domain: %v", wideErr)
	}
	adapter, err := NewCudd2Adapter(2, 0, 17, 31, 1024, HopLimits{MaxObjects: 256})
	if err != nil {
		t.Fatal(err)
	}
	x0, _ := adapter.IthVar(0)
	x1, _ := adapter.IthVar(1)
	if err = adapter.Register(Cudd2Reference{Identity: 17, Complemented: true}, x0); err != nil {
		t.Fatal(err)
	}
	if resolved, resolveErr := adapter.Resolve(Cudd2Reference{Identity: 17}); resolveErr != nil || resolved != x0.Not() {
		t.Fatal(resolved, resolveErr)
	}
	if err = adapter.Register(Cudd2Reference{Identity: 17}, x1); err != nil {
		t.Fatal(err)
	}
	if resolved, _ := adapter.Resolve(Cudd2Reference{Identity: 17, Complemented: true}); resolved != x0 {
		t.Fatal(resolved)
	}
	xor, _ := adapter.Xor(x0, x1)
	xnor, _ := adapter.Xnor(x0, x1)
	if equal, err := adapter.Equal(xor.Not(), xnor); err != nil || !equal {
		t.Fatal(equal, err)
	}
	if leq, err := adapter.Leq(xor, x0); err != nil || leq {
		t.Fatal(leq, err)
	}
	manager, _ := adapter.Manager()
	stats, _ := manager.Stats()
	if stats.POs != 2 {
		t.Fatalf("Cudd2 predicates recorded %d outputs, want 2", stats.POs)
	}
	conjunction, err := manager.CreateConjunction(2)
	if err != nil {
		t.Fatal(err)
	}
	requireHopTruth(t, manager, conjunction, 2, 0x8)
	disjunction, _ := manager.CreateDisjunction(2)
	parity, _ := manager.CreateParity(2)
	miter, err := manager.Miter([]HopEdge{disjunction, disjunction, parity, parity.Not()})
	if err != nil {
		t.Fatal(err)
	}
	requireHopTruth(t, manager, miter, 2, 0xf)
	limited, _ := NewHopManager(HopLimits{MaxObjects: 3, MaxVars: 2})
	a, _ := limited.IthVar(0)
	b, _ := limited.IthVar(1)
	before, _ := limited.Stats()
	if _, err = limited.Xor(a, b); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	after, _ := limited.Stats()
	if before.Created != after.Created || before.HashEntries != after.HashEntries {
		t.Fatal(before, after)
	}
	cancelled := false
	cancellable, _ := NewHopManager(HopLimits{MaxObjects: 32, Stop: func() bool { return cancelled }})
	cancelled = true
	if _, err = cancellable.IthVar(0); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if err = adapter.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = adapter.One(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestHopTransformationsExhaustiveAndBalanceContradiction(t *testing.T) {
	manager, _ := NewHopManager(HopLimits{MaxObjects: 1024, MaxVars: 4})
	x := make([]HopEdge, 4)
	for i := range x {
		x[i], _ = manager.IthVar(i)
	}
	left, _ := manager.And(x[0], x[1].Not())
	right, _ := manager.And(x[2], x[3])
	root, _ := manager.Or(left, right)

	composed, err := manager.Compose(root, x[3].Not(), 1)
	if err != nil {
		t.Fatal(err)
	}
	permutation := []int{3, 0, 2, 1}
	permuted, err := manager.Permute(root, permutation)
	if err != nil {
		t.Fatal(err)
	}
	remapped, err := manager.Remap(root, 0b1101, 4)
	if err != nil {
		t.Fatal(err)
	}
	balanced, err := manager.Balance(root)
	if err != nil {
		t.Fatal(err)
	}
	destination, _ := NewHopManager(HopLimits{MaxObjects: 1024, MaxVars: 4})
	transferred, err := manager.TransferTo(destination, root, 4)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 16; assignment++ {
		inputs := make([]bool, 4)
		for i := range inputs {
			inputs[i] = assignment&(1<<i) != 0
		}
		want := (inputs[0] && !inputs[1]) || (inputs[2] && inputs[3])
		gotBalanced, _ := manager.Evaluate(balanced, inputs)
		gotTransfer, _ := destination.Evaluate(transferred, inputs)
		if gotBalanced != want || gotTransfer != want {
			t.Fatalf("assignment %d balance=%v transfer=%v want=%v", assignment, gotBalanced, gotTransfer, want)
		}
		gotCompose, _ := manager.Evaluate(composed, inputs)
		wantCompose := inputs[0] && inputs[3] || inputs[2] && inputs[3]
		if gotCompose != wantCompose {
			t.Fatalf("assignment %d compose=%v want=%v", assignment, gotCompose, wantCompose)
		}
		gotPermute, _ := manager.Evaluate(permuted, inputs)
		wantPermute := inputs[permutation[0]] && !inputs[permutation[1]] || inputs[permutation[2]] && inputs[permutation[3]]
		if gotPermute != wantPermute {
			t.Fatalf("assignment %d permute=%v want=%v", assignment, gotPermute, wantPermute)
		}
		gotRemap, _ := manager.Evaluate(remapped, inputs)
		// Selected old variables 0, 2, and 3 map to compact PIs 0, 1,
		// and 2; old variable 1 maps to constant zero.
		wantRemap := inputs[0] || inputs[1] && inputs[2]
		if gotRemap != wantRemap {
			t.Fatalf("assignment %d remap=%v want=%v", assignment, gotRemap, wantRemap)
		}
	}

	xy, _ := manager.And(x[0], x[1])
	nxz, _ := manager.And(x[0].Not(), x[2])
	contradiction, _ := manager.And(xy, nxz)
	balancedContradiction, err := manager.Balance(contradiction)
	if err != nil || balancedContradiction != manager.Const0() {
		t.Fatalf("balanced contradiction=%d: %v", balancedContradiction, err)
	}
	if _, err = manager.CreatePO(contradiction); err != nil {
		t.Fatal(err)
	}
	balancedManager, err := manager.Balanced(true)
	if err != nil {
		t.Fatal(err)
	}
	output, _ := balancedManager.Object(balancedManager.pos[0])
	if output.Fanin0 != balancedManager.Const0() {
		t.Fatalf("manager balance failed to collapse contradiction: %d", output.Fanin0)
	}
}

func runHopTruthModel(data []byte) error {
	manager, err := NewHopManager(HopLimits{MaxObjects: 4096, MaxVars: 4})
	if err != nil {
		return err
	}
	edges := make([]HopEdge, 0, 68)
	masks := make([]uint16, 0, 68)
	for variable, mask := range []uint16{0xaaaa, 0xcccc, 0xf0f0, 0xff00} {
		edge, edgeErr := manager.IthVar(variable)
		if edgeErr != nil {
			return edgeErr
		}
		edges = append(edges, edge)
		masks = append(masks, mask)
	}
	for offset := 0; offset+2 < len(data) && len(edges) < 68; offset += 3 {
		leftIndex := int(data[offset]) % len(edges)
		rightIndex := int(data[offset+1]) % len(edges)
		left, right := edges[leftIndex], edges[rightIndex]
		leftMask, rightMask := masks[leftIndex], masks[rightIndex]
		if data[offset+2]&0x10 != 0 {
			left = left.Not()
			leftMask = ^leftMask
		}
		if data[offset+2]&0x20 != 0 {
			right = right.Not()
			rightMask = ^rightMask
		}
		var result HopEdge
		var resultMask uint16
		switch data[offset+2] & 3 {
		case 0:
			result, err = manager.And(left, right)
			resultMask = leftMask & rightMask
		case 1:
			result, err = manager.Or(left, right)
			resultMask = leftMask | rightMask
		case 2:
			result, err = manager.Xor(left, right)
			resultMask = leftMask ^ rightMask
		case 3:
			control := edges[(leftIndex+rightIndex)%len(edges)]
			controlMask := masks[(leftIndex+rightIndex)%len(edges)]
			result, err = manager.Mux(control, left, right)
			resultMask = (controlMask & leftMask) | (^controlMask & rightMask)
		}
		if err != nil {
			return err
		}
		edges = append(edges, result)
		masks = append(masks, resultMask)
		truth, truthErr := manager.Truth(result, 4, false)
		if truthErr != nil || len(truth) != 1 || uint16(truth[0]) != resultMask {
			return ErrInvariant
		}
	}
	return manager.Check()
}

func TestHopIndependentTruthEquivalenceCampaign(t *testing.T) {
	state := uint64(0xd1b54a32d192ed03)
	data := make([]byte, 96)
	for execution := 0; execution < 5000; execution++ {
		for index := range data {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			data[index] = byte(state)
		}
		if err := runHopTruthModel(data); err != nil {
			t.Fatalf("execution %d: %v", execution, err)
		}
	}
}

func FuzzHopTruthEquivalence(f *testing.F) {
	f.Add([]byte{0, 1, 0, 1, 2, 1, 2, 3, 2})
	f.Add([]byte{3, 2, 0x31, 1, 0, 0x12})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 384 {
			return
		}
		if err := runHopTruthModel(data); err != nil {
			t.Fatal(err)
		}
	})
}
