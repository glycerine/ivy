package abc

// Standalone tests for src/base/io/ioWriteBblif.c.

import (
	"errors"
	"reflect"
	"slices"
	"sync"
	"testing"
)

func abcBblifFixture(t testing.TB) (*ABCNetwork, []int) {
	t.Helper()
	network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	network.Name = "bbl"
	first, err := network.CreatePI()
	if err != nil {
		t.Fatal(err)
	}
	second, err := network.CreatePI()
	if err != nil {
		t.Fatal(err)
	}
	node, err := network.CreateLogicNode("11 1\n", ABCObjectEdge{ID: first}, ABCObjectEdge{ID: second})
	if err != nil {
		t.Fatal(err)
	}
	dangling, err := network.CreateLogicNode("1 1\n", ABCObjectEdge{ID: first})
	if err != nil {
		t.Fatal(err)
	}
	output, err := network.CreatePO()
	if err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(output, ABCObjectEdge{ID: node}); err != nil {
		t.Fatal(err)
	}
	return network, []int{first, second, node, dangling, output}
}

func TestABCBblifManagerOrderIDsFunctionsAndOwnership(t *testing.T) {
	network, ids := abcBblifFixture(t)
	source := network.Clone()
	manager, err := BblManagerFromABCNetwork(network, ABCBblifOptions{})
	if err != nil {
		t.Fatal(err)
	}
	objects, err := manager.Objects()
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) != 4 {
		t.Fatalf("objects=%v", objects)
	}
	for index, want := range []int{ids[0], ids[1], ids[2], ids[4]} {
		if objects[index].ID != want || objects[index].OriginalID != want {
			t.Fatalf("object %d=%+v", index, objects[index])
		}
	}
	if objects[0].Type != BblObjectCI || objects[1].Type != BblObjectCI || objects[2].Type != BblObjectNode || objects[3].Type != BblObjectCO {
		t.Fatalf("types=%v", objects)
	}
	if objects[2].SOP != "11 1\n" || !slices.Equal(objects[2].Fanins, ids[:2]) || !slices.Equal(objects[3].Fanins, []int{ids[2]}) {
		t.Fatalf("objects=%+v", objects)
	}
	if !reflect.DeepEqual(network, source) {
		t.Fatal("conversion mutated source")
	}
	if err = network.SetNodeFunction(ids[2], "00 0\n"); err != nil {
		t.Fatal(err)
	}
	again, found, err := manager.ObjectByID(ids[2])
	if err != nil || !found || again.SOP != "11 1\n" {
		t.Fatalf("owned object=%+v found=%v err=%v", again, found, err)
	}
}

func TestABCBblifMarshalMatchesDirectManager(t *testing.T) {
	network, ids := abcBblifFixture(t)
	got, err := MarshalABCBinaryBLIF(network, ABCBblifOptions{})
	if err != nil {
		t.Fatal(err)
	}
	direct, err := NewBblManager("bbl", BblLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		kind       BblObjectType
		id, fanins int
		sop        string
	}{{BblObjectCI, ids[0], 0, ""}, {BblObjectCI, ids[1], 0, ""}, {BblObjectNode, ids[2], 2, "11 1\n"}, {BblObjectCO, ids[4], 1, ""}} {
		if err = direct.AddObject(row.kind, row.id, row.fanins, row.sop); err != nil {
			t.Fatal(err)
		}
	}
	for _, edge := range [][2]int{{ids[2], ids[0]}, {ids[2], ids[1]}, {ids[4], ids[2]}} {
		if err = direct.AddFanin(edge[0], edge[1]); err != nil {
			t.Fatal(err)
		}
	}
	want, err := direct.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("binary mismatch\ngot  %x\nwant %x", got, want)
	}
	parsed, err := ParseBblBinary(got, BblLimits{})
	if err != nil {
		t.Fatal(err)
	}
	parsedObjects, err := parsed.Objects()
	if err != nil || len(parsedObjects) != 4 || parsedObjects[2].OriginalID != ids[2] {
		t.Fatalf("parsed=%+v err=%v", parsedObjects, err)
	}
}

