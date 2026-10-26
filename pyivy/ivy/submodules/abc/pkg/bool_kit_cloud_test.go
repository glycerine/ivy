package abc

// Tests for the Go port of src/bool/kit/cloud.c and cloud.h.

import (
	"errors"
	"math/bits"
	"testing"
)

func cloudMust(t testing.TB, operation func() (CloudEdge, error)) CloudEdge {
	t.Helper()
	edge, err := operation()
	if err != nil {
		t.Fatal(err)
	}
	return edge
}

func cloudTruthEdge(t testing.TB, manager *CloudManager, variables int, truth uint64) CloudEdge {
	t.Helper()
	var build func(int, uint64) CloudEdge
	build = func(variable int, assignment uint64) CloudEdge {
		if variable == variables {
			if truth>>assignment&1 != 0 {
				return manager.One()
			}
			return manager.Zero()
		}
		otherwise := build(variable+1, assignment)
		then := build(variable+1, assignment|uint64(1)<<variable)
		return cloudMust(t, func() (CloudEdge, error) { return manager.MakeNode(variable, then, otherwise) })
	}
	return build(0, 0)
}

func cloudEdgeTruth(t testing.TB, manager *CloudManager, edge CloudEdge, variables int) uint64 {
	t.Helper()
	var truth uint64
	assignment := make([]bool, variables)
	for value := 0; value < 1<<variables; value++ {
		for variable := range assignment {
			assignment[variable] = value>>variable&1 != 0
		}
		result, err := manager.Evaluate(edge, assignment)
		if err != nil {
			t.Fatal(err)
		}
		if result {
			truth |= uint64(1) << value
		}
	}
	return truth
}

func TestCloudAllFourVariableFunctionsCanonical(t *testing.T) {
	manager, err := NewCloudManager(4, 1<<16)
	if err != nil {
		t.Fatal(err)
	}
	edges := make(map[CloudEdge]uint16, 1<<16)
	for truth := 0; truth < 1<<16; truth++ {
		edge := cloudTruthEdge(t, manager, 4, uint64(truth))
		if previous, found := edges[edge]; found && previous != uint16(truth) {
			t.Fatalf("noncanonical collision %04x/%04x", previous, truth)
		}
		edges[edge] = uint16(truth)
		if got := cloudEdgeTruth(t, manager, edge, 4); got != uint64(truth) {
			t.Fatalf("truth %04x became %04x", truth, got)
		}
	}
	if len(edges) != 1<<16 {
		t.Fatalf("canonical edges %d", len(edges))
	}
}

func TestCloudExhaustiveThreeVariableOperations(t *testing.T) {
	manager, err := NewCloudManager(3, 1<<14)
	if err != nil {
		t.Fatal(err)
	}
	edges := make([]CloudEdge, 256)
	for truth := range edges {
		edges[truth] = cloudTruthEdge(t, manager, 3, uint64(truth))
	}
	const mask = uint64(0xff)
	for first := range edges {
		for second := range edges {
			and := cloudMust(t, func() (CloudEdge, error) { return manager.And(edges[first], edges[second], CloudOptions{}) })
			or := cloudMust(t, func() (CloudEdge, error) { return manager.Or(edges[first], edges[second], CloudOptions{}) })
			xor := cloudMust(t, func() (CloudEdge, error) { return manager.Xor(edges[first], edges[second], CloudOptions{}) })
			if got := cloudEdgeTruth(t, manager, and, 3); got != uint64(first&second) {
				t.Fatalf("and %02x %02x: %02x", first, second, got)
			}
			if got := cloudEdgeTruth(t, manager, or, 3); got != uint64(first|second) {
				t.Fatalf("or %02x %02x: %02x", first, second, got)
			}
			if got := cloudEdgeTruth(t, manager, xor, 3); got != uint64(first^second)&mask {
				t.Fatalf("xor %02x %02x: %02x", first, second, got)
			}
		}
	}
}

