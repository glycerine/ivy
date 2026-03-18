package integ_test

import (
	"fmt"
	"testing"
)

func TestDebugClientServerInit(t *testing.T) {
	mod := compileIvyFile(t, "client_server.ivy")
	
	fmt.Printf("InitCond: %v\n", mod.InitCond)
	fmt.Printf("InitCond is nil: %v\n", mod.InitCond == nil)
	if mod.InitCond != nil {
		fmt.Printf("InitCond fmlas: %d\n", len(mod.InitCond.Fmlas))
		for i, f := range mod.InitCond.Fmlas {
			fmt.Printf("  fmla[%d]: %s\n", i, f)
		}
		fmt.Printf("InitCond defs: %d\n", len(mod.InitCond.Defs))
	}
	fmt.Printf("LabeledInits: %d\n", len(mod.LabeledInits))
	for i, li := range mod.LabeledInits {
		fmt.Printf("  init[%d]: %s\n", i, li.Formula)
	}
	fmt.Printf("Initializers: %d\n", len(mod.Initializers))
	for i, na := range mod.Initializers {
		fmt.Printf("  initializer[%d]: name=%s action=%T\n", i, na.Name, na.Action)
	}
}
