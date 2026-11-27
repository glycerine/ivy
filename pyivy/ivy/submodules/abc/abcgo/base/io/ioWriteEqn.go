package io

import (
	"bytes"
	"errors"
	"fmt"
	stdio "io"
	"os"
	"strings"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
)

// WriteEqn writes a netlist in EQN form (Io_WriteEqn); a netlist with other
// functions is converted to AIG first, as Io_Write does. When a name cannot
// appear in EQN it prints ABC's message, writes nothing and returns
// errEqnNames.
func WriteEqn(w stdio.Writer, n *abc.Network) error {
	data, err := ioWriteEqnData(n)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

var errEqnNames = errors.New("io: names cannot be written in EQN")

func ioWriteEqnData(n *abc.Network) ([]byte, error) {
	if n == nil {
		return nil, fmt.Errorf("nil network")
	}
	if abc.Abc_NtkHasAig(n) == 0 && abc.Abc_NtkToAig(n) == 0 {
		return nil, fmt.Errorf("io: cannot convert network to AIG functions")
	}
	if abc.Abc_NtkLatchNum(n) > 0 {
		fmt.Printf("Warning: only combinational portion is being written.\n")
	}
	if !Io_NtkWriteEqnCheck(n) {
		return nil, errEqnNames
	}
	var out bytes.Buffer
	fmt.Fprintf(&out, "# Equations for \"%s\" written by ABC on %s\n", n.Name, extra.Extra_TimeStamp())
	Io_NtkWriteEqnOne(&out, n)
	out.WriteString("\n")
	return out.Bytes(), nil
}

// WriteEqnFile is Io_Write's EQN branch: netlist, AIG functions, Io_WriteEqn.
func WriteEqnFile(n *abc.Network, filename string) error {
	net := abc.Abc_NtkToNetlist(n)
	if net == nil {
		fmt.Printf("Converting to netlist has failed.\n")
		return nil
	}
	defer abc.Abc_NtkDelete(net)
	data, err := ioWriteEqnData(net)
	if err == errEqnNames {
		return nil
	}
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// Io_NtkWriteEqnOne writes the orders and one equation per node.
func Io_NtkWriteEqnOne(w *bytes.Buffer, n *abc.Network) {
	w.WriteString("INORDER =")
	Io_NtkWriteEqnCis(w, n)
	w.WriteString(";\n")
	w.WriteString("OUTORDER =")
	Io_NtkWriteEqnCos(w, n)
	w.WriteString(";\n")
	manager := n.FunctionManager.(*hop.Hop_Man_t)
	levels := hop.Vec_VecAlloc(10)
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		if abc.Abc_ObjId(node) == 0 {
			continue
		}
		fmt.Fprintf(w, "%s = ", abc.Abc_ObjName(abc.Abc_ObjFanout0(node)))
		for i, fanin := range node.Fanins() {
			hop.Hop_ObjSetData(hop.Hop_IthVar(manager, int32(i)), abc.Abc_ObjName(fanin.Object))
		}
		hop.Hop_ObjPrintEqn(w, node.Data.(*hop.Hop_Obj_t), levels, 0)
		w.WriteString(";\n")
	}
}

func ioWriteEqnNames(w *bytes.Buffer, names []string, start int) {
	lineLength, nameCounter := start, 0
	for _, name := range names {
		added := len(name) + 1
		if nameCounter != 0 && lineLength+added+3 > IO_WRITE_LINE_LENGTH {
			w.WriteString(" \n")
			lineLength, nameCounter = 0, 0
		}
		w.WriteString(" " + name)
		lineLength += added
		nameCounter++
	}
}

// Io_NtkWriteEqnCis writes the CI nets, wrapping lines as ABC does.
func Io_NtkWriteEqnCis(w *bytes.Buffer, n *abc.Network) {
	var names []string
	for i := int32(0); i < abc.Abc_NtkCiNum(n); i++ {
		names = append(names, abc.Abc_ObjName(abc.Abc_ObjFanout0(abc.Abc_NtkCi(n, i))))
	}
	ioWriteEqnNames(w, names, 9)
}

// Io_NtkWriteEqnCos writes the CO nets, wrapping lines as ABC does.
func Io_NtkWriteEqnCos(w *bytes.Buffer, n *abc.Network) {
	var names []string
	for i := int32(0); i < abc.Abc_NtkCoNum(n); i++ {
		names = append(names, abc.Abc_ObjName(abc.Abc_ObjFanin0(abc.Abc_NtkCo(n, i))))
	}
	ioWriteEqnNames(w, names, 10)
}

// Io_NtkWriteEqnCheck rejects names EQN cannot carry (a leading 0 or 1, or
// any of "()!*+"), printing ABC's message for the first one.
func Io_NtkWriteEqnCheck(n *abc.Network) bool {
	for i := int32(0); i < abc.Abc_NtkObjNumMax(n); i++ {
		obj := abc.Abc_NtkObj(n, i)
		if obj == nil {
			continue
		}
		name := obj.Name()
		if name == "" {
			continue
		}
		if name[0] == '0' || name[0] == '1' || strings.ContainsAny(name, "()!*+") {
			fmt.Printf("The network cannot be written in the EQN format because object %d has name \"%s\".\n", i, name)
			fmt.Printf("Consider renaming the objects using command \"short_names\" and trying again.\n")
			return false
		}
	}
	return true
}
func Io_WriteEqn(n *abc.Network, filename string) error { return WriteEqnFile(n, filename) }
func init()                                             { RegisterFormat(IO_FILE_EQN, nil, WriteEqnFile) }
