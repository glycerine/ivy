package abc

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func TestGiaCutEnumerationNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_cut_enumerate.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Mode, Size, Limit, Compute, Minimize int
		Statistics                           []uint64
		Objects                              [][3]int
		Cuts                                 [][]struct {
			Function, Tree int
			Sign           uint64
			Leaves         []GiaID
		}
		Truths [][]uint64
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 400 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		source := NewGiaManager()
		for _, object := range row.Objects[1:] {
			var e error
			switch object[0] {
			case 1:
				_, e = source.AppendCI()
			case 2:
				_, e = source.AppendCO(GiaLit(object[1]))
			case 3:
				_, e = source.AppendAnd(GiaLit(object[1]), GiaLit(object[2]))
			case 4:
				_, e = source.AppendXorReal(GiaLit(object[1]), GiaLit(object[2]))
			}
			if e != nil {
				t.Fatal(e)
			}
		}
		before := slices.Clone(source.objects)
		r, e := source.EnumerateCuts(GiaCutEnumerationOptions{CutSize: row.Size, CutLimit: row.Limit, ComputeTruth: row.Compute != 0, MinimizeSupport: row.Minimize != 0})
		if e != nil {
			t.Fatal(index, e)
		}
		store, e := NewGiaCutStore(source, GiaCutEnumerationOptions{CutSize: row.Size, CutLimit: row.Limit, ComputeTruth: row.Compute != 0, MinimizeSupport: row.Minimize != 0})
		if e != nil {
			t.Fatal(e)
		}
		for step := 0; step < index%7; step++ {
			if _, e := store.Advance(); e != nil {
				t.Fatal(index, step, e)
			}
		}
		incremental, e := store.Finish()
		if e != nil {
			t.Fatal(index, e)
		}
		assertGiaCutEnumerationsEqual(t, r, incremental)
		stats := []uint64{r.Statistics.Pairs, r.Statistics.Merges, r.Statistics.Evaluations, r.Statistics.Cuts, r.Statistics.SaturatedNodes}
		if !slices.Equal(stats, row.Statistics) || len(r.Cuts) != len(row.Cuts) || r.TruthCount() != len(row.Truths) {
			t.Fatal("enumeration metadata", index, stats, row.Statistics, r.TruthCount(), len(row.Truths))
		}
		for id, list := range row.Cuts {
			if len(r.Cuts[id]) != len(list) {
				t.Fatal("cut count", index, id, len(r.Cuts[id]), len(list))
			}
			for j, expected := range list {
				got := r.Cuts[id][j]
				if !slices.Equal(got.Leaves, expected.Leaves) || got.Function != expected.Function || got.TreeLeaves != expected.Tree || got.Signature != expected.Sign {
					t.Fatal("ranked cut", index, id, j, got, expected)
				}
			}
		}
		for id, expected := range row.Truths {
			got, e := r.truths.Read(id)
			if e != nil || !slices.Equal(got, expected) {
				t.Fatal("truth IDs", index, id, got, expected, e)
			}
		}
		if !slices.Equal(source.objects, before) || source.Hashing() {
			t.Fatal("source changed")
		}
	}
}
func TestGiaCutEnumerationLimitsAndOwnership(t *testing.T) {
	source := NewGiaManager()
	a, _ := source.AppendCI()
	b, _ := source.AppendCI()
	c, _ := source.AppendCI()
	n, _ := source.AppendAnd(a, b)
	root, _ := source.AppendXorReal(n, c)
	source.AppendCO(root)
	before := slices.Clone(source.objects)
	for _, options := range []GiaCutEnumerationOptions{
		{MaxObjects: source.ObjectCount() - 1}, {MaxStoredCuts: 1}, {MaxPairs: 1},
		{ComputeTruth: true, MaxTruthEntries: 1}, {ComputeTruth: true, MaxTruthEntries: 2}, {ComputeTruth: true, MaxTruthWords: 2},
	} {
		if r, e := source.EnumerateCuts(options); r != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(options, r, e)
		}
	}
	for _, options := range []GiaCutEnumerationOptions{{CutSize: 1}, {CutSize: 9}, {CutLimit: 1}, {CutLimit: 51}, {MaxObjects: -1}, {MaxPairs: -1}, {MaxStoredCuts: -1}, {MaxTruthWords: -1}, {MaxTruthEntries: -1}} {
		if r, e := source.EnumerateCuts(options); r != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(options, r, e)
		}
	}
	if r, e := (*GiaManager)(nil).EnumerateCuts(GiaCutEnumerationOptions{}); r != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	for _, when := range []int{1, 4, 7} {
		calls := 0
		r, e := source.EnumerateCuts(GiaCutEnumerationOptions{ComputeTruth: true, Stop: func() bool { calls++; return calls == when }})
		if r != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(when, r, e)
		}
	}
	options := GiaCutEnumerationOptions{CutSize: 8, CutLimit: 4, ComputeTruth: true, MinimizeSupport: true}
	r, e := source.EnumerateCuts(options)
	if e != nil {
		t.Fatal(e)
	}
	cut := r.Cuts[root.ID()][0]
	truth, e := r.TruthTable(cut)
	if e != nil {
		t.Fatal(e)
	}
	saved := truth.Words()
	truth.Set(0, true)
	fresh, e := r.TruthTable(cut)
	if e != nil || !slices.Equal(fresh.Words(), saved) {
		t.Fatal("truth result alias", e)
	}
	r.Cuts[root.ID()][0].Leaves[0] = InvalidGiaID
	again, e := source.EnumerateCuts(options)
	if e != nil || again.Cuts[root.ID()][0].Leaves[0] == InvalidGiaID || !slices.Equal(source.objects, before) {
		t.Fatal("enumeration ownership", e)
	}
	plain, e := source.EnumerateCuts(GiaCutEnumerationOptions{})
	if e != nil || plain.TruthCount() != 0 {
		t.Fatal(e)
	}
	if _, e := plain.TruthTable(plain.Cuts[n.ID()][0]); !errors.Is(e, ErrInvariant) {
		t.Fatal(e)
	}
	buffer := source.Clone()
	buffer.AppendBuffer(a)
	if r, e := buffer.EnumerateCuts(options); r != nil || !errors.Is(e, ErrUnsupported) {
		t.Fatal("native buffer precondition", r, e)
	}
	mux := source.Clone()
	mux.AppendMuxReal(c, a, b)
	if r, e := mux.EnumerateCuts(options); r != nil || !errors.Is(e, ErrUnsupported) {
		t.Fatal("native mux precondition", r, e)
	}
}
func FuzzGiaCutEnumeration(f *testing.F) {
	f.Add(uint8(4), []byte{1, 2, 3, 4, 5})
	regressionBytes, e := os.ReadFile("test_vectors/gia_cut_enumeration_regression.json")
	if e != nil {
		f.Fatal(e)
	}
	var regression struct {
		Flags uint8
		Data  string
	}
	if e = json.Unmarshal(regressionBytes, &regression); e != nil {
		f.Fatal(e)
	}
	f.Add(regression.Flags, []byte(regression.Data))
	f.Add(uint8(255), []byte{255, 0, 17, 42})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 20 {
			data = data[:20]
		}
		m := buildGiaCutEnumerationTestGraph(t, data)
		before := slices.Clone(m.objects)
		options := GiaCutEnumerationOptions{CutSize: 2 + int(flags%7), CutLimit: 2 + int((flags/7)%8), ComputeTruth: flags&128 != 0, MinimizeSupport: flags&64 != 0}
		r, e := m.EnumerateCuts(options)
		if e != nil {
			t.Fatal(e)
		}
		store, e := NewGiaCutStore(m, options)
		if e != nil {
			t.Fatal(e)
		}
		for step := uint8(0); step < flags%6; step++ {
			if _, e := store.Advance(); e != nil {
				t.Fatal(e)
			}
		}
		resumed, e := store.Finish()
		if e != nil {
			t.Fatal(e)
		}
		assertGiaCutEnumerationsEqual(t, r, resumed)
		for id, cuts := range r.Cuts {
			if len(cuts) > r.CutLimit {
				t.Fatal("cut cap")
			}
			for i, cut := range cuts {
				if len(cut.Leaves) > r.CutSize || !slices.IsSorted(cut.Leaves) || cut.Signature != giaCutSignature(cut.Leaves) {
					t.Fatal("cut shape", id, cut)
				}
				for j := 1; j < len(cut.Leaves); j++ {
					if cut.Leaves[j] == cut.Leaves[j-1] {
						t.Fatal("duplicate leaf")
					}
				}
				for j, other := range cuts {
					if i == j || len(cut.Leaves) == 1 && cut.Leaves[0] == GiaID(id) || len(other.Leaves) == 1 && other.Leaves[0] == GiaID(id) {
						continue
					}
					if giaCutContains(cut.Leaves, other.Leaves) {
						t.Fatal("dominated cuts", id, cuts)
					}
				}
			}
		}
		if options.ComputeTruth {
			for id, cuts := range r.Cuts {
				for _, cut := range cuts {
					allCIs := true
					for _, leaf := range cut.Leaves {
						if !m.objects[leaf].IsCI() {
							allCIs = false
							break
						}
					}
					if !allCIs {
						continue
					} // Internal functional-cut slots need not be independent.
					actual, e := r.TruthTable(cut)
					if e != nil {
						t.Fatal(e)
					}
					expected, e := m.CutTruth(GiaLit(id)<<1, cut.Leaves, GiaCutTruthOptions{})
					if e == nil {
						if !actual.Equal(expected.Table) {
							t.Fatal("free CI-cut truth", id, cut, actual.Words(), expected.Table.Words())
						}
					} else if !options.MinimizeSupport || !errors.Is(e, ErrDimension) {
						t.Fatal("invalid cut boundary", id, cut, e)
					}
				}
			}
			for assignment := 0; assignment < 16; assignment++ {
				ci := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
				values, e := m.evaluateValues(ci)
				if e != nil {
					t.Fatal(e)
				}
				for id, cuts := range r.Cuts {
					for _, cut := range cuts {
						truth, e := r.TruthTable(cut)
						if e != nil {
							t.Fatal(e)
						}
						local := 0
						for i, leaf := range cut.Leaves {
							if values[leaf] {
								local |= 1 << uint(i)
							}
						}
						got, e := truth.Get(uint64(local))
						if e != nil || got != values[id] {
							t.Fatal("cut function", assignment, id, cut, truth.Words(), e)
						}
					}
				}
			}
		} else {
			for id, cuts := range r.Cuts {
				for _, cut := range cuts {
					boundary := make([]bool, m.ObjectCount())
					boundary[0] = true
					for _, leaf := range cut.Leaves {
						boundary[leaf] = true
					}
					stack := []GiaID{GiaID(id)}
					for len(stack) > 0 {
						node := stack[len(stack)-1]
						stack = stack[:len(stack)-1]
						if boundary[node] {
							continue
						}
						if m.objects[node].IsCI() {
							t.Fatal("uncovered structural cut")
						}
						children, n, e := m.crossCutFanins(node)
						if e != nil {
							t.Fatal(e)
						}
						stack = append(stack, children[:n]...)
					}
				}
			}
		}
		if !slices.Equal(m.objects, before) {
			t.Fatal("source changed")
		}
	})
}

