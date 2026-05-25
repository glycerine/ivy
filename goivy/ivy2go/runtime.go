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

// emitRuntimeHelpers writes the always-on helpers requested by every
// emitted program (ivyAssert, ivyAssume, ivyChoose, RNG source).
// Conditional helpers triggered by OnceGlobals flags are emitted
// later by emitRuntimeHelpersLate so emit* methods that run after
// runtime — like emitMain — can still request them.
func (g *Generator) emitRuntimeHelpers(w *goWriter) {
	g.emitRuntimePreamble(w)
}

// emitRuntimeHelpersLate writes the conditional helpers any earlier
// emit method has requested via Ctx.OnceGlobals. Called from the
// generate() orchestration AFTER every other emit step (state /
// actions / init / repl / main / native / …) has had a chance to
// register requirements.
func (g *Generator) emitRuntimeHelpersLate(w *goWriter) {
	if g.Ctx == nil {
		return
	}
	if g.Ctx.OnceGlobals["__need_uint128"] {
		g.emitUint128Helpers(w)
	}
	if g.Ctx.OnceGlobals["__need_bigint"] {
		g.emitBigIntHelpers(w)
	}
	if g.Ctx.OnceGlobals["__need_mixhash"] {
		g.emitMixHashHelper(w)
	}
	if g.Ctx.OnceGlobals["__need_lessord"] {
		g.emitLessOrdHelper(w)
	}
	if g.Ctx.OnceGlobals["__need_pickinput"] {
		g.emitPickInputHelpers(w)
	}
	if g.Ctx.OnceGlobals["__need_musthelpers"] {
		g.emitMustHelpers(w)
	}
	if g.Ctx.OnceGlobals["__need_testflags"] {
		g.emitTestFlagsHelper(w)
	}
	// ite_<type> helpers: walk OnceGlobals keys, find any starting
	// with "ite_", emit one per. Each call site recorded the helper
	// name in OnceGlobals via requestIteHelper.
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
	// os.Stdout; tests / embedders can override. Always declared
	// for non-class targets so action-prologue traces (`< name(…)`)
	// and the test-driver traces (`> name(…)`) emitted by the
	// test main always have a sink.
	if g.Config.RequestedTarget != "class" {
		g.Ctx.AddImport("runtime", "io", "")
		g.Ctx.AddImport("runtime", "os", "")
		w.line("// ivyTraceOut is the io.Writer used by all trace lines.")
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

// emitMixHashHelper writes a generic mixHash(uint64, any) → uint64
// FNV-1a step. We use any rather than generics per D6 (no generics
// in emitted code) — the function lives once and dispatches by type.
func (g *Generator) emitMixHashHelper(w *goWriter) {
	w.line("// mixHash mixes v into h via an FNV-1a step. Primitive types")
	w.line("// hash directly; struct types must implement Hash() uint64.")
	w.line("func mixHash(h uint64, v any) uint64 {")
	w.line("\tconst prime uint64 = 1099511628211")
	w.line("\tvar bits uint64")
	w.line("\tswitch x := v.(type) {")
	w.line("\tcase bool:")
	w.line("\t\tif x { bits = 1 } else { bits = 0 }")
	w.line("\tcase uint32:")
	w.line("\t\tbits = uint64(x)")
	w.line("\tcase uint64:")
	w.line("\t\tbits = x")
	w.line("\tcase int:")
	w.line("\t\tbits = uint64(x)")
	w.line("\tcase int64:")
	w.line("\t\tbits = uint64(x)")
	w.line("\tcase string:")
	w.line("\t\tfor i := 0; i < len(x); i++ {")
	w.line("\t\t\th = (h ^ uint64(x[i])) * prime")
	w.line("\t\t}")
	w.line("\t\treturn h")
	w.line("\tdefault:")
	w.line("\t\tif hv, ok := v.(interface{ Hash() uint64 }); ok {")
	w.line("\t\t\tbits = hv.Hash()")
	w.line("\t\t}")
	w.line("\t}")
	w.line("\treturn (h ^ bits) * prime")
	w.line("}")
	w.blank()
}

// emitLessOrdHelper writes a generic lessOrd(a, b) bool that orders
// primitive types and delegates to Less() for struct types.
func (g *Generator) emitLessOrdHelper(w *goWriter) {
	w.line("// lessOrd orders comparable primitives. Struct types must")
	w.line("// implement Less(other) bool.")
	w.line("func lessOrd(a, b any) bool {")
	w.line("\tswitch x := a.(type) {")
	w.line("\tcase bool:")
	w.line("\t\treturn !x && b.(bool)")
	w.line("\tcase uint32:")
	w.line("\t\treturn x < b.(uint32)")
	w.line("\tcase uint64:")
	w.line("\t\treturn x < b.(uint64)")
	w.line("\tcase int:")
	w.line("\t\treturn x < b.(int)")
	w.line("\tcase int64:")
	w.line("\t\treturn x < b.(int64)")
	w.line("\tcase string:")
	w.line("\t\treturn x < b.(string)")
	w.line("\t}")
	w.line("\tif lv, ok := a.(interface{ Less(any) bool }); ok {")
	w.line("\t\treturn lv.Less(b)")
	w.line("\t}")
	w.line("\treturn false")
	w.line("}")
	w.blank()
}

// emitPickInputHelpers writes the per-sort model-or-fallback helpers
// used by action_gen.go's Generate methods. Both helpers attempt to
// evaluate `sym` in the supplied ModelResult; on failure (nil model
// or non-bool / non-numeric result) they reroll via ivyChoose.
//
// String-parsing is the chosen extraction strategy because Z3 values
// always have a stable String() representation (mirroring how the
// generator package interacts with smt.Z3Expr).
func (g *Generator) emitPickInputHelpers(w *goWriter) {
	g.Ctx.AddImport("runtime", g.Config.GoivyImportPath, "")
	g.Ctx.AddImport("runtime", "strings", "")
	g.Ctx.AddImport("runtime", "strconv", "")
	// ModelResult.Solver is *smt.Z3Solver, not *goivy.Solver — to
	// reach the Translator we accept the goivy.Solver explicitly.
	w.line("// pickBoolOrChoose returns the model's value of sym (true/false)")
	w.line("// when the model evaluates to one, else falls back to ivyChoose.")
	w.line("func pickBoolOrChoose(sol *goivy.Solver, mr *goivy.ModelResult, sym *goivy.Const) bool {")
	w.line("\tif mr == nil || sol == nil { return ivyChoose(2) == 1 }")
	w.line("\tz, err := sol.Translator().TermToZ3(sym)")
	w.line("\tif err != nil { return ivyChoose(2) == 1 }")
	w.line("\tv, ok := mr.Eval(z)")
	w.line("\tif !ok { return ivyChoose(2) == 1 }")
	w.line("\ts := strings.TrimSpace(v.String())")
	w.line(`	if s == "true" { return true }`)
	w.line(`	if s == "false" { return false }`)
	w.line("\treturn ivyChoose(2) == 1")
	w.line("}")
	w.blank()
	w.line("// pickUintOrChoose returns the model's value of sym (as a uint64)")
	w.line("// clamped to [0, card), else falls back to ivyChoose.")
	w.line("func pickUintOrChoose(sol *goivy.Solver, mr *goivy.ModelResult, sym *goivy.Const, card int) uint64 {")
	w.line("\tif mr == nil || sol == nil { return uint64(ivyChoose(card)) }")
	w.line("\tz, err := sol.Translator().TermToZ3(sym)")
	w.line("\tif err != nil { return uint64(ivyChoose(card)) }")
	w.line("\tv, ok := mr.Eval(z)")
	w.line("\tif !ok { return uint64(ivyChoose(card)) }")
	w.line("\ts := strings.TrimSpace(v.String())")
	w.line("\tn, err := strconv.ParseUint(s, 10, 64)")
	w.line("\tif err != nil { return uint64(ivyChoose(card)) }")
	w.line("\tif card > 0 { n = n % uint64(card) }")
	w.line("\treturn n")
	w.line("}")
	w.blank()
}

// emitTestFlagsHelper writes parseTestItersFlag — used by the
// test-loop main to decide how many iterations to run. Flag
// precedence: `--iters=N` on argv → IVY_ITERS env var → default.
// Mirrors ivy2cpp's `test_iters` argv plumbing.
func (g *Generator) emitTestFlagsHelper(w *goWriter) {
	g.Ctx.AddImport("runtime", "os", "")
	g.Ctx.AddImport("runtime", "strconv", "")
	g.Ctx.AddImport("runtime", "strings", "")
	w.line("// parseTestItersFlag returns the iteration count for the")
	w.line("// test loop. Argv `--iters=N`, env `IVY_ITERS`, then the")
	w.line("// emit-time default (in that order).")
	w.line("func parseTestItersFlag(defaultIters int) int {")
	w.line("\tfor _, a := range os.Args[1:] {")
	w.line(`		if strings.HasPrefix(a, "--iters=") {`)
	w.line(`			if n, err := strconv.Atoi(a[len("--iters="):]); err == nil { return n }`)
	w.line("\t\t}")
	w.line("\t}")
	w.line(`	if v := os.Getenv("IVY_ITERS"); v != "" {`)
	w.line("\t\tif n, err := strconv.Atoi(v); err == nil { return n }")
	w.line("\t}")
	w.line("\treturn defaultIters")
	w.line("}")
	w.blank()
}

// emitMustHelpers writes the panic-on-failure facades for goivy
// constructors that return (T, error). The reifier in action_gen.go
// uses them so the emitted Pre-construction code reads cleanly
// (no error-plumbing per node).
func (g *Generator) emitMustHelpers(w *goWriter) {
	g.Ctx.AddImport("runtime", g.Config.GoivyImportPath, "")
	w.line("// mustApply wraps goivy.NewApply; panics on error.")
	w.line("// Used by the reified Pre-clause builders in action_gen.")
	w.line("func mustApply(fn goivy.Expr, args ...goivy.Expr) goivy.Expr {")
	w.line("\tres, err := goivy.NewApply(fn, args...)")
	w.line(`	if err != nil { panic(fmt.Sprintf("ivy reify: NewApply: %s", err)) }`)
	w.line("\treturn res")
	w.line("}")
	w.blank()
	w.line("// mustNewVariable wraps goivy.NewVariable; panics on error.")
	w.line("func mustNewVariable(name string, sort goivy.Sort) *goivy.LogicVariable {")
	w.line("\tres, err := goivy.NewVariable(name, sort)")
	w.line(`	if err != nil { panic(fmt.Sprintf("ivy reify: NewVariable %q: %s", name, err)) }`)
	w.line("\treturn res")
	w.line("}")
	w.blank()
	w.line("// mustNewIte wraps goivy.NewIte; panics on error.")
	w.line("func mustNewIte(cond, t, e goivy.Expr) goivy.Expr {")
	w.line("\tres, err := goivy.NewIte(cond, t, e)")
	w.line(`	if err != nil { panic(fmt.Sprintf("ivy reify: NewIte: %s", err)) }`)
	w.line("\treturn res")
	w.line("}")
	w.blank()
	w.line("// mustNewFunctionSort wraps goivy.NewFunctionSort; panics on")
	w.line("// error. Takes one or more sorts; the last is the range.")
	w.line("func mustNewFunctionSort(sorts ...goivy.Sort) *goivy.LogicFunctionSort {")
	w.line("\tres, err := goivy.NewFunctionSort(sorts...)")
	w.line(`	if err != nil { panic(fmt.Sprintf("ivy reify: NewFunctionSort: %s", err)) }`)
	w.line("\treturn res")
	w.line("}")
	w.blank()
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
