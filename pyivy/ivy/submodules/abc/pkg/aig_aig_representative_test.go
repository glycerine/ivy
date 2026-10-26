package abc

// Standalone tests for src/aig/aig/aigRepr.c representative tables.

import (
	"errors"
	"slices"
	"sync"
	"testing"
)

func aigRepresentativeFixture(t testing.TB) (*AigManager, []AigID) {
	t.Helper()
	graph := NewAigManager()
	a, _ := graph.CreateCI()
	b, _ := graph.CreateCI()
	first, err := graph.And(a, b)
	if err != nil {
		t.Fatal(err)
	}
	second, err := graph.And(a.Not(), b.Not())
	if err != nil {
		t.Fatal(err)
	}
	return graph, []AigID{a.ID(), b.ID(), first.ID(), second.ID()}
}

func TestAigCompiledSourceInventory(t *testing.T) {
	got := AigCompiledSourceInventory()
	want := []string{
		"aigCheck.c", "aigCanon.c", "aigCuts.c", "aigDfs.c", "aigDup.c",
		"aigFanout.c", "aigFrames.c", "aigInter.c", "aigJust.c", "aigMan.c",
		"aigMem.c", "aigMffc.c", "aigObj.c", "aigOper.c", "aigOrder.c",
		"aigPack.c", "aigPart.c", "aigPartReg.c", "aigPartSat.c", "aigRepr.c",
		"aigRet.c", "aigRetF.c", "aigScl.c", "aigShow.c", "aigSplit.c",
		"aigTable.c", "aigTiming.c", "aigTruth.c", "aigTsim.c", "aigUtil.c",
		"aigWin.c",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("inventory=%v", got)
	}
	got[0] = "changed"
	if AigCompiledSourceInventory()[0] != "aigCheck.c" {
		t.Fatal("inventory aliases caller")
	}
}

