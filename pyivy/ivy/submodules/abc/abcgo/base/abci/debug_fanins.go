package abci

// Debug fanin enumeration from src/base/abci/abcDebug.c.
import "github.com/glycerine/abcgo/base/abc"

func debugForEachFanin(n *abc.Network, visit func(*abc.Object, *abc.Object) bool) {
	for _, obj := range n.Objects() {
		if abc.Abc_ObjIsNode(obj) == 0 && abc.Abc_ObjIsPo(obj) == 0 {
			continue
		}
		if abc.Abc_ObjIsPo(obj) != 0 && abc.Abc_NtkPoNum(n) == 1 {
			continue
		}
		for _, edge := range obj.Fanins() {
			if abc.Abc_ObjIsNode(obj) != 0 && abc.Abc_NodeIsConst(edge.Object) != 0 {
				continue
			}
			if !visit(obj, edge.Object) {
				return
			}
		}
	}
}
func Abc_NtkCountFaninsTotal(n *abc.Network) int32 {
	var count int32
	debugForEachFanin(n, func(*abc.Object, *abc.Object) bool { count++; return true })
	return count
}
func Abc_NtkFindGivenFanin(n *abc.Network, step int32, outObj, outFanin **abc.Object) int32 {
	var count int32
	var found bool
	debugForEachFanin(n, func(obj, fanin *abc.Object) bool {
		if count == step {
			*outObj, *outFanin = obj, fanin
			found = true
			return false
		}
		count++
		return true
	})
	if found {
		return 1
	}
	return 0
}