func buildGiaCutEnumerationTestGraph(t *testing.T, data []byte) *GiaManager {
	t.Helper()
	m := NewGiaManager()
	pool := []GiaLit{GiaConst0, GiaConst1}
	for i := 0; i < 4; i++ {
		lit, _ := m.AppendCI()
		pool = append(pool, lit)
	}
	for i, value := range data {
		a := pool[int(value)%len(pool)]
		b := pool[int(value>>2)%len(pool)]
		if a.ID() == b.ID() {
			b = pool[(int(value)%len(pool)+2)%len(pool)]
			if a.ID() == b.ID() {
				b = pool[2]
			}
			if a.ID() == b.ID() {
				b = pool[3]
			}
		}
		var lit GiaLit
		var e error
		if value&1 != 0 {
			lit, e = m.AppendXorReal(a.NotIf(value&16 != 0), b.NotIf(value&32 != 0))
		} else {
			lit, e = m.AppendAnd(a.NotIf(value&16 != 0), b.NotIf(value&32 != 0))
		}
		if e != nil {
			t.Fatal(i, e)
		}
		pool = append(pool, lit)
	}
	m.AppendCO(pool[len(pool)-1])

	return m
}

func TestGiaCutEnumerationCorrelatedFunctionalRegression(t *testing.T) {
	data, e := os.ReadFile("test_vectors/gia_cut_enumeration_regression.json")
	if e != nil {
		t.Fatal(e)
	}
	var regression struct {
		Flags                       uint8
		Data                        string
		Node                        GiaID
		Leaves                      []GiaID
		Function, GlobalAssignments int
		Truth, StructuralTruth      uint64
	}
	if e = json.Unmarshal(data, &regression); e != nil {
		t.Fatal(e)
	}
	m := buildGiaCutEnumerationTestGraph(t, []byte(regression.Data))
	r, e := m.EnumerateCuts(GiaCutEnumerationOptions{CutSize: 3, CutLimit: 3, ComputeTruth: true, MinimizeSupport: true})
	if e != nil {
		t.Fatal(e)
	}
	var truth *TruthTable
	for _, cut := range r.Cuts[regression.Node] {
		if slices.Equal(cut.Leaves, regression.Leaves) {
			if cut.Function != regression.Function {
				t.Fatal("native function ID", cut)
			}
			truth, e = r.TruthTable(cut)
			if e != nil {
				t.Fatal(e)
			}
			break
		}
	}
	if truth == nil || truth.words[0] != regression.Truth {
		t.Fatal("native functional truth", truth)
	}
	structural, e := m.CutTruth(GiaLit(regression.Node)<<1, regression.Leaves, GiaCutTruthOptions{})
	if e != nil || structural.Table.words[0] != regression.StructuralTruth || structural.Table.Equal(truth) {
		t.Fatal("functional/structural distinction", structural, e)
	}
	reached := uint8(0)
	for assignment := 0; assignment < regression.GlobalAssignments; assignment++ {
		values, e := m.evaluateValues([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0})
		if e != nil {
			t.Fatal(e)
		}
		local := 0
		for i, leaf := range regression.Leaves {
			if values[leaf] {
				local |= 1 << uint(i)
			}
		}
		reached |= 1 << uint(local)
		actual, _ := truth.Get(uint64(local))
		other, _ := structural.Table.Get(uint64(local))
		if actual != values[regression.Node] || other != actual {
			t.Fatal("realizable assignment mismatch", assignment, local)
		}
	}
	if reached != 7 {
		t.Fatal("expected only local assignments 00,01,10", reached)
	}
	copy := *r
	copy.CutSize = 100
	if _, e := copy.TruthTable(GiaCut{Function: 2, Leaves: make([]GiaID, 9)}); !errors.Is(e, ErrDimension) {
		t.Fatal("invalid truth metadata", e)
	}
}
func TestGiaCutEnumerationMutatedTruthWidth(t *testing.T) {
	m := buildGiaCutEnumerationTestGraph(t, []byte{1, 2, 3, 4})
	for _, size := range []int{2, 7, 8} {
		original, err := m.EnumerateCuts(GiaCutEnumerationOptions{CutSize: size, ComputeTruth: true})
		if err != nil {
			t.Fatal(err)
		}
		for _, changed := range []int{2, 7, 8} {
			if changed == size {
				continue
			}
			mutated := *original
			mutated.CutSize = changed
			// Public metadata must never reinterpret a differently sized truth record.
			for _, phase := range []int{0, 1} {
				if _, err := mutated.TruthTable(GiaCut{Function: 2 + phase, Leaves: make([]GiaID, changed)}); !errors.Is(err, ErrDimension) {
					t.Fatal("mutated storage width", size, changed, phase, err)
				}
			}
		}
		if _, err := original.TruthTable(original.Cuts[m.cis[0]][0]); err != nil {
			t.Fatal("original damaged", err)
		}
	}
}

