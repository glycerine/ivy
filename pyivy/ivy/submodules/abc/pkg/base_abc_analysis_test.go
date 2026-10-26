package abc

// Tests for the Go port of src/base/abc/{abcCheck.c,abcDfs.c,abcRefs.c}.

import (
	"errors"
	"slices"
	"testing"
)

func mustABCAnalysisNetwork(t *testing.T) (*ABCNetwork, []int, []int, []int) {
	t.Helper()
	network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{MaxObjects: 10000, MaxEdges: 30000})
	if err != nil {
		t.Fatal(err)
	}
	pis := make([]int, 3)
	for i := range pis {
		pis[i], err = network.CreatePI()
		if err != nil {
			t.Fatal(err)
		}
		if err = network.SetObjectName(pis[i], string(rune('a'+i))); err != nil {
			t.Fatal(err)
		}
	}
	n1, err := network.CreateAND(ABCObjectEdge{ID: pis[0]}, ABCObjectEdge{ID: pis[1]})
	if err != nil {
		t.Fatal(err)
	}
	n2, err := network.CreateOR(ABCObjectEdge{ID: n1}, ABCObjectEdge{ID: pis[2]})
	if err != nil {
		t.Fatal(err)
	}
	po, err := network.CreatePO()
	if err != nil {
		t.Fatal(err)
	}
	if err = network.SetObjectName(po, "out"); err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(po, ABCObjectEdge{ID: n2}); err != nil {
		t.Fatal(err)
	}
	return network, pis, []int{n1, n2}, []int{po}
}

func independentABCPostorder(network *ABCNetwork, roots []int) []int {
	seen := make([]bool, len(network.objects))
	result := make([]int, 0)
	var visit func(int)
	visit = func(id int) {
		if seen[id] {
			return
		}
		seen[id] = true
		object := network.objects[id]
		if abcAnalysisBoundary(object.Type) {
			return
		}
		for _, edge := range object.Fanins {
			visit(edge.ID)
		}
		if abcAnalysisInternal(object.Type) {
			result = append(result, id)
		}
	}
	for _, root := range roots {
		object := network.objects[root]
		if abcIsCO(object.Type) {
			for _, edge := range object.Fanins {
				visit(edge.ID)
			}
		} else {
			visit(root)
		}
	}
	return result
}

func independentABCSupport(network *ABCNetwork, roots []int) []int {
	seen := make([]bool, len(network.objects))
	result := make([]int, 0)
	var visit func(int)
	visit = func(id int) {
		if seen[id] {
			return
		}
		seen[id] = true
		object := network.objects[id]
		if abcIsCI(object.Type) {
			result = append(result, id)
			return
		}
		for _, edge := range object.Fanins {
			visit(edge.ID)
		}
	}
	for _, root := range roots {
		object := network.objects[root]
		if abcIsCO(object.Type) {
			for _, edge := range object.Fanins {
				visit(edge.ID)
			}
		} else {
			visit(root)
		}
	}
	return result
}

func independentABCMFFC(network *ABCNetwork, root int, leaves []int) (members, boundary []int) {
	inCone := make([]bool, len(network.objects))
	blocked := make([]bool, len(network.objects))
	for _, id := range leaves {
		blocked[id] = true
	}
	inCone[root] = true
	for changed := true; changed; {
		changed = false
		for consumer, inside := range inCone {
			if !inside {
				continue
			}
			for _, edge := range network.objects[consumer].Fanins {
				candidate := edge.ID
				if inCone[candidate] || blocked[candidate] || abcAnalysisBoundary(network.objects[candidate].Type) {
					continue
				}
				allInternal := true
				for _, fanout := range network.objects[candidate].Fanouts {
					if !inCone[fanout] {
						allInternal = false
						break
					}
				}
				if allInternal {
					inCone[candidate] = true
					changed = true
				}
			}
		}
	}
	boundarySeen := make([]bool, len(network.objects))
	for id, inside := range inCone {
		if !inside {
			continue
		}
		members = append(members, id)
		for _, edge := range network.objects[id].Fanins {
			if !inCone[edge.ID] && !boundarySeen[edge.ID] {
				boundarySeen[edge.ID] = true
				boundary = append(boundary, edge.ID)
			}
		}
	}
	return members, boundary
}