func TestAigRepresentativesDirectTransitiveClearAndOwnership(t *testing.T) {
	graph, ids := aigRepresentativeFixture(t)
	table, err := NewAigRepresentatives(graph, AigRepresentativeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err = table.SetRepresentative(ids[0], ids[2]); err != nil {
		t.Fatal(err)
	}
	if err = table.SetRepresentative(ids[2], ids[3]); err != nil {
		t.Fatal(err)
	}
	if direct, ok, err := table.Representative(ids[3]); err != nil || !ok || direct != ids[2] {
		t.Fatalf("direct=%d ok=%v err=%v", direct, ok, err)
	}
	if transitive, ok, err := table.TransitiveRepresentative(ids[3]); err != nil || !ok || transitive != ids[0] {
		t.Fatalf("transitive=%d ok=%v err=%v", transitive, ok, err)
	}
	if table.Count() != 2 {
		t.Fatalf("count=%d", table.Count())
	}
	snapshot := table.Snapshot()
	snapshot[ids[3]] = InvalidAigID
	if table.Count() != 2 {
		t.Fatal("snapshot aliases table")
	}
	if err = table.Clear(ids[3]); err != nil || table.Count() != 1 {
		t.Fatalf("clear count=%d err=%v", table.Count(), err)
	}
}

func TestAigRepresentativesResolveMappedPhase(t *testing.T) {
	graph, ids := aigRepresentativeFixture(t)
	table, err := NewAigRepresentatives(graph, AigRepresentativeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	low, high := ids[2], ids[3]
	if low > high {
		low, high = high, low
	}
	if err = table.SetRepresentative(low, high); err != nil {
		t.Fatal(err)
	}
	mapped := make([]AigLit, graph.ObjectCount())
	for id := range mapped {
		mapped[id] = AigLit(id) << 1
	}
	phase := graph.objects[high].Phase != graph.objects[low].Phase
	if !phase {
		t.Fatal("fixture must exercise representative phase correction")
	}
	for _, complemented := range []bool{false, true} {
		source := (AigLit(high) << 1).NotIf(complemented)
		got, resolveErr := table.ResolveMapped(source, mapped)
		if resolveErr != nil {
			t.Fatal(resolveErr)
		}
		want := mapped[low].NotIf(phase != complemented)
		if got != want {
			t.Fatalf("complemented=%v got=%v want=%v", complemented, got, want)
		}
	}
	for _, complemented := range []bool{false, true} {
		source := (AigLit(ids[0]) << 1).NotIf(complemented)
		got, resolveErr := table.ResolveMapped(source, mapped)
		if resolveErr != nil {
			t.Fatal(resolveErr)
		}
		if want := mapped[ids[0]].NotIf(complemented); got != want {
			t.Fatalf("unrepresented complemented=%v got=%v want=%v", complemented, got, want)
		}
	}
}

func TestAigRepresentativesLimitsErrorsAndAtomicity(t *testing.T) {
	graph, ids := aigRepresentativeFixture(t)
	for _, options := range []AigRepresentativeOptions{{MaxObjects: 1}, {MaxWork: 1}, {MaxBytes: 1}, {Stop: func() bool { return true }}} {
		if got, err := NewAigRepresentatives(graph, options); err == nil || got != nil {
			t.Fatalf("options=%+v got=%v err=%v", options, got, err)
		}
	}
	table, err := NewAigRepresentatives(graph, AigRepresentativeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]AigID{{ids[2], ids[0]}, {ids[0], ids[0]}, {ids[0], 999}} {
		before := table.Snapshot()
		if err = table.SetRepresentative(pair[0], pair[1]); err == nil || !slices.Equal(before, table.Snapshot()) {
			t.Fatalf("pair=%v err=%v", pair, err)
		}
	}
	if err = table.SetLowerPair(ids[3], ids[2]); err != nil {
		t.Fatal(err)
	}
	if representative, ok, err := table.Representative(ids[3]); err != nil || !ok || representative != ids[2] {
		t.Fatalf("representative=%d ok=%v err=%v", representative, ok, err)
	}
	if _, _, err = table.Representative(999); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("invalid err=%v", err)
	}
}

func TestAigRepresentativesConcurrentReaders(t *testing.T) {
	graph, ids := aigRepresentativeFixture(t)
	table, err := NewAigRepresentatives(graph, AigRepresentativeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err = table.SetRepresentative(ids[0], ids[2]); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for iteration := 0; iteration < 100; iteration++ {
				representative, ok, readErr := table.TransitiveRepresentative(ids[2])
				if readErr != nil || !ok || representative != ids[0] || table.Count() != 1 {
					t.Errorf("representative=%d ok=%v err=%v", representative, ok, readErr)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func FuzzAigRepresentatives(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4})
	f.Add([]byte{9, 0, 7, 2, 5})
	f.Fuzz(func(t *testing.T, operations []byte) {
		if len(operations) > 64 {
			t.Skip()
		}
		graph := NewAigManager()
		for index := 0; index < 8; index++ {
			if _, err := graph.CreateCI(); err != nil {
				t.Fatal(err)
			}
		}
		table, err := NewAigRepresentatives(graph, AigRepresentativeOptions{MaxWork: 1024})
		if err != nil {
			t.Fatal(err)
		}
		model := make([]AigID, graph.ObjectCount())
		for index := range model {
			model[index] = InvalidAigID
		}
		for _, operation := range operations {
			object := AigID(int(operation)%8 + 1)
			if operation&1 == 0 {
				representative := AigID(int(operation/8) % int(object))
				if err = table.SetRepresentative(representative, object); err != nil {
					t.Fatal(err)
				}
				model[object] = representative
			} else {
				if err = table.Clear(object); err != nil {
					t.Fatal(err)
				}
				model[object] = InvalidAigID
			}
		}
		if got := table.Snapshot(); !slices.Equal(got, model) {
			t.Fatalf("got=%v want=%v", got, model)
		}
	})
}
