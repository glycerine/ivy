package abc

// Go-only frontier liveness, round-trip, and fuzz checks for src/aig/gia/giaFront.c.

import (
	"errors"
	"math/rand"
	"slices"
	"testing"
)

func giaFrontierFixture(t testing.TB) *GiaManager {
	t.Helper()
	manager := NewGiaManager()
	manager.Name, manager.Spec = "frontier", "frontier.aig"
	a, _ := manager.AppendCI()
	b, _ := manager.AppendCI()
	r, _ := manager.AppendCI()
	and, _ := manager.AppendAnd(a, b.Not())
	xor, _ := manager.AppendXorReal(and, r)
	manager.AppendCO(xor)
	manager.AppendCO(and.Not())
	manager.AppendCO(xor.Not())
	if err := manager.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err := manager.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	return manager
}

func independentGiaFrontierPeak(t testing.TB, manager *GiaManager) int {
	t.Helper()
	uses := make([][]int, manager.ObjectCount())
	for id, object := range manager.objects {
		if object.IsCO() || object.IsAnd() {
			first, err := object.Fanin0(GiaID(id))
			if err != nil {
				t.Fatal(err)
			}
			uses[first.ID()] = append(uses[first.ID()], id)
		}
		if object.IsAnd() {
			second, err := object.Fanin1(GiaID(id))
			if err != nil {
				t.Fatal(err)
			}
			uses[second.ID()] = append(uses[second.ID()], id)
		}
	}
	peak := 0
	for step := range manager.objects {
		live := 0
		for object := 0; object <= step; object++ {
			if slices.ContainsFunc(uses[object], func(use int) bool { return use >= step }) {
				live++
			}
		}
		if live > peak {
			peak = live
		}
	}
	return peak
}

func TestGiaFrontierCrossCutSlotsAndRoundTrip(t *testing.T) {
	manager := NewGiaManager()
	a, _ := manager.AppendCI()
	b, _ := manager.AppendCI()
	and, _ := manager.AppendAnd(a, b)
	manager.AppendCO(and.Not())
	peak, err := manager.FrontierCrossCut(GiaFrontierOptions{})
	if err != nil || peak != 3 || peak != independentGiaFrontierPeak(t, manager) {
		t.Fatal(peak, err)
	}
	frontier, err := manager.EncodeGiaFrontier(GiaFrontierOptions{})
	if err != nil || frontier.Peak != 3 || frontier.Width != 4 {
		t.Fatal(frontier, err)
	}
	objects := frontier.Objects()
	if len(objects) != 5 || objects[0].Slot != 0 || objects[1].Slot != 0 || objects[2].Slot != 1 || objects[3].Slot != 2 || objects[3].Fanin0 != 0 || objects[3].Fanin1 != 1 || objects[4].Fanin0 != 2 || !objects[4].Complement0 {
		t.Fatal(objects)
	}
	rebuilt, err := frontier.Decode(GiaFrontierOptions{})
	if err != nil || !slices.Equal(manager.objects, rebuilt.objects) {
		t.Fatal(rebuilt, err)
	}
	ok, err := manager.FrontierRoundTrip(GiaFrontierOptions{})
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
	objects[1].Slot = 99
	if frontier.Objects()[1].Slot == 99 {
		t.Fatal("object snapshots alias frontier storage")
	}
}

