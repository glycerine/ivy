package cec

// SAT-sweeping orchestration follows cecSolve.c and uses ABC's bsat port.
import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func cecResetPatterns(p *Cec_ManPat_t) {
	if p == nil {
		return
	}
	p.iStart = int32(vec.Vec_StrSize(p.vStorage))
	p.nPats = 0
	p.nPatLits = 0
	p.nPatLitsMin = 0
}
func cecSetOutputStatus(output *Gia_Obj_t, status int32) {
	var sat, proved int32
	if status == 0 {
		sat = 1
	}
	if status == 1 {
		proved = 1
	}
	gia.Gia_ObjSetMark0(output, sat)
	gia.Gia_ObjSetMark1(output, proved)
}

// Cec_ManSatSolve checks each output, marks its status, and optionally stores
// witnesses, original equivalence pairs, and sensitizing patterns.
func Cec_ManSatSolve(patterns *Cec_ManPat_t, graph *Gia_Man_t, pars *Cec_ParSat_t, orig, pairs, equiv *Vec_Int_t, proveZero int32) {
	graph.ClearSequentialCounterexamples()
	cecResetPatterns(patterns)
	gia.Gia_ManSetPhase(graph)
	gia.Gia_ManLevelNum(graph)
	gia.Gia_ManIncrementTravId(graph)
	sat := Cec_ManSatCreate(graph, pars)
	defer Cec_ManSatStop(sat)
	pars = sat.pPars
	classic := newCecClassicSolver(graph, pars)
	var models []*util.Abc_Cex_t
	if pars.fSaveCexes != 0 {
		models = make([]*util.Abc_Cex_t, int(gia.Gia_ManCoNum(graph)))
	}
	for i, lit := range graph.OutputLiterals() {
		output := gia.Gia_ManCo(graph, int32(i))
		var status int32
		if lit == 0 {
			status = 1
		} else if lit == 1 {
			status = 0
		} else {
			status = classic.check(lit)
			if status == 0 {
				var err error
				sat.values, err = sat.proof.Model().Values(classic.inputs())
				if err != nil {
					panic(err)
				}
			}
		}
		cecSetOutputStatus(output, status)
		if lit > 1 && status == 1 && orig != nil {
			if pairs == nil || equiv == nil {
				panic("cec: missing original equivalence pair vectors")
			}
			a, b := vec.Vec_IntEntry(pairs, 2*i), vec.Vec_IntEntry(pairs, 2*i+1)
			x, y := vec.Vec_IntEntry(orig, int(a)), vec.Vec_IntEntry(orig, int(b))
			if x < 0 || y < 0 {
				panic("cec: negative original node id")
			}
			vec.Vec_IntPushTwo(equiv, x, y)
		}
		if models != nil && status == 1 {
			models[i] = util.AbcCexSentinel
		}
		if models != nil && status == 0 {
			if lit == 1 {
				models[i] = Cex_ManGenSimple(sat, int32(i))
			} else {
				models[i] = Cex_ManGenCex(sat, int32(i))
			}
		}
		if lit > 1 && proveZero != 0 && status == 1 {
			gia.Gia_ManPatchCoDriver(graph, int32(i), 0)
		}
		if lit > 1 && status == 0 {
			if patterns != nil {
				start := util.Abc_Clock()
				Cec_ManPatSavePattern(patterns, sat, output)
				patterns.timeTotalSave += util.Abc_Clock() - start
			}
			if pars.fCheckMiter != 0 {
				break
			}
		}
	}
	if models != nil {
		if err := graph.SetCombinationalCounterexamples(models); err != nil {
			panic(err)
		}
	}
	if pars.fVerbose != 0 {
		classic.printStats()
	}
}

func Cec_ManSatSolveExractPattern(store *Vec_Int_t, start int32, pat *Vec_Int_t) int32 {
	vec.Vec_IntClear(pat)
	start++
	n := vec.Vec_IntEntry(store, int(start))
	start++
	for i := int32(0); i < n; i++ {
		vec.Vec_IntPush(pat, vec.Vec_IntEntry(store, int(start)))
		start++
	}
	return start
}
func Cec_ManSatSolveCSat(patterns *Cec_ManPat_t, graph *Gia_Man_t, pars *Cec_ParSat_t) {
	var statuses *Vec_Str_t
	store := gia.Cbs_ManSolveMiterNc(graph, pars.nBTLimit, &statuses, 0, 0)
	pat := vec.Vec_IntAlloc(1000)
	cecResetPatterns(patterns)
	var start int32
	if vec.Vec_StrSize(statuses) != int(gia.Gia_ManCoNum(graph)) {
		panic("cec: CBS status count mismatch")
	}
	for i := 0; i < vec.Vec_StrSize(statuses); i++ {
		status := int32(int8(vec.Vec_StrEntry(statuses, i)))
		cecSetOutputStatus(gia.Gia_ManCo(graph, int32(i)), status)
		if vec.Vec_IntSize(store) > 0 && status != 1 {
			start = Cec_ManSatSolveExractPattern(store, start, pat)
		}
		if status != 0 {
			continue
		}
		if patterns != nil && vec.Vec_IntSize(pat) > 0 {
			clock := util.Abc_Clock()
			Cec_ManPatSavePatternCSat(patterns, pat)
			patterns.timeTotalSave += util.Abc_Clock() - clock
		}
		if pars.fCheckMiter != 0 {
			break
		}
	}
	if start != int32(vec.Vec_IntSize(store)) && pars.fCheckMiter == 0 {
		panic("cec: CBS counterexample store has trailing entries")
	}
}