func TestCloudSupportCollectionCubeAndFormatting(t *testing.T) {
	manager, err := NewCloudManager(4, 256)
	if err != nil {
		t.Fatal(err)
	}
	x0 := cloudMust(t, func() (CloudEdge, error) { return manager.Variable(0) })
	x1 := cloudMust(t, func() (CloudEdge, error) { return manager.Variable(1) })
	x2 := cloudMust(t, func() (CloudEdge, error) { return manager.Variable(2) })
	and12 := cloudMust(t, func() (CloudEdge, error) { return manager.And(x1, x2, CloudOptions{}) })
	root := cloudMust(t, func() (CloudEdge, error) { return manager.Xor(x0, and12, CloudOptions{}) })
	support, err := manager.Support(root, CloudOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{0, 1, 2}; len(support) != len(want) || support[0] != 0 || support[1] != 1 || support[2] != 2 {
		t.Fatalf("support %v", support)
	}
	if size, sizeErr := manager.SupportSize(root, CloudOptions{}); sizeErr != nil || size != 3 {
		t.Fatalf("support size %d %v", size, sizeErr)
	}
	cube := cloudMust(t, func() (CloudEdge, error) { return manager.SupportCube(root, CloudOptions{}) })
	if got := cloudEdgeTruth(t, manager, cube, 4); got != uint64(1)<<7|uint64(1)<<15 {
		t.Fatalf("support cube %04x", got)
	}
	nodes, err := manager.DagCollect(root, CloudOptions{})
	if err != nil {
		t.Fatal(err)
	}
	size, err := manager.DagSize(root, CloudOptions{})
	if err != nil || size != len(nodes) || len(nodes) == 0 || nodes[0] != manager.One() {
		t.Fatalf("collect=%v size=%d err=%v", nodes, size, err)
	}
	sharing, err := manager.SharingSize([]CloudEdge{root, and12}, CloudOptions{})
	if err != nil || sharing > size+3 || sharing < size {
		t.Fatalf("sharing %d size %d err %v", sharing, size, err)
	}
	oneCube := cloudMust(t, func() (CloudEdge, error) { return manager.GetOneCube(root, CloudOptions{}) })
	if oneCube == manager.Zero() {
		t.Fatal("missing cube")
	}
	for assignment := 0; assignment < 16; assignment++ {
		cubeValue := cloudEdgeTruth(t, manager, oneCube, 4)>>assignment&1 != 0
		rootValue := cloudEdgeTruth(t, manager, root, 4)>>assignment&1 != 0
		if cubeValue && !rootValue {
			t.Fatalf("cube escapes at %d", assignment)
		}
	}
	if text, err := manager.FormatCube(oneCube); err != nil || text == "" {
		t.Fatalf("cube text %q %v", text, err)
	}
	if text, err := manager.FormatDNF(root, CloudOptions{}); err != nil || text == "" || text[len(text)-1] != '\n' {
		t.Fatalf("dnf %q %v", text, err)
	}
	if text, _ := manager.FormatDNF(manager.Zero(), CloudOptions{}); text != "Constant 0.\n" {
		t.Fatalf("zero %q", text)
	}
}

func TestCloudRestartLimitsCancellationAndLifecycle(t *testing.T) {
	manager, err := NewCloudManager(2, 4)
	if err != nil {
		t.Fatal(err)
	}
	x0 := cloudMust(t, func() (CloudEdge, error) { return manager.Variable(0) })
	x1 := cloudMust(t, func() (CloudEdge, error) { return manager.Variable(1) })
	before, _ := manager.Statistics()
	and := cloudMust(t, func() (CloudEdge, error) { return manager.And(x0, x1, CloudOptions{}) })
	if and == x0 || and == x1 {
		t.Fatal("and was not created")
	}
	if _, err = manager.Or(x0, x1, CloudOptions{}); !errors.Is(err, ErrLimit) {
		t.Fatalf("limit %v", err)
	}
	after, _ := manager.Statistics()
	if after.Nodes != before.Nodes+1 {
		t.Fatalf("failed operation was not atomic: %d -> %d", before.Nodes, after.Nodes)
	}
	if _, err = manager.And(x0, x1, CloudOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel %v", err)
	}
	if err = manager.AllocateCache(CloudOperationAnd, 3); err != nil {
		t.Fatal(err)
	}
	if err = manager.AllocateCacheRatio(CloudOperationXor, 1); err != nil {
		t.Fatal(err)
	}
	if err = manager.Restart(); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Evaluate(x0, []bool{false, false}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("stale edge %v", err)
	}
	newX0 := cloudMust(t, func() (CloudEdge, error) { return manager.Variable(0) })
	if newX0 == x0 || newX0.generation() != x0.generation()+1 {
		t.Fatalf("restart edge %#x -> %#x", x0, newX0)
	}
	statistics, err := manager.Close()
	if err != nil || statistics.Nodes != 3 || statistics.Format() == "" {
		t.Fatalf("close %+v %v", statistics, err)
	}
	if _, err = manager.Variable(0); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed manager %v", err)
	}
}

