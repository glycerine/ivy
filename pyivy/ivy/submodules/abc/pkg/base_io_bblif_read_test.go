package abc

// Standalone tests for Bbl_ManToAbc and Bbl_ManDfs in ioReadBblif.c.

import (
	"errors"
	"reflect"
	"slices"
	"sync"
	"testing"
)

func TestBblManagerNodeOrderAndABCReconstruction(t *testing.T) {
	source, ids := abcBblifFixture(t)
	manager, err := BblManagerFromABCNetwork(source, ABCBblifOptions{})
	if err != nil {
		t.Fatal(err)
	}
	danglingID := 20
	if err = manager.AddObject(BblObjectNode, danglingID, 1, "0 1\n"); err != nil {
		t.Fatal(err)
	}
	if err = manager.AddFanin(danglingID, ids[0]); err != nil {
		t.Fatal(err)
	}
	before, err := manager.Objects()
	if err != nil {
		t.Fatal(err)
	}
	order, err := BblManagerNodeOrder(manager, ABCBblifOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(order, []int{ids[2], danglingID}) {
		t.Fatalf("order=%v", order)
	}
	network, err := ABCNetworkFromBblManager(manager, ABCBblifOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if network.Name != "bbl" || len(network.pis) != 2 || len(network.pos) != 1 || network.objects[network.pis[0]].Name != "pi0" || network.objects[network.pis[1]].Name != "pi1" || network.objects[network.pos[0]].Name != "po0" {
		t.Fatalf("network=%+v", network)
	}
	if network.ObjectCountByType(ABCObjectNode) != 2 {
		t.Fatalf("node count=%d", network.ObjectCountByType(ABCObjectNode))
	}
	convertedNode := network.objects[3]
	if convertedNode.Function != "11 1\n" || !slices.Equal(convertedNode.Fanins, []ABCObjectEdge{{ID: network.pis[0]}, {ID: network.pis[1]}}) {
		t.Fatalf("node=%+v", convertedNode)
	}
	after, err := manager.Objects()
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("manager mutated: err=%v", err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		inputs := map[int]bool{ids[0]: assignment&1 != 0, ids[1]: assignment&2 != 0}
		want := evaluateBblifManager(t, manager, inputs)
		converted, convertErr := ABCNetworkToAIG(network, ABCFunctionalOptions{})
		if convertErr != nil {
			t.Fatal(convertErr)
		}
		got, evaluateErr := converted.Graph.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0})
		if evaluateErr != nil || len(got) != 1 || got[0] != want {
			t.Fatalf("assignment=%d got=%v want=%v err=%v", assignment, got, want, evaluateErr)
		}
	}
}

func TestBblManagerNodeOrderForwardReferencesAndCycles(t *testing.T) {
	manager, err := NewBblManager("forward", BblLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		kind       BblObjectType
		id, fanins int
		sop        string
	}{{BblObjectNode, 7, 1, "1 1\n"}, {BblObjectCI, 2, 0, ""}, {BblObjectCO, 9, 1, ""}} {
		if err = manager.AddObject(row.kind, row.id, row.fanins, row.sop); err != nil {
			t.Fatal(err)
		}
	}
	if err = manager.AddFanin(7, 2); err != nil {
		t.Fatal(err)
	}
	if err = manager.AddFanin(9, 7); err != nil {
		t.Fatal(err)
	}
	if order, orderErr := BblManagerNodeOrder(manager, ABCBblifOptions{}); orderErr != nil || !slices.Equal(order, []int{7}) {
		t.Fatalf("order=%v err=%v", order, orderErr)
	}
	if network, convertErr := ABCNetworkFromBblManager(manager, ABCBblifOptions{}); convertErr != nil || network.ObjectCount() != 3 {
		t.Fatalf("network=%v err=%v", network, convertErr)
	}

	cycle, _ := NewBblManager("cycle", BblLimits{})
	_ = cycle.AddObject(BblObjectNode, 1, 1, "1 1\n")
	_ = cycle.AddObject(BblObjectNode, 2, 1, "1 1\n")
	_ = cycle.AddFanin(1, 2)
	_ = cycle.AddFanin(2, 1)
	if order, orderErr := BblManagerNodeOrder(cycle, ABCBblifOptions{}); order != nil || !errors.Is(orderErr, ErrInvariant) {
		t.Fatalf("cycle order=%v err=%v", order, orderErr)
	}
	if network, convertErr := ABCNetworkFromBblManager(cycle, ABCBblifOptions{}); network != nil || !errors.Is(convertErr, ErrInvariant) {
		t.Fatalf("cycle network=%v err=%v", network, convertErr)
	}
}

