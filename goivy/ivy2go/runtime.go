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
	g.emitRuntimeImplPreamble(w)
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
	if g.Ctx.OnceGlobals["__need_progress_check"] {
		g.emitProgressCheckHelper(w)
	}
	// ite_<type> helpers: walk OnceGlobals keys, find any starting
	// with "ite_", emit one per. Each call site recorded the helper
	// name in OnceGlobals via requestIteHelper.
	for _, h := range g.collectIteHelpers() {
		g.emitIteHelper(w, h)
	}
}

// emitRuntimeImplPreamble writes the package-level helpers every emitted
// program needs at runtime.
func (g *Generator) emitRuntimeImplPreamble(w *goWriter) {
	// Import list for runtime.go.
	if g.Ctx != nil {
		g.Ctx.AddImport("runtime", "encoding/binary", "")
		g.Ctx.AddImport("runtime", "fmt", "")
		g.Ctx.AddImport("runtime", "math/rand/v2", "rand")
	}

	w.line("// ivyRand is the package-wide ChaCha8 RNG used by ivyChoose,")
	w.line("// the test-loop weighted scheduler, and any nondet helpers.")
	w.line("// Same algorithm + seeding convention as ivy_to_cpp (Python")
	w.line("// + Go-ported C++ chacha8c.hpp) so the random stream is")
	w.line("// byte-equivalent across all three tools for any given")
	w.line("// `seed=N` argument. Reseeded from `--seed=N` / IVY_SEED by")
	w.line("// applyTestSeedFlag in the test main.")
	w.line("var ivyRand = rand.NewChaCha8(ivySeedBytes(1))")
	w.blank()
	w.line("// ivySeedBytes mirrors cpp's `std::memcpy(seed32, &seed, sizeof(seed))`")
	w.line("// pattern: the seed integer's little-endian bytes occupy the")
	w.line("// first 4 bytes of the 32-byte ChaCha8 key; the rest stay zero.")
	w.line("func ivySeedBytes(seed uint32) [32]byte {")
	w.line("\tvar key [32]byte")
	w.line("\tbinary.LittleEndian.PutUint32(key[:4], seed)")
	w.line("\treturn key")
	w.line("}")
	w.blank()
	w.line("// ivyRand31 mirrors cpp's `chacha8c::Rand()` —")
	w.line("// `static_cast<int>(Uint64() >> 33)` — yielding a non-negative")
	w.line("// 31-bit int in [0, 2^31). Pairs with biased `% N` reduction")
	w.line("// the same way cpp does so action selection sequences match.")
	w.line("func ivyRand31() int {")
	w.line("\treturn int(ivyRand.Uint64() >> 33)")
	w.line("}")
	w.blank()
	w.line("// ivyRandMaxPlus1 is the cpp `(double)RAND_MAX + 1.0` constant")
	w.line("// (2^31 = 2147483648.0). Used by the weighted action picker to")
	w.line("// scale ivyRand31() into [0, 1.0) the same way cpp does.")
	w.line("const ivyRandMaxPlus1 = 2147483648.0")
	w.blank()

	w.line("// ivyRandomRange mirrors ivy_z3_gen.hpp's `random_range` —")
	w.line("// a single chacha8 Uint64() reduced into [lo, hi]. The cpp")
	w.line("// implementation does `res = __chacha8c_rng.Uint64(); if (card")
	w.line("// != -1) res = res % (card+1) + lo`. Both ivy2cpp-emitted and")
	w.line("// ivy2go-emitted binaries consume one chacha8 word per action")
	w.line("// input under matching seeds, so the per-input randomization")
	w.line("// preferences pinned into the solver are byte-equivalent.")
	w.line("func ivyRandomRange(lo, hi uint64) uint64 {")
	w.line("\tres := ivyRand.Uint64()")
	w.line("\tcard := hi - lo")
	w.line("\tif card != ^uint64(0) {")
	w.line("\t\tres = res%(card+1) + lo")
	w.line("\t}")
	w.line("\treturn res")
	w.line("}")
	w.blank()
	w.line("func ivyRandomIntRange(lo, hi int64) int64 {")
	w.line("\tif hi < lo { return lo }")
	w.line("\tspan := uint64(hi - lo)")
	w.line("\treturn lo + int64(ivyRandomRange(0, span))")
	w.line("}")
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

	// ivyAssume panics on failure — the per-action solver (see
	// actionGen_*.generate in action_gen.go) is the single source of
	// truth for whether an action can fire. UNSAT preconditions
	// cause generate() to return false and execute() is never
	// called, so a live panic from ivyAssume means either the
	// solver lost a precondition or some non-test-driver caller
	// invoked the action directly. Either way it's a real bug, and
	// panicking (rather than log.Fatalf) keeps the stack trace.
	w.open("func ivyAssume(cond bool, label string) {")
	w.open("if !cond {")
	w.line(`panic(fmt.Sprintf("ivy assume failed: %s (solver precondition was not enforced)", label))`)
	w.close("")
	w.close("")
	w.blank()

	// ivyChoose mirrors cpp's `___ivy_choose(0, …)` reduction —
	// `chacha8c::Rand() % N` (biased modulo) — so the per-call output
	// matches cpp's sequence byte-for-byte under the same seed.
	w.open("func ivyChoose(rng int) int {")
	w.open("if rng <= 0 {")
	w.line("return 0")
	w.close("")
	w.line("return ivyRand31() % rng")
	w.close("")
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
	// callers may pass uint32 / uint64 / *big.Int — each has a
	// distinct path. We use interface{} to admit them all.
	w.line("// toBigInt coerces a value to *big.Int. Accepts uint32,")
	w.line("// uint64, *big.Int, or any int-typed Go value.")
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
	case "uint32", "uint64", "uint8", "uint16", "int", "int64", "bool", "string":
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
	w.line("func pickRangeOrChoose(sol *goivy.Solver, mr *goivy.ModelResult, sym *goivy.Const, lo, hi int64) int64 {")
	w.line("\tif hi < lo { return lo }")
	w.line("\tchoose := func() int64 { return lo + int64(ivyChoose(int(hi-lo+1))) }")
	w.line("\tif mr == nil || sol == nil { return choose() }")
	w.line("\tz, err := sol.Translator().TermToZ3(sym)")
	w.line("\tif err != nil { return choose() }")
	w.line("\tv, ok := mr.Eval(z)")
	w.line("\tif !ok { return choose() }")
	w.line("\ts := strings.TrimSpace(v.String())")
	w.line("\tn, err := strconv.ParseInt(s, 10, 64)")
	w.line("\tif err != nil { return choose() }")
	w.line("\tif n < lo { return lo }")
	w.line("\tif hi < n { return hi }")
	w.line("\treturn n")
	w.line("}")
	w.blank()
	w.line("// pickEnumOrChoose is the enum-aware variant of pickUintOrChoose.")
	w.line("// goivy's model returns enum values as the symbolic name (e.g.")
	w.line("// \"green\"), not the integer index. We look up the name in")
	w.line("// `names` to recover the index. Fallback path matches")
	w.line("// pickUintOrChoose's ivyChoose call so PRNG consumption stays")
	w.line("// aligned with the cpp side when the solver model is absent.")
	w.line("func pickEnumOrChoose(sol *goivy.Solver, mr *goivy.ModelResult, sym *goivy.Const, names []string) uint64 {")
	w.line("\tcard := len(names)")
	w.line("\tif mr == nil || sol == nil { return uint64(ivyChoose(card)) }")
	w.line("\tz, err := sol.Translator().TermToZ3(sym)")
	w.line("\tif err != nil { return uint64(ivyChoose(card)) }")
	w.line("\tv, ok := mr.Eval(z)")
	w.line("\tif !ok { return uint64(ivyChoose(card)) }")
	w.line("\ts := strings.TrimSpace(v.String())")
	w.line("\tfor i, n := range names {")
	w.line("\t\tif n == s { return uint64(i) }")
	w.line("\t}")
	w.line("\t// Numeric fallback — some translators still encode enums as ints.")
	w.line("\tif n, err := strconv.ParseUint(s, 10, 64); err == nil {")
	w.line("\t\tif card > 0 { n = n % uint64(card) }")
	w.line("\t\treturn n")
	w.line("\t}")
	w.line("\treturn uint64(ivyChoose(card))")
	w.line("}")
	w.blank()
}

