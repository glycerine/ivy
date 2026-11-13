package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

// Abc_NtkExpandCubes expands each SOP output against its corresponding GIA
// cone and removes node fanins when the resulting function is constant.
func Abc_NtkExpandCubes(network *abc.Network, graph *gia.Gia_Man_t, verbose int32) {
	if network == nil || graph == nil || abc.Abc_NtkIsSopLogic(network) == 0 || len(network.Inputs()) != int(gia.Gia_ManCiNum(graph)) || len(network.Outputs()) != int(gia.Gia_ManCoNum(graph)) {
		panic("bmc: incompatible SOP and GIA networks")
	}
	for i, output := range network.Outputs() {
		if abc.Abc_ObjFaninNum(output) == 0 {
			continue
		}
		node := abc.Abc_ObjFanin0(output)
		if abc.Abc_ObjIsNode(node) == 0 || abc.Abc_ObjFaninNum(node) == 0 {
			continue
		}
		if abc.Abc_ObjFaninNum(node) != gia.Gia_ManCiNum(graph) {
			panic("bmc: SOP support does not match GIA inputs")
		}
		sop, ok := node.Data.(string)
		if !ok {
			panic("bmc: node has no SOP")
		}
		raw := append([]byte(sop), 0)
		v := vec.Vec_StrAllocArrayCopy(raw, len(raw))
		cone := gia.Gia_ManDupCones(graph, []int32{int32(i)}, 1, 0)
		status := Abc_ObjExpandCubes(v, cone, abc.Abc_ObjFaninNum(node))
		if status < 0 {
			panic("bmc: cube expansion failed")
		}
		if status > 0 {
			abc.Abc_ObjRemoveFanins(node)
		}
		data := vec.Vec_StrArray(v)
		node.Data = string(data[:len(data)-1])
	}
	abc.Abc_NtkSortSops(network)
	_ = verbose
}
