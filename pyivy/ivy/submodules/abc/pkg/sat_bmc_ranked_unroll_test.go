package abc

import (
	"errors"
	"slices"
	"testing"
)

func compareRankedAndSimpleUnroll(t testing.TB, source *GiaManager, frames int) {
	t.Helper()
	ranked, err := source.UnrollRanked(frames, GiaRankedUnrollOptions{})
	if err != nil {
		t.Fatal(err)
	}
	simple, err := source.Unroll(frames)
	if err != nil {
		t.Fatal(err)
	}
	if ranked.InputCount() != simple.Graph().InputCount() || ranked.OutputCount() != simple.Graph().OutputCount() {
		t.Fatal("terminal counts differ", ranked.InputCount(), ranked.OutputCount(), simple.Graph().InputCount(), simple.Graph().OutputCount())
	}
	if err = ranked.Check(); err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 1<<ranked.InputCount(); assignment++ {
		inputs := make([]bool, ranked.InputCount())
		for i := range inputs {
			inputs[i] = assignment&(1<<i) != 0
		}
		got, gotErr := ranked.Evaluate(inputs)
		want, wantErr := simple.Graph().Evaluate(inputs)
		if gotErr != nil || wantErr != nil || !slices.Equal(got, want) {
			t.Fatal("ranked/simple semantics differ", assignment, got, want, gotErr, wantErr)
		}
	}
}