func TestGiaCutEnumerationEightInputFunctions(t *testing.T) {
	m := NewGiaManager()
	inputs := make([]GiaLit, 8)
	for i := range inputs {
		inputs[i], _ = m.AppendCI()
	}
	root := inputs[0]
	for i := 1; i < len(inputs); i++ {
		root, _ = m.AppendAnd(root.NotIf(i%3 == 0), inputs[i].NotIf(i&1 != 0))
	}
	m.AppendCO(root)
	for _, minimize := range []bool{false, true} {
		r, e := m.EnumerateCuts(GiaCutEnumerationOptions{CutSize: 8, ComputeTruth: true, MinimizeSupport: minimize})
		if e != nil {
			t.Fatal(e)
		}
		for assignment := 0; assignment < 256; assignment++ {
			ci := make([]bool, 8)
			for i := range ci {
				ci[i] = assignment&(1<<uint(i)) != 0
			}
			values, e := m.evaluateValues(ci)
			if e != nil {
				t.Fatal(e)
			}
			for id, cuts := range r.Cuts {
				for _, cut := range cuts {
					truth, e := r.TruthTable(cut)
					if e != nil {
						t.Fatal(e)
					}
					local := 0
					for i, leaf := range cut.Leaves {
						if values[leaf] {
							local |= 1 << uint(i)
						}
					}
					got, e := truth.Get(uint64(local))
					if e != nil || got != values[id] {
						t.Fatal("multiword cut function", minimize, id, cut, assignment, e)
					}
				}
			}
		}
	}
}