func TestBblManagerToABCLimitsUnsupportedAndConcurrency(t *testing.T) {
	source, _ := abcBblifFixture(t)
	manager, err := BblManagerFromABCNetwork(source, ABCBblifOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, options := range []ABCBblifOptions{
		{MaxObjects: 3},
		{MaxObjectID: 2},
		{MaxFanins: 1},
		{MaxWork: 1},
		{MaxBytes: 1},
		{Stop: func() bool { return true }},
		{MaxWork: -1},
	} {
		if network, convertErr := ABCNetworkFromBblManager(manager, options); network != nil || convertErr == nil {
			t.Fatalf("options=%+v network=%v err=%v", options, network, convertErr)
		}
	}
	boxed, _ := NewBblManager("box", BblLimits{})
	_ = boxed.AddObject(BblObjectCI, 1, 0, "")
	boxed.objects[0].box = true
	if network, convertErr := ABCNetworkFromBblManager(boxed, ABCBblifOptions{}); network != nil || !errors.Is(convertErr, ErrUnsupported) {
		t.Fatalf("boxed network=%v err=%v", network, convertErr)
	}

	var wait sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for iteration := 0; iteration < 20; iteration++ {
				network, convertErr := ABCNetworkFromBblManager(manager, ABCBblifOptions{})
				if convertErr != nil || network.ObjectCount() != 4 {
					t.Errorf("network=%v err=%v", network, convertErr)
					return
				}
			}
		}()
	}
	wait.Wait()
}

func TestBblManagerToAIGReadVerifyAndRoundTrip(t *testing.T) {
	source, ids := abcBblifFixture(t)
	manager, err := BblManagerFromABCNetwork(source, ABCBblifOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err = manager.AddObject(BblObjectNode, 20, 1, "0 1\n"); err != nil {
		t.Fatal(err)
	}
	if err = manager.AddFanin(20, ids[0]); err != nil {
		t.Fatal(err)
	}
	before, _ := manager.Objects()
	aig, err := AigManagerFromBblManager(manager, ABCBblifOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if aig.Name != "bbl" || aig.InputCount() != 2 || aig.OutputCount() != 1 || aig.AndCount() != 1 {
		t.Fatalf("aig inputs=%d outputs=%d ands=%d name=%q", aig.InputCount(), aig.OutputCount(), aig.AndCount(), aig.Name)
	}
	for assignment := 0; assignment < 4; assignment++ {
		got, evaluateErr := aig.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0})
		if evaluateErr != nil || len(got) != 1 || got[0] != (assignment == 3) {
			t.Fatalf("assignment=%d got=%v err=%v", assignment, got, evaluateErr)
		}
	}
	after, _ := manager.Objects()
	if !reflect.DeepEqual(before, after) {
		t.Fatal("AIG conversion mutated BBLIF manager")
	}
	data, err := MarshalABCBinaryBLIF(source, ABCBblifOptions{})
	if err != nil {
		t.Fatal(err)
	}
	read, err := ReadABCBinaryBLIF(data, true, ABCBblifOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if read.Type != ABCNetworkStrash || read.Function != ABCFunctionAIG || read.Name != "bbl" || read.objects[read.pis[0]].Name != "pi0" || read.objects[read.pos[0]].Name != "po0" {
		t.Fatalf("read network=%+v", read)
	}
	verification, err := VerifyBblifABCNetworks(source, read, ABCBblifOptions{})
	if err != nil || verification.Status != CECEquivalent {
		t.Fatalf("verification=%+v err=%v", verification, err)
	}
	round, verification, err := RoundTripABCBinaryBLIF(source, ABCBblifOptions{})
	if err != nil || round == nil || verification.Status != CECEquivalent {
		t.Fatalf("round=%v verification=%+v err=%v", round, verification, err)
	}
	different := source.Clone()
	if err = different.SetNodeFunction(ids[2], "00 0\n"); err != nil {
		t.Fatal(err)
	}
	verification, err = VerifyBblifABCNetworks(source, different, ABCBblifOptions{})
	if err != nil || verification.Status != CECDifferent || verification.Counter == nil {
		t.Fatalf("different verification=%+v err=%v", verification, err)
	}
}

func TestBblManagerToAIGAndReaderLimits(t *testing.T) {
	source, _ := abcBblifFixture(t)
	manager, err := BblManagerFromABCNetwork(source, ABCBblifOptions{})
	if err != nil {
		t.Fatal(err)
	}
	data, err := MarshalABCBinaryBLIF(source, ABCBblifOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, options := range []ABCBblifOptions{{MaxObjects: 3}, {MaxObjectID: 2}, {MaxFanins: 1}, {MaxWork: 1}, {MaxBytes: 1}, {Stop: func() bool { return true }}} {
		if graph, convertErr := AigManagerFromBblManager(manager, options); graph != nil || convertErr == nil {
			t.Fatalf("AIG options=%+v graph=%v err=%v", options, graph, convertErr)
		}
	}
	for _, options := range []ABCBblifOptions{{MaxObjects: 3}, {MaxObjectID: 2}, {MaxFanins: 1}, {MaxSOPBytes: 4}, {MaxBinaryBytes: len(data) - 1}, {MaxWork: 1}, {MaxBytes: 1}, {Stop: func() bool { return true }}} {
		if network, readErr := ReadABCBinaryBLIF(data, true, options); network != nil || readErr == nil {
			t.Fatalf("read options=%+v network=%v err=%v", options, network, readErr)
		}
	}
	corrupt := slices.Clone(data)
	corrupt[len(corrupt)-1] ^= 0xff
	if network, readErr := ReadABCBinaryBLIF(corrupt, true, ABCBblifOptions{}); network != nil || readErr == nil {
		t.Fatalf("corrupt network=%v err=%v", network, readErr)
	}
}

func FuzzBblManagerToABC(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3})
	f.Add([]byte{9, 7, 5})
	f.Fuzz(func(t *testing.T, operations []byte) {
		if len(operations) > 12 {
			t.Skip()
		}
		manager, err := NewBblManager("fuzz", BblLimits{})
		if err != nil {
			t.Fatal(err)
		}
		available := []int{10, 20, 30}
		for _, id := range available {
			if err = manager.AddObject(BblObjectCI, id, 0, ""); err != nil {
				t.Fatal(err)
			}
		}
		nextID := 40
		for _, operation := range operations {
			first := available[int(operation)%len(available)]
			second := available[int(operation/7)%len(available)]
			sop := "11 1\n"
			if operation&1 != 0 {
				sop = "10 1\n01 1\n"
			}
			if err = manager.AddObject(BblObjectNode, nextID, 2, sop); err != nil {
				t.Fatal(err)
			}
			if err = manager.AddFanin(nextID, first); err != nil {
				t.Fatal(err)
			}
			if err = manager.AddFanin(nextID, second); err != nil {
				t.Fatal(err)
			}
			available = append(available, nextID)
			nextID++
		}
		if err = manager.AddObject(BblObjectCO, nextID, 1, ""); err != nil {
			t.Fatal(err)
		}
		if err = manager.AddFanin(nextID, available[len(available)-1]); err != nil {
			t.Fatal(err)
		}
		network, err := ABCNetworkFromBblManager(manager, ABCBblifOptions{})
		if err != nil {
			t.Fatal(err)
		}
		converted, err := ABCNetworkToAIG(network, ABCFunctionalOptions{MaxOperations: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			inputMap := map[int]bool{10: assignment&1 != 0, 20: assignment&2 != 0, 30: assignment&4 != 0}
			want := evaluateBblifManager(t, manager, inputMap)
			got, evaluateErr := converted.Graph.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0})
			if evaluateErr != nil || len(got) != 1 || got[0] != want {
				t.Fatalf("assignment=%d got=%v want=%v err=%v", assignment, got, want, evaluateErr)
			}
		}
	})
}