// emitTestFlagsHelper writes parseTestItersFlag + applyTestSeedFlag.
// Mirrors ivy2cpp's `test_iters` + `seed` argv plumbing so cpp / go
// emitted binaries accept the same flags AND produce the same random
// sequence under the same seed (ChaCha8 + ivySeedBytes convention).
// Flag precedence: `--iters=N` / `--seed=N` on argv → IVY_ITERS /
// IVY_SEED env vars → the emit-time default.
//
// Also accepts the cpp-style positional `seed=N` form (no leading
// `--`) so the same argv literally invokes both binaries the same
// way for cross-tool comparison.
func (g *Generator) emitTestFlagsHelper(w *goWriter) {
	g.Ctx.AddImport("runtime", "os", "")
	g.Ctx.AddImport("runtime", "strconv", "")
	g.Ctx.AddImport("runtime", "strings", "")
	w.line("// parseTestItersFlag returns the iteration count for the")
	w.line("// test loop. Argv `--iters=N` / `iters=N`, env `IVY_ITERS`,")
	w.line("// then the emit-time default (in that order).")
	w.line("func parseTestItersFlag(defaultIters int) int {")
	w.line("\tfor _, a := range os.Args[1:] {")
	w.line(`		for _, pfx := range []string{"--iters=", "iters="} {`)
	w.line("\t\t\tif strings.HasPrefix(a, pfx) {")
	w.line("\t\t\t\tif n, err := strconv.Atoi(a[len(pfx):]); err == nil { return n }")
	w.line("\t\t\t}")
	w.line("\t\t}")
	w.line("\t}")
	w.line(`	if v := os.Getenv("IVY_ITERS"); v != "" {`)
	w.line("\t\tif n, err := strconv.Atoi(v); err == nil { return n }")
	w.line("\t}")
	w.line("\treturn defaultIters")
	w.line("}")
	w.blank()
	w.line("// applyTestSeedFlag re-seeds ivyRand when `--seed=N` or the")
	w.line("// cpp-style `seed=N` positional is on argv (or IVY_SEED is")
	w.line("// set). Uses the same ivySeedBytes packing cpp uses so the")
	w.line("// ChaCha8 stream is byte-identical for the same seed across")
	w.line("// ivy_to_cpp, ivy2cpp, and ivy2go binaries.")
	w.line("func applyTestSeedFlag() {")
	w.line("\tapply := func(s string) bool {")
	w.line("\t\tn, err := strconv.ParseUint(s, 10, 32)")
	w.line("\t\tif err != nil { return false }")
	w.line("\t\tivyRand = rand.NewChaCha8(ivySeedBytes(uint32(n)))")
	w.line("\t\treturn true")
	w.line("\t}")
	w.line("\tfor _, a := range os.Args[1:] {")
	w.line(`		for _, pfx := range []string{"--seed=", "seed="} {`)
	w.line("\t\t\tif strings.HasPrefix(a, pfx) {")
	w.line("\t\t\t\tif apply(a[len(pfx):]) { return }")
	w.line("\t\t\t}")
	w.line("\t\t}")
	w.line("\t}")
	w.line(`	if v := os.Getenv("IVY_SEED"); v != "" {`)
	w.line("\t\tapply(v)")
	w.line("\t}")
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

// emitProgressCheckHelper emits the ivyCheckProgress runtime helper
// invoked by the Tick() method when a progress counter exceeds its
// implied-rely bound. Mirrors the cpp `ivy_check_progress` invariant
// check — it panics with a clear message when the bound is violated,
// because (per CLAUDE.md guidance) panic + stack trace beats a
// silent liveness failure.
func (g *Generator) emitProgressCheckHelper(w *goWriter) {
	g.Ctx.AddImport("runtime", "fmt", "")
	w.line("// ivyCheckProgress is invoked from State.Tick() to enforce a")
	w.line("// rely-derived upper bound on a progress counter. Panics when")
	w.line("// the bound is exceeded — liveness violation detected at runtime.")
	w.line("func ivyCheckProgress(counter, bound int) {")
	w.line("\tif counter > bound {")
	w.line(`		panic(fmt.Sprintf("ivy liveness check failed: progress counter %d exceeds rely bound %d", counter, bound))`)
	w.line("\t}")
	w.line("}")
	w.blank()
}

// avoid an unused-import complaint until M5 wires ivyAssume into more code paths
var _ goivy.Sort
