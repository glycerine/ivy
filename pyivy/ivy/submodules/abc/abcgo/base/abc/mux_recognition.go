package abc

import "github.com/glycerine/abcgo/misc/vec"

func muxChildren(node *Object) (*Object, *Object, bool) {
	if Abc_AigNodeIsAnd(node) == 0 || !node.fanins[0].Complement || !node.fanins[1].Complement {
		return nil, nil, false
	}
	a, b := node.fanins[0].Object, node.fanins[1].Object
	return a, b, Abc_AigNodeIsAnd(a) != 0 && Abc_AigNodeIsAnd(b) != 0
}

func Abc_NodeIsExorType(node *Object) int32 {
	a, b, ok := muxChildren(node)
	if !ok {
		return 0
	}
	return abcBool(a.fanins[0].Object == b.fanins[0].Object && a.fanins[1].Object == b.fanins[1].Object && a.fanins[0].Complement != b.fanins[0].Complement && a.fanins[1].Complement != b.fanins[1].Complement)
}

func Abc_NodeIsMuxType(node *Object) int32 {
	a, b, ok := muxChildren(node)
	if !ok {
		return 0
	}
	for _, x := range a.fanins {
		for _, y := range b.fanins {
			if x.Object == y.Object && x.Complement != y.Complement {
				return 1
			}
		}
	}
	return 0
}

func Abc_NtkCountMuxes(n *Network) int32 {
	var count int32
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		count += Abc_NodeIsMuxType(node)
	}
	return count
}
func Abc_NtkGetMuxNum(n *Network) int32 { return Abc_NtkCountMuxes(n) }
func Abc_NtkGetExorNum(n *Network) int32 {
	var count int32
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		count += Abc_NodeIsExorType(node)
	}
	return count
}

func Abc_NodeIsMuxControlType(node *Object) int32 {
	if node == nil || len(node.fanouts) != 2 {
		return 0
	}
	a, b := node.fanouts[0], node.fanouts[1]
	return abcBool(len(a.fanouts) == 1 && len(b.fanouts) == 1 && a.fanouts[0] == b.fanouts[0])
}

// Abc_NodeRecognizeMux returns the positive control and data edges of node.
func Abc_NodeRecognizeMux(node *Object, thenEdge, elseEdge *Edge) *Object {
	a, b, ok := muxChildren(node)
	if !ok || thenEdge == nil || elseEdge == nil {
		panic("abc: invalid mux")
	}
	for i, x := range a.fanins {
		for j, y := range b.fanins {
			if x.Object != y.Object || x.Complement == y.Complement {
				continue
			}
			if x.Complement {
				*thenEdge, *elseEdge = b.fanins[1-j].Not(), a.fanins[1-i].Not()
			} else {
				*thenEdge, *elseEdge = a.fanins[1-i].Not(), b.fanins[1-j].Not()
			}
			return x.Object
		}
	}
	panic("abc: node is not a mux")
}

func Abc_NodeCollectFanins(node *Object, nodes *vec.Vec_Ptr_t) {
	vec.Vec_PtrClear(nodes)
	for _, edge := range node.fanins {
		vec.Vec_PtrPush(nodes, edge.Object)
	}
}
func Abc_NodeCollectFanouts(node *Object, nodes *vec.Vec_Ptr_t) {
	vec.Vec_PtrClear(nodes)
	for _, fanout := range node.fanouts {
		vec.Vec_PtrPush(nodes, fanout)
	}
}

// Abc_NtkCollectLatches retains ABC's behavior of returning every object.
func Abc_NtkCollectLatches(n *Network) *vec.Vec_Ptr_t { return Abc_NtkCollectObjects(n) }