func FuzzABCBinaryBLIFRoundTrip(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4})
	f.Add([]byte{9, 0, 7})
	f.Fuzz(func(t *testing.T, operations []byte) {
		if len(operations) > 8 {
			t.Skip()
		}
		network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
		if err != nil {
			t.Fatal(err)
		}
		network.Name = "round"
		available := make([]int, 3, 11)
		for index := range available {
			available[index], err = network.CreatePI()
			if err != nil {
				t.Fatal(err)
			}
		}
		for _, operation := range operations {
			first := available[int(operation)%len(available)]
			second := available[int(operation/5)%len(available)]
			sop := "11 1\n"
			if operation&1 != 0 {
				sop = "10 1\n01 1\n"
			}
			node, createErr := network.CreateLogicNode(sop, ABCObjectEdge{ID: first}, ABCObjectEdge{ID: second})
			if createErr != nil {
				t.Fatal(createErr)
			}
			available = append(available, node)
		}
		output, err := network.CreatePO()
		if err != nil {
			t.Fatal(err)
		}
		if err = network.AddFanin(output, ABCObjectEdge{ID: available[len(available)-1]}); err != nil {
			t.Fatal(err)
		}
		data, err := MarshalABCBinaryBLIF(network, ABCBblifOptions{})
		if err != nil {
			t.Fatal(err)
		}
		read, err := ReadABCBinaryBLIF(data, true, ABCBblifOptions{})
		if err != nil {
			t.Fatal(err)
		}
		left, err := ABCNetworkToAIG(network, ABCFunctionalOptions{})
		if err != nil {
			t.Fatal(err)
		}
		right, err := ABCNetworkToAIG(read, ABCFunctionalOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			want, leftErr := left.Graph.Evaluate(inputs)
			got, rightErr := right.Graph.Evaluate(inputs)
			if leftErr != nil || rightErr != nil || !slices.Equal(got, want) {
				t.Fatalf("assignment=%d got=%v want=%v left=%v right=%v", assignment, got, want, leftErr, rightErr)
			}
		}
	})
}
