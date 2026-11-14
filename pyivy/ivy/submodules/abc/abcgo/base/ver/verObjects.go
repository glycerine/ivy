// Network and box predicates from verCore.c.
package ver

import "github.com/glycerine/abcgo/base/abc"

func Ver_NtkIsDefined(n *abc.Network) int32 {
	if n == nil {
		panic("ver: nil network")
	}
	if abc.Abc_NtkPiNum(n) != 0 || abc.Abc_NtkPoNum(n) != 0 {
		return 1
	}
	return 0
}
func Ver_ObjIsConnected(o *abc.Object) int32 {
	if o == nil || abc.Abc_ObjIsBox(o) == 0 {
		panic("ver: expected box")
	}
	if abc.Abc_ObjFaninNum(o) != 0 || abc.Abc_ObjFanoutNum(o) != 0 {
		return 1
	}
	return 0
}
