// Standard gate primitives from verCore.c.
package ver

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
)

func Ver_ParseGateStandard(p *Ver_Man_t, n *abc.Network, kind Ver_GateType_t) int32 {
	if Ver_ParseConvertNetwork(p, n, p.fMapped) == 0 {
		return 0
	}
	Ver_StreamMove(p.pReader)
	if Ver_StreamPopChar(p.pReader) != '(' {
		return verRangeError(p, "Cannot parse a standard gate (expected opening parenthesis).")
	}
	Ver_ParseSkipComments(p)
	node := abc.Abc_NtkCreateNode(n)
	for {
		name := Ver_ParseGetName(p)
		if name == nil {
			return 0
		}
		net := Ver_ParseFindNet(n, *name)
		if net == nil {
			return verRangeError(p, fmt.Sprintf("Net is missing in gate %s.", *name))
		}
		if abc.Abc_ObjFanoutNum(node) == 0 {
			abc.Abc_ObjAddFanin(net, node)
		} else {
			abc.Abc_ObjAddFanin(node, net)
		}
		Ver_ParseSkipComments(p)
		symbol := Ver_StreamPopChar(p.pReader)
		if symbol == ')' {
			break
		}
		if symbol != ',' {
			return verRangeError(p, fmt.Sprintf("Cannot parse a standard gate %s (expected closing parenthesis).", abc.Abc_ObjName(abc.Abc_ObjFanout0(node))))
		}
		Ver_ParseSkipComments(p)
	}
	if (kind == VER_GATE_BUF || kind == VER_GATE_NOT) && abc.Abc_ObjFaninNum(node) != 1 {
		return verRangeError(p, fmt.Sprintf("Buffer or interver with multiple fanouts %s (currently not supported).", abc.Abc_ObjName(abc.Abc_ObjFanout0(node))))
	}
	Ver_ParseSkipComments(p)
	if Ver_StreamPopChar(p.pReader) != ';' {
		return verRangeError(p, fmt.Sprintf("Cannot read standard gate %s (expected closing semicolumn).", abc.Abc_ObjName(abc.Abc_ObjFanout0(node))))
	}
	manager := n.FunctionManager.(*hop.Hop_Man_t)
	count := abc.Abc_ObjFaninNum(node)
	var function *hop.Hop_Obj_t
	switch kind {
	case VER_GATE_AND, VER_GATE_NAND, VER_GATE_BUF, VER_GATE_NOT:
		function = hop.Hop_CreateAnd(manager, count)
	case VER_GATE_OR, VER_GATE_NOR:
		function = hop.Hop_CreateOr(manager, count)
	case VER_GATE_XOR, VER_GATE_XNOR:
		function = hop.Hop_CreateExor(manager, count)
	}
	if kind == VER_GATE_NAND || kind == VER_GATE_NOR || kind == VER_GATE_XNOR || kind == VER_GATE_NOT {
		function = hop.Hop_Not(function)
	}
	node.Data = function
	return 1
}

func Ver_ParseFlopStandard(p *Ver_Man_t, n *abc.Network) int32 {
	if Ver_ParseConvertNetwork(p, n, p.fMapped) == 0 {
		return 0
	}
	if Ver_StreamPopChar(p.pReader) != '(' {
		return verRangeError(p, "Cannot parse a standard gate (expected opening parenthesis).")
	}
	Ver_ParseSkipComments(p)
	name := Ver_ParseGetName(p)
	if name == nil {
		return 0
	}
	output := Ver_ParseFindNet(n, *name)
	if output == nil {
		return verRangeError(p, fmt.Sprintf("Net is missing in gate %s.", *name))
	}
	Ver_ParseSkipComments(p)
	if Ver_StreamPopChar(p.pReader) != ',' {
		return verRangeError(p, "Cannot parse the flop.")
	}
	Ver_ParseSkipComments(p)
	name = Ver_ParseGetName(p)
	if name == nil {
		return 0
	}
	input := Ver_ParseFindNet(n, *name)
	if input == nil {
		return verRangeError(p, fmt.Sprintf("Net is missing in gate %s.", *name))
	}
	Ver_ParseSkipComments(p)
	if Ver_StreamPopChar(p.pReader) != ')' {
		return verRangeError(p, "Cannot parse the flop.")
	}
	Ver_ParseSkipComments(p)
	if Ver_StreamPopChar(p.pReader) != ';' {
		return verRangeError(p, "Cannot parse the flop.")
	}
	abc.Abc_LatchSetInit0(Ver_ParseCreateLatch(n, input, output))
	return 1
}