func TestABCAnalysisTraversalSupportLevels(t *testing.T) {
	network, pis, nodes, pos := mustABCAnalysisNetwork(t)
	if err := CheckABCNetwork(network, ABCAnalysisOptions{}); err != nil {
		t.Fatal(err)
	}
	order, err := ABCForwardDFS(network, nil, false, ABCAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if want := independentABCPostorder(network, pos); !slices.Equal(order, want) || !slices.Equal(order, nodes) {
		t.Fatalf("postorder=%v want=%v", order, want)
	}
	support, err := ABCNetworkSupport(network, nil, ABCAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if want := independentABCSupport(network, pos); !slices.Equal(support, want) || !slices.Equal(support, pis) {
		t.Fatalf("support=%v want=%v", support, want)
	}
	levels, maximum, err := ABCNetworkLevels(network, ABCAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if levels[nodes[0]] != 1 || levels[nodes[1]] != 2 || levels[pos[0]] != 2 || maximum != 2 {
		t.Fatalf("levels=%v max=%d", levels, maximum)
	}
	reverse, err := ABCReverseDFS(network, nil, ABCAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(reverse) != 2 || reverse[len(reverse)-1] != nodes[0] {
		t.Fatalf("reverse=%v", reverse)
	}
	acyclic, cycle, err := ABCNetworkAcyclic(network, ABCAnalysisOptions{})
	if err != nil || !acyclic || len(cycle) != 0 {
		t.Fatalf("acyclic=%v cycle=%v err=%v", acyclic, cycle, err)
	}
}

func TestABCAnalysisMFFCAndLeaves(t *testing.T) {
	network, pis, nodes, _ := mustABCAnalysisNetwork(t)
	result, err := ABCNodeMFFC(network, nodes[1], nil, false, ABCAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Size != 2 || !slices.Equal(result.Cone, nodes) || !slices.Equal(result.Support, pis) {
		t.Fatalf("result=%+v", result)
	}
	limited, err := ABCNodeMFFC(network, nodes[1], []int{nodes[0]}, false, ABCAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if limited.Size != 1 || !slices.Equal(limited.Cone, nodes[1:]) || !slices.Equal(limited.Support, []int{nodes[0], pis[2]}) {
		t.Fatalf("limited=%+v", limited)
	}
	other, err := network.CreateXOR(ABCObjectEdge{ID: nodes[0]}, ABCObjectEdge{ID: pis[2]})
	if err != nil {
		t.Fatal(err)
	}
	shared, err := ABCNodeMFFC(network, nodes[1], nil, false, ABCAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if shared.Size != 1 || !slices.Contains(shared.Support, nodes[0]) {
		t.Fatalf("shared=%+v other=%d", shared, other)
	}
	phaseRoot, err := network.CreateAND(ABCObjectEdge{ID: nodes[0], Complemented: true}, ABCObjectEdge{ID: pis[2]})
	if err != nil {
		t.Fatal(err)
	}
	stopped, err := ABCNodeMFFC(network, phaseRoot, nil, true, ABCAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if stopped.Size != 1 || !slices.Contains(stopped.Support, nodes[0]) {
		t.Fatalf("stopped=%+v", stopped)
	}
}

func TestABCAnalysisCycleCancellationAndOwnership(t *testing.T) {
	network, _, nodes, _ := mustABCAnalysisNetwork(t)
	before := network.Clone()
	calls := 0
	_, err := ABCForwardDFS(network, nil, true, ABCAnalysisOptions{Stop: func() bool {
		calls++
		return calls > 1
	}})
	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel err=%v", err)
	}
	if err = network.Validate(); err != nil {
		t.Fatal(err)
	}
	for id := range network.objects {
		if !slices.Equal(network.objects[id].Fanins, before.objects[id].Fanins) || !slices.Equal(network.objects[id].Fanouts, before.objects[id].Fanouts) {
			t.Fatalf("analysis mutated object %d", id)
		}
	}
	if err = network.AddFanin(nodes[0], ABCObjectEdge{ID: nodes[1]}); err != nil {
		t.Fatal(err)
	}
	acyclic, cycle, err := ABCNetworkAcyclic(network, ABCAnalysisOptions{})
	if err != nil || acyclic || len(cycle) < 2 || cycle[0] != cycle[len(cycle)-1] {
		t.Fatalf("acyclic=%v cycle=%v err=%v", acyclic, cycle, err)
	}
	if err = CheckABCNetwork(network, ABCAnalysisOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("check err=%v", err)
	}
}

func TestABCAnalysisNamesAndCompare(t *testing.T) {
	first, _, _, _ := mustABCAnalysisNetwork(t)
	second := first.Clone()
	equal, err := ABCCompareSignals(first, second, false, false)
	if err != nil || !equal {
		t.Fatalf("equal=%v err=%v", equal, err)
	}
	if err = second.SetObjectName(second.pos[0], "different"); err != nil {
		t.Fatal(err)
	}
	equal, err = ABCCompareSignals(first, second, false, false)
	if err != nil || equal {
		t.Fatalf("equal=%v err=%v", equal, err)
	}
	names, err := ABCSortedSignalNames(first, append(first.PIs(), first.POs()...))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(names, []string{"a", "b", "c", "out"}) {
		t.Fatalf("names=%v", names)
	}

	// The source sorts terminal arrays by name before comparison. Preserve the
	// non-mutating equivalent even when creation order differs.
	reordered, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	var inputs []int
	for _, name := range []string{"c", "b", "a"} {
		id, createErr := reordered.CreatePI()
		if createErr != nil {
			t.Fatal(createErr)
		}
		if createErr = reordered.SetObjectName(id, name); createErr != nil {
			t.Fatal(createErr)
		}
		inputs = append(inputs, id)
	}
	node, err := reordered.CreateAND(ABCObjectEdge{ID: inputs[0]}, ABCObjectEdge{ID: inputs[1]})
	if err != nil {
		t.Fatal(err)
	}
	node, err = reordered.CreateOR(ABCObjectEdge{ID: node}, ABCObjectEdge{ID: inputs[2]})
	if err != nil {
		t.Fatal(err)
	}
	po, err := reordered.CreatePO()
	if err != nil {
		t.Fatal(err)
	}
	if err = reordered.SetObjectName(po, "out"); err != nil {
		t.Fatal(err)
	}
	if err = reordered.AddFanin(po, ABCObjectEdge{ID: node}); err != nil {
		t.Fatal(err)
	}
	if equal, err = ABCCompareSignals(first, reordered, false, false); err != nil || !equal {
		t.Fatalf("name-sorted equal=%v err=%v", equal, err)
	}
}

func abcAnalysisSequentialSignals(t *testing.T, init ABCLatchInit, outputName string) (*ABCNetwork, int) {
	t.Helper()
	network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	pi, err := network.CreatePI()
	if err != nil {
		t.Fatal(err)
	}
	if err = network.SetObjectName(pi, "in"); err != nil {
		t.Fatal(err)
	}
	bi, latch, bo, err := network.CreateLatchBox(init)
	if err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(bi, ABCObjectEdge{ID: pi}); err != nil {
		t.Fatal(err)
	}
	if err = network.SetObjectName(bo, outputName); err != nil {
		t.Fatal(err)
	}
	po, err := network.CreatePO()
	if err != nil {
		t.Fatal(err)
	}
	if err = network.SetObjectName(po, "out"); err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(po, ABCObjectEdge{ID: bo}); err != nil {
		t.Fatal(err)
	}
	return network, latch
}

func TestABCAnalysisCompareBoxesAndLatchShape(t *testing.T) {
	first, latch := abcAnalysisSequentialSignals(t, ABCInitZero, "state")
	second, _ := abcAnalysisSequentialSignals(t, ABCInitOne, "state")
	equal, err := ABCCompareSignals(first, second, false, true)
	if err != nil || !equal {
		t.Fatalf("box output match equal=%v err=%v", equal, err)
	}
	secondBO := second.objects[second.boxes[0]].Fanouts[0]
	if err = second.SetObjectName(secondBO, "state2"); err != nil {
		t.Fatal(err)
	}
	if equal, err = ABCCompareSignals(first, second, false, false); err != nil || !equal {
		t.Fatalf("non-combinational box skip equal=%v err=%v", equal, err)
	}
	if equal, err = ABCCompareSignals(first, second, false, true); err != nil || equal {
		t.Fatalf("combinational box mismatch equal=%v err=%v", equal, err)
	}

	first.objects[latch].Init = ABCInitNone
	if err = CheckABCNetwork(first, ABCAnalysisOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("unset latch init err=%v", err)
	}
	first.objects[latch].Init = ABCInitZero
	extraBO, err := first.CreateBO()
	if err != nil {
		t.Fatal(err)
	}
	if err = first.AddFanin(extraBO, ABCObjectEdge{ID: latch}); err != nil {
		t.Fatal(err)
	}
	if err = CheckABCNetwork(first, ABCAnalysisOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("multiple latch outputs err=%v", err)
	}
}

func TestABCAnalysisNetlistTraversalAndLevels(t *testing.T) {
	network, err := NewABCNetwork(ABCNetworkNetlist, ABCFunctionSOP, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	pi, _ := network.CreatePI()
	net0, _ := network.CreateNet()
	if err = network.AddFanin(net0, ABCObjectEdge{ID: pi}); err != nil {
		t.Fatal(err)
	}
	node1, _ := network.CreateNode()
	if err = network.SetNodeFunction(node1, "1 1\n"); err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(node1, ABCObjectEdge{ID: net0}); err != nil {
		t.Fatal(err)
	}
	net1, _ := network.CreateNet()
	if err = network.AddFanin(net1, ABCObjectEdge{ID: node1}); err != nil {
		t.Fatal(err)
	}
	node2, _ := network.CreateNode()
	if err = network.SetNodeFunction(node2, "1 1\n"); err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(node2, ABCObjectEdge{ID: net1}); err != nil {
		t.Fatal(err)
	}
	net2, _ := network.CreateNet()
	if err = network.AddFanin(net2, ABCObjectEdge{ID: node2}); err != nil {
		t.Fatal(err)
	}
	po, _ := network.CreatePO()
	if err = network.AddFanin(po, ABCObjectEdge{ID: net2}); err != nil {
		t.Fatal(err)
	}
	levels, maximum, err := ABCNetworkLevels(network, ABCAnalysisOptions{})
	if err != nil || levels[node1] != 1 || levels[node2] != 2 || levels[po] != 2 || maximum != 2 {
		t.Fatalf("levels=%v max=%d err=%v", levels, maximum, err)
	}
	reverse, err := ABCReverseDFS(network, nil, ABCAnalysisOptions{})
	if err != nil || !slices.Equal(reverse, []int{node2, node1}) {
		t.Fatalf("reverse=%v err=%v", reverse, err)
	}

	// Collect-all follows the source node iteration. A transparent net that was
	// allocated before its driver must not pull that driver ahead of earlier
	// dangling nodes.
	dangling, err := NewABCNetwork(ABCNetworkNetlist, ABCFunctionSOP, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	earlyNet, _ := dangling.CreateNet()
	middle, _ := dangling.CreateNode()
	driver, _ := dangling.CreateNode()
	if err = dangling.SetNodeFunction(middle, " 0\n"); err != nil {
		t.Fatal(err)
	}
	if err = dangling.SetNodeFunction(driver, " 1\n"); err != nil {
		t.Fatal(err)
	}
	if err = dangling.AddFanin(earlyNet, ABCObjectEdge{ID: driver}); err != nil {
		t.Fatal(err)
	}
	all, err := ABCForwardDFS(dangling, nil, true, ABCAnalysisOptions{})
	if err != nil || !slices.Equal(all, []int{middle, driver}) {
		t.Fatalf("collect-all=%v err=%v", all, err)
	}
}

func TestABCAnalysisLatchBreaksCombinationalCycle(t *testing.T) {
	network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	bi, _, bo, err := network.CreateLatchBox(ABCInitZero)
	if err != nil {
		t.Fatal(err)
	}
	node, err := network.CreateBuffer(ABCObjectEdge{ID: bo})
	if err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(bi, ABCObjectEdge{ID: node}); err != nil {
		t.Fatal(err)
	}
	acyclic, cycle, err := ABCNetworkAcyclic(network, ABCAnalysisOptions{})
	if err != nil || !acyclic || len(cycle) != 0 {
		t.Fatalf("acyclic=%v cycle=%v err=%v", acyclic, cycle, err)
	}
	if err = CheckABCNetwork(network, ABCAnalysisOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestABCAnalysisDeepIterative(t *testing.T) {
	network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{MaxObjects: 5000, MaxEdges: 10000})
	if err != nil {
		t.Fatal(err)
	}
	last, err := network.CreatePI()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3000; i++ {
		last, err = network.CreateBuffer(ABCObjectEdge{ID: last})
		if err != nil {
			t.Fatal(err)
		}
	}
	po, _ := network.CreatePO()
	if err = network.AddFanin(po, ABCObjectEdge{ID: last}); err != nil {
		t.Fatal(err)
	}
	order, err := ABCForwardDFS(network, nil, false, ABCAnalysisOptions{MaxVisited: 4000})
	if err != nil || len(order) != 3000 {
		t.Fatalf("len=%d err=%v", len(order), err)
	}
	if _, err = ABCForwardDFS(network, nil, false, ABCAnalysisOptions{MaxVisited: 100}); !errors.Is(err, ErrLimit) {
		t.Fatalf("limit err=%v", err)
	}
}

func abcAnalysisModelCase(t *testing.T, data []byte) {
	t.Helper()
	network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{MaxObjects: 256, MaxEdges: 1024})
	if err != nil {
		t.Fatal(err)
	}
	piCount := 1
	if len(data) != 0 {
		piCount += int(data[0] % 5)
	}
	ids := make([]int, 0, 64)
	for i := 0; i < piCount; i++ {
		id, createErr := network.CreatePI()
		if createErr != nil {
			t.Fatal(createErr)
		}
		ids = append(ids, id)
	}
	for i := 1; i+1 < len(data) && len(ids) < 60; i += 2 {
		a := ids[int(data[i])%len(ids)]
		b := ids[int(data[i+1])%len(ids)]
		id, createErr := network.CreateAND(ABCObjectEdge{ID: a, Complemented: data[i]&0x80 != 0}, ABCObjectEdge{ID: b, Complemented: data[i+1]&0x80 != 0})
		if createErr != nil {
			t.Fatal(createErr)
		}
		ids = append(ids, id)
	}
	po, err := network.CreatePO()
	if err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(po, ABCObjectEdge{ID: ids[len(ids)-1]}); err != nil {
		t.Fatal(err)
	}
	if err = CheckABCNetwork(network, ABCAnalysisOptions{}); err != nil {
		t.Fatal(err)
	}
	got, err := ABCForwardDFS(network, []int{po}, false, ABCAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if want := independentABCPostorder(network, []int{po}); !slices.Equal(got, want) {
		t.Fatalf("order=%v want=%v", got, want)
	}
	support, err := ABCNetworkSupport(network, []int{po}, ABCAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if want := independentABCSupport(network, []int{po}); !slices.Equal(support, want) {
		t.Fatalf("support=%v want=%v", support, want)
	}
	if len(got) != 0 {
		root := got[len(got)-1]
		result, mffcErr := ABCNodeMFFC(network, root, nil, false, ABCAnalysisOptions{})
		if mffcErr != nil {
			t.Fatal(mffcErr)
		}
		if result.Size == 0 || result.Cone[len(result.Cone)-1] != root {
			t.Fatalf("mffc=%+v", result)
		}
		wantCone, wantSupport := independentABCMFFC(network, root, nil)
		gotCone := slices.Clone(result.Cone)
		gotSupport := slices.Clone(result.Support)
		slices.Sort(gotCone)
		slices.Sort(gotSupport)
		slices.Sort(wantCone)
		slices.Sort(wantSupport)
		if !slices.Equal(gotCone, wantCone) || !slices.Equal(gotSupport, wantSupport) {
			t.Fatalf("mffc cone=%v/%v support=%v/%v", gotCone, wantCone, gotSupport, wantSupport)
		}
	}
}

func TestABCAnalysisIndependentModel(t *testing.T) {
	state := uint64(0x7e215d78a3c491bf)
	for run := 0; run < 5000; run++ {
		data := make([]byte, 20)
		for i := range data {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			data[i] = byte(state)
		}
		abcAnalysisModelCase(t, data)
	}
}

func FuzzABCAnalysisModel(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{3, 1, 2, 4, 8, 16, 32})
	f.Add([]byte{5, 255, 128, 64, 32, 16, 8, 4, 2, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 96 {
			data = data[:96]
		}
		abcAnalysisModelCase(t, data)
	})
}
