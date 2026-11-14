package abci

import (
	"fmt"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cas"
	"github.com/glycerine/abcgo/bdd/cudd"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/extra"
)

// Abc_NtkCascade decomposes the shared output BDDs into a LUT cascade.
// As in ABC's original implementation, this currently returns a checked copy
// of the input network; the experiment emits the cascade files as a side effect.
func Abc_NtkCascade(pNtk *Abc_Ntk_t, nLutSize, fCheck, fVerbose int32) *Abc_Ntk_t {
	if pNtk == nil || abc.Abc_NtkIsStrash(pNtk) == 0 {
		return nil
	}

	inputs := pNtk.Inputs()
	outputs := pNtk.Outputs()

	// Build the same global BDD representation used by the other ABC BDD
	// commands. The 500000-node guard and dynamic reordering match abcCas.c.
	dd := core.Cudd_Init(uint32(len(inputs)), 0, core.CUDD_UNIQUE_SLOTS, core.CUDD_CACHE_SLOTS, 0)
	if dd == nil {
		return nil
	}
	defer core.Cudd_Quit(dd)
	cudd.Cudd_AutodynEnable(dd, cudd.CUDD_REORDER_SYMM_SIFT)

	values := make(map[*abc.Object]*cudd.DdNode, len(pNtk.Objects()))
	for i, input := range inputs {
		values[input] = cudd.Cudd_bddIthVar(dd, int32(i))
	}
	objects, err := pNtk.Topological(outputs)
	if err != nil {
		return nil
	}
	one := cudd.Cudd_ReadOne(dd)
	for _, object := range objects {
		if object.Kind == abc.ObjectConst1 {
			values[object] = one
			continue
		}
		if object.IsCI() {
			continue
		}
		fanins := object.Fanins()
		if object.IsCO() {
			if len(fanins) != 1 {
				return nil
			}
			values[object] = bddEdgeValue(values, fanins[0], one)
			continue
		}
		if len(fanins) == 0 {
			values[object] = one
			continue
		}
		value := one
		for _, fanin := range fanins {
			value = core.Cudd_bddAnd(dd, value, bddEdgeValue(values, fanin, one))
			if value == nil {
				return nil
			}
		}
		values[object] = value
		if cudd.Cudd_ReadKeys(dd)-cudd.Cudd_ReadDead(dd) > 500000 {
			return nil
		}
	}

	globalOutputs := make([]*cudd.DdNode, len(outputs))
	for i, output := range outputs {
		value, ok := values[output]
		if !ok || value == nil {
			return nil
		}
		globalOutputs[i] = value
	}
	if fVerbose != 0 {
		fmt.Printf("Shared BDD size = %6d nodes.\n", cudd.Cudd_ReadKeys(dd)-cudd.Cudd_ReadDead(dd))
	}
	base := extra.Extra_FileNameGeneric(pNtk.Spec)
	cas.Abc_CascadeExperiment(base, dd, globalOutputs, len(inputs), len(outputs), int(nLutSize), fCheck != 0, fVerbose != 0)

	result := abc.Abc_NtkDup(pNtk)
	if result == nil || result.Check() != nil {
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
