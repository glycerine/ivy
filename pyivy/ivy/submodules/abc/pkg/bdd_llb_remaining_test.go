package abc

// Standalone tests for the LLB incidence-matrix, scheduling, and clustering
// semantics represented by llb1Matrix.c, llb1Sched.c, and llb1Cluster.c.

import (
	"errors"
	"slices"
	"sync"
	"testing"
)

func llbTestSpec() LLBMatrixSpec {
	return LLBMatrixSpec{
		PrimaryInputs: 1,
		FlipFlops:     1,
		Rows:          5,
		Columns: []LLBPartition{
			{ID: 0, Variables: []int{1}},
			{ID: 10, Variables: []int{0, 1, 2}},
			{ID: 11, Variables: []int{2, 3}},
			{ID: 12, Variables: []int{1, 3, 4}},
			{ID: 99, Variables: []int{4}},
		},
	}
}

func TestLLBCompiledSourceInventory(t *testing.T) {
	want := []string{
		"llb1Cluster.c", "llb1Constr.c", "llb1Core.c", "llb1Group.c", "llb1Hint.c", "llb1Man.c",
		"llb1Matrix.c", "llb1Pivot.c", "llb1Reach.c", "llb1Sched.c", "llb2Bad.c", "llb2Core.c",
		"llb2Driver.c", "llb2Dump.c", "llb2Flow.c", "llb2Image.c", "llb3Image.c", "llb3Nonlin.c",
		"llb4Cex.c", "llb4Image.c", "llb4Nonlin.c", "llb4Sweep.c",
	}
	got := LLBCompiledSources()
	if !slices.Equal(got, want) {
		t.Fatalf("inventory mismatch\n got %v\nwant %v", got, want)
	}
	got[0] = "changed"
	if LLBCompiledSources()[0] != want[0] {
		t.Fatal("inventory aliases caller storage")
	}
}

