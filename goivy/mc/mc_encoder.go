package mc

import (
	"fmt"
	"math"
	"strconv"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/theory"
)

// Encoder wraps an Aiger circuit with multi-bit encoding for finite sorts.
// Non-boolean sorts are represented as multiple AIGER bits (binary encoding).
// Python: ivy_mc.py Encoder class — inputs/latches/outputs are [Symbol],
// encoding maps Symbol → [Symbol].
type Encoder struct {
	Inputs             []*lg.Const                // original multi-bit input symbols
	Latches            []*lg.Const                // original multi-bit latch symbols
	Outputs            []*lg.Const                // original multi-bit output symbols
	Encoding           map[lg.NodeKey][]string    // lg.Key(sym) -> list of sub-bit names
	Sub                *Aiger                     // underlying AIGER circuit
	Ops                map[string]ArithOp         // arithmetic operations
	IsConstructor      func(*lg.Const) bool       // checks if symbol is a constructor
	ConstructorIndexFn func(*lg.Const) (int, int) // returns (index, total) for constructor
	Interp             map[string]interface{}     // sort interpretations for DecodeVal
}

// ArithOp is a function type for multi-bit arithmetic operations.
type ArithOp func(nbits int, x, y []int) []int

// NewEncoder creates a new Encoder with the given inputs, latches, and outputs.
// Each symbol is expanded into multiple bits based on its sort.
// Python: Encoder.__init__(inputs, latches, outputs) — all [Symbol].
func NewEncoder(inputs, latches, outputs []*lg.Const) *Encoder {
	enc := &Encoder{
		Inputs:   inputs,
		Latches:  latches,
		Outputs:  outputs,
		Encoding: make(map[lg.NodeKey][]string),
	}

	subInputs := encodeVars(inputs, enc.Encoding)
	subLatches := encodeVars(latches, enc.Encoding)
	subOutputs := encodeVars(outputs, enc.Encoding)
	enc.Sub = NewAiger(subInputs, subLatches, subOutputs)

	return enc
}