func TestCloudAlgebraAndStatistics(t *testing.T) {
	manager, err := NewCloudManager(3, 256)
	if err != nil {
		t.Fatal(err)
	}
	x := make([]CloudEdge, 3)
	for variable := range x {
		variable := variable
		x[variable] = cloudMust(t, func() (CloudEdge, error) { return manager.Variable(variable) })
	}
	for _, first := range append(append([]CloudEdge{}, x...), manager.Zero(), manager.One()) {
		if got := cloudMust(t, func() (CloudEdge, error) { return manager.And(first, first, CloudOptions{}) }); got != first {
			t.Fatal("and idempotence")
		}
		if got := cloudMust(t, func() (CloudEdge, error) { return manager.Xor(first, first, CloudOptions{}) }); got != manager.Zero() {
			t.Fatal("xor self")
		}
		if got := cloudMust(t, func() (CloudEdge, error) { return manager.Or(first, first.Not(), CloudOptions{}) }); got != manager.One() {
			t.Fatal("excluded middle")
		}
	}
	first := cloudMust(t, func() (CloudEdge, error) { return manager.And(x[0], x[1], CloudOptions{}) })
	second := cloudMust(t, func() (CloudEdge, error) { return manager.And(x[1], x[0], CloudOptions{}) })
	if first != second {
		t.Fatal("commutative result not canonical")
	}
	occupancy, err := manager.HashOccupancy()
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, occupied := range occupancy {
		if occupied {
			count++
		}
	}
	statistics, err := manager.Statistics()
	if err != nil || count != statistics.Nodes-1 || statistics.UniqueMisses == 0 || statistics.MemoryBytes == 0 {
		t.Fatalf("stats=%+v occupancy=%d err=%v", statistics, count, err)
	}
}

func FuzzCloudOperations(f *testing.F) {
	f.Add([]byte{2, 0, 1, 2, 3, 4, 5, 6})
	f.Add([]byte{4, 3, 3, 2, 1, 0, 255, 17, 91})
	f.Add([]byte{1, 4, 4, 4, 4})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		variables := 1 + int(data[0]%4)
		manager, err := NewCloudManager(variables, 1<<14)
		if err != nil {
			t.Fatal(err)
		}
		mask := uint64(1)<<(1<<variables) - 1
		edges := []CloudEdge{manager.Zero(), manager.One()}
		truths := []uint64{0, mask}
		for variable := 0; variable < variables; variable++ {
			variable := variable
			edges = append(edges, cloudMust(t, func() (CloudEdge, error) { return manager.Variable(variable) }))
			var truth uint64
			for assignment := 0; assignment < 1<<variables; assignment++ {
				if assignment>>variable&1 != 0 {
					truth |= uint64(1) << assignment
				}
			}
			truths = append(truths, truth)
		}
		canonical := make(map[uint64]CloudEdge)
		for index, truth := range truths {
			canonical[truth] = edges[index]
		}
		for position := 1; position < len(data) && position < 193; position++ {
			code := data[position]
			first := int(code) % len(edges)
			second := int(bits.Reverse8(code)^byte(position)) % len(edges)
			var edge CloudEdge
			var truth uint64
			switch code % 4 {
			case 0:
				edge, err = manager.And(edges[first], edges[second], CloudOptions{MaxOperations: 1 << 16})
				truth = truths[first] & truths[second]
			case 1:
				edge, err = manager.Or(edges[first], edges[second], CloudOptions{MaxOperations: 1 << 16})
				truth = truths[first] | truths[second]
			case 2:
				edge, err = manager.Xor(edges[first], edges[second], CloudOptions{MaxOperations: 1 << 16})
				truth = truths[first] ^ truths[second]
			default:
				edge = edges[first].Not()
				truth = ^truths[first] & mask
			}
			if err != nil {
				t.Fatal(err)
			}
			truth &= mask
			if old, found := canonical[truth]; found && old != edge {
				t.Fatalf("noncanonical truth %x: %#x %#x", truth, old, edge)
			}
			canonical[truth] = edge
			if got := cloudEdgeTruth(t, manager, edge, variables); got != truth {
				t.Fatalf("truth %x got %x", truth, got)
			}
			edges = append(edges, edge)
			truths = append(truths, truth)
		}
	})
}
