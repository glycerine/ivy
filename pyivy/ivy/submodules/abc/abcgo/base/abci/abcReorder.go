package abci

import (
	"fmt"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/reo"
)

// Abc_NodeBddReorder reorders the local BDD of node and applies the resulting
// variable permutation to the node's fanins. The shared implementation in
// base/abc owns the network and object representation.
func Abc_NodeBddReorder(reorder *reo.Manager, node *Abc_Obj_t) {
	abc.Abc_NodeBddReorder(reorder, node)
}

// Abc_NtkBddReorder reorders the local BDDs of all nodes with at least three
// fanins. Duplicate fanins are removed and each node's BDD support is reduced
// before running the REO engine, matching the C implementation.
func Abc_NtkBddReorder(network *Abc_Ntk_t, verbose int32) {
	abc.Abc_NtkRemoveDupFanins(network)
	abc.Abc_NtkMinimumBase(network)

	maxFanins := int(abc.Abc_NtkGetFaninMax(network))
	reorder := reo.Extra_ReorderInit(maxFanins, 100)
	defer reo.Extra_ReorderQuit(reorder)

	for _, node := range network.ObjectsOfKind(abc.ObjectNode) {
		if len(node.Fanins()) < 3 {
			continue
		}
		before := 0
		if verbose != 0 {
			fmt.Printf("%10s: ", node.Name())
			before = core.Cudd_DagSize(node.Data.(*core.DdNode))
			fmt.Printf("Before = %5d  BDD nodes.  ", before)
		}
		Abc_NodeBddReorder(reorder, node)
		if verbose != 0 {
			after := core.Cudd_DagSize(node.Data.(*core.DdNode))
			fmt.Printf("After = %5d  BDD nodes.\n", after)
		}
	}
}
