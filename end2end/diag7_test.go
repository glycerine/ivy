package end2end

import (
	"fmt"
	"testing"
)

func TestDiagnostic7(t *testing.T) {
	mod := compileIvyFile(t, "client_server.ivy")
	fmt.Printf("Actions: %d\n", len(mod.Actions))
	for k, v := range mod.Actions {
		fmt.Printf("  %q: %T\n", k, v)
	}
	fmt.Printf("PublicActions: %d\n", len(mod.PublicActions))
	for k := range mod.PublicActions {
		fmt.Printf("  %q\n", k)
	}
}
