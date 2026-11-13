package bmc

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestUnrollGIAEntryPoints(t *testing.T) {
	g := aig.NewGraph()
	pi, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	p := gia.Gia_ManFromAig(g)
	manager := Unr_ManUnrollStart(p, 0)
	if manager == nil {
		t.Fatal("missing manager")
	}
	graphs := []*gia.Gia_Man_t{Unr_ManUnrollFrame(manager, 1), Unr_ManUnroll(p, 2), Unr_ManUnrollSimple(p, 2)}
	for _, u := range graphs {
		if u == nil {
			t.Fatal("missing unrolling")
		}
		out, err := u.Evaluate([]bool{true, false})
		if err != nil || !reflect.DeepEqual(out, []bool{false, true, false}) {
			t.Fatalf("unrolling %v: %v", out, err)
		}
	}
	if Unr_ManUnrollFrame(manager, -1) != nil || Unr_ManUnroll(nil, 1) != nil || Unr_ManUnrollStart(nil, 0) != nil {
		t.Fatal("invalid argument accepted")
	}
	Unr_ManFree(manager)
}

func TestBmciGIAEntryPoint(t *testing.T) {
	p := gia.Gia_ManFromAig(stateRegister(true))
	for _, tc := range []struct {
		initial      []int32
		frames, want int32
	}{
		{[]int32{0}, 1, 1}, {[]int32{1}, 1, 0}, {[]int32{2}, 1, 0}, {[]int32{1}, 0, 1}, {nil, 1, -1}, {[]int32{-1}, 1, -1},
	} {
		init := vec.Vec_IntAllocArrayCopy(tc.initial, len(tc.initial))
		if got := Gia_ManBmciTest(p, init, tc.frames, 1, 0, 0, 0); got != tc.want {
			t.Errorf("init=%v frames=%d got=%d want=%d", tc.initial, tc.frames, got, tc.want)
		}
	}
}

func TestLoadTestGIAOutputs(t *testing.T) {
	p := gia.Gia_ManStart(8)
	input := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, input)
	gia.Gia_ManAppendCo(p, 0)
	for _, mode := range []int32{0, 1} {
		read, write, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		old := os.Stdout
		os.Stdout = write
		Bmc_LoadTest(p, mode, 1)
		os.Stdout = old
		write.Close()
		output, err := io.ReadAll(read)
		read.Close()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(output), "Output 0: SAT") || !strings.Contains(string(output), "Output 1: UNSAT") {
			t.Fatalf("mode=%d output=%q", mode, output)
		}
	}
}

func TestBMCGIAWrappers(t *testing.T) {
	implementations := []struct {
		name string
		run  func(*gia.Gia_Man_t, *Bmc_AndPar_t) int32
	}{
		{"bmcg-one", Bmcg_ManPerformOne}, {"bmcg", Bmcg_ManPerform}, {"bmcs-one", Bmcs_ManPerformOne}, {"bmcs", Bmcs_ManPerform},
	}
	for _, impl := range implementations {
		for _, bad := range []bool{false, true} {
			t.Run(impl.name+map[bool]string{false: "-safe", true: "-fails"}[bad], func(t *testing.T) {
				g := aig.NewGraph()
				input, state := g.NewInput(), g.NewInput()
				g.AddOutput(state)
				next := g.Const0()
				if bad {
					next = input
				}
				g.AddOutput(next)
				aig.Aig_ManSetRegNum(g, 1)
				p := gia.Gia_ManFromAig(g)
				params := &Bmc_AndPar_t{nFramesMax: 2}
				got := impl.run(p, params)
				if bad {
					cex := p.SequentialCounterexample()
					if got != 0 || cex == nil || cex.IFrame != 1 || cex.IPo != 0 {
						t.Fatalf("result=%d cex=%+v", got, cex)
					}
				} else if got != 1 || p.SequentialCounterexample() != nil {
					t.Fatalf("result=%d cex=%+v", got, p.SequentialCounterexample())
				}
			})
		}
	}
}
