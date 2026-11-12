package io

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/base/abc"
)

// ReadBlifAsAIG parses the same checked netlist as ReadBLIF, then structurally
// hashes its actual combinational hierarchy. The AIGER model bridge provides
// the common native ABC strash-network construction.
func ReadBlifAsAIG(filename string, check bool) (*abc.Network, error) {
	n, err := ReadBLIF(filename, false, check)
	if err != nil {
		return nil, err
	}
	graph := abc.Abc_NtkDeriveFlatGia(n)
	if graph == nil {
		return nil, fmt.Errorf("%s: cannot derive Boolean AIG", filename)
	}
	model, err := graph.AigerModel()
	if err != nil {
		return nil, err
	}
	result, err := NetworkFromAigerModel(model)
	if err != nil {
		return nil, err
	}
	result.Name = n.Name
	result.Spec = filename
	return result, nil
}
func Io_ReadBlifAsAig(filename string, check int32) *abc.Network {
	n, err := ReadBlifAsAIG(filename, check != 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return n
}
