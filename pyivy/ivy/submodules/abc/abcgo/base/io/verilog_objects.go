package io

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/mio"
)

func Io_WriteVerilogObjects(w io.Writer, n *abc.Network, onlyAnds int32) {
	boxes := n.Boxes()
	count := 0
	for _, box := range boxes {
		if box.Kind != abc.ObjectLatch {
			count++
		}
	}
	width := len(strconv.Itoa(count))
	index := 0
	for _, box := range boxes {
		if box.Kind == abc.ObjectLatch {
			continue
		}
		model, ok := box.Data.(*abc.Network)
		if !ok || model == nil {
			panic("io: box has no network model")
		}
		fmt.Fprintf(w, "  %s box%0*d(", model.Name, width, index)
		index++
		var pins []string
		for i, port := range model.ObjectsOfKind(abc.ObjectPI) {
			formal := abc.Abc_ObjFanout0(port)
			actual := abc.Abc_ObjFanin0(abc.Abc_ObjFanin(box, int32(i)))
			pins = append(pins, fmt.Sprintf(".%s(%s)", Io_WriteVerilogGetName(abc.Abc_ObjName(formal)), Io_WriteVerilogGetName(abc.Abc_ObjName(actual))))
		}
		for i, port := range model.ObjectsOfKind(abc.ObjectPO) {
			formal := abc.Abc_ObjFanin0(port)
			actual := abc.Abc_ObjFanout0(abc.Abc_ObjFanout(box, int32(i)))
			pins = append(pins, fmt.Sprintf(".%s(%s)", Io_WriteVerilogGetName(abc.Abc_ObjName(formal)), Io_WriteVerilogGetName(abc.Abc_ObjName(actual))))
		}
		fmt.Fprintln(w, strings.Join(pins, ", ")+");")
	}
	nodes := n.ObjectsOfKind(abc.ObjectNode)
	if abc.Abc_NtkHasMapping(n) != 0 {
		lib, ok := n.FunctionManager.(*mio.Mio_Library_t)
		if !ok || lib == nil {
			panic("io: mapped network has no gate library")
		}
		length := mio.Mio_LibraryReadGateNameMax(lib)
		width = len(strconv.Itoa(len(nodes)))
		index = 0
		for _, node := range nodes {
			gate, ok := node.Data.(*mio.Mio_Gate_t)
			if !ok || gate == nil {
				panic("io: mapped node has no gate")
			}
			name := mio.Mio_GateReadName(gate)
			out := Io_WriteVerilogGetName(abc.Abc_ObjName(abc.Abc_ObjFanout0(node)))
			if abc.Abc_ObjFaninNum(node) == 0 && (name == "_const0_" || name == "_const1_") {
				value := 0
				if name == "_const1_" {
					value = 1
				}
				fmt.Fprintf(w, "  %-*s %s = 1'b%d;\n", length, "assign", out, value)
				continue
			}
			fmt.Fprintf(w, "  %-*s g%0*d(", length, name, width, index)
			index++
			i := int32(0)
			for pin := mio.Mio_GateReadPins(gate); pin != nil; pin = mio.Mio_PinReadNext(pin) {
				if i >= abc.Abc_ObjFaninNum(node) {
					panic("io: gate pin count mismatch")
				}
				fmt.Fprintf(w, ".%s(%s), ", Io_WriteVerilogGetName(mio.Mio_PinReadName(pin)), Io_WriteVerilogGetName(abc.Abc_ObjName(abc.Abc_ObjFanin(node, i))))
				i++
			}
			if i != abc.Abc_ObjFaninNum(node) {
				panic("io: gate pin count mismatch")
			}
			fmt.Fprintf(w, ".%s(%s));\n", Io_WriteVerilogGetName(mio.Mio_GateReadOutName(gate)), out)
		}
		return
	}
	manager, ok := n.FunctionManager.(*hop.Hop_Man_t)
	if !ok || manager == nil {
		panic("io: network has no AIG function manager")
	}
	for _, node := range nodes {
		function, ok := node.Data.(*hop.Hop_Obj_t)
		if !ok || function == nil {
			panic("io: node has no AIG function")
		}
		fmt.Fprintf(w, "  assign %s = ", Io_WriteVerilogGetName(abc.Abc_ObjName(abc.Abc_ObjFanout0(node))))
		for i, fanin := range node.Fanins() {
			hop.Hop_ObjSetData(hop.Hop_IthVar(manager, int32(i)), Io_WriteVerilogGetName(abc.Abc_ObjName(fanin.Object)))
		}
		hop.Hop_ObjPrintVerilog(w, function, nil, 0, onlyAnds)
		fmt.Fprintln(w, ";")
		for i := range node.Fanins() {
			hop.Hop_ObjSetData(hop.Hop_IthVar(manager, int32(i)), nil)
		}
	}
}
