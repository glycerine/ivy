package io

import (
	"fmt"
	stdio "io"
	"os"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/amap/formula"
)

func WriteEqn(w stdio.Writer, n *abc.Network) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	if len(n.Boxes()) > 0 {
		return fmt.Errorf("EQN requires a combinational flattened network")
	}
	for _, signal := range n.ObjectsOfKind(abc.ObjectNet) {
		name := blifObjectName(signal)
		expression, err := formula.Parse(name)
		if err != nil {
			return fmt.Errorf("signal %q cannot be represented in EQN", name)
		}
		variables := expression.Variables()
		if len(variables) != 1 || variables[0] != name {
			return fmt.Errorf("signal %q cannot be represented in EQN", name)
		}
	}

	var out strings.Builder
	out.WriteString("INORDER =")
	for _, port := range n.ObjectsOfKind(abc.ObjectPI) {
		fmt.Fprint(&out, " ", blifPortName(port, true))
	}
	out.WriteString(";\nOUTORDER =")
	for _, port := range n.ObjectsOfKind(abc.ObjectPO) {
		fmt.Fprint(&out, " ", blifPortName(port, false))
	}
	out.WriteString(";\n")
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		outputs := node.Fanouts()
		if len(outputs) != 1 {
			return fmt.Errorf("node must drive one net")
		}
		equation, err := sopEquation(node)
		if err != nil {
			return err
		}
		fmt.Fprintf(&out, "%s = %s;\n", blifObjectName(outputs[0]), equation)
	}
	return writeText(w, out.String())
}
func WriteEqnFile(n *abc.Network, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = WriteEqn(file, n)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func Io_WriteEqn(n *abc.Network, filename string) error { return WriteEqnFile(n, filename) }
func init()                                             { RegisterFormat(IO_FILE_EQN, nil, WriteEqnFile) }
