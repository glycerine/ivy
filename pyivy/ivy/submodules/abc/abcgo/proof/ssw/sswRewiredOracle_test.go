package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestCorrespondenceHandlesRewiredNodeOrder(t *testing.T) {
	for _, mode := range []int32{0, 1, 2, 3} {
		g := aig.NewGraph()
		a, b, q := g.NewInput(), g.NewInput(), g.NewInput()
		early := g.AND(a, q)
		late := g.AND(b, q)
		aig.Aig_ObjDisconnect(g, early)
		aig.Aig_ObjConnect(g, early, late, b)
		g.AddOutput(early)
		if mode == 3 {
			g.AddOutput(g.Const0())
		}
		g.AddOutput(a)
		aig.Aig_ManSetRegNum(g, 1)
		if mode == 3 {
			aig.Aig_ManSetConstrNum(g, 1)
		}
		if err := g.Check(); err != nil {
			t.Fatal(err)
		}
		p := new(Ssw_Pars_t)
		Ssw_ManSetDefaultParams(p)
		p.nStepsMax = 2
		if mode == 3 {
			p.fConstrs = 1
		}
		if mode > 0 && mode < 3 {
			p.fDynamic = 1
			p.fLocalSim = mode - 1
		}
		result := Ssw_SignalCorrespondence(g, p)
		if result == nil {
			t.Fatal("correspondence returned nil")
		}
		for mask := 0; mask < 64; mask++ {
			oldState, newState := false, false
			for f := 0; f < 3; f++ {
				x, y := mask&(1<<(2*f)) != 0, mask&(1<<(2*f+1)) != 0
				old, err := g.Evaluate([]bool{x, y, oldState})
				if err != nil {
					t.Fatal(err)
				}
				args := []bool{x, y}
				if result.RegisterCount() != 0 {
					args = append(args, newState)
				}
				out, err := result.Evaluate(args)
				if err != nil {
					t.Fatal(err)
				}
				if old[0] != out[0] {
					t.Fatalf("trace %d frame %d mismatch", mask, f)
				}
				oldState = old[len(old)-1]
				if result.RegisterCount() != 0 {
					newState = out[len(out)-1]
				}
			}
		}
	}
}

func TestRewiredUnframedRepresentativeIsDiscarded(t *testing.T) {
	for _, mode := range []string{"frame", "ordinary", "constrained"} {
		t.Run(mode, func(t *testing.T) {
			g := aig.NewGraph()
			a, b, q := g.NewInput(), g.NewInput(), g.NewInput()
			early, late := g.AND(a, q), g.AND(b, q)
			aig.Aig_ObjDisconnect(g, early)
			aig.Aig_ObjConnect(g, early, late, b)
			g.AddOutput(early)
			g.AddOutput(a)
			aig.Aig_ManSetRegNum(g, 1)
			classes := Ssw_ClassesStart(g)
			Ssw_ObjAddClass(classes, early, []aig.Edge{early, late}, 2)
			aig.Aig_ObjSetRepr(g, late, early)
			frames := aig.NewGraph()
			p := &Ssw_Man_t{pAig: g, pFrames: frames, pPars: &Ssw_Pars_t{}, ppClasses: classes, nFrames: 1, pNodeToFrames: make([]aig.Edge, len(g.Nodes()))}
			for i := range p.pNodeToFrames {
				p.pNodeToFrames[i] = aig.InvalidEdge
			}
			Ssw_ObjSetFrame(p, late, 0, frames.NewInput())
			switch mode {
			case "frame":
				Ssw_FramesConstrainNode(p, frames, g, late, 0, 1)
			case "ordinary":
				if Ssw_ManSweepNode(p, late, 0, 1, nil) != 1 {
					t.Fatal("expected refinement")
				}
			case "constrained":
				if Ssw_ManSweepNodeConstr(p, late, 0, 1) != 1 {
					t.Fatal("expected refinement")
				}
			}
			if classes.nClasses != 0 || classes.nLits != 0 || aig.Aig_ObjRepr(g, late) != aig.InvalidEdge {
				t.Fatal("unframed representative was retained")
			}
		})
	}
}

func TestDynamicPrebuiltFaninRemainsCorrectAfterMerge(t *testing.T) {
	g := aig.NewGraph()
	a, b, q := g.NewInput(), g.NewInput(), g.NewInput()
	consumer := g.AND(a, q)
	repr := g.AND(b, q)
	candidate := g.AND(repr, b)
	aig.Aig_ObjDisconnect(g, consumer)
	aig.Aig_ObjConnect(g, consumer, candidate, a)
	g.AddOutput(consumer)
	g.AddOutput(a)
	aig.Aig_ManSetRegNum(g, 1)
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	pars := new(Ssw_Pars_t)
	Ssw_ManSetDefaultParams(pars)
	pars.fDynamic = 1
	pars.nFramesK = 1
	p := Ssw_ManCreate(g, pars)
	p.ppClasses = Ssw_ClassesStart(g)
	Ssw_ObjAddClass(p.ppClasses, repr, []aig.Edge{repr, candidate}, 2)
	aig.Aig_ObjSetRepr(g, candidate, repr)
	p.pSml = Ssw_SmlStart(g, 0, 2, 1)
	p.pMSat = Ssw_SatStart(0)
	if got := Ssw_ManSweepDyn(p); got != 0 {
		t.Fatalf("equivalent fanin unexpectedly refined: %d", got)
	}
	if p.nSatProof == 0 || Ssw_ObjFrame(p, candidate, 1) != Ssw_ObjFrame(p, repr, 1) {
		t.Fatal("later fanin was not proved and merged")
	}
	p.pFrames.AddOutput(Ssw_ObjFrame(p, consumer, 1))
	if p.pFrames.InputCount() != 5 {
		t.Fatalf("unexpected framed input count %d", p.pFrames.InputCount())
	}
	for mask := 0; mask < 32; mask++ {
		input := make([]bool, 5)
		for i := range input {
			input[i] = mask&(1<<i) != 0
		}
		out, err := p.pFrames.Evaluate(input)
		if err != nil {
			t.Fatal(err)
		}
		want := input[3] && input[4] && input[1]
		if out[len(out)-1] != want {
			t.Fatalf("input %v: got %v want %v", input, out, want)
		}
	}
	Ssw_SatStop(p.pMSat)
	p.pMSat = nil
	Ssw_ManCleanup(p)
	Ssw_ManStop(p)
}
