package abc

// Go-only model and fuzz tests for src/aig/gia/giaPack.c.

import (
	"errors"
	"math/rand"
	"reflect"
	"slices"
	"testing"
)

type giaPackFixtureData struct {
	manager             *GiaManager
	mapping             *GiaLUTMapping
	first, second       GiaID
	join, top, dangling GiaID
}

func giaPackFixture(t *testing.T) giaPackFixtureData {
	t.Helper()
	manager := NewGiaManager()
	inputs := make([]GiaLit, 4)
	for index := range inputs {
		var err error
		inputs[index], err = manager.AppendCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	first, _ := manager.AppendAnd(inputs[0], inputs[1])
	second, _ := manager.AppendAnd(inputs[2], inputs[3])
	join, _ := manager.AppendAnd(first, second)
	top, _ := manager.AppendAnd(join, inputs[0])
	dangling, _ := manager.AppendAnd(inputs[1], inputs[2])
	if _, err := manager.AppendCO(top.Not()); err != nil {
		t.Fatal(err)
	}
	records := []GiaLUTRecord{
		{Root: first.ID(), Fanins: []GiaID{inputs[0].ID(), inputs[1].ID()}},
		{Root: second.ID(), Fanins: []GiaID{inputs[2].ID(), inputs[3].ID()}},
		{Root: join.ID(), Fanins: []GiaID{first.ID(), second.ID()}},
		{Root: top.ID(), Fanins: []GiaID{join.ID(), inputs[0].ID()}},
		{Root: dangling.ID(), Fanins: []GiaID{inputs[1].ID(), inputs[2].ID()}},
	}
	mapping, err := NewGiaLUTMapping(manager.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	return giaPackFixtureData{manager, mapping, first.ID(), second.ID(), join.ID(), top.ID(), dangling.ID()}
}

func TestGiaLUTPackOrdersAndDelays(t *testing.T) {
	fixture := giaPackFixture(t)
	beforeObjects := slices.Clone(fixture.manager.objects)
	beforeMapping := fixture.mapping.Packed()
	reachable, err := fixture.mapping.LUTReachableOrder(fixture.manager, GiaLUTPackOptions{})
	if err != nil || !slices.Equal(reachable, []GiaID{fixture.second, fixture.first, fixture.join, fixture.top}) {
		t.Fatal(reachable, err)
	}
	order, err := fixture.mapping.LUTDistanceOrder(fixture.manager, GiaLUTPackOptions{})
	wantOrder := []GiaID{fixture.first, fixture.second, fixture.join, fixture.top, fixture.dangling}
	if err != nil || !slices.Equal(order, wantOrder) {
		t.Fatal(order, wantOrder, err)
	}
	result, err := fixture.mapping.Pack(fixture.manager, 2, 10, 2, GiaLUTPackOptions{})
	if err != nil || result.Packing == nil || result.MaximumDelay != 22 || !slices.Equal(result.Order, wantOrder) {
		t.Fatal(result, err)
	}
	wantGroups := [][]GiaID{{fixture.first}, {fixture.second}, {fixture.top, fixture.join}, {fixture.dangling}}
	if !reflect.DeepEqual(result.Packing.Groups(), wantGroups) {
		t.Fatal(result.Packing.Groups(), wantGroups)
	}
	for id, want := range map[GiaID]int{fixture.first: 10, fixture.second: 10, fixture.join: 20, fixture.top: 22, fixture.dangling: 10} {
		if result.Delays[id] != want {
			t.Fatal(id, result.Delays[id], want)
		}
	}
	for id, want := range map[GiaID]int{fixture.first: 0, fixture.second: 1, fixture.join: 2, fixture.top: 2, fixture.dangling: 3} {
		if result.Blocks[id] != want {
			t.Fatal(id, result.Blocks[id], want)
		}
	}
	if !slices.Equal(beforeObjects, fixture.manager.objects) || !slices.Equal(beforeMapping, fixture.mapping.Packed()) {
		t.Fatal("packing mutated graph or mapping")
	}
}

func TestGiaLUTPackWideBlocks(t *testing.T) {
	manager := NewGiaManager()
	a, _ := manager.AppendCI()
	b, _ := manager.AppendCI()
	chain, _ := manager.AppendAnd(a, b)
	records := []GiaLUTRecord{{Root: chain.ID(), Fanins: []GiaID{a.ID(), b.ID()}}}
	roots := []GiaID{chain.ID()}
	for index := 0; index < 4; index++ {
		previous := chain
		chain, _ = manager.AppendAnd(previous, a.NotIf(index&1 != 0))
		records = append(records, GiaLUTRecord{Root: chain.ID(), Fanins: []GiaID{previous.ID(), a.ID()}})
		roots = append(roots, chain.ID())
	}
	manager.AppendCO(chain)
	mapping, err := NewGiaLUTMapping(manager.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	result, err := mapping.Pack(manager, 4, 10, 2, GiaLUTPackOptions{})
	if err != nil || result.MaximumDelay != 26 {
		t.Fatal(result, err)
	}
	want := [][]GiaID{{roots[3], roots[2], roots[1], roots[0]}, {roots[4]}}
	if !reflect.DeepEqual(result.Packing.Groups(), want) || result.Packing.Statistics().BySize[4] != 1 {
		t.Fatal(result.Packing.Groups(), want, result.Packing.Statistics())
	}
	if _, err = result.Packing.Edges(GiaEdgeOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
}

func TestGiaLUTPackErrorsLimitsAndCancellation(t *testing.T) {
	fixture := giaPackFixture(t)
	for _, parameters := range [][3]int{{1, 10, 2}, {2, 0, 2}, {2, 10, 0}} {
		if result, err := fixture.mapping.Pack(fixture.manager, parameters[0], parameters[1], parameters[2], GiaLUTPackOptions{}); err == nil || result.Packing != nil {
			t.Fatal(parameters, result, err)
		}
	}
	for _, options := range []GiaLUTPackOptions{{MaxObjects: -1}, {MaxVisits: -1}, {MaxEntries: -1}, {MaxObjects: 1}, {MaxVisits: 1}, {MaxEntries: 1}} {
		if result, err := fixture.mapping.Pack(fixture.manager, 2, 10, 2, options); err == nil || result.Packing != nil {
			t.Fatal(options, result, err)
		}
	}
	for threshold := 1; threshold < 12; threshold++ {
		calls := 0
		result, err := fixture.mapping.Pack(fixture.manager, 2, 10, 2, GiaLUTPackOptions{Stop: func() bool { calls++; return calls == threshold }})
		if err != nil && !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, err)
		}
		if errors.Is(err, ErrCancelled) && result.Packing != nil {
			t.Fatal(threshold, result)
		}
	}
	if _, err := (*GiaLUTMapping)(nil).LUTDistanceOrder(fixture.manager, GiaLUTPackOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := fixture.mapping.LUTReachableOrder(nil, GiaLUTPackOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := fixture.mapping.LUTReachableOrder(fixture.manager, GiaLUTPackOptions{MaxVisits: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	manager := NewGiaManager()
	a, _ := manager.AppendCI()
	b, _ := manager.AppendCI()
	root, _ := manager.AppendAnd(a, b)
	manager.AppendCO(root)
	missing, _ := NewGiaLUTMapping(manager.ObjectCount(), nil, 0)
	if _, err := missing.LUTReachableOrder(manager, GiaLUTPackOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	one, _ := NewGiaLUTMapping(manager.ObjectCount(), []GiaLUTRecord{{Root: root.ID(), Fanins: []GiaID{a.ID()}}}, 0)
	if result, err := one.Pack(manager, 2, 10, 2, GiaLUTPackOptions{}); !errors.Is(err, ErrDimension) || result.Packing != nil {
		t.Fatal(result, err)
	}
}

func giaPackReference(mapping *GiaLUTMapping, manager *GiaManager, blockSize, routeDelay, directDelay int) ([]GiaID, [][]GiaID, []int, int) {
	roots := mapping.Roots()
	distance := make([]int, manager.ObjectCount())
	for _, output := range manager.cos {
		driver, _ := manager.objects[output].Fanin0(output)
		distance[driver.ID()] = 1
	}
	for index := len(roots) - 1; index >= 0; index-- {
		record, _ := mapping.Record(roots[index])
		for _, fanin := range record.Fanins {
			distance[fanin] = max(distance[fanin], distance[record.Root]+1)
		}
	}
	costs := make([]int, len(roots))
	for index, root := range roots {
		costs[index] = -distance[root]
	}
	permutation := IntCostPermutation(costs)
	order := make([]GiaID, len(roots))
	for index, position := range permutation {
		order[index] = roots[position]
	}
	delays, blocks := make([]int, manager.ObjectCount()), make([]int, manager.ObjectCount())
	blockSizes := make([]int, 0)
	maximum := 0
	for _, root := range order {
		record, _ := mapping.Record(root)
		fanins := slices.Clone(record.Fanins)
		values := make([]int, len(fanins))
		for index, fanin := range fanins {
			values[index] = delays[fanin]
		}
		for index := 0; index < len(fanins)-1; index++ {
			best := index
			for candidate := index + 1; candidate < len(fanins); candidate++ {
				if values[candidate] > values[best] {
					best = candidate
				}
			}
			fanins[index], fanins[best] = fanins[best], fanins[index]
			values[index], values[best] = values[best], values[index]
		}
		block := blocks[fanins[0]]
		if values[0] > 0 && values[0] > values[1] && blockSizes[block] < blockSize {
			delays[root], blocks[root] = values[0]+directDelay, block
			blockSizes[block]++
		} else {
			delays[root], blocks[root] = values[0]+routeDelay, len(blockSizes)
			blockSizes = append(blockSizes, 1)
		}
		for index := 1; index < len(values); index++ {
			delays[root] = max(delays[root], values[index]+routeDelay)
		}
		maximum = max(maximum, delays[root])
	}
	groups := make([][]GiaID, len(blockSizes))
	for index := len(order) - 1; index >= 0; index-- {
		root := order[index]
		groups[blocks[root]] = append(groups[blocks[root]], root)
	}
	return order, groups, delays, maximum
}

func giaPackReachableReference(mapping *GiaLUTMapping, manager *GiaManager) []GiaID {
	seen := make([]bool, manager.ObjectCount())
	order := make([]GiaID, 0)
	for _, output := range manager.cos {
		driver, _ := manager.objects[output].Fanin0(output)
		if manager.objects[driver.ID()].IsAnd() && !seen[driver.ID()] {
			seen[driver.ID()] = true
			order = append(order, driver.ID())
		}
	}
	for index := 0; index < len(order); index++ {
		record, _ := mapping.Record(order[index])
		for _, fanin := range record.Fanins {
			if manager.objects[fanin].IsAnd() && !seen[fanin] {
				seen[fanin] = true
				order = append(order, fanin)
			}
		}
	}
	slices.Reverse(order)
	return order
}

func FuzzGiaLUTPack(f *testing.F) {
	for _, seed := range []uint64{1, 2, 3, 17, 1001} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		manager := NewGiaManager()
		literals := make([]GiaLit, 2+random.Intn(5))
		for index := range literals {
			literals[index], _ = manager.AppendCI()
		}
		records := make([]GiaLUTRecord, 0)
		for gate := 0; gate < 5+random.Intn(24); gate++ {
			leftIndex := random.Intn(len(literals))
			rightIndex := random.Intn(len(literals) - 1)
			if rightIndex >= leftIndex {
				rightIndex++
			}
			left := literals[leftIndex].NotIf(random.Intn(2) != 0)
			right := literals[rightIndex].NotIf(random.Intn(2) != 0)
			root, err := manager.AppendAnd(left, right)
			if err != nil {
				t.Fatal(err)
			}
			records = append(records, GiaLUTRecord{Root: root.ID(), Fanins: []GiaID{left.ID(), right.ID()}})
			literals = append(literals, root)
		}
		for output := 0; output < 1+random.Intn(4); output++ {
			root := records[random.Intn(len(records))].Root
			manager.AppendCO(GiaLit(root) << 1)
		}
		mapping, err := NewGiaLUTMapping(manager.ObjectCount(), records, 0)
		if err != nil {
			t.Fatal(err)
		}
		blockSize := 2 + random.Intn(6)
		routeDelay := 1 + random.Intn(16)
		directDelay := 1 + random.Intn(16)
		actual, err := mapping.Pack(manager, blockSize, routeDelay, directDelay, GiaLUTPackOptions{})
		if err != nil {
			t.Fatal(err)
		}
		order, groups, delays, maximum := giaPackReference(mapping, manager, blockSize, routeDelay, directDelay)
		if !slices.Equal(actual.Order, order) || !reflect.DeepEqual(actual.Packing.Groups(), groups) || !slices.Equal(actual.Delays, delays) || actual.MaximumDelay != maximum {
			t.Fatal(seed, actual.Packing.Groups(), groups, actual.MaximumDelay, maximum)
		}
		reachable, err := mapping.LUTReachableOrder(manager, GiaLUTPackOptions{})
		if err != nil || !slices.Equal(reachable, giaPackReachableReference(mapping, manager)) {
			t.Fatal(seed, reachable, err)
		}
	})
}
