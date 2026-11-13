package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func inductionOneRegister(property, next aig.Edge) *aig.Graph {
	g := aig.Aig_ManStart(8)
	aig.Aig_ObjCreateCi(g)
	aig.Aig_ObjCreateCo(g, property)
	aig.Aig_ObjCreateCo(g, next)
	aig.Aig_ManSetRegNum(g, 1)
	return g
}

func TestInductionBoundedConstantProperties(t *testing.T) {
	for _, tc := range []struct {
		name     string
		property aig.Edge
		want     int32
		cex      bool
	}{
		{name: "always-false", property: 1, want: 1},
		{name: "always-true", property: 0, want: 0, cex: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := inductionOneRegister(tc.property, 0)
			defer aig.Aig_ManStop(g)
			got := Saig_ManInduction(g, 0, 1, 0, 0, 0, inductionBoolInt(tc.cex), 0, 0)
			if got != tc.want {
				t.Fatalf("induction returned %d, want %d", got, tc.want)
			}
			if tc.cex && Saig_ManInductionCounterexample(g) == nil {
				t.Fatal("missing induction model")
			}
		})
	}
}

func TestInductionRejectsInvalidRegisterlessGraph(t *testing.T) {
	g := aig.Aig_ManStart(4)
	aig.Aig_ObjCreateCo(g, 0)
	defer aig.Aig_ManStop(g)
	if got := Saig_ManInduction(g, 0, 1, 0, 0, 0, 0, 0, 0); got != -1 {
		t.Fatalf("got %d, want invalid-input status -1", got)
	}
}

func inductionBoolInt(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

func TestInductionModelValidatesCompactLayout(t *testing.T) {
	g := aig.Aig_ManStart(8)
	aig.Aig_ObjCreateCi(g)
	aig.Aig_ObjCreateCi(g)
	aig.Aig_ObjCreateCi(g)
	aig.Aig_ObjCreateCi(g)
	aig.Aig_ObjCreateCo(g, aig.Aig_ManConst1(g))
	aig.Aig_ObjCreateCo(g, aig.Aig_ManConst1(g))
	aig.Aig_ObjCreateCo(g, aig.Aig_ManConst1(g))
	aig.Aig_ManSetRegNum(g, 2)
	defer aig.Aig_ManStop(g)
	valid := util.Abc_CexAlloc(1, 2, 1)
	if err := g.SetInductionCounterexample(valid); err != nil {
		t.Fatalf("valid compact witness rejected: %v", err)
	}
	invalid := util.Abc_CexAlloc(2, 2, 1)
	if err := g.SetInductionCounterexample(invalid); err == nil {
		t.Fatal("full-register witness accepted for compact induction layout")
	}
	if g.InductionCounterexample() != valid {
		t.Fatal("invalid replacement changed the saved model")
	}
}

func TestInductionUniquenessExcludesEqualStates(t *testing.T) {
	sat := bsat.NewSolver()
	if sat.NewVariable() != 0 || sat.NewVariable() != 1 {
		t.Fatal("unexpected state variable IDs")
	}
	states := vec.Vec_IntAlloc(3)
	vec.Vec_IntPush(states, -1) // state zero is not part of the comparison
	vec.Vec_IntPush(states, 0)
	vec.Vec_IntPush(states, 1)
	defer vec.Vec_IntFree(states)
	if sat.Solve(bsat.Positive(0), bsat.Positive(1)) != bsat.Satisfiable || Saig_ManStatesAreEqual(sat, states, 1, 1, 2) == 0 {
		t.Fatal("identical states were not recognized")
	}
	satVars, clauses := int32(2), int32(0)
	if Saig_ManAddUniqueness(sat, states, 1, 1, 2, &satVars, &clauses, 0) != 0 {
		t.Fatal("adding a valid uniqueness constraint reported an insertion conflict")
	}
	if got := sat.Solve(); got != bsat.Satisfiable || Saig_ManStatesAreEqual(sat, states, 1, 1, 2) != 0 {
		t.Fatalf("uniqueness did not force distinct states: status=%v", got)
	}
}

func TestInductionCounterexampleCopiesTopPrimaryInput(t *testing.T) {
	g := aig.Aig_ManStart(8)
	pi := aig.Aig_ObjCreateCi(g)
	aig.Aig_ObjCreateCi(g)
	aig.Aig_ObjCreateCo(g, pi)
	aig.Aig_ObjCreateCo(g, aig.Aig_ManConst1(g))
	aig.Aig_ManSetRegNum(g, 1)
	defer aig.Aig_ManStop(g)
	if got := Saig_ManInduction(g, 0, 1, 0, 0, 0, 1, 0, 0); got != 0 {
		t.Fatalf("induction returned %d, want bounded SAT result 0", got)
	}
	cex := Saig_ManInductionCounterexample(g)
	if cex == nil || cex.NBits != 1 || !util.Abc_InfoHasBit(cex.PData, 0) {
		t.Fatalf("top PI value was not preserved in compact model: %+v", cex)
	}
}
