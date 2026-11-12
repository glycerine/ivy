package io

import (
	"fmt"
	"io"

	"github.com/glycerine/abcgo/base/abc"
)

func verilogNames(w io.Writer, objects []*abc.Object, start int32) {
	length, written := int(start), 0
	for i, obj := range objects {
		name := Io_WriteVerilogGetName(abc.Abc_ObjName(obj))
		added := len(name) + 2
		if written != 0 && length+added+3 > 78 {
			fmt.Fprint(w, "\n   ")
			length, written = 3, 0
		}
		fmt.Fprint(w, " ", name)
		if i+1 < len(objects) {
			fmt.Fprint(w, ",")
		}
		length += added
		written++
	}
}

func Io_WriteVerilogPis(w io.Writer, n *abc.Network, start int32) {
	var nets []*abc.Object
	for _, port := range n.ObjectsOfKind(abc.ObjectPI) {
		nets = append(nets, abc.Abc_ObjFanout0(port))
	}
	verilogNames(w, nets, start)
}

func Io_WriteVerilogPos(w io.Writer, n *abc.Network, start int32) {
	var nets []*abc.Object
	var skipped *abc.Object
	count := 0
	for _, port := range n.ObjectsOfKind(abc.ObjectPO) {
		net := abc.Abc_ObjFanin0(port)
		if abc.Abc_ObjIsPi(abc.Abc_ObjFanin0(net)) != 0 {
			count++
			skipped = net
			continue
		}
		nets = append(nets, net)
	}
	verilogNames(w, nets, start)
	if count != 0 {
		fmt.Printf("Io_WriteVerilogPos(): Omitted %d feedthrough nets from output list of module (e.g. %s).\n", count, abc.Abc_ObjName(skipped))
	}
}

func verilogInternalNets(n *abc.Network) []*abc.Object {
	var nets []*abc.Object
	isOutput := func(net *abc.Object) bool {
		return abc.Abc_ObjFanoutNum(net) > 0 && abc.Abc_ObjIsCo(abc.Abc_ObjFanout0(net)) != 0
	}
	for i, node := range n.ObjectsOfKind(abc.ObjectNode) {
		if i == 0 {
			continue
		}
		net := abc.Abc_ObjFanout0(node)
		if !isOutput(net) {
			nets = append(nets, net)
		}
	}
	for _, latch := range n.ObjectsOfKind(abc.ObjectLatch) {
		nets = append(nets, abc.Abc_ObjFanin0(abc.Abc_ObjFanin0(latch)))
	}
	for _, box := range n.Boxes() {
		if box.Kind == abc.ObjectLatch {
			continue
		}
		for _, term := range box.Fanins() {
			nets = append(nets, abc.Abc_ObjFanin0(term.Object))
		}
		for _, term := range box.Fanouts() {
			net := abc.Abc_ObjFanout0(term)
			if !isOutput(net) {
				nets = append(nets, net)
			}
		}
	}
	return nets
}

func Io_WriteVerilogWiresCount(n *abc.Network) int32 { return int32(len(verilogInternalNets(n))) }
func Io_WriteVerilogWires(w io.Writer, n *abc.Network, start int32) {
	verilogNames(w, verilogInternalNets(n), start)
}
func Io_WriteVerilogRegs(w io.Writer, n *abc.Network, start int32) {
	var nets []*abc.Object
	for _, latch := range n.ObjectsOfKind(abc.ObjectLatch) {
		nets = append(nets, abc.Abc_ObjFanout0(abc.Abc_ObjFanout0(latch)))
	}
	verilogNames(w, nets, start)
}
