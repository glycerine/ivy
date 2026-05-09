package smt

import "fmt"

// ErrMsg reports a Z3 API error observed at the smt boundary.
//
// Z3 reports API misuse and other external errors through the context error
// code and the registered error handler. The handler must not throw across the
// wasm/native boundary; smt checks the context error state after Z3 calls and
// raises this error from normal Go control flow when the existing API shape has
// no error return.
type ErrMsg struct {
	Op   string
	Code int
	Msg  string
}

func (e *ErrMsg) Error() string {
	if e == nil {
		return "z3: <nil>"
	}
	if e.Op != "" {
		return fmt.Sprintf("z3 %s: %s", e.Op, e.Msg)
	}
	return fmt.Sprintf("z3: %s", e.Msg)
}
