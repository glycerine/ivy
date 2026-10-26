package abc

import (
	"slices"
	"testing"
)

func FuzzGiaConeAnalysis(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
	f.Add([]byte{3, 255, 23, 9, 17, 43, 92, 6, 71})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 48 {
			data = data[:48]
		}
		get := func(i int) byte { return data[i%len(data)] }
		g := NewGiaManager()
		signals := make([]GiaLit, 4)
		for i := range signals {
			var err error
			signals[i], err = g.AppendCI()
			if err != nil {
				t.Fatal(err)
			}
		}
		type gate struct {
			kind    byte
			args    []GiaLit
			inverse bool
		}
		records := make(map[GiaID]gate)
		for step := 0; step < 12; step++ {
			a := int(get(3*step)) % len(signals)
			b := int(get(3*step+1)) % len(signals)
			if b == a {
				b = (b + 1) % len(signals)
			}
			c := (b + 1) % len(signals)
			for c == a || c == b {
				c = (c + 1) % len(signals)
			}
			la, lb, lc := signals[a], signals[b], signals[c]
			if get(step)&16 != 0 {
				la = la.Not()
			}
			if get(step)&32 != 0 {
				lb = lb.Not()
			}
			kind := get(3*step+2) % 4
			var result GiaLit
			var err error
			args := []GiaLit{la, lb}
			switch kind {
			case 0:
				result, err = g.AppendAnd(la, lb)
			case 1:
				result, err = g.AppendXorReal(la, lb)
			case 2:
				args = []GiaLit{lc, la, lb}
				result, err = g.AppendMuxReal(lc, la, lb)
			case 3:
				args = []GiaLit{la}
				result, err = g.AppendBuffer(la)
			}
			if err != nil {
				t.Fatal(err)
			}
			records[result.ID()] = gate{kind: kind, args: args, inverse: result.Complemented()}
			signals = append(signals, result)
		}
		root := signals[len(signals)-1].ID()
		extra := signals[int(get(0))%len(signals)]
		if _, err := g.AppendCO(GiaLit(root) << 1); err != nil {
			t.Fatal(err)
		}
		if _, err := g.AppendCO(extra); err != nil {
			t.Fatal(err)
		}
		// Independent edge ledger comes from constructor arguments, not manager
		// fanin decoding, reference counting or the static fanout implementation.
		consumers := make(map[GiaID][]GiaID)
		for id, record := range records {
			for _, arg := range record.args {
				consumers[arg.ID()] = append(consumers[arg.ID()], id)
			}
		}
		consumers[root] = append(consumers[root], GiaID(g.ObjectCount()-2))
		consumers[extra.ID()] = append(consumers[extra.ID()], GiaID(g.ObjectCount()-1))
		threshold := int(get(0) % 4)
		selected, err := g.CofactorCandidates(threshold, 0)
		if err != nil {
			t.Fatal(err)
		}
		expectedSelected := make([]GiaID, 0)
		for id := GiaID(1); int(id) < g.ObjectCount()-2; id++ {
			if len(consumers[id]) >= threshold {
				expectedSelected = append(expectedSelected, id)
			}
		}
		if !slices.Equal(selected, expectedSelected) {
			t.Fatal("independent pivot selection", selected, expectedSelected)
		}
		dangling, err := g.DanglingRoots(0)
		if err != nil {
			t.Fatal(err)
		}
		expectedDangles := make(map[GiaID]bool)
		for id := range records {
			if len(consumers[id]) == 0 {
				expectedDangles[id] = true
			}
		}
		if len(dangling) != len(expectedDangles) {
			t.Fatal("independent dangling root count", dangling, expectedDangles)
		}
		for _, id := range dangling {
			if !expectedDangles[id] {
				t.Fatal("independent dangling root", id)
			}
		}
		wantReverse := make([]uint32, g.ObjectCount())
		wantForward := make([]uint32, g.ObjectCount())
		forwardMaximum := uint32(0)
		for id := GiaID(5); int(id) < g.ObjectCount()-2; id++ {
			record := records[id]
			delay := uint32(1)
			if record.kind == 1 || record.kind == 2 {
				delay = 2
			} else if record.kind == 3 {
				delay = 0
			}
			level := uint32(0)
			for _, arg := range record.args {
				level = max(level, wantForward[arg.ID()])
			}
			wantForward[id] = level + delay
			forwardMaximum = max(forwardMaximum, level+delay)
		}
		for id := GiaID(g.ObjectCount() - 3); id >= 5; id-- {
			record := records[id]
			delay := uint32(1)
			if record.kind == 1 || record.kind == 2 {
				delay = 2
			} else if record.kind == 3 {
				delay = 0
			}
			for _, arg := range record.args {
				wantReverse[arg.ID()] = max(wantReverse[arg.ID()], wantReverse[id]+delay)
			}
		}
		weighted, err := g.WeightedReverseLevels(0)
		if err != nil {
			t.Fatal(err)
		}
		reverseMaximum := uint32(0)
		for i := 1; i <= 4; i++ {
			reverseMaximum = max(reverseMaximum, wantReverse[i])
		}
		wantForward[g.ObjectCount()-2] = wantForward[root]
		wantForward[g.ObjectCount()-1] = wantForward[extra.ID()]
		required, err := g.RequiredLevels(0)
		if err != nil {
			t.Fatal(err)
		}
		slacks, err := g.LevelSlacks(0)
		if err != nil {
			t.Fatal(err)
		}
		for i := range weighted {
			if weighted[i] != wantReverse[i] || required[i] != int64(reverseMaximum)-int64(wantReverse[i]) || slacks[i] != int64(forwardMaximum)-int64(wantForward[i])-int64(wantReverse[i]) {
				t.Fatal("independent weighted depth", i)
			}
		}
		// Independent reachability on constructor records checks cardinalities.
		reach := make(map[GiaID]bool)
		pending := []GiaID{root}
		wantInputs, wantGates := 0, 0
		for len(pending) > 0 {
			id := pending[len(pending)-1]
			pending = pending[:len(pending)-1]
			if reach[id] {
				continue
			}
			reach[id] = true
			if record, gate := records[id]; gate {
				wantGates++
				for _, arg := range record.args {
					pending = append(pending, arg.ID())
				}
			} else if id != 0 {
				wantInputs++
			}
		}
		stats, err := g.ConeStatistics([]GiaID{root, GiaID(g.ObjectCount() - 2)}, 0)
		if err != nil || stats.Inputs != wantInputs || stats.Gates != wantGates {
			t.Fatal("independent cone statistics", stats, wantInputs, wantGates, err)
		}
		// Independent postdominator-set intersection, without parent climbing.
		dominators, err := g.ClosestDominators(0)
		if err != nil {
			t.Fatal(err)
		}
		post := make([]map[GiaID]bool, g.ObjectCount())
		for index := g.ObjectCount() - 1; index >= 0; index-- {
			id := GiaID(index)
			edges := consumers[id]
			common := make(map[GiaID]bool)
			if len(edges) > 0 {
				for candidate := range post[edges[0]] {
					common[candidate] = true
				}
				for _, consumer := range edges[1:] {
					for candidate := range common {
						if !post[consumer][candidate] {
							delete(common, candidate)
						}
					}
				}
			}
			expected := InvalidGiaID
			if len(edges) > 0 {
				expected = id
				for candidate := range common {
					if expected == id || candidate < expected {
						expected = candidate
					}
				}
			} else if index >= g.ObjectCount()-2 {
				expected = id
			}
			if dominators[index] != expected {
				t.Fatal("independent closest dominator", index, dominators[index], expected)
			}
			common[id] = true
			post[index] = common
		}
		cone := map[GiaID]bool{root: true}
		for changed := true; changed; {
			changed = false
			for id, record := range records {
				if !cone[id] {
					continue
				}
				for _, arg := range record.args {
					child := arg.ID()
					if _, gate := records[child]; !gate || cone[child] {
						continue
					}
					internal := true
					for _, consumer := range consumers[child] {
						if !cone[consumer] {
							internal = false
						}
					}
					if internal {
						cone[child] = true
						changed = true
					}
				}
			}
		}
		boundary := make(map[GiaID]int64)
		for id := range cone {
			for _, arg := range records[id].args {
				if !cone[arg.ID()] {
					boundary[arg.ID()]--
				}
			}
		}
		got, err := g.Mffc(root, 0)
		if err != nil {
			t.Fatal(err)
		}
		if got.Size != len(cone) || len(got.Boundary) != len(boundary) {
			t.Fatal("independent MFFC", got, cone, boundary)
		}
		for i, leaf := range got.Boundary {
			delta, ok := boundary[leaf]
			if !ok || delta != got.BoundaryReferenceDeltas[i] {
				t.Fatal("independent boundary", got, boundary)
			}
		}
		// Independently evaluate the constructor ledger, then all cofactor rows.
		guardPivot := GiaID(1 + int(get(2))%(g.ObjectCount()-3))
		guardFixed := get(3)&1 != 0
		guardTruth := make([]bool, 16)
		allTruth := make([][]bool, g.ObjectCount())
		for i := range allTruth {
			allTruth[i] = make([]bool, 16)
		}
		truth := make([]bool, 16)
		extraTruth := make([]bool, 16)
		for assignment := 0; assignment < 16; assignment++ {
			values := make([]bool, g.ObjectCount())
			for i := 0; i < 4; i++ {
				values[i+1] = assignment&(1<<i) != 0
			}
			for id := GiaID(5); int(id) < g.ObjectCount()-2; id++ {
				record := records[id]
				read := func(l GiaLit) bool { return values[l.ID()] != l.Complemented() }
				value := read(record.args[0])
				switch record.kind {
				case 0:
					value = value && read(record.args[1])
				case 1:
					value = value != read(record.args[1])
				case 2:
					if value {
						value = read(record.args[1])
					} else {
						value = read(record.args[2])
					}
				}
				values[id] = value != record.inverse
			}
			for id, value := range values {
				allTruth[id][assignment] = value
			}
			guardTruth[assignment] = values[guardPivot]
			truth[assignment] = values[root]
			extraTruth[assignment] = values[extra.ID()] != extra.Complemented()
		}
		var limited GiaCutExtraction
		var limitErr error
		if get(0)&1 == 0 {
			limited, limitErr = g.ExtractDepthLimitedCones([]GiaID{root, 5}, int(get(1)%4), GiaCofactorOptions{MaxObjects: 4096, MaxMappingEntries: 4096})
		} else {
			limited, limitErr = g.ExtractLevelLimitedCones([]GiaID{root, 5}, uint32(get(1)%4), GiaCofactorOptions{MaxObjects: 4096, MaxMappingEntries: 4096})
		}
		if limitErr != nil {
			t.Fatal(limitErr)
		}
		allIDs := make([]GiaID, len(g.objects))
		supportMasks := make([]uint8, len(g.objects))
		for i := 0; i < 4; i++ {
			supportMasks[i+1] = 1 << i
		}
		for i := range allIDs {
			allIDs[i] = GiaID(i)
			if record, ok := records[GiaID(i)]; ok {
				for _, arg := range record.args {
					supportMasks[i] |= supportMasks[arg.ID()]
				}
			}
		}
		supportMasks[g.cos[0]] = supportMasks[root]
		supportMasks[g.cos[1]] = supportMasks[extra.ID()]
		sets, supportErr := g.NodeSupports(allIDs, 0, 0)
		if supportErr != nil {
			t.Fatal(supportErr)
		}
		for i, set := range sets {
			expected := make([]int, 0)
			for ci := 0; ci < 4; ci++ {
				if supportMasks[i]&(1<<ci) != 0 {
					expected = append(expected, ci)
				}
			}
			if !slices.Equal(set, expected) {
				t.Fatal("independent structural support", i, set, expected)
			}
		}
		reduced, reduceErr := g.ReduceBySupport(get(0)&1 != 0, GiaCofactorOptions{})
		if reduceErr != nil {
			t.Fatal(reduceErr)
		}
		for assignment := 0; assignment < 16; assignment++ {
			for output, support := range reduced.Supports {
				projected := make([]bool, reduced.Graph.InputCount())
				for i, ci := range support {
					projected[i] = assignment&(1<<ci) != 0
				}
				values, err := reduced.Graph.Evaluate(projected)
				want := truth[assignment]
				if output == 1 {
					want = extraTruth[assignment]
				}
				if err != nil || values[reduced.Outputs[output]] != want {
					t.Fatal("independent support reduction", assignment, output, err)
				}
			}
		}
		exposed, _, exposeErr := g.DuplicateWithGateOutputs(GiaCofactorOptions{})
		if exposeErr != nil {
			t.Fatal(exposeErr)
		}
		for assignment := 0; assignment < 16; assignment++ {
			values, err := exposed.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0})
			if err != nil || values[0] != truth[assignment] || values[1] != extraTruth[assignment] {
				t.Fatal("exposed original outputs", err)
			}
			index := 2
			for id, obj := range g.objects {
				if obj.IsAnd() {
					if values[index] != allTruth[id][assignment] {
						t.Fatal("independent exposed gate", id, assignment)
					}
					index++
				}
			}
		}
		sliceLimit := int(get(1) % 5)
		sliced, sliceErr := g.SliceBySupport(sliceLimit, GiaCofactorOptions{})
		if sliceErr != nil {
			t.Fatal(sliceErr)
		}
		countSupport := func(mask uint8) int {
			count := 0
			for mask != 0 {
				count += int(mask & 1)
				mask >>= 1
			}
			return count
		}
		expectedRoots := make([]GiaID, 0)
		for id, obj := range g.objects {
			if !obj.IsAnd() || countSupport(supportMasks[id]) > sliceLimit {
				continue
			}
			for _, consumer := range consumers[GiaID(id)] {
				if g.objects[consumer].IsCO() || countSupport(supportMasks[consumer]) > sliceLimit {
					expectedRoots = append(expectedRoots, GiaID(id))
					break
				}
			}
		}
		if !slices.Equal(sliced.Roots, expectedRoots) {
			t.Fatal("independent slice boundaries", sliced.Roots, expectedRoots)
		}
		for assignment := 0; assignment < 16; assignment++ {
			values, err := sliced.Graph.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0})
			if err != nil {
				t.Fatal(err)
			}
			for output, id := range sliced.Roots {
				if values[output] != allTruth[id][assignment] {
					t.Fatal("independent slice truth", id, assignment)
				}
			}
		}
		listInputs := slices.Clone(g.cis)
		slices.Reverse(listInputs)
		listGates := make([]GiaID, 0)
		for id, obj := range g.objects {
			if obj.IsAnd() {
				listGates = append(listGates, GiaID(id))
			}
		}
		listed, _, listErr := g.DuplicateFromLists(listInputs, listGates, g.cos, 0, GiaCofactorOptions{})
		if listErr != nil {
			t.Fatal(listErr)
		}
		for assignment := 0; assignment < 16; assignment++ {
			values, err := listed.Evaluate([]bool{assignment&8 != 0, assignment&4 != 0, assignment&2 != 0, assignment&1 != 0})
			if err != nil || values[0] != truth[assignment] || values[1] != extraTruth[assignment] {
				t.Fatal("independent list duplication", assignment, err)
			}
		}
		levelized, _, levelErr := g.DuplicateLevelized(GiaCofactorOptions{})
		if levelErr != nil {
			t.Fatal(levelErr)
		}
		for assignment := 0; assignment < 16; assignment++ {
			values, err := levelized.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0})
			if err != nil || values[0] != truth[assignment] || values[1] != extraTruth[assignment] {
				t.Fatal("independent levelized duplication", assignment, err)
			}
		}
		for assignment := 0; assignment < 16; assignment++ {
			projected := make([]bool, 0, len(limited.Inputs))
			for _, id := range limited.Inputs {
				projected = append(projected, allTruth[id][assignment])
			}
			values, err := limited.Graph.Evaluate(projected)
			if err != nil || values[0] != truth[assignment] || values[1] != allTruth[5][assignment] {
				t.Fatal("independent limited cone projection", assignment, err)
			}
		}
		extracted, err := g.ExtractGateCones([]GiaID{root, 1}, get(0)&1 != 0, GiaCofactorOptions{MaxObjects: 4096, MaxMappingEntries: 4096})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 16; assignment++ {
			projected := make([]bool, 0, len(extracted.PrimaryInputs))
			for _, pi := range extracted.PrimaryInputs {
				projected = append(projected, assignment&(1<<pi) != 0)
			}
			values, err := extracted.Graph.Evaluate(projected)
			if err != nil || values[0] != truth[assignment] || values[1] != (assignment&1 != 0) {
				t.Fatal("independent gate-root extraction", assignment, err)
			}
		}
		guarded, err := g.CofactorObjectGuarded(guardPivot, guardFixed, GiaCofactorOptions{MaxObjects: 4096, MaxMappingEntries: 4096})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 16; assignment++ {
			values, err := guarded.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0})
			allowed := guardTruth[assignment] == guardFixed
			if err != nil || values[0] != (truth[assignment] && allowed) || values[1] != (extraTruth[assignment] && allowed) {
				t.Fatal("independent guarded cofactor", assignment, guardPivot, guardFixed, err)
			}
		}
		rebuilt, err := g.CofactorRebuildAll([]GiaID{root, 5, 1}, GiaMultiCofactorOptions{GiaCofactorOptions: GiaCofactorOptions{MaxObjects: 4096, MaxMappingEntries: 4096}})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 16; assignment++ {
			values, err := rebuilt.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0})
			if err != nil || values[0] != truth[assignment] || values[1] != extraTruth[assignment] {
				t.Fatal("independent multi-pivot reconstruction", assignment, err)
			}
		}
		piIndex := int(get(0) % 4)
		fixed := get(1)&1 != 0
		cofactor, err := g.CofactorPrimaryInput(piIndex, fixed, GiaCofactorOptions{MaxObjects: 4096, MaxMappingEntries: 4096})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 16; assignment++ {
			forced := assignment &^ (1 << piIndex)
			if fixed {
				forced |= 1 << piIndex
			}
			values, err := cofactor.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0})
			if err != nil || values[0] != truth[forced] || values[1] != extraTruth[forced] {
				t.Fatal("independent fixed PI cofactor", assignment, piIndex, fixed, err)
			}
		}
		variables := int(get(1) % 5)
		derived, err := g.DeriveDominatedCone(root, []GiaID{1, 2, 3, 4}, variables, 4096)
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 16; assignment++ {
			values, err := derived.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0})
			if err != nil {
				t.Fatal(err)
			}
			mask := (1 << variables) - 1
			for cofactor, value := range values {
				if value != truth[(assignment&^mask)|cofactor] {
					t.Fatal("independent cofactor", assignment, cofactor, variables)
				}
			}
		}
	})
}
