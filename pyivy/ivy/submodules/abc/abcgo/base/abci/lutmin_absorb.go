package abci

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

// Abc_ObjCheckAbsorb collects the fanins after replacing pivot by its fanins.
// On failure, fanins retains the partially collected set, as in ABC.
func Abc_ObjCheckAbsorb(obj, pivot *abc.Object, lutSize int, fanins *vec.Vec_Ptr_t) int32 {
	if abc.Abc_ObjIsNode(obj) == 0 || abc.Abc_ObjIsNode(pivot) == 0 {
		panic("Abc_ObjCheckAbsorb: expected nodes")
	}
	vec.Vec_PtrClear(fanins)
	for i := int32(0); i < abc.Abc_ObjFaninNum(obj); i++ {
		fanin := abc.Abc_ObjFanin(obj, i)
		if fanin != pivot {
			vec.Vec_PtrPush(fanins, fanin)
		}
	}
	for i := int32(0); i < abc.Abc_ObjFaninNum(pivot); i++ {
		vec.Vec_PtrPushUnique(fanins, abc.Abc_ObjFanin(pivot, i))
		if vec.Vec_PtrSize(fanins) > lutSize {
			return 0
		}
	}
	return 1
}

// lutAbsorbCounts follows Abc_NtkCheckAbsorb's two passes over the nodes.
func lutAbsorbCounts(network *abc.Network, lutSize int) (absorbed, fully int) {
	counts := make([]int32, abc.Abc_NtkObjNumMax(network))
	fanins := vec.Vec_PtrAlloc(100)
	nodes := network.ObjectsOfKind(abc.ObjectNode)
	for _, node := range nodes {
		for i := int32(0); i < abc.Abc_ObjFaninNum(node); i++ {
			fanin := abc.Abc_ObjFanin(node, i)
			if abc.Abc_ObjIsNode(fanin) != 0 && Abc_ObjCheckAbsorb(node, fanin, lutSize, fanins) != 0 {
				counts[abc.Abc_ObjId(fanin)]++
				absorbed++
			}
		}
	}
	for _, node := range nodes {
		if counts[abc.Abc_ObjId(node)] == abc.Abc_ObjFanoutNum(node) {
			fully++
		}
	}
	return
}

// Abc_NtkCheckAbsorb reports the source's LUT absorption statistics.
func Abc_NtkCheckAbsorb(network *abc.Network, lutSize int) {
	started := time.Now()
	absorbed, fully := lutAbsorbCounts(network, lutSize)
	nodes := float64(abc.Abc_NtkNodeNum(network))
	fmt.Printf("Absorted = %6d. (%6.2f %%)   Fully = %6d. (%6.2f %%)  Time = %.2f sec\n", absorbed, 100*float64(absorbed)/nodes, fully, 100*float64(fully)/nodes, time.Since(started).Seconds())
}
