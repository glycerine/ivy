package abc

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"slices"
	"testing"
)

func unrollTestMachine(t testing.TB) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	g.Name = "shift-xor"
	giaTestHash(t, g)
	input, a, b := giaTestCI(t, g), giaTestCI(t, g), giaTestCI(t, g)
	bad := giaTestAnd(t, g, a, b)
	constraint, err := g.HashOr(input, a.Not())
	if err != nil {
		t.Fatal(err)
	}
	next, err := g.HashXor(a, input)
	if err != nil {
		t.Fatal(err)
	}
	giaTestCO(t, g, bad)
	giaTestCO(t, g, constraint)
	giaTestCO(t, g, b.Not())
	giaTestCO(t, g, next)
	if err := g.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	if err := g.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	return g
}
func TestGiaUnrollExhaustiveInitialAndTraceSemantics(t *testing.T) {
	source := unrollTestMachine(t)
	before := source.Clone()
	cases := [][]LatchInit{nil, {LatchInitOne, LatchInitZero}, {LatchInitZero, LatchInitOne}, {LatchInitUnknown, LatchInitOne}, {LatchInitUnknown, LatchInitUnknown}}
	for _, initial := range cases {
		u, err := source.UnrollWithOptions(GiaUnrollOptions{Frames: 4, Initial: initial})
		if err != nil {
			t.Fatal(err)
		}
		giaTestCheck(t, u.Graph())
		if u.Graph().RegisterCount() != 0 || u.Graph().ConstraintCount() != 0 || u.Graph().OutputCount() != 8 {
			t.Fatal("unrolling terminal counts")
		}
		if len(initial) == 0 {
			initial = []LatchInit{LatchInitZero, LatchInitZero}
		}
		unknown := 0
		for _, v := range initial {
			if v == LatchInitUnknown {
				unknown++
			}
		}
		if u.Graph().InputCount() != unknown+4 {
			t.Fatal("unknown input ordering")
		}
		for assignment := 0; assignment < 1<<(unknown+4); assignment++ {
			inputs := make([]bool, unknown+4)
			for i := range inputs {
				inputs[i] = assignment&(1<<i) != 0
			}
			outputs, err := u.Graph().Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			values, err := u.Graph().evaluateValues(inputs)
			if err != nil {
				t.Fatal(err)
			}
			state := make([]bool, 2)
			nextUnknown := 0
			for i, init := range initial {
				if init == LatchInitUnknown {
					state[i] = inputs[nextUnknown]
					nextUnknown++
				} else {
					state[i] = init == LatchInitOne
				}
			}
			for i, l := range u.InitialLiterals() {
				if (values[l.ID()] != l.Complemented()) != state[i] {
					t.Fatal("initial literal mapping")
				}
			}
			for frame := 0; frame < 4; frame++ {
				pi := inputs[unknown+frame]
				wantBad, wantConstraint := state[0] && state[1], pi || !state[0]
				if outputs[2*frame] != wantBad || outputs[2*frame+1] != wantConstraint {
					t.Fatal("unrolled sequential behavior", initial, assignment, frame)
				}
				frameInputs, err := u.FrameInputs(frame)
				if err != nil || len(frameInputs) != 1 || values[frameInputs[0].ID()] != pi {
					t.Fatal("frame PI mapping", err)
				}
				props, err := u.PropertyOutputs(frame)
				if err != nil || len(props) != 1 || values[props[0]] != wantBad {
					t.Fatal("frame property mapping", err)
				}
				constrs, err := u.ConstraintOutputs(frame)
				if err != nil || len(constrs) != 1 || values[constrs[0]] != wantConstraint {
					t.Fatal("frame constraint mapping", err)
				}
				originals, _ := source.evaluateValues([]bool{pi, state[0], state[1]})
				for id, want := range originals {
					l, err := u.ObjectLiteral(frame, GiaID(id))
					if err != nil {
						continue
					}
					if (values[l.ID()] != l.Complemented()) != want {
						t.Fatal("per-frame object mapping", frame, id)
					}
				}
				// Independent simultaneous update; do not use graph Step as the oracle.
				state = []bool{!state[1], state[0] != pi}
			}
		}
		for i := range source.objects {
			if source.objects[i] != before.objects[i] {
				t.Fatal("unrolling mutated source scratch")
			}
		}
	}
}
func TestGiaUnrollNativeNodesAndCombinationalGraph(t *testing.T) {
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
	buf, err := g.AppendBuffer(mux)
	if err != nil {
		t.Fatal(err)
	}
	giaTestCO(t, g, buf)
	u, err := g.Unroll(2)
	if err != nil {
		t.Fatal(err)
	}
	if u.Graph().XorCount() != 0 || u.Graph().MuxCount() != 0 || u.Graph().BufferCount() != 0 {
		t.Fatal("unrolled graph must lower native gates")
	}
	for assignment := 0; assignment < 64; assignment++ {
		input := make([]bool, 6)
		for i := range input {
			input[i] = assignment&(1<<i) != 0
		}
		outputs, _ := u.Graph().Evaluate(input)
		for frame := 0; frame < 2; frame++ {
			a, b, c := input[3*frame], input[3*frame+1], input[3*frame+2]
			want := b
			if c {
				want = a != b
			}
			if outputs[frame] != want {
				t.Fatal("native-node unroll")
			}
		}
	}
	empty, err := NewGiaManager().Unroll(3)
	if err != nil || empty.Graph().ObjectCount() != 1 {
		t.Fatal("empty unroll", err)
	}
}
func TestGiaUnrollLimitsAndOwnership(t *testing.T) {
	source := unrollTestMachine(t)
	for _, options := range []GiaUnrollOptions{
		{Frames: 0}, {Frames: -1}, {Frames: math.MaxInt}, {Frames: 1, MaxObjects: -1}, {Frames: 1, MaxMappingEntries: -1},
		{Frames: 4, MaxObjects: 1}, {Frames: 4, MaxMappingEntries: 1}, {Frames: 1, Initial: []LatchInit{LatchInitZero}},
		{Frames: 1, Initial: []LatchInit{LatchInit(4), LatchInitZero}},
	} {
		if _, err := source.UnrollWithOptions(options); err == nil {
			t.Fatal("accepted invalid/excessive unrolling", options)
		}
	}
	if _, err := NewGiaManagerWithLimit(0); err == nil {
		t.Fatal("zero object limit")
	}
	if _, err := NewGiaManagerWithLimit(int(MaxGiaID) + 2); err == nil {
		t.Fatal("oversize object limit")
	}
	bounded, err := NewGiaManagerWithLimit(2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bounded.AppendCI(); err != nil {
		t.Fatal(err)
	}
	if _, err := bounded.AppendCI(); err == nil || bounded.ObjectCount() != 2 {
		t.Fatal("object limit not enforced")
	}
	giaTestCheck(t, bounded)
	clone := bounded.Clone()
	if _, err := clone.AppendCI(); err == nil {
		t.Fatal("clone lost arena limit")
	}
	u, err := source.Unroll(2)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := u.FrameInputs(0)
	a[0] = InvalidGiaLit
	b, _ := u.FrameInputs(0)
	if slices.Equal(a, b) {
		t.Fatal("PI mapping alias")
	}
	if _, err := u.FrameInputs(-1); err == nil {
		t.Fatal("frame bounds")
	}
	if _, err := u.FrameOutputs(2); err == nil {
		t.Fatal("frame bounds")
	}
	if _, err := u.ObjectLiteral(0, InvalidGiaID); err == nil {
		t.Fatal("object bounds")
	}
}
func TestGiaNativeUnrollFixture(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Unroll struct {
			Frames      int
			SourceWords [][3]uint32 `json:"source_words"`
			Words       [][3]uint32
			Truths      []uint32
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	g := unrollTestMachine(t)
	if len(g.objects) != len(fixture.Unroll.SourceWords) {
		t.Fatal("native unroll source size")
	}
	for i, obj := range g.objects {
		if obj.Words() != fixture.Unroll.SourceWords[i] {
			t.Fatal("native unroll source", i, obj.Words(), fixture.Unroll.SourceWords[i])
		}
	}
	u, err := g.Unroll(fixture.Unroll.Frames)
	if err != nil {
		t.Fatal(err)
	}
	if len(u.graph.objects) != len(fixture.Unroll.Words) {
		t.Fatal("native unrolled size", len(u.graph.objects), len(fixture.Unroll.Words))
	}
	for i, obj := range u.graph.objects {
		if obj.Words() != fixture.Unroll.Words[i] {
			t.Fatal("native unrolled object", i, obj.Words(), fixture.Unroll.Words[i])
		}
	}
	for assignment := 0; assignment < 16; assignment++ {
		input := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
		outputs, _ := u.graph.Evaluate(input)
		for i, table := range fixture.Unroll.Truths {
			if outputs[i] != (table&(1<<assignment) != 0) {
				t.Fatal("native unrolled truth")
			}
		}
	}
}
func FuzzGiaUnroll(f *testing.F) {
	for _, seed := range [][]byte{{}, {0, 1, 2, 3}, {255, 254, 253, 252, 251, 250}, {3, 2, 0, 0, 1, 1, 0, 1}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 32 {
			return
		}
		g := NewGiaManager()
		giaTestHash(t, g)
		roots := []GiaLit{GiaConst0, GiaConst1, giaTestCI(t, g), giaTestCI(t, g), giaTestCI(t, g)}
		tables := []uint8{0, 255, 0xaa, 0xcc, 0xf0}
		for i := 0; i+2 < len(data); i += 3 {
			ai, bi := int(data[i+1])%len(roots), int(data[i+2])%len(roots)
			a, b, av, bv := roots[ai], roots[bi], tables[ai], tables[bi]
			if data[i]&8 != 0 {
				a = a.Not()
				av = ^av
			}
			if data[i]&16 != 0 {
				b = b.Not()
				bv = ^bv
			}
			var l GiaLit
			var table uint8
			var err error
			switch data[i] % 3 {
			case 0:
				l, err = g.HashAnd(a, b)
				table = av & bv
			case 1:
				l, err = g.HashXorReal(a, b)
				table = av ^ bv
			case 2:
				l, err = g.HashOr(a, b)
				table = av | bv
			}
			if err != nil {
				t.Fatal(err)
			}
			roots = append(roots, l)
			tables = append(tables, table)
		}
		n := len(roots)
		giaTestCO(t, g, roots[n-1])
		giaTestCO(t, g, roots[n-2])
		giaTestCO(t, g, roots[n-3])
		g.SetRegisterCount(2)
		initial := []LatchInit{LatchInitZero, LatchInitOne}
		if len(data) > 0 {
			initial[0] = LatchInit(data[0] % 3)
		}
		if len(data) > 1 {
			initial[1] = LatchInit(data[1] % 3)
		}
		u, err := g.UnrollWithOptions(GiaUnrollOptions{Frames: 3, Initial: initial})
		if err != nil {
			t.Fatal(err)
		}
		unknown := 0
		for _, v := range initial {
			if v == LatchInitUnknown {
				unknown++
			}
		}
		for assignment := 0; assignment < 1<<(unknown+3); assignment++ {
			input := make([]bool, unknown+3)
			for i := range input {
				input[i] = assignment&(1<<i) != 0
			}
			state := 0
			nextUnknown := 0
			for i, v := range initial {
				value := v == LatchInitOne
				if v == LatchInitUnknown {
					value = input[nextUnknown]
					nextUnknown++
				}
				if value {
					state |= 1 << i
				}
			}
			outputs, _ := u.graph.Evaluate(input)
			for frame := 0; frame < 3; frame++ {
				index := state << 1
				if input[unknown+frame] {
					index |= 1
				}
				if outputs[frame] != (tables[n-1]&(1<<index) != 0) {
					t.Fatal("independent unrolling truth")
				}
				state = 0
				if tables[n-2]&(1<<index) != 0 {
					state |= 1
				}
				if tables[n-3]&(1<<index) != 0 {
					state |= 2
				}
			}
		}
	})
}

func TestGiaUnrollCancellation(t *testing.T) {
	source := unrollTestMachine(t)
	before := source.Clone()
	for _, threshold := range []int{1, 3, 10} {
		calls := 0
		u, err := source.UnrollWithOptions(GiaUnrollOptions{Frames: 4, Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(err, ErrCancelled) || u != nil {
			t.Fatal("unroll cancellation", threshold, u, err)
		}
		if !slices.Equal(source.objects, before.objects) {
			t.Fatal("cancellation changed source")
		}
	}
}
