package abc

// Owned reached-state BLIF output from src/bdd/llb/llb2Dump.c.

import (
	"fmt"
	"strings"
)

type LLBDumpOptions struct {
	MaxBytes, MaxNodes, MaxOperations int
	Stop                              func() bool
}

// LLBDummyName is the owned and reentrant form of Llb_ManGetDummyName.
func LLBDummyName(prefix string, number, digits int) (string, error) {
	if strings.IndexByte(prefix, 0) >= 0 || number < 0 || digits < 0 || digits > 18 {
		return "", ErrDimension
	}
	return prefix + fmt.Sprintf("%0*d", digits, number), nil
}

// LLBMarshalReachedBLIF emits the reached-state function with the source's
// ffNNN input names and Reached output name. The fixed-order manager makes the
// source's in-place heap reordering unnecessary and keeps frozen managers safe
// for concurrent readers.
func LLBMarshalReachedBLIF(manager *ExtraBDDManager, reached ExtraBDDLit, model string, options LLBDumpOptions) ([]byte, error) {
	if err := manager.checkLiteral(reached); err != nil {
		return nil, err
	}
	if options.MaxBytes < 0 || options.MaxNodes < 0 || options.MaxOperations < 0 {
		return nil, ErrDimension
	}
	digits := HopBase10Log(uint(manager.variables))
	inputNames := make([]string, manager.variables)
	for variable := range inputNames {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		name, err := LLBDummyName("ff", variable, digits)
		if err != nil {
			return nil, err
		}
		inputNames[variable] = name
	}
	text, err := CuddDumpBDDs(manager, []ExtraBDDLit{reached}, CuddDumpOptions{
		Format: CuddDumpBLIF, ModelName: model, InputNames: inputNames, OutputNames: []string{"Reached"},
		MaxBytes: options.MaxBytes, MaxNodes: options.MaxNodes, MaxOperations: options.MaxOperations, Stop: options.Stop,
	})
	if err != nil {
		return nil, err
	}
	return []byte(text), nil
}
