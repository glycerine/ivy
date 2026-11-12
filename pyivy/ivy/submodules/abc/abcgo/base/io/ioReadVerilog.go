package io

import (
	"fmt"
	"sync"

	"github.com/glycerine/abcgo/base/abc"
)

var verilogParser struct {
	sync.RWMutex
	parse func(string, int32) *abc.Design
}

// RegisterVerilogParser connects the Verilog parser without an import cycle.
func RegisterVerilogParser(parse func(string, int32) *abc.Design) {
	verilogParser.Lock()
	verilogParser.parse = parse
	verilogParser.Unlock()
}

// Io_ReadVerilog selects the first top-level module from a parsed design.
func Io_ReadVerilog(filename string, check int32) *abc.Network {
	verilogParser.RLock()
	parse := verilogParser.parse
	verilogParser.RUnlock()
	if parse == nil {
		panic("io: Verilog parser is not registered")
	}
	design := parse(filename, check)
	if design == nil {
		return nil
	}
	count := abc.Abc_DesFindTopLevelModels(design)
	if count == 0 || len(design.Modules) == 0 {
		abc.Abc_DesFree(design, nil)
		return nil
	}
	root := design.Tops[0]
	if count > 1 {
		fmt.Printf("Warning: The design has %d root-level modules:", count)
		for _, top := range design.Tops {
			fmt.Printf(" %s", top.Name)
		}
		fmt.Printf("\nThe first one (%s) will be used.\n", root.Name)
	}
	root.Design = design
	design.FunctionManager = nil
	if len(design.Modules) == 1 {
		abc.Abc_DesFree(design, root)
		root.Design = nil
		root.Spec = filename
	} else {
		abc.Abc_NtkIsAcyclicHierarchy(root)
	}
	return root
}

func init() {
	RegisterFormat(IO_FILE_VERILOG, func(filename string, check bool) (*abc.Network, error) {
		checkInt := int32(0)
		if check {
			checkInt = 1
		}
		n := Io_ReadVerilog(filename, checkInt)
		if n == nil {
			return nil, fmt.Errorf("%s: Verilog parse failed", filename)
		}
		return n, nil
	}, nil)
}