func TestGiaRankedUnrollExhaustiveAndIncremental(t *testing.T) {
	source := unrollTestMachine(t)
	compareRankedAndSimpleUnroll(t, source, 4)

	u, err := NewGiaRankedUnroller(source, GiaRankedUnrollOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if stats := u.Statistics(); stats.Frames != 0 || stats.SourceObjects != source.ObjectCount() || stats.MaxTent < 1 || stats.ReachableObjects == 0 {
		t.Fatal("initial statistics", stats)
	}
	for frames := 1; frames <= 4; frames++ {
		if err = u.AppendFrames(1); err != nil {
			t.Fatal(err)
		}
		got, snapshotErr := u.Snapshot(true)
		if snapshotErr != nil {
			t.Fatal(snapshotErr)
		}
		want, simpleErr := source.Unroll(frames)
		if simpleErr != nil {
			t.Fatal(simpleErr)
		}
		for assignment := 0; assignment < 1<<got.InputCount(); assignment++ {
			inputs := make([]bool, got.InputCount())
			for i := range inputs {
				inputs[i] = assignment&(1<<i) != 0
			}
			actual, _ := got.Evaluate(inputs)
			expected, _ := want.Graph().Evaluate(inputs)
			if !slices.Equal(actual, expected) {
				t.Fatal("incremental semantics", frames, assignment, actual, expected)
			}
		}
	}
	if u.Frames() != 4 || u.Statistics().TargetOutputs != 8 || len(u.RankProfile()) == 0 {
		t.Fatal("incremental accounting", u.Frames(), u.Statistics(), u.RankProfile())
	}
	u.Stop()
	u.Stop()
	if u.Frames() != 0 || u.RankProfile() != nil || !errors.Is(u.AppendFrames(1), ErrInvariant) {
		t.Fatal("stopped lifecycle")
	}
}

func TestGiaRankedUnrollNativeNodesOwnershipAndAtomicFailure(t *testing.T) {
	g := NewGiaManager()
	giaTestHash(t, g)
	a, b, c := giaTestCI(t, g), giaTestCI(t, g), giaTestCI(t, g)
	x, err := g.HashXorReal(a, b)
	if err != nil {
		t.Fatal(err)
	}
	mux, err := g.HashMuxReal(c, x, b)
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := g.AppendBuffer(mux)
	if err != nil {
		t.Fatal(err)
	}
	giaTestCO(t, g, buffer)
	u, err := NewGiaRankedUnroller(g, GiaRankedUnrollOptions{})
	if err != nil {
		t.Fatal(err)
	}
	// Corrupt the caller-owned graph after construction. The unroller must use
	// its checked private snapshot, not borrow this storage.
	g.objects[1] = GiaObject{}
	if err = u.AppendFrames(2); err != nil {
		t.Fatal("source storage was borrowed", err)
	}
	result, err := u.Snapshot(true)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 64; assignment++ {
		inputs := make([]bool, 6)
		for i := range inputs {
			inputs[i] = assignment&(1<<i) != 0
		}
		outputs, evalErr := result.Evaluate(inputs)
		if evalErr != nil {
			t.Fatal(evalErr)
		}
		for frame := 0; frame < 2; frame++ {
			av, bv, cv := inputs[3*frame], inputs[3*frame+1], inputs[3*frame+2]
			want := bv
			if cv {
				want = av != bv
			}
			if outputs[frame] != want {
				t.Fatal("native node semantics", assignment, frame)
			}
		}
	}

	source := unrollTestMachine(t)
	calls := 0
	cancellable, err := NewGiaRankedUnroller(source, GiaRankedUnrollOptions{})
	if err != nil {
		t.Fatal(err)
	}
	before := cancellable.Statistics()
	cancellable.options.Stop = func() bool { calls++; return calls == 3 }
	if err = cancellable.AppendFrames(2); !errors.Is(err, ErrCancelled) {
		t.Fatal("cancellation", err)
	}
	if after := cancellable.Statistics(); after != before {
		t.Fatal("failed append was not atomic", before, after)
	}
}

func TestGiaRankedUnrollLimits(t *testing.T) {
	source := unrollTestMachine(t)
	for _, options := range []GiaRankedUnrollOptions{
		{MaxObjects: -1}, {MaxRank: -1}, {MaxMappingEntries: -1},
		{MaxRank: 1}, {MaxMappingEntries: source.ObjectCount() - 1},
		{Stop: func() bool { return true }},
	} {
		if _, err := NewGiaRankedUnroller(source, options); err == nil {
			t.Fatal("accepted invalid or excessive options", options)
		}
	}
	if _, err := source.UnrollRanked(0, GiaRankedUnrollOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal("frame bounds", err)
	}
	objectLimited, err := NewGiaRankedUnroller(source, GiaRankedUnrollOptions{MaxObjects: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err = objectLimited.AppendFrames(1); !errors.Is(err, ErrLimit) || objectLimited.Frames() != 0 {
		t.Fatal("object limit or atomicity", err, objectLimited.Frames())
	}
	u, err := NewGiaRankedUnroller(source, GiaRankedUnrollOptions{MaxMappingEntries: source.ObjectCount()})
	if err != nil {
		t.Fatal(err)
	}
	if err = u.AppendFrames(1); err != nil {
		t.Fatal("exact mapping limit", err)
	}
	before := u.Statistics()
	if err = u.AppendFrames(1); !errors.Is(err, ErrLimit) || u.Statistics() != before {
		t.Fatal("mapping limit or atomicity", err, before, u.Statistics())
	}
}

func FuzzGiaRankedUnroll(f *testing.F) {
	for _, seed := range [][]byte{{}, {0, 1, 2}, {2, 7, 5, 11, 3, 1}, {255, 254, 253, 252}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 30 {
			return
		}
		g := NewGiaManager()
		giaTestHash(t, g)
		roots := []GiaLit{GiaConst0, GiaConst1, giaTestCI(t, g), giaTestCI(t, g)}
		for i := 0; i+2 < len(data); i += 3 {
			a := roots[int(data[i+1])%len(roots)]
			b := roots[int(data[i+2])%len(roots)]
			if data[i]&4 != 0 {
				a = a.Not()
			}
			if data[i]&8 != 0 {
				b = b.Not()
			}
			var literal GiaLit
			var err error
			switch data[i] % 3 {
			case 0:
				literal, err = g.HashAnd(a, b)
			case 1:
				literal, err = g.HashXorReal(a, b)
			case 2:
				control := roots[int(data[i])%len(roots)]
				literal, err = g.HashMuxReal(control, a, b)
			}
			if err != nil {
				t.Fatal(err)
			}
			roots = append(roots, literal)
		}
		giaTestCO(t, g, roots[len(roots)-1])
		giaTestCO(t, g, roots[len(roots)/2])
		compareRankedAndSimpleUnroll(t, g, 3)
	})
}
