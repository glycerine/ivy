// Net lookup from verCore.c.
package ver

import "github.com/glycerine/abcgo/base/abc"

func Ver_ParseFindNet(n *abc.Network, name string) *abc.Object {
	if net := abc.Abc_NtkFindNet(n, name); net != nil {
		return net
	}
	if name == "1'b0" || name == "1'bx" {
		return abc.Abc_NtkFindOrCreateNet(n, "1'b0")
	}
	if name == "1'b1" {
		return abc.Abc_NtkFindOrCreateNet(n, "1'b1")
	}
	return nil
}

func Ver_ParseFindOrCreateNetwork(p *Ver_Man_t, name string) *abc.Network {
	if n := abc.Abc_DesFindModelByName(p.pDesign, name); n != nil {
		return n
	}
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_NETLIST, abc.ABC_FUNC_BLACKBOX, p.fUseMemMan)
	n.Name = name
	abc.Abc_DesAddModel(p.pDesign, n)
	return n
}
