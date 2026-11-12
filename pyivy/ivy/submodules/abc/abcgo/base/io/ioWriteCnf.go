package io

import (
	"fmt"
	stdio "io"
	"os"

	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/base/abc"
)

func WriteCNF(w stdio.Writer, n *abc.Network) error {
	if n == nil {
		return fmt.Errorf("nil network")
	}
	if len(n.ObjectsOfKind(abc.ObjectLatch)) > 0 {
		return fmt.Errorf("CNF requires a combinational network")
	}
	var model *ioa.Model
	var err error
	if n.Type == uint8(abc.ABC_NTK_STRASH) {
		model, err = aigerModelFromNetwork(n)
	} else {
		netlist := n
		if n.Type != uint8(abc.ABC_NTK_NETLIST) {
			netlist = abc.Abc_NtkToNetlist(n)
			if netlist == nil {
				return fmt.Errorf("cannot convert logic to netlist")
			}
		}
		graph := abc.Abc_NtkDeriveFlatGia(netlist)
		if graph == nil {
			return fmt.Errorf("cannot derive Boolean AIG")
		}
		model, err = graph.AigerModel()
		if err == nil {
			if model.Symbols == nil {
				model.Symbols = map[ioa.Symbol]string{}
			}
			for i, port := range n.ObjectsOfKind(abc.ObjectPI) {
				model.Symbols[ioa.Symbol{Kind: 'i', Index: uint32(i)}] = port.Name()
			}
		}
	}
	if err != nil {
		return err
	}
	return WriteAigerCNF(w, model)
}
func WriteCNFFile(n *abc.Network, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = WriteCNF(file, n)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func Io_WriteCnf(n *abc.Network, filename string, allPrimes int32) int32 {
	if err := WriteCNFFile(n, filename); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 0
	}
	return 1
}

// Explicit network context replaces the source's process-global current-miter
// pointer, so concurrent exports cannot mix PI names and variable mappings.
func Io_WriteCnfOutputPiMapping(w stdio.Writer, increment int32, n *abc.Network) error {
	model, err := aigerModelFromNetwork(n)
	if err != nil {
		return err
	}
	if _, err = fmt.Fprintln(w, "c PI variable numbers: <PI_name> <SAT_var_number>"); err != nil {
		return err
	}
	offset := int32(0)
	if increment > 0 {
		offset = 1
	}
	for i, literal := range model.Inputs {
		if _, err = fmt.Fprintf(w, "c %s %d\n", blifObjectName(n.ObjectsOfKind(abc.ObjectPI)[i]), int32(literal>>1)-1+offset); err != nil {
			return err
		}
	}
	return nil
}
func init() { RegisterFormat(IO_FILE_CNF, nil, WriteCNFFile) }
