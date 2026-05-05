// Package mc implements model checking for Ivy, converting programs to
// AIGER format for hardware model checkers.
//
// Ported to Go from ivy_mc.py.
package goivy

import (
	"fmt"
	"strings"
)

// Aiger is a low-level AIGER circuit representation.
// AIGER uses integer literals: even = positive, odd = negated.
// Literal 0 = false, 1 = true.
type Aiger struct {
	Inputs  []string // input variable names
	Latches []string // latch (state) variable names
	Outputs []string // output variable names

	Gates  [][3]int       // each gate: (output_lit, input0, input1)
	VarMap map[string]int // symbol name -> literal
	NextID int            // next available variable ID
	Values map[string]int // latch/output -> next-state literal

	// Simulation state
	State map[int]byte // literal (even) -> '0' or '1'
}

// NewAiger creates a new Aiger circuit.
// A bogus input is added to work around an ABC model checker bug.
func NewAiger(inputs, latches, outputs []string) *Aiger {
	// Add bogus input to work around ABC bug
	allInputs := make([]string, len(inputs)+1)
	copy(allInputs, inputs)
	allInputs[len(inputs)] = "%%bogus%%"

	a := &Aiger{
		Inputs:  allInputs,
		Latches: latches,
		Outputs: outputs,
		VarMap:  make(map[string]int),
		NextID:  1,
		Values:  make(map[string]int),
	}

	// Assign literals to inputs and latches
	for _, x := range allInputs {
		a.VarMap[x] = a.NextID * 2
		a.NextID++
	}
	for _, x := range latches {
		a.VarMap[x] = a.NextID * 2
		a.NextID++
	}

	return a
}

// True returns the AIGER literal for true (constant 1).
func (a *Aiger) True() int {
	return 1
}

// False returns the AIGER literal for false (constant 0).
func (a *Aiger) False() int {
	return 0
}

// Lit returns the literal for the given symbol name.
func (a *Aiger) Lit(sym string) (int, bool) {
	v, ok := a.VarMap[sym]
	return v, ok
}

// MustLit returns the literal for the given symbol name, panicking if not found.
func (a *Aiger) MustLit(sym string) int {
	v, ok := a.VarMap[sym]
	if !ok {
		panic(fmt.Sprintf("no literal for symbol: %s", sym))
	}
	return v
}

// Define maps a symbol name to an AIGER literal value.
func (a *Aiger) Define(sym string, val int) {
	a.VarMap[sym] = val
}

// Andl returns the AND of all arguments. Empty args returns True.
func (a *Aiger) Andl(args ...int) int {
	if len(args) == 0 {
		return a.True()
	}
	res := args[0]
	for _, x := range args[1:] {
		tmp := a.NextID * 2
		a.Gates = append(a.Gates, [3]int{tmp, res, x})
		a.NextID++
		res = tmp
	}
	return res
}

// Notl returns the negation of an AIGER literal.
func (a *Aiger) Notl(arg int) int {
	return 2*(arg/2) + (1 - arg%2)
}

// Orl returns the OR of all arguments. Empty args returns False.
func (a *Aiger) Orl(args ...int) int {
	if len(args) == 0 {
		return a.False()
	}
	negated := make([]int, len(args))
	for i, x := range args {
		negated[i] = a.Notl(x)
	}
	return a.Notl(a.Andl(negated...))
}

// Ite returns if-then-else: (x ? y : z).
func (a *Aiger) Ite(x, y, z int) int {
	return a.Orl(a.Andl(x, y), a.Andl(a.Notl(x), z))
}

// Iff returns the biconditional (equivalence) of x and y.
func (a *Aiger) Iff(x, y int) int {
	return a.Orl(a.Andl(x, y), a.Andl(a.Notl(x), a.Notl(y)))
}

// ImpliesL returns x implies y.
func (a *Aiger) ImpliesL(x, y int) int {
	return a.Orl(a.Notl(x), y)
}

// Xor returns the exclusive or of x and y.
func (a *Aiger) Xor(x, y int) int {
	return a.Notl(a.Iff(x, y))
}

// Set sets the next-state value for a latch.
func (a *Aiger) Set(sym string, val int) {
	a.Values[sym] = val
}

