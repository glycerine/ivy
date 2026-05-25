package ivy2go

import (
	"sort"

	"github.com/glycerine/ivy/goivy"
)

// runtime.go: per-program runtime helper emission. Mirrors the role of
// ivy2cpp/runtime.go but the C++ approach (linking against external
// ivy_*.hpp headers) is replaced in Go by inlining the helpers into
// each emitted package's runtime.go (per ARCHITECTURE_TODO.md §3.5.8).
//
// emitRuntimeHelpers is called by generator.go's emitRuntime() AFTER
// all other emitters have run, so any helper requests they recorded
// in Ctx.OnceGlobals are visible.

// emitRuntimeHelpers writes the helpers requested by other emitters
// during this generation pass. Helper requests are made via
// requireUint128, requireBigInt, requestIteHelper, etc.
func (g *Generator) emitRuntimeHelpers(w *goWriter) {
	// Always-on helpers: ivyAssert / ivyAssume / ivyChoose / a
	// package-level random source. Generated programs need these for
	// any target other than `class`.
	g.emitRuntimePreamble(w)

	// Specialised helpers: emit only those marked in OnceGlobals.
	if g.Ctx == nil {
		return
	}
	if g.Ctx.OnceGlobals["__need_uint128"] {
		g.emitUint128Helpers(w)
	}
	if g.Ctx.OnceGlobals["__need_bigint"] {
		g.emitBigIntHelpers(w)
	}
	// ite_<type> helpers: walk OnceGlobals keys, find any starting
	// with "ite_", emit one per. We don't know the result type from
	// the key alone — store (name, typeExpr) on g for emission.
	for _, h := range g.collectIteHelpers() {
		g.emitIteHelper(w, h)
	}
}

// emitRuntimePreamble writes the package-level helpers every emitted
// program needs at runtime.
func (g *Generator) emitRuntimePreamble(w *goWriter) {
	// Import list for runtime.go.
	if g.Ctx != nil {
		g.Ctx.AddImport("runtime", "fmt", "")
		g.Ctx.AddImport("runtime", "math/rand", "")
	}

	w.line("// ivyRand is the package-wide RNG used by ivyChoose and any")
	w.line("// nondet helpers. Seeded from --seed by main.go.")
	w.line("var ivyRand = rand.New(rand.NewSource(1))")
	w.blank()

	// ivyTraceOut is the io.Writer trace writes target. Default is
	// os.Stdout; tests / embedders can override.
	if g.Config.Trace {
		g.Ctx.AddImport("runtime", "io", "")
		g.Ctx.AddImport("runtime", "os", "")
		w.line("// ivyTraceOut is the io.Writer used by trace-LHS writes.")
		w.line("// Defaults to os.Stdout; override by reassigning before Init.")
		w.line("var ivyTraceOut io.Writer = os.Stdout")
		w.blank()
	}

	w.open("func ivyAssert(cond bool, label string) {")
	w.open("if !cond {")
	w.line(`panic(fmt.Sprintf("ivy assert failed: %s", label))`)
	w.close("")
	w.close("")
	w.blank()

	w.open("func ivyAssume(cond bool, label string) {")
	w.open("if !cond {")
	w.line(`panic(fmt.Sprintf("ivy assume failed: %s", label))`)
	w.close("")
	w.close("")
	w.blank()

	w.open("func ivyChoose(rng int) int {")
	w.open("if rng <= 0 {")
	w.line("return 0")
	w.close("")
	w.line("return ivyRand.Intn(rng)")
	w.close("")
	w.blank()
}