func assertGiaCutEnumerationsEqual(t *testing.T, a, b *GiaCutEnumeration) {
	t.Helper()
	if a.CutSize != b.CutSize || a.CutLimit != b.CutLimit || a.Statistics != b.Statistics || !reflect.DeepEqual(a.Cuts, b.Cuts) || a.TruthCount() != b.TruthCount() {
		t.Fatal("store/batch results differ")
	}
	for id := 0; id < a.TruthCount(); id++ {
		left, err := a.truths.Read(id)
		if err != nil {
			t.Fatal(err)
		}
		right, err := b.truths.Read(id)
		if err != nil || !slices.Equal(left, right) {
			t.Fatal("store truth IDs differ", id, err)
		}
	}
}

func TestGiaCutStoreLifecycle(t *testing.T) {
	source := buildGiaCutEnumerationTestGraph(t, []byte{1, 2, 3, 4, 5, 6})
	options := GiaCutEnumerationOptions{ComputeTruth: true, MinimizeSupport: true}
	expected, err := source.EnumerateCuts(options)
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewGiaCutStore(source, options)
	if err != nil {
		t.Fatal(err)
	}
	// Source mutations after construction must not reach the store snapshot.
	source.AppendCI()
	source.objects[1].value = 123
	previous := 1
	for {
		done, err := store.Advance()
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for id := previous; id < store.next; id++ {
			if store.source.objects[id].IsAnd() {
				count++
			}
		}
		if count > 1 {
			t.Fatal("advance computed multiple gates", count)
		}
		previous = store.next
		if done {
			break
		}
	}
	if done, err := store.Advance(); !done || err != nil {
		t.Fatal("repeated completion", done, err)
	}
	actual, err := store.Finish()
	if err != nil {
		t.Fatal(err)
	}
	assertGiaCutEnumerationsEqual(t, expected, actual)
	if again, err := store.Finish(); again != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal("double transfer", again, err)
	}
	if store.source != nil || store.result != nil || store.references != nil || store.options.Stop != nil {
		t.Fatal("retained completed storage")
	}
	store.Close()
	var missing *GiaCutStore
	missing.Close()
	if _, err := missing.Advance(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := missing.Finish(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewGiaCutStore(nil, options); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	var zero GiaCutStore
	if _, err := zero.Advance(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	empty, err := NewGiaCutStore(NewGiaManager(), options)
	if err != nil {
		t.Fatal(err)
	}
	if done, err := empty.Advance(); !done || err != nil {
		t.Fatal(done, err)
	}
	if result, err := empty.Finish(); err != nil || len(result.Cuts) != 1 {
		t.Fatal(result, err)
	}
}

func TestGiaCutStoreFailureCloses(t *testing.T) {
	source := NewGiaManager()
	a, _ := source.AppendCI()
	b, _ := source.AppendCI()
	c, _ := source.AppendCI()
	d, _ := source.AppendCI()
	left, _ := source.AppendAnd(a, b)
	right, _ := source.AppendXorReal(c, d)
	root, _ := source.AppendAnd(left, right)
	source.AppendCO(root)
	for _, options := range []GiaCutEnumerationOptions{
		{MaxPairs: 1}, {MaxStoredCuts: 5}, {ComputeTruth: true, MaxTruthEntries: 2},
	} {
		store, err := NewGiaCutStore(source, options)
		if err != nil {
			t.Fatal("expected gate-time failure", err)
		}
		if result, err := store.Finish(); result != nil || !errors.Is(err, ErrLimit) {
			t.Fatal(result, err)
		}
		if _, err := store.Advance(); !errors.Is(err, ErrLimit) {
			t.Fatal("lost failure", err)
		}
		if store.source != nil || store.result != nil || store.references != nil {
			t.Fatal("retained failed storage")
		}
	}
	stop := false
	store, err := NewGiaCutStore(source, GiaCutEnumerationOptions{Stop: func() bool { return stop }})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Advance(); err != nil {
		t.Fatal(err)
	}
	stop = true
	if result, err := store.Finish(); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	stop = false
	if _, err := store.Advance(); !errors.Is(err, ErrCancelled) {
		t.Fatal("cancelled store restarted", err)
	}
	abandoned, err := NewGiaCutStore(source, GiaCutEnumerationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	abandoned.Close()
	abandoned.Close()
	if result, err := abandoned.Finish(); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
}

func TestGiaCutStoreMidNodeCancellation(t *testing.T) {
	source := buildGiaCutEnumerationTestGraph(t, []byte{1, 2, 3, 4, 5, 6})
	calls, active := 0, false
	store, err := NewGiaCutStore(source, GiaCutEnumerationOptions{ComputeTruth: true, Stop: func() bool {
		if !active {
			return false
		}
		calls++
		return calls == 2
	}})
	if err != nil {
		t.Fatal(err)
	}
	active = true
	if _, err := store.Advance(); !errors.Is(err, ErrCancelled) {
		t.Fatal("mid-node cancellation", calls, err)
	}
	if store.result != nil || store.source != nil || store.references != nil || store.options.Stop != nil {
		t.Fatal("failed store retained storage")
	}
}