func TestGiaFrontierSequentialMetadataTruthAndDangling(t *testing.T) {
	manager := giaFrontierFixture(t)
	dangling, _ := manager.AppendAnd(GiaLit(manager.cis[0])<<1, GiaLit(manager.cis[1])<<1)
	_ = dangling
	frontier, err := manager.EncodeGiaFrontier(GiaFrontierOptions{})
	if err != nil {
		t.Fatal(err)
	}
	rebuilt, err := frontier.Decode(GiaFrontierOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if rebuilt.Name != manager.Name || rebuilt.Spec != manager.Spec || rebuilt.RegisterCount() != 1 || rebuilt.ConstraintCount() != 1 || !slices.Equal(rebuilt.objects, manager.objects) {
		t.Fatal("metadata or object mismatch")
	}
	for assignment := 0; assignment < 1<<manager.InputCount(); assignment++ {
		inputs := make([]bool, manager.InputCount())
		for index := range inputs {
			inputs[index] = assignment&(1<<index) != 0
		}
		want, wantErr := manager.Evaluate(inputs)
		got, gotErr := rebuilt.Evaluate(inputs)
		if wantErr != nil || gotErr != nil || !slices.Equal(got, want) {
			t.Fatal(assignment, want, got, wantErr, gotErr)
		}
	}
	if frontier.Peak != independentGiaFrontierPeak(t, manager) {
		t.Fatal(frontier.Peak)
	}
}

func TestGiaFrontierErrorsLimitsCancellationAndDeep(t *testing.T) {
	manager := giaFrontierFixture(t)
	if _, err := (*GiaManager)(nil).EncodeGiaFrontier(GiaFrontierOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := manager.EncodeGiaFrontier(GiaFrontierOptions{MaxObjects: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := manager.EncodeGiaFrontier(GiaFrontierOptions{MaxFrontier: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := manager.EncodeGiaFrontier(GiaFrontierOptions{MaxObjects: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := manager.EncodeGiaFrontier(GiaFrontierOptions{MaxFrontier: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	for threshold := 1; threshold <= manager.ObjectCount()*3; threshold++ {
		calls := 0
		_, err := manager.EncodeGiaFrontier(GiaFrontierOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if err != nil && !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, calls, err)
		}
	}
	buffered := NewGiaManager()
	a, _ := buffered.AppendCI()
	buffer, _ := buffered.AppendBuffer(a)
	buffered.AppendCO(buffer)
	if _, err := buffered.EncodeGiaFrontier(GiaFrontierOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	muxed := NewGiaManager()
	c, _ := muxed.AppendCI()
	x, _ := muxed.AppendCI()
	y, _ := muxed.AppendCI()
	mux, _ := muxed.AppendMuxReal(c, x, y)
	muxed.AppendCO(mux)
	if _, err := muxed.EncodeGiaFrontier(GiaFrontierOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}

	frontier, err := manager.EncodeGiaFrontier(GiaFrontierOptions{})
	if err != nil {
		t.Fatal(err)
	}
	broken := *frontier
	broken.objects = slices.Clone(frontier.objects)
	broken.objects[1].Slot = broken.Width
	if _, err = broken.Decode(GiaFrontierOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}

	deep := NewGiaManager()
	left, _ := deep.AppendCI()
	right, _ := deep.AppendCI()
	chain, _ := deep.AppendAnd(left, right)
	for index := 0; index < 10000; index++ {
		chain, _ = deep.AppendAnd(chain, left)
	}
	deep.AppendCO(chain)
	ok, err := deep.FrontierRoundTrip(GiaFrontierOptions{})
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
}

func FuzzGiaFrontierRoundTrip(f *testing.F) {
	f.Add(uint64(1))
	f.Add(uint64(0x9e3779b97f4a7c15))
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		manager := NewGiaManager()
		signals := []GiaLit{GiaConst0, GiaConst1}
		for index := 0; index < 1+random.Intn(5); index++ {
			literal, _ := manager.AppendCI()
			signals = append(signals, literal)
		}
		for index := 0; index < random.Intn(12); index++ {
			first := signals[random.Intn(len(signals))].NotIf(random.Intn(2) != 0)
			second := signals[random.Intn(len(signals))].NotIf(random.Intn(2) != 0)
			if first.ID() == second.ID() {
				continue
			}
			var literal GiaLit
			var err error
			if random.Intn(4) == 0 {
				literal, err = manager.AppendXorReal(first, second)
			} else {
				literal, err = manager.AppendAnd(first, second)
			}
			if err == nil {
				signals = append(signals, literal)
			}
		}
		for index := 0; index < 1+random.Intn(4); index++ {
			manager.AppendCO(signals[random.Intn(len(signals))].NotIf(random.Intn(2) != 0))
		}
		frontier, err := manager.EncodeGiaFrontier(GiaFrontierOptions{MaxObjects: 4096, MaxFrontier: 4096})
		if err != nil {
			t.Fatal(seed, err)
		}
		if frontier.Peak != independentGiaFrontierPeak(t, manager) {
			t.Fatal(seed, frontier.Peak)
		}
		rebuilt, err := frontier.Decode(GiaFrontierOptions{MaxObjects: 4096, MaxFrontier: 4096})
		if err != nil || !slices.Equal(manager.objects, rebuilt.objects) {
			t.Fatal(seed, err)
		}
		for assignment := 0; assignment < 1<<manager.InputCount(); assignment++ {
			inputs := make([]bool, manager.InputCount())
			for input := range inputs {
				inputs[input] = assignment&(1<<input) != 0
			}
			want, _ := manager.Evaluate(inputs)
			got, evalErr := rebuilt.Evaluate(inputs)
			if evalErr != nil || !slices.Equal(got, want) {
				t.Fatal(seed, assignment, want, got, evalErr)
			}
		}
	})
}
