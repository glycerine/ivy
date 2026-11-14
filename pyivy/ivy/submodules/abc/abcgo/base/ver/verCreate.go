// Netlist terminal and latch construction from verCore.c.
package ver

import "github.com/glycerine/abcgo/base/abc"

func Ver_ParseCreatePi(n *abc.Network, name string) *abc.Object {
	net := abc.Abc_NtkFindOrCreateNet(n, name)
	term := abc.Abc_NtkCreatePi(n)
	abc.Abc_ObjAddFanin(net, term)
	return term
}
func Ver_ParseCreatePo(n *abc.Network, name string) *abc.Object {
	net := abc.Abc_NtkFindOrCreateNet(n, name)
	term := abc.Abc_NtkCreatePo(n)
	abc.Abc_ObjAddFanin(term, net)
	return term
}
func Ver_ParseCreateLatch(n *abc.Network, input, output *abc.Object) *abc.Object {
	term := abc.Abc_NtkCreateBi(n)
	abc.Abc_ObjAddFanin(term, input)
	latch := abc.Abc_NtkCreateLatch(n)
	abc.Abc_ObjAddFanin(latch, term)
	term = abc.Abc_NtkCreateBo(n)
	abc.Abc_ObjAddFanin(term, latch)
	abc.Abc_ObjAddFanin(output, term)
	abc.Abc_ObjAssignName(latch, abc.Abc_ObjName(output), "L")
	abc.Abc_LatchSetInitDc(latch)
	return latch
}
func Ver_ParseCreateInv(n *abc.Network, net *abc.Object) *abc.Object {
	node := abc.Abc_NtkCreateNodeInv(n, net)
	result := abc.Abc_NtkCreateNet(n)
	abc.Abc_ObjAddFanin(result, node)
	return result
}
