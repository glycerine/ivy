package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/bdd/bbr"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/pdr"
	"github.com/glycerine/abcgo/sat/bmc"
	"time"
)

// Saig_ManCexRefine checks an abstraction using the requested backend and
// adds the abstracted registers needed to explain a counterexample.
func Saig_ManCexRefine(g, abstraction *aig.Graph, flops *vec.Vec_Int_t, frames, conflicts, useBDDs, usePDR, verbose int32, start, status, framesDone *int32) *aig.Graph {
	*status = -1
	if usePDR != 0 && saig.Saig_ManRegNum(abstraction) > 0 {
		combined := saig.Saig_ManDupOrpos(abstraction)
		options := pdr.DefaultOptions()
		options.Timeout = 10 * time.Second
		options.Verbose = verbose != 0
		if verbose != 0 {
			fmt.Println("Running property directed reachability...")
		}
		result, err := pdr.Solve(combined, &options)
		if err != nil {
			return nil
		}
		if model := combined.SequentialCounterexample(); model != nil {
			copyModel := *model
			copyModel.IPo = saig.Saig_ManFindFailedPoCex(abstraction, model)
			if err := abstraction.SetSequentialCounterexample(&copyModel); err != nil {
				panic(err)
			}
		}
		if result.Status == pdr.Proved {
			*status = 1
		}
	} else if useBDDs != 0 && saig.Saig_ManRegNum(abstraction) > 0 && saig.Saig_ManRegNum(abstraction) <= 80 {
		var options bbr.Saig_ParBbr_t
		bbr.Bbr_ManSetDefaultParams(&options)
		options.TimeLimit = 0
		options.NBddMax = 1000000
		options.NIterMax = frames
		options.FPartition = 1
		options.FReorder = 1
		options.FReorderImage = 1
		options.FVerbose = verbose
		options.FSilent = 0
		if bbr.Aig_ManVerifyUsingBdds(abstraction, &options) == 1 {
			*status = 1
		}
	} else {
		from := int32(0)
		if start != nil {
			from = *start
		}
		bmc.Saig_BmcPerform(abstraction, from, frames, 2000, 0, conflicts, 0, verbose, 0, framesDone, 0, 0)
	}
	cex := abstraction.SequentialCounterexample()
	if cex == nil {
		return nil
	}
	if start != nil {
		*start = cex.IFrame
	}
	selected := Saig_ManExtendCounterExampleTest3(abstraction, Saig_ManCexFirstFlopPi(g, abstraction), cex, verbose)
	if selected == nil {
		return nil
	}
	if vec.Vec_IntSize(selected) == 0 {
		fmt.Println("Discovered a true counter-example!")
		model := Saig_ManCexRemap(g, abstraction, cex)
		if model != nil {
			if err := g.SetSequentialCounterexample(model); err != nil {
				panic(err)
			}
		}
		*status = 0
		return nil
	}
	if verbose != 0 {
		fmt.Printf("Adding %d registers to the abstraction (total = %d).\n\n", vec.Vec_IntSize(selected), int(saig.Saig_ManRegNum(abstraction))+vec.Vec_IntSize(selected))
	}
	for _, input := range vec.Vec_IntArray(selected) {
		if input < 0 || int(input) >= len(abstraction.InputNumbers) {
			panic("invalid abstract input")
		}
		original := abstraction.InputNumbers[input]
		piCount := saig.Saig_ManPiNum(g)
		if original < piCount || int(original) >= len(g.Inputs()) {
			panic("abstract input is not a register")
		}
		vec.Vec_IntPush(flops, original-piCount)
	}
	vec.Vec_IntSort(flops, 0)
	values := vec.Vec_IntArray(flops)
	for i := 1; i < len(values); i++ {
		if values[i-1] == values[i] {
			panic("duplicate abstraction register")
		}
	}
	return saig.Saig_ManDupAbstraction(g, flops)
}

// Saig_ManCexAbstractionFlops repeatedly checks and refines an abstraction.
func Saig_ManCexAbstractionFlops(g *aig.Graph, params *Gia_ParAbs_t) *vec.Vec_Int_t {
	if saig.Saig_ManRegNum(g) <= 0 {
		panic("abstraction requires registers")
	}
	if params.fVerbose != 0 {
		fmt.Println("Performing counter-example-based refinement.")
	}
	// A previous run's witness must not be reported for this run.
	g.ClearSequentialCounterexample()
	flops := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(flops, 0)
	abstract := saig.Saig_ManDupAbstraction(g, flops)
	if params.fVerbose == 0 {
		fmt.Print("Init : ")
		aig.Aig_ManPrintStats(abstract)
	}
	fmt.Println("Refining abstraction...")
	var start int32
	for iteration := 0; ; iteration++ {
		var startPtr *int32
		if params.fUseStart != 0 {
			startPtr = &start
		}
		next := Saig_ManCexRefine(g, abstract, flops, params.nFramesBmc, params.nConfMaxBmc, params.fUseBdds, params.fUseDprove, params.fVerbose, startPtr, &params.Status, &params.nFramesDone)
		if next == nil {
			// A real abstract counterexample may have extra pseudo-inputs;
			// Saig_ManCexRefine has already installed its remapped version.
			if g.SequentialCounterexample() == nil {
				if model := abstract.SequentialCounterexample(); model != nil {
					_ = g.SetSequentialCounterexample(model)
				}
			}
			break
		}
		abstract = next
		fmt.Printf("ITER %4d : ", iteration)
		if params.fVerbose == 0 {
			aig.Aig_ManPrintStats(abstract)
		}
		if err := ioa.Ioa_WriteAiger(abstract, "gabs.aig", 0, 0); err != nil {
			fmt.Printf("Could not write intermediate abstraction: %v\n", err)
		}
		if 100.0*float64(saig.Saig_ManRegNum(g)-saig.Saig_ManRegNum(abstract))/float64(saig.Saig_ManRegNum(g)) < float64(params.nRatio) {
			fmt.Printf("Refinements is stopped because flop reduction is less than %d%%\n", params.nRatio)
			return nil
		}
	}
	return flops
}