// emitUint128Helpers writes the Uint128 struct + its methods. Lands
// in runtime.go only when bv_expr.go's emission recorded a request
// via requireUint128.
func (g *Generator) emitUint128Helpers(w *goWriter) {
	w.line("// Uint128 is a 128-bit unsigned integer used for bv[N] with")
	w.line("// 65 <= N <= 128. Mirrors ivy2cpp's `unsigned __int128`.")
	w.open("type Uint128 struct {")
	w.line("Hi, Lo uint64")
	w.close("")
	w.blank()

	w.open("func Uint128FromString(s string) Uint128 {")
	w.line("// Minimal parser: handles decimal and 0x-prefixed hex; for")
	w.line("// larger inputs callers should reach for big.Int. Mirrors")
	w.line("// ivy2cpp/ivy_uint128_from_string.")
	w.line("var v Uint128")
	w.line(`if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {`)
	w.line("for _, r := range s[2:] {")
	w.line("v = v.Shl(4)")
	w.line(`if r >= '0' && r <= '9' { v.Lo |= uint64(r - '0') }`)
	w.line(`if r >= 'a' && r <= 'f' { v.Lo |= uint64(r - 'a' + 10) }`)
	w.line(`if r >= 'A' && r <= 'F' { v.Lo |= uint64(r - 'A' + 10) }`)
	w.line("}")
	w.line("return v")
	w.line("}")
	w.line("for _, r := range s {")
	w.line(`if r >= '0' && r <= '9' { v = v.MulU64(10).AddU64(uint64(r - '0')) }`)
	w.line("}")
	w.line("return v")
	w.close("")
	w.blank()

	w.line("func Uint128Mask(bits int) Uint128 {")
	w.line("\tif bits <= 0 { return Uint128{} }")
	w.line("\tif bits >= 128 { return Uint128{Hi: ^uint64(0), Lo: ^uint64(0)} }")
	w.line("\tif bits <= 64 { return Uint128{Lo: (uint64(1) << uint(bits)) - 1} }")
	w.line("\treturn Uint128{Hi: (uint64(1) << uint(bits-64)) - 1, Lo: ^uint64(0)}")
	w.line("}")
	w.blank()

	w.line("func Uint128SignMask(bits int) Uint128 {")
	w.line("\tif bits <= 0 || bits > 128 { return Uint128{} }")
	w.line("\tif bits <= 64 { return Uint128{Lo: uint64(1) << uint(bits-1)} }")
	w.line("\treturn Uint128{Hi: uint64(1) << uint(bits-65)}")
	w.line("}")
	w.blank()

	// Arithmetic methods. Implemented in terms of math/bits when
	// helpful. M3 lands the small set bv_expr.go uses; later
	// milestones extend as needed.
	w.line("func (a Uint128) Shl(n uint) Uint128 {")
	w.line("\tif n == 0 { return a }")
	w.line("\tif n >= 128 { return Uint128{} }")
	w.line("\tif n >= 64 { return Uint128{Hi: a.Lo << (n - 64)} }")
	w.line("\treturn Uint128{Hi: (a.Hi << n) | (a.Lo >> (64 - n)), Lo: a.Lo << n}")
	w.line("}")
	w.blank()

	w.line("func (a Uint128) AddU64(v uint64) Uint128 {")
	w.line("\tlo := a.Lo + v")
	w.line("\thi := a.Hi")
	w.line("\tif lo < a.Lo { hi++ }")
	w.line("\treturn Uint128{Hi: hi, Lo: lo}")
	w.line("}")
	w.blank()

	w.line("func (a Uint128) MulU64(v uint64) Uint128 {")
	w.line("\t// Schoolbook 64x64 -> 128 using two halves of v.")
	w.line("\tloLo, loHi := splitUint64(a.Lo)")
	w.line("\thi := a.Hi * v")
	w.line("\tp0 := loLo * v")
	w.line("\tp1 := loHi * v")
	w.line("\thi += p1 >> 32")
	w.line("\tloProd := (p1 << 32) + p0")
	w.line("\tif loProd < p0 { hi++ }")
	w.line("\treturn Uint128{Hi: hi, Lo: loProd}")
	w.line("}")
	w.blank()

	w.line("func splitUint64(x uint64) (uint64, uint64) {")
	w.line("\treturn x & 0xFFFFFFFF, x >> 32")
	w.line("}")
	w.blank()

	w.line("func (a Uint128) MaskTo(bits int) Uint128 {")
	w.line("\tm := Uint128Mask(bits)")
	w.line("\treturn Uint128{Hi: a.Hi & m.Hi, Lo: a.Lo & m.Lo}")
	w.line("}")
	w.blank()
}

