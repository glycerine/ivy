package abc

import (
	"fmt"
	"github.com/glycerine/abcgo/map/mio"
	"os"
)

func Abc_ObjPrint(pFile *os.File, pObj *Object) {
	fmt.Fprintf(pFile, "Object %5d : ", Abc_ObjId(pObj))
	kind := map[ObjectKind]string{ObjectNone: "NONE   ", ObjectConst1: "Const1 ", ObjectPI: "PI     ", ObjectPO: "PO     ", ObjectBI: "BI     ", ObjectBO: "BO     ", ObjectNet: "Net    ", ObjectNode: "Node   ", ObjectLatch: "Latch  ", ObjectWhitebox: "Whitebox", ObjectBlackbox: "Blackbox"}
	label, ok := kind[pObj.Kind]
	assert(ok, "known object kind")
	fmt.Fprint(pFile, label, " Fanins ( ")
	for _, edge := range pObj.Fanins() {
		fmt.Fprintf(pFile, "%d ", Abc_ObjId(edge.Object))
	}
	fmt.Fprint(pFile, ") ")
	if Abc_ObjIsNode(pObj) != 0 && Abc_NtkHasSop(pObj.Network()) != 0 {
		fmt.Fprintf(pFile, " %s", pObj.Data)
	} else if Abc_ObjIsNode(pObj) != 0 && Abc_NtkHasMapping(pObj.Network()) != 0 {
		if gate, ok := pObj.Data.(*mio.Mio_Gate_t); ok {
			fmt.Fprintf(pFile, " %s\n", mio.Mio_GateReadName(gate))
		} else {
			fmt.Fprintln(pFile)
		}
	} else {
		fmt.Fprintln(pFile)
	}
}
