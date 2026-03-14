package mc

import (
	"fmt"
)

// Encoder wraps an Aiger circuit with multi-bit encoding for finite sorts.
// Non-boolean sorts are represented as multiple AIGER bits (binary encoding).
type Encoder struct {
	Inputs   []string          // original multi-bit input names
	Latches  []string          // original multi-bit latch names
	Outputs  []string          // original multi-bit output names
	Encoding map[string][]string // symbol -> list of sub-bit symbol names
	Sub      *Aiger            // underlying AIGER circuit
	Ops      map[string]ArithOp // arithmetic operations
}

// ArithOp is a function type for multi-bit arithmetic operations.
type ArithOp func(nbits int, x, y []int) []int

// NewEncoder creates a new Encoder with the given inputs, latches, and outputs.
// Each symbol is expanded into multiple bits based on its bit width.
func NewEncoder(inputs, latches, outputs []string, bitWidths map[string]int) *Encoder {
	enc := &Encoder{
		Inputs:   inputs,
		Latches:  latches,
		Outputs:  outputs,
		Encoding: make(map[string][]string),
	}

	subInputs := encodeVarNames(inputs, bitWidths, enc.Encoding)
	subLatches := encodeVarNames(latches, bitWidths, enc.Encoding)
	subOutputs := encodeVarNames(outputs, bitWidths, enc.Encoding)
	enc.Sub = NewAiger(subInputs, subLatches, subOutputs)

	return enc
}

// encodeVarNames expands variable names into bit-level names.
func encodeVarNames(syms []string, bitWidths map[string]int, encoding map[string][]string) []string {
	var res []string
	for _, sym := range syms {
		n := 1
		if bw, ok := bitWidths[sym]; ok {
			n = bw
		}
		vs := make([]string, n)
		for i := 0; i < n; i++ {
			vs[i] = fmt.Sprintf("%s[%d]", sym, i)
		}
		encoding[sym] = vs
		res = append(res, vs...)
	}
	return res
}

// True returns [1] (multi-bit true).
func (e *Encoder) True() []int {
	return []int{e.Sub.True()}
}

// False returns [0] (multi-bit false).
func (e *Encoder) False() []int {
	return []int{e.Sub.False()}
}

// Lit returns the multi-bit literal for a symbol.
func (e *Encoder) Lit(sym string) ([]int, bool) {
	enc, ok := e.Encoding[sym]
	if !ok {
		return nil, false
	}
	res := make([]int, len(enc))
	for i, s := range enc {
		v, ok := e.Sub.Lit(s)
		if !ok {
			return nil, false
		}
		res[i] = v
	}
	return res, true
}

// MustLit returns the multi-bit literal, panicking if not found.
func (e *Encoder) MustLit(sym string) []int {
	res, ok := e.Lit(sym)
	if !ok {
		panic(fmt.Sprintf("no encoding for symbol: %s", sym))
	}
	return res
}

// DefineSym maps a symbol to multi-bit values by adding encoding entries.
func (e *Encoder) DefineSym(sym string, val []int) {
	// Create encoding if it doesn't exist
	if _, ok := e.Encoding[sym]; !ok {
		vs := make([]string, len(val))
		for i := range val {
			vs[i] = fmt.Sprintf("%s[%d]", sym, i)
		}
		e.Encoding[sym] = vs
	}
	enc := e.Encoding[sym]
	for i, v := range val {
		if i < len(enc) {
			e.Sub.Define(enc[i], v)
		}
	}
}

// AndlMulti returns the bit-wise AND of multi-bit arguments.
func (e *Encoder) AndlMulti(args ...[]int) []int {
	if len(args) == 0 {
		return e.True()
	}
	width := len(args[0])
	res := make([]int, width)
	for i := 0; i < width; i++ {
		bits := make([]int, len(args))
		for j, a := range args {
			if i < len(a) {
				bits[j] = a[i]
			}
		}
		res[i] = e.Sub.Andl(bits...)
	}
	return res
}

// OrlMulti returns the bit-wise OR of multi-bit arguments.
func (e *Encoder) OrlMulti(args ...[]int) []int {
	if len(args) == 0 {
		return e.False()
	}
	width := len(args[0])
	res := make([]int, width)
	for i := 0; i < width; i++ {
		bits := make([]int, len(args))
		for j, a := range args {
			if i < len(a) {
				bits[j] = a[i]
			}
		}
		res[i] = e.Sub.Orl(bits...)
	}
	return res
}

// NotlMulti returns the bit-wise NOT of a multi-bit value.
func (e *Encoder) NotlMulti(arg []int) []int {
	res := make([]int, len(arg))
	for i, x := range arg {
		res[i] = e.Sub.Notl(x)
	}
	return res
}

// ImpliesMulti returns bit-wise implies.
func (e *Encoder) ImpliesMulti(x, y []int) []int {
	res := make([]int, len(x))
	for i := range x {
		res[i] = e.Sub.ImpliesL(x[i], y[i])
	}
	return res
}

// IffMulti returns bit-wise iff.
func (e *Encoder) IffMulti(x, y []int) []int {
	res := make([]int, len(x))
	for i := range x {
		res[i] = e.Sub.Iff(x[i], y[i])
	}
	return res
}

