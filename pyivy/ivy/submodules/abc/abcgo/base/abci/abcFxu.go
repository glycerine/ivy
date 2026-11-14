package abci

import (
	"fmt"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/opt/fxu"
)

// Fxu_Data_t preserves the C command's public parameter names. The cover
// vectors are owned by opt/fxu in Go; these fields are retained for ABC's
// command-line compatibility and are populated only by the source algorithm.
type Fxu_Data_t struct {
	nSingleMax, nPairsMax, nNodesExt  int32
	WeightMin, LitCountMax            int32
	fOnlyS, fOnlyD, fUse0             int32
	fUseCompl, fVerbose, fVeryVerbose int32
	fCanonDivs                        int32
	nNodesOld, nNodesNew              int32
	algorithm                         *fxu.Fxu_Data_t
}

func Abc_NtkSetDefaultFxParams(p *Fxu_Data_t) {
	*p = Fxu_Data_t{nSingleMax: 20000, nPairsMax: 30000, nNodesExt: 1000000, LitCountMax: 4, fUseCompl: 1}
}

func Abc_NtkFastExtract(ntk *Abc_Ntk_t, p *Fxu_Data_t) int32 {
	if baseabc.Abc_NtkIsLogic(ntk) == 0 {
		return 0
	}
	if baseabc.Abc_NtkIsSopLogic(ntk) != 0 {
		// ABC's source leaves the SOP->BDD->SOP SCC normalization disabled.
	}
	if baseabc.Abc_NtkToSop(ntk, -1, 1000000000) == 0 {
		fmt.Printf("Abc_NtkFastExtract(): Converting to SOPs has failed.\n")
		return 0
	}
	if Abc_NtkFxuCheck(ntk) == 0 {
		fmt.Printf("Abc_NtkFastExtract: Nodes have duplicated or complemented fanins. FXU is not performed.\n")
		return 0
	}
	baseabc.Abc_NtkCleanup(ntk, 0)
	params := &fxu.Fxu_Data_t{
		OnlyS: p.fOnlyS != 0, OnlyD: p.fOnlyD != 0, Use0: p.fUse0 != 0,
		UseCompl: p.fUseCompl != 0, Verbose: p.fVerbose != 0, VeryVerbose: p.fVeryVerbose != 0,
		NodesExt: p.nNodesExt, SingleMax: p.nSingleMax, PairsMax: p.nPairsMax,
		WeightMin: p.WeightMin, LiteralCountMax: p.LitCountMax,
		CanonicalDivisors: p.fCanonDivs != 0,
	}
	p.algorithm = params
	changed := fxu.Abc_NtkFastExtract(ntk, params)
	p.nNodesOld = params.OldNodeCount()
	p.nNodesNew = params.NewNodeCount()
	if changed == 0 {
		fmt.Printf("Warning: The network has not been changed by \"fx\".\n")
		return 0
	}
	if baseabc.Abc_NtkCheck(ntk) == 0 {
		fmt.Printf("Abc_NtkFastExtract: The network check has failed.\n")
	}
	return 1
}

func Abc_NtkFxuCheck(ntk *Abc_Ntk_t) int32 {
	for _, node := range ntk.ObjectsOfKind(baseabc.ObjectNode) {
		for i, edge := range node.Fanins() {
			if i < 2 && edge.Complement {
				return 0
			}
			for k, other := range node.Fanins() {
				if i != k && edge.Object == other.Object {
					return 0
				}
			}
		}
	}
	return 1
}

// In Go the FXU algorithm's temporary arrays are garbage collected. This
// retains the command's C cleanup entry point and releases its parameter state.
func Abc_NtkFxuFreeInfo(p *Fxu_Data_t) {
	if p.algorithm != nil {
		fxu.Abc_NtkFxuFreeInfo(p.algorithm)
	}
	p.algorithm = nil
}
