package abc

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/misc/vec"
)

func Abc_NtkTopoHasBeg(node *Object) int32 { return node.network.Topo[2*node.ID] }
func Abc_NtkTopoHasEnd(node *Object) int32 { return node.network.Topo[2*node.ID+1] }
func Abc_NtkTopoSetBeg(node *Object)       { node.network.Topo[2*node.ID] = int32(len(node.network.Topo)) }
func Abc_NtkTopoSetEnd(node *Object)       { node.network.Topo[2*node.ID+1] = int32(len(node.network.Topo)) }

func Abc_NtkReverseTopoOrder_rec(node *Object, pivot int32) {
	if Abc_NodeIsTravIdCurrent(node) != 0 {
		return
	}
	Abc_NodeSetTravIdCurrent(node)
	n := node.network
	if node.Kind == ObjectPO {
		n.Topo = append(n.Topo, int32(node.ID))
		return
	}
	if node.Kind != ObjectNode {
		panic("abc: reverse topological traversal requires logic nodes")
	}
	if pivot != 0 {
		Abc_NtkTopoSetBeg(node)
	}
	var first *Object
	for _, fanout := range node.fanouts {
		if Abc_NtkTopoHasBeg(fanout) == 0 {
			if Abc_NtkTopoHasEnd(fanout) != 0 {
				panic("abc: invalid topological interval")
			}
			Abc_NtkReverseTopoOrder_rec(fanout, 1)
			first = fanout
			break
		}
	}
	for _, fanout := range node.fanouts {
		if fanout != first {
			Abc_NtkReverseTopoOrder_rec(fanout, 0)
		}
	}
	if pivot != 0 {
		Abc_NtkTopoSetEnd(node)
	}
	n.Topo = append(n.Topo, int32(node.ID))
}

func Abc_NtkReverseTopoOrder(n *Network) {
	if n.Topo != nil {
		panic("abc: reverse topological index already exists")
	}
	n.Topo = make([]int32, 2*len(n.objects))
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if Abc_NtkTopoHasBeg(node) != 0 {
			continue
		}
		Abc_NtkIncrementTravId(n)
		Abc_NtkReverseTopoOrder_rec(node, 1)
	}
	ratio := 0.0
	if count := len(n.ObjectsOfKind(ObjectNode)); count != 0 {
		ratio = float64(len(n.Topo)) / float64(count)
	}
	fmt.Printf("Nodes = %d.   Size = %d.  Ratio = %f.\n", len(n.ObjectsOfKind(ObjectNode)), len(n.Topo), ratio)
}

func Abc_NtkReverse_rec(node *Object, visited *vec.Vec_Int_t) {
	if Abc_NodeIsTravIdCurrent(node) != 0 {
		return
	}
	Abc_NodeSetTravIdCurrent(node)
	for _, fanout := range node.fanouts {
		Abc_NtkReverse_rec(fanout, visited)
	}
	vec.Vec_IntPush(visited, int32(node.ID))
}

func Abc_NtkReverseTopoOrderTest(n *Network) {
	started := time.Now()
	Abc_NtkReverseTopoOrder(n)
	n.Topo = nil
	fmt.Printf("Time = %s\n", time.Since(started))
	started = time.Now()
	visited := vec.Vec_IntAlloc(1000)
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		vec.Vec_IntClear(visited)
		Abc_NtkIncrementTravId(n)
		Abc_NtkReverse_rec(node, visited)
	}
	fmt.Printf("Time = %s\n", time.Since(started))
}