// emitBigIntHelpers writes the bigInt mask helpers + wide-BV
// arithmetic helpers needed for BV widths >64. Always pairs with a
// math/big import.
func (g *Generator) emitBigIntHelpers(w *goWriter) {
	if g.Ctx != nil {
		g.Ctx.AddImport("runtime", "math/big", "")
	}
	w.line("func bigIntMask(bits int) *big.Int {")
	w.line("\tm := new(big.Int).Lsh(big.NewInt(1), uint(bits))")
	w.line("\treturn m.Sub(m, big.NewInt(1))")
	w.line("}")
	w.blank()
	w.line("func bigIntSignMask(bits int) *big.Int {")
	w.line("\treturn new(big.Int).Lsh(big.NewInt(1), uint(bits-1))")
	w.line("}")
	w.blank()
	// toBigInt: coerce any integer-ish value to *big.Int. Generated
	// callers may pass uint32 / uint64 / Uint128 / *big.Int — each
	// has a distinct path. We use interface{} to admit them all.
	w.line("// toBigInt coerces a value to *big.Int. Accepts uint32,")
	w.line("// uint64, Uint128, *big.Int, or any int-typed Go value.")
	w.open("func toBigInt(v interface{}) *big.Int {")
	w.line("switch x := v.(type) {")
	w.line("case *big.Int:")
	w.line("\treturn new(big.Int).Set(x)")
	w.line("case uint64:")
	w.line("\treturn new(big.Int).SetUint64(x)")
	w.line("case uint32:")
	w.line("\treturn new(big.Int).SetUint64(uint64(x))")
	w.line("case int:")
	w.line("\treturn big.NewInt(int64(x))")
	w.line("case Uint128:")
	w.line("\thi := new(big.Int).SetUint64(x.Hi)")
	w.line("\thi.Lsh(hi, 64)")
	w.line("\treturn hi.Or(hi, new(big.Int).SetUint64(x.Lo))")
	w.line("}")
	w.line(`panic(fmt.Sprintf("toBigInt: unsupported %T", v))`)
	w.close("")
	w.blank()
	// Arithmetic helpers — each masks to `bits` after the op.
	for _, op := range []string{"Add", "Sub", "Mul"} {
		w.linef("func wideBV%s(a, b *big.Int, bits int) *big.Int {", op)
		w.linef("\tr := new(big.Int).%s(a, b)", op)
		w.line("\treturn r.And(r, bigIntMask(bits))")
		w.line("}")
		w.blank()
	}
	w.line("func wideBVDiv(a, b *big.Int, bits int) *big.Int {")
	w.line("\tif b.Sign() == 0 { return new(big.Int) }")
	w.line("\tr := new(big.Int).Quo(a, b)")
	w.line("\treturn r.And(r, bigIntMask(bits))")
	w.line("}")
	w.blank()
	w.line("func wideBVMod(a, b *big.Int, bits int) *big.Int {")
	w.line("\tif b.Sign() == 0 { return new(big.Int) }")
	w.line("\tr := new(big.Int).Rem(a, b)")
	w.line("\treturn r.And(r, bigIntMask(bits))")
	w.line("}")
	w.blank()
	for _, op := range []string{"And", "Or", "Xor"} {
		w.linef("func wideBV%s(a, b *big.Int, bits int) *big.Int {", op)
		w.linef("\treturn new(big.Int).%s(a, b).And(new(big.Int).%s(a, b), bigIntMask(bits))", op, op)
		w.line("}")
		w.blank()
	}
	w.line("func wideBVNot(a *big.Int, bits int) *big.Int {")
	w.line("\tm := bigIntMask(bits)")
	w.line("\treturn new(big.Int).Xor(a, m)")
	w.line("}")
	w.blank()
	w.line("func wideBVNeg(a *big.Int, bits int) *big.Int {")
	w.line("\tr := new(big.Int).Neg(a)")
	w.line("\treturn r.And(r, bigIntMask(bits))")
	w.line("}")
	w.blank()
	w.line("func wideBVShl(a, b *big.Int, bits int) *big.Int {")
	w.line("\tn := uint(b.Uint64())")
	w.line("\tif n >= uint(bits) { return new(big.Int) }")
	w.line("\tr := new(big.Int).Lsh(a, n)")
	w.line("\treturn r.And(r, bigIntMask(bits))")
	w.line("}")
	w.blank()
	w.line("func wideBVShrLogical(a, b *big.Int, bits int) *big.Int {")
	w.line("\tn := uint(b.Uint64())")
	w.line("\tif n >= uint(bits) { return new(big.Int) }")
	w.line("\treturn new(big.Int).Rsh(a, n)")
	w.line("}")
	w.blank()
	w.line("func wideBVShrArith(a, b *big.Int, bits int) *big.Int {")
	w.line("\t// Treat a's high bit (at position bits-1) as the sign;")
	w.line("\t// fill from there on logical right-shift.")
	w.line("\tn := uint(b.Uint64())")
	w.line("\tsign := bigIntSignMask(bits)")
	w.line("\tneg := new(big.Int).And(a, sign).Sign() != 0")
	w.line("\tif n >= uint(bits) {")
	w.line("\t\tif neg { return bigIntMask(bits) }")
	w.line("\t\treturn new(big.Int)")
	w.line("\t}")
	w.line("\tr := new(big.Int).Rsh(a, n)")
	w.line("\tif neg {")
	w.line("\t\thi := new(big.Int).Lsh(bigIntMask(int(n)), uint(bits)-n)")
	w.line("\t\tr.Or(r, hi)")
	w.line("\t}")
	w.line("\treturn r.And(r, bigIntMask(bits))")
	w.line("}")
	w.blank()
	w.line("func wideBVConcat(a, b *big.Int, bWidth, bits int) *big.Int {")
	w.line("\tr := new(big.Int).Lsh(a, uint(bWidth))")
	w.line("\tr.Or(r, b)")
	w.line("\treturn r.And(r, bigIntMask(bits))")
	w.line("}")
	w.blank()
	w.line("func wideBVMask(a *big.Int, bits int) *big.Int {")
	w.line("\treturn new(big.Int).And(a, bigIntMask(bits))")
	w.line("}")
	w.blank()
}

