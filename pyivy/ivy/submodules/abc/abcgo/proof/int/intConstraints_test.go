package int

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

// Constraint outputs signal forbidden behavior: x is both the bad output
// and a constraint violation, so there is no legal initial counterexample.
func TestInitialConstraintPolarity(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.NewInput()
	g.AddOutput(x)
	g.AddOutput(x)
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	if got := Inter_ManCheckInitialState(g); got != 0 {
		t.Fatalf("constraint-violating input accepted: %d", got)
	}
}

func TestConstrainedInterpolation(t *testing.T) {
	for _, forbidden := range []bool{false, true} {
		g := aig.NewGraph()
		x := g.NewInput()
		g.NewInput()
		g.AddOutput(x)
		if forbidden {
			g.AddOutput(x)
		} else {
			g.AddOutput(g.Const0())
		}
		g.AddOutput(g.Const0())
		aig.Aig_ManSetRegNum(g, 1)
		aig.Aig_ManSetConstrNum(g, 1)
		var pars Inter_ManParams_t
		Inter_ManSetDefaultParams(&pars)
		pars.NFramesMax = 4
		var frame int32
		got := Inter_ManPerformInterpolation(g, &pars, &frame)
		want := int32(0)
		if forbidden {
			want = 1
		}
		if got != want {
			t.Fatalf("forbidden=%v got %d want %d frame %d", forbidden, got, want, frame)
		}
	}
}

func TestConstrainedSequentialWitness(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(x.Not()) // legal frames require x=1
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	var pars Inter_ManParams_t
	Inter_ManSetDefaultParams(&pars)
	pars.NFramesMax = 4
	var frame int32
	if got := Inter_ManPerformInterpolation(g, &pars, &frame); got != 0 || frame != 1 {
		t.Fatalf("got %d frame %d", got, frame)
	}
	cex := g.SequentialCounterexample()
	if cex == nil {
		t.Fatal("missing witness")
	}
	for f := int32(0); f <= cex.IFrame; f++ {
		if !util.Abc_InfoHasBit(cex.PData, 1+f) {
			t.Fatalf("constraint violated at frame %d", f)
		}
	}
}

func TestConstrainedUnrollRejectsEarlierConstraintViolation(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(x) // each legal frame requires x=false
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	frames := Inter_ManFramesBmc(g, 2)
	for mask := 0; mask < 4; mask++ {
		out, err := frames.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
		if err != nil {
			t.Fatal(err)
		}
		if out[0] {
			t.Fatalf("constraint-violating sequence %02b accepted", mask)
		}
	}
	if got := Inter_ManGetCounterExample(g, 2, 0); got != nil {
		t.Fatalf("unexpected witness %+v", got)
	}
}

func TestConstrainedWitnessHonorsExpiredDeadline(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(x.Not())
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	if got := interManGetCounterExample(g, 2, 0, util.Abc_Clock()-1); got != nil {
		t.Fatalf("expired deadline yielded witness %+v", got)
	}
}

func TestWitnessChecksEveryConstraintInEveryFrame(t *testing.T) {
	baseline := aig.NewGraph()
	bx, bs := baseline.NewInput(), baseline.NewInput()
	baseline.AddOutput(bs)
	baseline.AddOutput(baseline.Const0())
	baseline.AddOutput(bx)
	aig.Aig_ManSetRegNum(baseline, 1)
	aig.Aig_ManSetConstrNum(baseline, 1)
	if got := Inter_ManGetCounterExample(baseline, 2, 0); got == nil {
		t.Fatal("first constraint alone should permit a witness")
	}
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(g.Const0())  // first constraint allows every input
	g.AddOutput(state.Not()) // second constraint excludes the initial state
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 2)
	if got := Inter_ManGetCounterExample(g, 2, 0); got != nil {
		t.Fatalf("contradictory constraints produced witness %+v", got)
	}
}