func TestLLBMatrixOwnershipShapeAndVariableOrder(t *testing.T) {
	spec := llbTestSpec()
	m, err := NewLLBIncidenceMatrix(spec, LLBLimits{})
	if err != nil {
		t.Fatal(err)
	}
	spec.Columns[1].Variables[0] = 4
	snapshot, err := m.Snapshot(LLBLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(snapshot.Columns[1].Variables, []int{0, 1, 2}) {
		t.Fatal(snapshot)
	}
	snapshot.Columns[1].Variables[0] = 4
	againSnapshot, err := m.Snapshot(LLBLimits{})
	if err != nil || againSnapshot.Columns[1].Variables[0] != 0 {
		t.Fatal("snapshot aliases matrix")
	}
	order, err := m.VariableOrder(LLBLimits{})
	if err != nil {
		t.Fatal(err)
	}
	// Last uses are: row 0 -> 1, row 2 -> 2, rows 1/3 -> 3, row 4 -> 4.
	if !slices.Equal(order, []int{0, 2, 1, 3, 4}) {
		t.Fatalf("order %v", order)
	}

	invalid := []LLBMatrixSpec{
		{},
		{PrimaryInputs: 0, FlipFlops: 1, Rows: 2, Columns: []LLBPartition{{ID: 0, Variables: []int{0}}, {ID: 1, Variables: []int{1}}}},
		{PrimaryInputs: 1, FlipFlops: 1, Rows: 3, Columns: []LLBPartition{{ID: 0, Variables: []int{0}}, {ID: 1, Variables: []int{2}}}},
		{PrimaryInputs: 0, FlipFlops: 1, Rows: 2, Columns: []LLBPartition{{ID: 0, Variables: []int{0}}, {ID: 0, Variables: []int{0, 1}}, {ID: 2, Variables: []int{1}}}},
		{PrimaryInputs: 0, FlipFlops: 1, Rows: 2, Columns: []LLBPartition{{ID: 0, Variables: []int{0}}, {ID: -1, Variables: []int{0, 1}}, {ID: 2, Variables: []int{1}}}},
		{PrimaryInputs: 0, FlipFlops: 1, Rows: 2, Columns: []LLBPartition{{ID: 0, Variables: []int{0}}, {ID: 1, Variables: []int{0, 0}}, {ID: 2, Variables: []int{1}}}},
		{PrimaryInputs: 0, FlipFlops: 1, Rows: 2, Columns: []LLBPartition{{ID: 0, Variables: []int{0}}, {ID: 1, Variables: []int{0, 2}}, {ID: 2, Variables: []int{1}}}},
	}
	for i, bad := range invalid {
		if _, err := NewLLBIncidenceMatrix(bad, LLBLimits{}); !errors.Is(err, ErrDimension) {
			t.Fatalf("case %d: %v", i, err)
		}
	}
}

func llbModelColumns(spec LLBMatrixSpec) (map[int][]bool, []int) {
	columns := make(map[int][]bool, len(spec.Columns))
	rowSums := make([]int, spec.Rows)
	for _, column := range spec.Columns {
		bits := make([]bool, spec.Rows)
		for _, row := range column.Variables {
			bits[row] = true
			rowSums[row]++
		}
		columns[column.ID] = bits
	}
	return columns, rowSums
}

func llbModelTrueIndices(values []bool) []int {
	result := make([]int, 0)
	for i, value := range values {
		if value {
			result = append(result, i)
		}
	}
	return result
}

func llbModelBest(remainingIDs []int, columns map[int][]bool, rowSums []int, live []bool, remaining []int, nonNextStateRows int) int {
	best, bestWeight := -1, -int(^uint(0)>>1)
	for row := 0; row < nonNextStateRows; row++ {
		if rowSums[row] < 2 || !live[row] || remaining[row] != 1 {
			continue
		}
		candidate := -1
		for _, id := range remainingIDs {
			if columns[id][row] {
				if candidate >= 0 {
					candidate = -2
					break
				}
				candidate = id
			}
		}
		if candidate < 0 {
			continue
		}
		weight := 0
		for k := range live {
			if live[k] && columns[candidate][k] && remaining[k] == 1 {
				weight += 2
			}
			if !live[k] && columns[candidate][k] {
				weight--
			}
		}
		if weight > 0 && weight > bestWeight {
			best, bestWeight = candidate, weight
		}
	}
	if best >= 0 {
		return best
	}
	bestCost, bestSecondary := int(^uint(0)>>1), int(^uint(0)>>1)
	for _, id := range remainingIDs {
		cost, secondary := 0, 0
		for row := range live {
			if !live[row] && columns[id][row] {
				cost++
				secondary += remaining[row]
			}
		}
		if cost < bestCost || (cost == bestCost && secondary > bestSecondary) {
			best, bestCost, bestSecondary = id, cost, secondary
		}
	}
	return best
}

func llbCheckScheduleModel(t testing.TB, spec LLBMatrixSpec, got LLBScheduleResult) {
	t.Helper()
	if len(got.ColumnIDs) != len(spec.Columns) || got.ColumnIDs[0] != spec.Columns[0].ID || got.ColumnIDs[len(got.ColumnIDs)-1] != spec.Columns[len(spec.Columns)-1].ID {
		t.Fatalf("bad endpoints: %v", got.ColumnIDs)
	}
	columns, rowSums := llbModelColumns(spec)
	live := make([]bool, spec.Rows)
	remaining := append([]int(nil), rowSums...)
	for row := spec.PrimaryInputs; row < spec.PrimaryInputs+spec.FlipFlops; row++ {
		live[row] = true
		remaining[row]--
	}
	unscheduled := make([]int, 0, len(spec.Columns)-2)
	for _, column := range spec.Columns[1 : len(spec.Columns)-1] {
		unscheduled = append(unscheduled, column.ID)
	}
	peak := spec.FlipFlops
	for step, chosen := range got.ColumnIDs[1 : len(got.ColumnIDs)-1] {
		want := llbModelBest(unscheduled, columns, rowSums, live, remaining, spec.Rows-spec.FlipFlops)
		if chosen != want {
			t.Fatalf("step %d chose %d want %d", step, chosen, want)
		}
		quantified := make([]int, 0)
		for row, present := range columns[chosen] {
			if !present {
				continue
			}
			if live[row] && remaining[row] == 1 {
				live[row], remaining[row] = false, 0
				quantified = append(quantified, row)
			} else {
				if !live[row] {
					live[row], remaining[row] = true, rowSums[row]
				}
				remaining[row]--
			}
		}
		liveRows := llbModelTrueIndices(live)
		if !slices.Equal(got.LiveAfter[step], liveRows) || !slices.Equal(got.QuantifiedAfter[step], quantified) {
			t.Fatalf("step %d got live=%v quant=%v want live=%v quant=%v", step, got.LiveAfter[step], got.QuantifiedAfter[step], liveRows, quantified)
		}
		if len(liveRows) > peak {
			peak = len(liveRows)
		}
		chosenIndex := -1
		for i, id := range unscheduled {
			if id == chosen {
				chosenIndex = i
				break
			}
		}
		if chosenIndex < 0 {
			t.Fatalf("chosen column %d was not pending", chosen)
		}
		// The source scheduler swaps the selected column into the current
		// position, so the displaced first candidate occupies its old slot.
		unscheduled[0], unscheduled[chosenIndex] = unscheduled[chosenIndex], unscheduled[0]
		unscheduled = unscheduled[1:]
	}
	if peak != got.PeakLive || len(unscheduled) != 0 {
		t.Fatalf("peak=%d want=%d remaining=%v", got.PeakLive, peak, unscheduled)
	}
}

func TestLLBScheduleAgainstIndependentModel(t *testing.T) {
	spec := llbTestSpec()
	m, err := NewLLBIncidenceMatrix(spec, LLBLimits{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.Schedule(LLBLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Work == 0 {
		t.Fatal("missing work accounting")
	}
	llbCheckScheduleModel(t, spec, got)
}

func llbModelCluster(spec LLBMatrixSpec, clusterMax int) []LLBClusterGroup {
	cols := make([][]bool, len(spec.Columns))
	colSums := make([]int, len(cols))
	rowSums := make([]int, spec.Rows)
	members := make([][]int, len(cols))
	active := make([]bool, len(cols))
	for i, column := range spec.Columns {
		cols[i] = make([]bool, spec.Rows)
		for _, row := range column.Variables {
			cols[i][row] = true
			colSums[i]++
			rowSums[row]++
		}
		members[i] = []int{column.ID}
		active[i] = true
	}
	merge := func(i, k int) {
		for row := range rowSums {
			if cols[i][row] && cols[k][row] {
				rowSums[row]--
			}
			if !cols[i][row] && cols[k][row] {
				cols[i][row] = true
				colSums[i]++
			}
		}
		members[i] = append(members[i], members[k]...)
		colSums[k], active[k] = 0, false
	}
	nonNS := spec.Rows - spec.FlipFlops
	for {
		for {
			bestI, bestK, bestScore := -1, -1, -100000
			for i := 1; i < len(cols)-1; i++ {
				if !active[i] || colSums[i] == 0 || colSums[i] > clusterMax {
					continue
				}
				for k := i + 1; k < len(cols)-1; k++ {
					if !active[k] || colSums[k] == 0 || colSums[k] > clusterMax {
						continue
					}
					score := 0
					for row := 0; row < nonNS; row++ {
						a, b := cols[i][row], cols[k][row]
						if a && b && rowSums[row] == 2 {
							score += 2
						} else if a != b {
							score--
						}
					}
					if score > 0 && score > bestScore {
						bestI, bestK, bestScore = i, k, score
					}
				}
			}
			if bestI < 0 {
				break
			}
			merge(bestI, bestK)
		}
		bestI, bestK, bestDiff, bestTotal := -1, -1, 0, 1
		for i := 1; i < len(cols)-1; i++ {
			if !active[i] || colSums[i] == 0 || colSums[i] > clusterMax {
				continue
			}
			for k := i + 1; k < len(cols)-1; k++ {
				if !active[k] || colSums[k] == 0 || colSums[k] > clusterMax {
					continue
				}
				common, different := 0, 0
				for row := 0; row < nonNS; row++ {
					a, b := cols[i][row], cols[k][row]
					if a && b {
						common++
					} else if a != b {
						different++
					}
				}
				total := common + different
				if total != 0 && (bestI < 0 || different*bestTotal < bestDiff*total) {
					bestI, bestK, bestDiff, bestTotal = i, k, different, total
				}
			}
		}
		if bestI < 0 {
			break
		}
		merge(bestI, bestK)
	}
	groups := make([]LLBClusterGroup, 0, len(cols))
	for i := range cols {
		if active[i] {
			groups = append(groups, LLBClusterGroup{PartitionIDs: members[i], Variables: llbModelTrueIndices(cols[i])})
		}
	}
	return groups
}

func TestLLBClusteringDeterministicAndOwned(t *testing.T) {
	spec := LLBMatrixSpec{
		PrimaryInputs: 1, FlipFlops: 1, Rows: 5,
		Columns: []LLBPartition{
			{ID: 0, Variables: []int{1}},
			{ID: 10, Variables: []int{0, 1, 2}},
			{ID: 11, Variables: []int{2, 3}},
			{ID: 12, Variables: []int{1, 3, 4}},
			{ID: 99, Variables: []int{4}},
		},
	}
	m, err := NewLLBIncidenceMatrix(spec, LLBLimits{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.Cluster(3, LLBLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Work == 0 || len(got.Groups) != 4 {
		t.Fatalf("clusters %+v", got)
	}
	found := false
	seen := make(map[int]bool)
	for _, group := range got.Groups {
		for _, id := range group.PartitionIDs {
			if seen[id] {
				t.Fatalf("duplicate partition %d", id)
			}
			seen[id] = true
		}
		if slices.Equal(group.PartitionIDs, []int{11, 12}) {
			found = slices.Equal(group.Variables, []int{1, 2, 3, 4})
		}
	}
	if !found || len(seen) != len(spec.Columns) {
		t.Fatalf("clusters %+v", got.Groups)
	}
	got.Groups[0].Variables[0] = 4
	again, err := m.Cluster(3, LLBLimits{})
	if err != nil || again.Groups[0].Variables[0] != 1 {
		t.Fatal("result aliases matrix", again, err)
	}
}

func TestLLBLimitsCancellationAtomicityAndConcurrency(t *testing.T) {
	spec := llbTestSpec()
	for _, limits := range []LLBLimits{{MaxWork: -1}, {MaxObjects: -1}, {MaxBytes: -1}} {
		if got, err := NewLLBIncidenceMatrix(spec, limits); !errors.Is(err, ErrDimension) || got != nil {
			t.Fatal(got, err)
		}
	}
	if got, err := NewLLBIncidenceMatrix(spec, LLBLimits{MaxBytes: 1}); !errors.Is(err, ErrLimit) || got != nil {
		t.Fatal(got, err)
	}
	if got, err := NewLLBIncidenceMatrix(spec, LLBLimits{MaxObjects: 1}); !errors.Is(err, ErrLimit) || got != nil {
		t.Fatal(got, err)
	}
	if got, err := NewLLBIncidenceMatrix(spec, LLBLimits{MaxObjects: 19}); !errors.Is(err, ErrLimit) || got != nil {
		t.Fatal(got, err)
	}
	if got, err := NewLLBIncidenceMatrix(spec, LLBLimits{MaxObjects: 20}); err != nil || got == nil {
		t.Fatal(got, err)
	}
	if got, err := NewLLBIncidenceMatrix(spec, LLBLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) || got != nil {
		t.Fatal(got, err)
	}
	m, err := NewLLBIncidenceMatrix(spec, LLBLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := m.Snapshot(LLBLimits{MaxObjects: 14}); !errors.Is(err, ErrLimit) || got.Columns != nil {
		t.Fatal(got, err)
	}
	if got, err := m.Snapshot(LLBLimits{MaxObjects: 15}); err != nil || len(got.Columns) != 5 {
		t.Fatal(got, err)
	}
	if got, err := m.VariableOrder(LLBLimits{MaxObjects: 9}); !errors.Is(err, ErrLimit) || got != nil {
		t.Fatal(got, err)
	}
	if got, err := m.VariableOrder(LLBLimits{MaxObjects: 10}); err != nil || len(got) != 5 {
		t.Fatal(got, err)
	}
	if got, err := m.Schedule(LLBLimits{MaxObjects: 94}); !errors.Is(err, ErrLimit) || got.ColumnIDs != nil {
		t.Fatal(got, err)
	}
	if got, err := m.Schedule(LLBLimits{MaxObjects: 95}); err != nil || len(got.ColumnIDs) != 5 {
		t.Fatal(got, err)
	}
	if got, err := m.Cluster(3, LLBLimits{MaxObjects: 84}); !errors.Is(err, ErrLimit) || got.Groups != nil {
		t.Fatal(got, err)
	}
	if got, err := m.Cluster(3, LLBLimits{MaxObjects: 85}); err != nil || len(got.Groups) == 0 {
		t.Fatal(got, err)
	}
	var nilMatrix *LLBIncidenceMatrix
	if _, err := nilMatrix.Snapshot(LLBLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := nilMatrix.VariableOrder(LLBLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := nilMatrix.Schedule(LLBLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := nilMatrix.Cluster(3, LLBLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := m.Cluster(0, LLBLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	for _, limits := range []LLBLimits{{MaxWork: 1}, {MaxObjects: 1}, {MaxBytes: 1}, {Stop: func() bool { return true }}} {
		if got, err := m.Snapshot(limits); err == nil || got.Columns != nil {
			t.Fatal(got, err)
		}
		if got, err := m.Schedule(limits); err == nil || got.ColumnIDs != nil {
			t.Fatal(got, err)
		}
		if got, err := m.VariableOrder(limits); err == nil || got != nil {
			t.Fatal(got, err)
		}
		if got, err := m.Cluster(3, limits); err == nil || got.Groups != nil {
			t.Fatal(got, err)
		}
	}
	calls := 0
	if got, err := m.Schedule(LLBLimits{Stop: func() bool { calls++; return calls > 20 }}); !errors.Is(err, ErrCancelled) || got.ColumnIDs != nil {
		t.Fatal(got, err)
	}

	const readers = 8
	baseline, err := m.Schedule(LLBLimits{})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, readers)
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := m.Schedule(LLBLimits{})
			if err == nil && (!slices.Equal(got.ColumnIDs, baseline.ColumnIDs) || got.PeakLive != baseline.PeakLive) {
				err = ErrDimension
			}
			if err == nil {
				_, err = m.Snapshot(LLBLimits{})
			}
			if err == nil {
				_, err = m.VariableOrder(LLBLimits{})
			}
			if err == nil {
				_, err = m.Cluster(3, LLBLimits{})
			}
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	calls = 0
	if got, err := m.Cluster(3, LLBLimits{Stop: func() bool { calls++; return calls > 20 }}); !errors.Is(err, ErrCancelled) || got.Groups != nil {
		t.Fatal(got, err)
	}
}

func FuzzLLBMatrixSchedule(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7, 8})
	f.Add([]byte{0xff, 0, 0xaa, 0x55, 0x13})
	f.Add([]byte("X0(X("))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		pi := int(data[0] % 2)
		internals := int(data[0]>>1) % 3
		rows := pi + 2 + internals
		interiors := 1 + int(data[0]>>3)%4
		columns := make([]LLBPartition, interiors+2)
		columns[0] = LLBPartition{ID: 0, Variables: []int{pi}}
		for col := 0; col < interiors; col++ {
			vars := make([]int, 0, rows)
			for row := 0; row < rows; row++ {
				index := 1 + (col*rows+row)%max(1, len(data)-1)
				include := len(data) == 1 || data[index%len(data)]&(1<<uint((col+row)&7)) != 0
				if col == 0 && (row < pi || (row >= pi && row < pi+1+internals)) {
					include = true
				}
				if include {
					vars = append(vars, row)
				}
			}
			if len(vars) == 0 {
				vars = append(vars, pi)
			}
			columns[col+1] = LLBPartition{ID: col + 1, Variables: vars}
		}
		columns[len(columns)-1] = LLBPartition{ID: 100, Variables: []int{rows - 1}}
		spec := LLBMatrixSpec{PrimaryInputs: pi, FlipFlops: 1, Rows: rows, Columns: columns}
		m, err := NewLLBIncidenceMatrix(spec, LLBLimits{MaxWork: 100000, MaxBytes: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		got, err := m.Schedule(LLBLimits{MaxWork: 100000, MaxBytes: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		llbCheckScheduleModel(t, spec, got)
		clustered, err := m.Cluster(rows, LLBLimits{MaxWork: 100000, MaxBytes: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		byID := make(map[int][]int, len(columns))
		for _, column := range columns {
			byID[column.ID] = column.Variables
		}
		seen := make(map[int]bool, len(columns))
		for _, group := range clustered.Groups {
			union := make([]bool, rows)
			for _, id := range group.PartitionIDs {
				if seen[id] {
					t.Fatalf("duplicate partition %d", id)
				}
				seen[id] = true
				for _, row := range byID[id] {
					union[row] = true
				}
			}
			if !slices.Equal(group.Variables, llbModelTrueIndices(union)) {
				t.Fatalf("bad cluster union %+v", group)
			}
		}
		if len(seen) != len(columns) {
			t.Fatalf("lost partitions: got %d want %d", len(seen), len(columns))
		}
		wantGroups := llbModelCluster(spec, rows)
		if len(clustered.Groups) != len(wantGroups) {
			t.Fatalf("cluster count %d want %d", len(clustered.Groups), len(wantGroups))
		}
		for i := range wantGroups {
			if !slices.Equal(clustered.Groups[i].PartitionIDs, wantGroups[i].PartitionIDs) || !slices.Equal(clustered.Groups[i].Variables, wantGroups[i].Variables) {
				t.Fatalf("cluster %d = %+v want %+v", i, clustered.Groups[i], wantGroups[i])
			}
		}
	})
}
