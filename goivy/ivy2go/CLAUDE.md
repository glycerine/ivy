This package is `ivy2go`. It is a sibling of `ivy2cpp` that emits **Go**
source code from a goivy Module instead of C++.

See `ARCHITECTURE_TODO.md` in this directory for the full architecture
and milestone plan.

## A. Source of truth

`ivy2go` mirrors `~/ivy/goivy/ivy2cpp/` file-for-file. ivy2cpp in turn
mirrors `~/ivy/pyivy/ivy/ivy/ivy_to_cpp.py` line-for-line (see the
parent goivy/CLAUDE.md). So the dependency chain is:

    pyivy/ivy_to_cpp.py   (Python source of truth)
        ↓ mechanical port
    goivy/ivy2cpp/        (Go that emits C++)
        ↓ mechanical port, swap C++ → Go at the leaves
    goivy/ivy2go/         (Go that emits Go)

When a question of "what should this do?" comes up:

1. First look at the corresponding file in `goivy/ivy2cpp/`.
2. If the C++-emission code there has a `// Python ivy_to_cpp.py:NNN`
   comment, follow that to the Python.
3. Mechanically translate: swap C++ emission for Go emission, keep
   the call graph, function names, struct names, memoization caches,
   and `// Python ivy_to_cpp.py:NNN` comments intact.

## B. Mechanical port rules (specialised for ivy2go)

These supplement the parent `goivy/CLAUDE.md` rules.

1. **One ivy2cpp file → one ivy2go file** with the same base name.
   Exceptions allowed: `cpp_context.go → go_context.go`,
   `cpp_types.go → go_types.go`, and `config.go` is split out from
   `generator.go` for clarity (see `ARCHITECTURE_TODO.md` §3.3).
   `build_findvs*.go` has no analogue (Go has one toolchain).

2. **One ivy2cpp function → one ivy2go function** with the same name.
   `emitExpr` in ivy2cpp/expr.go has a counterpart `emitExpr` in
   ivy2go/expr.go. `emitHeader` becomes `emitTypes` + `emitState`
   etc. because there is no header/impl split — document the
   correspondence in a comment at each split point.

3. **Preserve Python-line comments.** When ivy2cpp has
   `// Python ivy_to_cpp.py:1948 ...`, ivy2go has the same comment
   on the analogous function, possibly with a "via ivy2cpp/foo.go:NN"
   suffix so reviewers can follow the chain.

4. **Generated Go must be `gofmt` clean.** Every generated `.go` file
   round-trips through `go/format.Source`. The test harness enforces
   this in Tier 3.

5. **No generics in generated code.** Helpers are specialised per
   sort during emission. The ivy2go *generator* package may use
   generics in its own internals, but the *emitted* Go must not.

6. **Z3 only through `goivy.Solver` / `goivy.Translator`.** Neither
   the generator nor the generated programs may import
   `goivy/smt/Z3Solver` directly. Emitted programs `import
   "github.com/glycerine/ivy/goivy"` and call into that facade.

7. **No global mutable state.** Per parent goivy/CLAUDE.md section C,
   all per-session state lives on the `Generator` struct or its
   `Config`. No package-level `var` for counters, caches, or flags.

8. **No git.** Per parent goivy/CLAUDE.md section D, the user commits
   in the background. Do not run git commands.

9. **Test scoping.** `go test ./ivy2go -count=1` for fast unit tests;
   `SLOW_GO_TEST=1 go test ./ivy2go -run TestSmoke` for build smoke.
   Never `go test ./...` — XTRACE makes the full suite painful
   (parent goivy/CLAUDE.md section 9).

10. **Audit hygiene.** Any `TODO`/`DEFER`/`XXX` comment in
    `ivy2go/*.go` must reference an item in `AUDIT_TODO_LIVE.md`.
    `ivy2go/comments_test.go` (M10) enforces this once it lands.

## C. Output package conventions

The generator emits a Go package directory per Ivy module (see
`ARCHITECTURE_TODO.md` §3.3 for the file split). Generated files
should:

- Use tab indentation (`go/format.Source` will rewrite anyway).
- Group imports in stdlib / external blocks.
- Name the exported state type `State`, the constructor `NewState`,
  and action methods `ActFoo` (PascalCase Ivy action name).
- Never embed a build timestamp or random suffix — keep emission
  deterministic so diffs are meaningful.

## D. Where to look first

| Question | Look here |
|----------|-----------|
| How does ivy2cpp do X? | `ivy2cpp/{file}.go` first; follow Python comments. |
| What's the test pattern? | `ivy2cpp/ivy2cpp_test.go` (200+ functions, table-driven). |
| What goivy types are available? | `~/ivy/goivy/` top level; especially `module.go`, `ivylogic_*.go`, `z3bridge_*.go`. |
| What's a current gap vs ivy2cpp? | `ivy2go/AUDIT_TODO_LIVE.md`. |
| What's the milestone roadmap? | `ivy2go/ARCHITECTURE_TODO.md` §4. |