// String returns the AIGER ASCII format (AAG) representation.
func (a *Aiger) String() string {
	var b strings.Builder
	numGates := a.NextID - 1 - (len(a.Inputs) + len(a.Latches))
	fmt.Fprintf(&b, "aag %d %d %d %d %d\n",
		a.NextID-1, len(a.Inputs), len(a.Latches), len(a.Outputs), numGates)

	for _, x := range a.Inputs {
		fmt.Fprintf(&b, "%d\n", a.VarMap[x])
	}
	for _, x := range a.Latches {
		fmt.Fprintf(&b, "%d %d\n", a.VarMap[x], a.Values[x])
	}
	for _, x := range a.Outputs {
		fmt.Fprintf(&b, "%d\n", a.Values[x])
	}
	for _, g := range a.Gates {
		fmt.Fprintf(&b, "%d %d %d\n", g[0], g[1], g[2])
	}
	return b.String()
}

// Reset initializes the simulation state: all latches = '0'.
func (a *Aiger) Reset() {
	a.State = make(map[int]byte)
	for _, x := range a.Latches {
		a.State[a.VarMap[x]] = '0'
	}
}

// GetIn reads a simulation value for an AIGER literal.
func (a *Aiger) GetIn(gi int) byte {
	if gi == 0 {
		return '0'
	}
	if gi == 1 {
		return '1'
	}
	v := a.State[gi&^1]
	if gi&1 != 0 {
		if v == '1' {
			return '0'
		}
		return '1'
	}
	return v
}

// Step simulates one clock cycle with the given input string.
// inp must be a string of '0'/'1' characters of length len(Inputs).
func (a *Aiger) Step(inp string) {
	if len(inp) != len(a.Inputs) {
		panic(fmt.Sprintf("step: input length %d != %d", len(inp), len(a.Inputs)))
	}
	for i, x := range a.Inputs {
		a.State[a.VarMap[x]] = inp[i]
	}
	for _, g := range a.Gates {
		out, in0, in1 := g[0], g[1], g[2]
		if a.GetIn(in0) == '1' && a.GetIn(in1) == '1' {
			a.State[out] = '1'
		} else {
			a.State[out] = '0'
		}
	}
}

// Advance moves the simulation to the next state (latches get next values).
func (a *Aiger) Advance() {
	for _, lt := range a.Latches {
		a.State[a.VarMap[lt]] = a.GetIn(a.Values[lt])
	}
}

// SymVals returns a string of simulation values for the given symbols.
func (a *Aiger) SymVals(syms []string) string {
	var b strings.Builder
	for _, sym := range syms {
		b.WriteByte(a.GetIn(a.VarMap[sym]))
	}
	return b.String()
}

// SymNextVals returns a string of next-state simulation values for the given symbols.
func (a *Aiger) SymNextVals(syms []string) string {
	var b strings.Builder
	for _, sym := range syms {
		b.WriteByte(a.GetIn(a.Values[sym]))
	}
	return b.String()
}

// LatchVals returns a string of simulation values for all latches.
func (a *Aiger) LatchVals() string {
	return a.SymVals(a.Latches)
}

// InputVals returns a string of simulation values for all inputs.
func (a *Aiger) InputVals() string {
	return a.SymVals(a.Inputs)
}

// GetState returns a map from latch name to truth value ('0', '1', or 'x')
// given a post-state string.
func (a *Aiger) GetState(post string) map[string]byte {
	res := make(map[string]byte)
	for i, v := range a.Latches {
		if i < len(post) {
			res[v] = post[i]
		}
	}
	return res
}

// Debug prints debug information about the circuit.
func (a *Aiger) Debug() string {
	var b strings.Builder
	fmt.Fprintf(&b, "inputs: %v\n", a.Inputs)
	fmt.Fprintf(&b, "latches: %v\n", a.Latches)
	fmt.Fprintf(&b, "outputs: %v\n", a.Outputs)
	fmt.Fprintln(&b, "map:")
	for x, y := range a.VarMap {
		fmt.Fprintf(&b, "  %s = %d\n", x, y)
	}
	fmt.Fprintln(&b, "values:")
	for x, y := range a.Values {
		fmt.Fprintf(&b, "  %s = %d\n", x, y)
	}
	fmt.Fprintln(&b, "self:")
	fmt.Fprint(&b, a.String())
	return b.String()
}
