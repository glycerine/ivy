package io

import (
	"fmt"
	stdio "io"
	"os"
	"strings"

	"github.com/glycerine/abcgo/base/abc"
)

func sopEquation(node *abc.Object) (string, error) {
	text, ok := node.Data.(string)
	if !ok {
		return "", fmt.Errorf("equation output requires SOP node functions")
	}
	inputs := node.Fanins()
	var cubes []string
	phase := ""
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		pattern, value := "", ""
		if len(fields) == 1 {
			value = fields[0]
		} else if len(fields) == 2 {
			pattern, value = fields[0], fields[1]
		} else {
			return "", fmt.Errorf("invalid SOP row")
		}
		if value != "0" && value != "1" {
			return "", fmt.Errorf("invalid SOP phase")
		}
		if phase != "" && phase != value {
			return "", fmt.Errorf("mixed SOP phases")
		}
		phase = value
		if pattern != "" && len(pattern) != len(inputs) {
			return "", fmt.Errorf("SOP width mismatch")
		}
		var literals []string
		for i, c := range pattern {
			if c == '-' {
				continue
			}
			if c != '0' && c != '1' {
				return "", fmt.Errorf("invalid SOP literal")
			}
			signal := blifObjectName(inputs[i].Object)
			if inputs[i].Complement != (c == '0') {
				signal = "!" + signal
			}
			literals = append(literals, signal)
		}
		cube := "1"
		if len(literals) > 0 {
			cube = "(" + strings.Join(literals, "*") + ")"
		}
		cubes = append(cubes, cube)
	}
	if len(cubes) == 0 {
		return "0", nil
	}
	result := "(" + strings.Join(cubes, "+") + ")"
	if phase == "0" {
		result = "!" + result
	}
	return result, nil
}
func WriteEqn(w stdio.Writer, n *abc.Network) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	if len(n.Boxes()) > 0 {
		return fmt.Errorf("EQN requires a combinational flattened network")
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