// SetSym sets the next-state values for a multi-bit symbol.
func (e *Encoder) SetSym(sym string, val []int) {
	enc := e.Encoding[sym]
	for i, v := range val {
		if i < len(enc) {
			e.Sub.Set(enc[i], v)
		}
	}
}

// String returns the underlying AIGER string.
func (e *Encoder) String() string {
	return e.Sub.String()
}

// BinEnc encodes an integer m as an n-bit binary literal.
// The MSB is first (big-endian).
func (e *Encoder) BinEnc(m, n int) []int {
	res := make([]int, n)
	for i := 0; i < n; i++ {
		if m&(1<<(n-1-i)) != 0 {
			res[i] = e.Sub.True()
		} else {
			res[i] = e.Sub.False()
		}
	}
	return res
}

// BinDec decodes a multi-bit literal to an integer,
// treating True (1) as 1 and anything else as 0.
func (e *Encoder) BinDec(bits []int) int {
	res := 0
	n := len(bits)
	for i, v := range bits {
		if v == e.Sub.True() {
			res += 1 << (n - 1 - i)
		}
	}
	return res
}

// GeBin returns the AIGER literal representing "bits >= n" (big-endian).
func (e *Encoder) GeBin(bits []int, n int) int {
	if n == 0 {
		return e.Sub.True()
	}
	if n >= (1 << len(bits)) {
		return e.Sub.False()
	}
	hval := 1 << (len(bits) - 1)
	if hval <= n {
		return e.Sub.Andl(bits[0], e.GeBin(bits[1:], n-hval))
	}
	return e.Sub.Orl(bits[0], e.GeBin(bits[1:], n))
}

// EncodeEquality returns a 1-bit result representing equality of two multi-bit values.
// nvals is the number of valid values in the domain.
func (e *Encoder) EncodeEquality(nvals int, eterms ...[]int) []int {
	if len(eterms) < 2 {
		return e.True()
	}
	x, y := eterms[0], eterms[1]
	iffBits := make([]int, len(x))
	for i := range x {
		iffBits[i] = e.Sub.Iff(x[i], y[i])
	}
	eqs := e.Sub.Andl(iffBits...)

	geBits := make([]int, len(eterms))
	for i, et := range eterms {
		geBits[i] = e.GeBin(et, nvals-1)
	}
	alt := e.Sub.Andl(geBits...)
	return []int{e.Sub.Orl(eqs, alt)}
}

// EncodePlusInt performs binary addition of x and y with carry-in cy.
// Returns (result, carry-out).
func (e *Encoder) EncodePlusInt(x, y []int, cy int) ([]int, int) {
	res := make([]int, len(x))
	for i := len(x) - 1; i >= 0; i-- {
		res[i] = e.Sub.Xor(e.Sub.Xor(x[i], y[i]), cy)
		cy = e.Sub.Orl(
			e.Sub.Andl(x[i], y[i]),
			e.Sub.Andl(x[i], cy),
			e.Sub.Andl(y[i], cy),
		)
	}
	return res, cy
}

// EncodePlus adds two multi-bit values.
func (e *Encoder) EncodePlus(x, y []int) []int {
	res, _ := e.EncodePlusInt(x, y, e.Sub.False())
	return res
}

// EncodeMinus subtracts y from x using two's complement.
func (e *Encoder) EncodeMinus(x, y []int) []int {
	ycom := e.NotlMulti(y)
	res, _ := e.EncodePlusInt(x, ycom, e.Sub.True())
	return res
}

// EncodeLt returns a 1-bit result: x < y.
func (e *Encoder) EncodeLt(x, y []int, cy int) []int {
	for i := len(x) - 1; i >= 0; i-- {
		cy = e.Sub.Orl(
			e.Sub.Andl(e.Sub.Notl(x[i]), y[i]),
			e.Sub.Andl(e.Sub.Iff(x[i], y[i]), cy),
		)
	}
	return []int{cy}
}

// EncodeLe returns a 1-bit result: x <= y.
func (e *Encoder) EncodeLe(x, y []int) []int {
	return e.EncodeLt(x, y, e.Sub.True())
}

// EncodeTimes multiplies two multi-bit values using shift-and-add.
func (e *Encoder) EncodeTimes(x, y []int) []int {
	n := len(x)
	res := make([]int, n)
	for i := range res {
		res[i] = e.Sub.False()
	}
	for i := 0; i < n; i++ {
		// shift result left by 1
		shifted := make([]int, n)
		copy(shifted, res[1:])
		shifted[n-1] = e.Sub.False()
		res = shifted
		// conditionally add y
		sum := e.EncodePlus(res, y)
		// ite on bit x[i]
		chosen := make([]int, n)
		for j := 0; j < n; j++ {
			chosen[j] = e.Sub.Ite(x[i], sum[j], res[j])
		}
		res = chosen
	}
	return res
}

// EncodeIte returns a multi-bit if-then-else.
func (e *Encoder) EncodeIte(cond int, thenBits, elseBits []int) []int {
	res := make([]int, len(thenBits))
	for i := range thenBits {
		res[i] = e.Sub.Ite(cond, thenBits[i], elseBits[i])
	}
	return res
}