// encodeVars expands typed symbols into bit-level names using sort for bit width.
// Python: encode_vars(syms, encoding) — uses get_encoding_bits(sym.sort).
func encodeVars(syms []*lg.Const, encoding map[lg.NodeKey][]string) []string {
	var res []string
	for _, sym := range syms {
		n := getEncodingBits(sym.CSort)
		vs := make([]string, n)
		for i := 0; i < n; i++ {
			vs[i] = fmt.Sprintf("%s[%d]", sym.Name, i)
		}
		encoding[lg.Key(sym)] = vs
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

// Lit returns the multi-bit literal for a typed symbol.
// Python: Encoder.lit(sym) — sym is a Symbol, looks up self.encoding[sym].
func (e *Encoder) Lit(sym *lg.Const) ([]int, bool) {
	enc, ok := e.Encoding[lg.Key(sym)]
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
func (e *Encoder) MustLit(sym *lg.Const) []int {
	res, ok := e.Lit(sym)
	if !ok {
		panic(fmt.Sprintf("no encoding for symbol: %s", sym.Name))
	}
	return res
}

// DefineSym maps a typed symbol to multi-bit values by adding encoding entries.
// Python: Encoder.define(sym, val) — sym is a Symbol.
func (e *Encoder) DefineSym(sym *lg.Const, val []int) {
	key := lg.Key(sym)
	if _, ok := e.Encoding[key]; !ok {
		vs := make([]string, len(val))
		for i := range val {
			vs[i] = fmt.Sprintf("%s[%d]", sym.Name, i)
		}
		e.Encoding[key] = vs
	}
	enc := e.Encoding[key]
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

// SetSym sets the next-state values for a multi-bit typed symbol.
// Python: Encoder.set(sym, val) — sym is a Symbol, looks up self.encoding[sym].
func (e *Encoder) SetSym(sym *lg.Const, val []int) {
	enc := e.Encoding[lg.Key(sym)]
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

// GetDefFunc is a callback for resolving undefined symbols during Eval.
type GetDefFunc func(sym *lg.Const) ([]int, error)

// Eval evaluates an Ivy logic expression to AIGER multi-bit literal(s).
// This is the core expression-to-circuit conversion.
//
// Handles: And, Or, Not, Implies, Iff, Eq, Ite, Apply (with constructors,
// numerals, arithmetic ops, and plain symbol lookup).
//
// Python: ivy_mc.py:313-359 Encoder.eval()
func (e *Encoder) Eval(expr lg.Expr, getdef GetDefFunc) ([]int, error) {
	res, err := e.evalRec(expr, getdef)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("eval produced empty result for: %v", expr)
	}
	return res, nil
}

func (e *Encoder) evalRec(expr lg.Expr, getdef GetDefFunc) ([]int, error) {
	switch t := expr.(type) {
	case *lg.Ite:
		cond, err := e.evalRec(t.Cond, getdef)
		if err != nil {
			return nil, err
		}
		thenTerm, err := e.evalRec(t.Then, getdef)
		if err != nil {
			return nil, err
		}
		elseTerm, err := e.evalRec(t.Else, getdef)
		if err != nil {
			return nil, err
		}
		return e.EncodeIte(cond[0], thenTerm, elseTerm), nil

	case *lg.Apply:
		sym, ok := t.Func.(*lg.Const)
		if !ok {
			return nil, fmt.Errorf("eval: non-const application: %v", expr)
		}

		// Constructor: encode as its index in the sort's constructors
		if e.IsConstructor != nil && e.IsConstructor(sym) {
			idx, total := e.ConstructorIndex(sym)
			bits := ceilLog2(total)
			return e.BinEnc(idx, bits), nil
		}

		// Numeral with interpreted sort
		if il.IsNumeral(sym) && isInterpretedSort(sym.CSort) {
			n := getEncodingBits(sym.CSort)
			val := parseNum(sym.Name)
			return e.BinEnc(val, n), nil
		}

		// Arithmetic operation on interpreted sort
		if opFn, ok := e.Ops[sym.Name]; ok {
			args := make([][]int, len(t.Terms))
			for i, arg := range t.Terms {
				v, err := e.evalRec(arg, getdef)
				if err != nil {
					return nil, err
				}
				args[i] = v
			}
			if len(args) >= 2 {
				nbits := len(args[0])
				return opFn(nbits, args[0], args[1]), nil
			}
			return nil, fmt.Errorf("eval: op %s requires at least 2 args", sym.Name)
		}

		// Plain symbol lookup (nullary)
		if len(t.Terms) == 0 {
			if lit, ok := e.Lit(sym); ok {
				return lit, nil
			}
			if getdef != nil {
				return getdef(sym)
			}
			return nil, fmt.Errorf("eval: no definition for %s in aiger output", sym.Name)
		}
		return nil, fmt.Errorf("eval: non-nullary application: %v", expr)

	case *lg.Const:
		// Plain symbol
		if lit, ok := e.Lit(t); ok {
			return lit, nil
		}
		if getdef != nil {
			return getdef(t)
		}
		return nil, fmt.Errorf("eval: no definition for %s", t.Name)

	case *lg.And:
		args := make([][]int, len(t.Terms))
		for i, term := range t.Terms {
			v, err := e.evalRec(term, getdef)
			if err != nil {
				return nil, err
			}
			args[i] = v
		}
		return e.AndlMulti(args...), nil

	case *lg.Or:
		args := make([][]int, len(t.Terms))
		for i, term := range t.Terms {
			v, err := e.evalRec(term, getdef)
			if err != nil {
				return nil, err
			}
			args[i] = v
		}
		return e.OrlMulti(args...), nil

	case *lg.Not:
		v, err := e.evalRec(t.Body, getdef)
		if err != nil {
			return nil, err
		}
		return e.NotlMulti(v), nil

	case *lg.Implies:
		lhs, err := e.evalRec(t.T1, getdef)
		if err != nil {
			return nil, err
		}
		rhs, err := e.evalRec(t.T2, getdef)
		if err != nil {
			return nil, err
		}
		return e.ImpliesMulti(lhs, rhs), nil

	case *lg.Iff:
		lhs, err := e.evalRec(t.T1, getdef)
		if err != nil {
			return nil, err
		}
		rhs, err := e.evalRec(t.T2, getdef)
		if err != nil {
			return nil, err
		}
		return e.IffMulti(lhs, rhs), nil

	case *lg.Eq:
		lhs, err := e.evalRec(t.T1, getdef)
		if err != nil {
			return nil, err
		}
		rhs, err := e.evalRec(t.T2, getdef)
		if err != nil {
			return nil, err
		}
		nvals := getSortSize(t.T1.NodeSort())
		return e.EncodeEquality(nvals, lhs, rhs), nil

	default:
		return nil, fmt.Errorf("eval: unimplemented op in aiger output: %T", expr)
	}
}

// DefList processes a list of definitions, evaluating each and defining the symbol.
// Python: ivy_mc.py:361-371
func (e *Encoder) DefList(defs []lg.Expr) error {
	dmap := make(map[lg.NodeKey]lg.Expr)
	for _, df := range defs {
		switch d := df.(type) {
		case *lg.Definition:
			if c, ok := d.Defines().(*lg.Const); ok {
				dmap[lg.Key(c)] = d.Rhs
			}
		case *lg.Eq:
			if c, ok := d.T1.(*lg.Const); ok {
				dmap[lg.Key(c)] = d.T2
			} else if app, ok := d.T1.(*lg.Apply); ok {
				if c, ok := app.Func.(*lg.Const); ok {
					dmap[lg.Key(c)] = d.T2
				}
			}
		}
	}

	var getdef GetDefFunc
	getdef = func(sym *lg.Const) ([]int, error) {
		body, ok := dmap[lg.Key(sym)]
		if !ok {
			return nil, fmt.Errorf("no definition for %s", sym.Name)
		}
		val, err := e.Eval(body, getdef)
		if err != nil {
			return nil, err
		}
		e.DefineSym(sym, val)
		return val, nil
	}

	for _, df := range defs {
		var sym *lg.Const
		var rhs lg.Expr
		switch d := df.(type) {
		case *lg.Definition:
			if c, ok := d.Defines().(*lg.Const); ok {
				sym = c
				rhs = d.Rhs
			}
		case *lg.Eq:
			if c, ok := d.T1.(*lg.Const); ok {
				sym = c
				rhs = d.T2
			} else if app, ok := d.T1.(*lg.Apply); ok {
				if c, ok := app.Func.(*lg.Const); ok {
					sym = c
					rhs = d.T2
				}
			}
		}
		if sym != nil && rhs != nil {
			val, err := e.Eval(rhs, getdef)
			if err != nil {
				return err
			}
			e.DefineSym(sym, val)
		}
	}
	return nil
}

// --- Encoder fields for constructor/sort integration ---

// IsConstructor checks if a symbol is a constructor. Set by caller.
var _ = (*Encoder)(nil) // ensure Encoder is used

// ConstructorIndex returns the index and total count for a constructor symbol.
func (e *Encoder) ConstructorIndex(sym *lg.Const) (int, int) {
	if e.ConstructorIndexFn != nil {
		return e.ConstructorIndexFn(sym)
	}
	return 0, 1
}

// ceilLog2 returns ceil(log2(n)), minimum 1.
func ceilLog2(n int) int {
	if n <= 1 {
		return 1
	}
	return int(math.Ceil(math.Log2(float64(n))))
}

// getEncodingBits returns the number of bits needed to encode a sort's values.
func getEncodingBits(s lg.Sort) int {
	if es, ok := s.(*lg.EnumeratedSort); ok {
		return ceilLog2(len(es.Extension))
	}
	// Default: 1 bit for boolean, more for numeric types
	return 1
}

// getSortSize returns the number of values in a sort (for equality encoding).
func getSortSize(s lg.Sort) int {
	if es, ok := s.(*lg.EnumeratedSort); ok {
		return len(es.Extension)
	}
	return 2 // boolean
}

// parseNum parses a numeral string to int.
func parseNum(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// isInterpretedSort checks if a sort is interpreted (has fixed meaning).
func isInterpretedSort(s lg.Sort) bool {
	if s == nil {
		return false
	}
	switch s.(type) {
	case *lg.EnumeratedSort:
		return true
	}
	// Check by name for common interpreted sorts
	name := s.String()
	return name == "int" || name == "nat" || name == "bool" || name == "bv"
}

// bitsToBoolExprs converts a string of '0'/'1' chars into a slice of
// lg.Expr where '1' → &lg.And{} (true) and '0' → &lg.Or{} (false).
// Python: [il.And() if b == '1' else il.Or() for b in abits]
func bitsToBoolExprs(s string) []lg.Expr {
	bits := make([]lg.Expr, len(s))
	for i, b := range s {
		if b == '1' {
			bits[i] = &lg.And{Terms: nil}
		} else {
			bits[i] = &lg.Or{Terms: nil}
		}
	}
	return bits
}

// binDecBool converts boolean expression bits to an integer.
// Each bit is &lg.And{} (true=1) or &lg.Or{} (false=0), LSB first.
// Python: Encoder.bindec(bits) in ivy_mc.py.
func binDecBool(bits []lg.Expr) int {
	res := 0
	for idx, bit := range bits {
		if isTrueNode(bit) {
			res += 1 << idx
		}
	}
	return res
}

// DecodeVal converts multi-bit simulation values to an Ivy expression
// based on the sort's theory.
// Python: ivy_mc.py:490-508 Encoder.decode_val(bits, v)
func (e *Encoder) DecodeVal(bits []lg.Expr, v *lg.Const) lg.Expr {
	interp := theory.GetSortTheory(v.CSort, e.Interp)
	switch s := interp.(type) {
	case *lg.EnumeratedSort:
		num := binDecBool(bits)
		vals := s.Extension
		if num >= len(vals) {
			num = len(vals) - 1
		}
		return lg.NewConst(vals[num], v.CSort)
	case *lg.RangeSort:
		num := binDecBool(bits)
		maxVal, _ := strconv.Atoi(s.Ub.BoundString())
		if num > maxVal {
			num = maxVal
		}
		return lg.NewConst(strconv.Itoa(num), v.CSort)
	case *theory.Theory:
		if s.Kind == theory.BitVectorKind {
			num := binDecBool(bits)
			return lg.NewConst(strconv.Itoa(num), v.CSort)
		}
	}
	if il.IsBooleanSort(v.CSort) {
		return bits[0]
	}
	panic(fmt.Sprintf("variable has unexpected sort: %s %v", v.Name, v.CSort))
}

// GetSym reads the current simulation value for a typed symbol.
// Returns nil if the symbol is not encoded in the circuit.
// Python: ivy_mc.py:520-524 Encoder.get_sym(v)
func (e *Encoder) GetSym(v *lg.Const) lg.Expr {
	enc, ok := e.Encoding[lg.Key(v)]
	if !ok || len(enc) == 0 {
		return nil
	}
	abits := e.Sub.SymVals(enc)
	bits := bitsToBoolExprs(abits)
	return e.DecodeVal(bits, v)
}

// GetNextSym reads the next-state simulation value for a typed symbol.
// Returns nil if the symbol is not encoded in the circuit.
// Python: ivy_mc.py:526-530 Encoder.get_next_sym(v)
func (e *Encoder) GetNextSym(v *lg.Const) lg.Expr {
	enc, ok := e.Encoding[lg.Key(v)]
	if !ok || len(enc) == 0 {
		return nil
	}
	abits := e.Sub.SymNextVals(enc)
	bits := bitsToBoolExprs(abits)
	return e.DecodeVal(bits, v)
}

// GetEncoderState returns a map from typed latch symbol to its decoded
// Ivy value, given a post-state string.
// Python: ivy_mc.py:511-518 Encoder.get_state(post)
func (e *Encoder) GetEncoderState(post string) map[lg.NodeKey]lg.Expr {
	subres := e.Sub.GetState(post)
	res := make(map[lg.NodeKey]lg.Expr)
	for _, v := range e.Latches {
		enc := e.Encoding[lg.Key(v)]
		bits := make([]lg.Expr, len(enc))
		for i, s := range enc {
			b := subres[s]
			if b == '1' {
				bits[i] = &lg.And{Terms: nil}
			} else if b == '0' {
				bits[i] = &lg.Or{Terms: nil}
			} else {
				bits[i] = nil
			}
		}
		val := e.DecodeVal(bits, v)
		res[lg.Key(v)] = val
	}
	return res
}