// iteHelperReq is a queued ite helper waiting to be emitted.
type iteHelperReq struct {
	Name string // e.g. "ite_uint32"
	Type string // the Go type expression for then/else operands
}

// collectIteHelpers walks Ctx.OnceGlobals for keys starting with "ite_"
// and reconstructs (name, type) pairs from the recorded suffix. Because
// requestIteHelper only stored the name, we re-derive the type via the
// inverse of iteHelperSuffix. This is lossy for compound types but the
// common cases (uint32, uint64, bool, named types) round-trip.
func (g *Generator) collectIteHelpers() []iteHelperReq {
	if g.Ctx == nil {
		return nil
	}
	var reqs []iteHelperReq
	for k := range g.Ctx.OnceGlobals {
		if len(k) < 4 || k[:4] != "ite_" {
			continue
		}
		suffix := k[4:]
		typ := iteHelperTypeFromSuffix(suffix)
		reqs = append(reqs, iteHelperReq{Name: k, Type: typ})
	}
	sort.Slice(reqs, func(i, j int) bool { return reqs[i].Name < reqs[j].Name })
	return reqs
}

// iteHelperTypeFromSuffix is the inverse of iteHelperSuffix for the
// shapes M3 actually uses (primitive names, named types). Compound
// types like arrays end up as no-op identifiers; a later milestone
// can refine if needed.
func iteHelperTypeFromSuffix(suffix string) string {
	switch suffix {
	case "uint32", "uint64", "uint8", "uint16", "int", "int64", "bool", "string", "Uint128":
		return suffix
	}
	return suffix
}

// emitIteHelper writes a single ite helper. ite(cond, t, f) returns t
// when cond is true, else f. Per D6 we specialise per type instead of
// using generics.
func (g *Generator) emitIteHelper(w *goWriter, req iteHelperReq) {
	w.linef("func %s(c bool, t, f %s) %s {", req.Name, req.Type, req.Type)
	w.line("\tif c { return t }")
	w.line("\treturn f")
	w.line("}")
	w.blank()
}

// avoid an unused-import complaint until M5 wires ivyAssume into more code paths
var _ goivy.Sort