func TestABCBblifValidationLimitsAndAtomicErrors(t *testing.T) {
	network, ids := abcBblifFixture(t)
	for _, options := range []ABCBblifOptions{
		{MaxObjects: 3},
		{MaxObjectID: ids[4] - 1},
		{MaxFanins: 1},
		{MaxSOPBytes: 4},
		{MaxBinaryBytes: 1},
		{MaxWork: 1},
		{MaxBytes: 1},
		{Stop: func() bool { return true }},
		{MaxObjects: -1},
	} {
		if got, err := BblManagerFromABCNetwork(network, options); got != nil || err == nil {
			t.Fatalf("options=%+v got=%v err=%v", options, got, err)
		}
	}
	wrong, err := NewABCNetwork(ABCNetworkStrash, ABCFunctionAIG, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if got, convertErr := BblManagerFromABCNetwork(wrong, ABCBblifOptions{}); got != nil || !errors.Is(convertErr, ErrUnsupported) {
		t.Fatalf("wrong network got=%v err=%v", got, convertErr)
	}
	complemented := network.Clone()
	complemented.objects[ids[2]].Fanins[0].Complemented = true
	if got, convertErr := BblManagerFromABCNetwork(complemented, ABCBblifOptions{}); got != nil || !errors.Is(convertErr, ErrUnsupported) {
		t.Fatalf("complemented got=%v err=%v", got, convertErr)
	}
	special := network.Clone()
	special.objects[ids[2]].Function = "11 x\n"
	specialManager, convertErr := BblManagerFromABCNetwork(special, ABCBblifOptions{})
	if convertErr != nil {
		t.Fatal(convertErr)
	}
	specialNode, found, convertErr := specialManager.ObjectByID(ids[2])
	if convertErr != nil || !found || specialNode.SOP != "11 x\n" {
		t.Fatalf("special=%+v found=%v err=%v", specialNode, found, convertErr)
	}
	specialBytes, convertErr := MarshalABCBinaryBLIF(special, ABCBblifOptions{})
	if convertErr != nil {
		t.Fatal(convertErr)
	}
	specialRoundTrip, convertErr := ParseBblBinary(specialBytes, BblLimits{})
	if convertErr != nil {
		t.Fatal(convertErr)
	}
	specialObjects, convertErr := specialRoundTrip.Objects()
	if convertErr != nil || len(specialObjects) != 4 || specialObjects[2].OriginalID != ids[2] || specialObjects[2].SOP != "11 x\n" {
		t.Fatalf("special round trip=%+v err=%v", specialObjects, convertErr)
	}
	malformed := network.Clone()
	malformed.objects[ids[2]].Function = "1 1\n"
	if got, convertErr := BblManagerFromABCNetwork(malformed, ABCBblifOptions{}); got != nil || convertErr == nil {
		t.Fatalf("malformed got=%v err=%v", got, convertErr)
	}
}

func TestABCBblifConcurrentFrozenSource(t *testing.T) {
	network, _ := abcBblifFixture(t)
	want, err := MarshalABCBinaryBLIF(network, ABCBblifOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for iteration := 0; iteration < 20; iteration++ {
				got, marshalErr := MarshalABCBinaryBLIF(network, ABCBblifOptions{})
				if marshalErr != nil || !slices.Equal(got, want) {
					t.Errorf("got=%x err=%v", got, marshalErr)
					return
				}
			}
		}()
	}
	wait.Wait()
}

func evaluateABCBblifNetwork(t testing.TB, network *ABCNetwork, inputs map[int]bool) bool {
	t.Helper()
	values := make(map[int]bool, len(network.objects))
	for id, value := range inputs {
		values[id] = value
	}
	for id, object := range network.objects {
		if !object.Alive || object.Type != ABCObjectNode {
			continue
		}
		sop, err := ParseABCSOP(object.Function, ABCMetadataOptions{})
		if err != nil {
			t.Fatal(err)
		}
		var assignment uint64
		for index, edge := range object.Fanins {
			if values[edge.ID] {
				assignment |= uint64(1) << uint(index)
			}
		}
		values[id], err = sop.Evaluate(assignment)
		if err != nil {
			t.Fatal(err)
		}
	}
	return values[network.objects[network.cos[0]].Fanins[0].ID]
}

func evaluateBblifManager(t testing.TB, manager *BblManager, inputs map[int]bool) bool {
	t.Helper()
	values := make([]bool, len(manager.objects))
	result := false
	for index, object := range manager.objects {
		switch object.kind {
		case BblObjectCI:
			values[index] = inputs[object.id]
		case BblObjectNode:
			sop, err := ParseABCSOP(manager.functions[object.function].sop, ABCMetadataOptions{})
			if err != nil {
				t.Fatal(err)
			}
			var assignment uint64
			for faninIndex, fanin := range object.fanins {
				if values[fanin] {
					assignment |= uint64(1) << uint(faninIndex)
				}
			}
			values[index], err = sop.Evaluate(assignment)
			if err != nil {
				t.Fatal(err)
			}
		case BblObjectCO:
			result = values[object.fanins[0]]
		}
	}
	return result
}

func FuzzABCBblifConversion(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3})
	f.Add([]byte{7, 4, 9})
	f.Fuzz(func(t *testing.T, operations []byte) {
		if len(operations) > 12 {
			t.Skip()
		}
		network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
		if err != nil {
			t.Fatal(err)
		}
		network.Name = "fuzz"
		available := make([]int, 3, 15)
		for index := range available {
			available[index], err = network.CreatePI()
			if err != nil {
				t.Fatal(err)
			}
		}
		for _, operation := range operations {
			first := available[int(operation)%len(available)]
			second := available[int(operation/7)%len(available)]
			sop := "11 1\n"
			switch operation & 3 {
			case 1:
				sop = "00 0\n"
			case 2:
				sop = "10 1\n01 1\n"
			case 3:
				sop = "1- 1\n-1 1\n"
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
		manager, err := BblManagerFromABCNetwork(network, ABCBblifOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			inputs := map[int]bool{}
			for index, id := range network.cis {
				inputs[id] = assignment&(1<<uint(index)) != 0
			}
			got := evaluateBblifManager(t, manager, inputs)
			want := evaluateABCBblifNetwork(t, network, inputs)
			if got != want {
				t.Fatalf("assignment=%d got=%v want=%v", assignment, got, want)
			}
		}
	})
}
