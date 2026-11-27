package abci

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/misc/util"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cas"
	"github.com/glycerine/abcgo/bdd/cudd"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/extra"
)

// Abc_NtkCascade decomposes the shared output BDDs into a LUT cascade.
// As in ABC's original implementation, this currently returns a checked copy
// of the input network; the experiment emits the cascade files as a side effect.
// Abc_NtkCascade is abcCas.c's: global BDDs (500000-node limit, dynamic
// reordering), the cascade experiment, and a copy of the network whatever
// the experiment found.
func Abc_NtkCascade(pNtk *Abc_Ntk_t, nLutSize, fCheck, fVerbose int32) *Abc_Ntk_t {
	if pNtk == nil || abc.Abc_NtkIsStrash(pNtk) == 0 {
		return nil
	}
	start := time.Now()
	if Abc_NtkBuildGlobalBdds(pNtk, 500000, 1, 1, 0, fVerbose) == nil {
		return nil
	}
	dd := Abc_NtkGlobalBddMan(pNtk)
	if fVerbose != 0 {
		fmt.Printf("Shared BDD size = %6d nodes.  ", cudd.Cudd_ReadKeys(dd)-cudd.Cudd_ReadDead(dd))
		util.Abc_PrintTime(1, "BDD construction time", time.Since(start).Nanoseconds())
	}
	cos := make([]*cudd.DdNode, 0, abc.Abc_NtkCoNum(pNtk))
	for i := int32(0); i < abc.Abc_NtkCoNum(pNtk); i++ {
		cos = append(cos, Abc_ObjGlobalBdd(abc.Abc_NtkCo(pNtk, i)))
	}
	base := extra.Extra_FileNameGeneric(pNtk.Spec)
	// a zero result means the LUT size is too small; the network is kept
	cas.Abc_CascadeExperiment(base, dd, cos, int(abc.Abc_NtkCiNum(pNtk)), int(abc.Abc_NtkCoNum(pNtk)), int(nLutSize), fCheck != 0, fVerbose != 0)
	result := abc.Abc_NtkDup(pNtk)
	Abc_NtkFreeGlobalBdds(pNtk, 1)
	if abc.Abc_NtkCheck(result) == 0 {
		fmt.Printf("Abc_NtkCollapse: The network check has failed.\n")
		return nil
	}
	return result
}

func bddEdgeValue(values map[*abc.Object]*cudd.DdNode, edge abc.Edge, _ *cudd.DdNode) *cudd.DdNode {
	value := values[edge.Object]
	if value == nil {
		return nil
	}
	return core.Cudd_NotCond(value, edge.Complement)
}
