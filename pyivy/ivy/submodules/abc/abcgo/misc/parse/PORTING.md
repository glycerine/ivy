# Boolean formula parser port

All files in this package compile as pure Go. Source references are
`src/misc/parse/parseCore.c`, `parseEqn.c`, and `parseStack.c`.

`ParseFormula` builds a backend-independent expression tree supporting the
original full formula operators, precedence and left associativity: constants,
variables, prefix `!`/`~`, suffix `'`, `&`/`*`, `^`/`<+>`, `+`/`|`, `<=>`, `=>`,
`<=`, implicit conjunction, grouping, and bracket rank shifts.
`ParseEquation` accepts the original stricter `!`, `*`, `^`, `+`, parentheses
grammar with explicit binary operators.

`Operations[T]` supplies typed graph constructors; `Compile` creates graph nodes
without importing an AIG or BDD implementation. `Shift(node, delta)` receives
`-rankWidth` for each bracket pair, matching `Extra_bddMove(dd,node,-nVars)`.
The concrete backend owns variable-index validity, permutation and graph
canonicalization. `Expression.Evaluate` supports ordinary Boolean assignments;
ranked expressions require a shift-capable graph backend.

The original-name generic adapters `Parse_FormulaParser` and
`Parse_FormulaParserEqn` preserve diagnostic writers and zero-result-on-error
behavior. The C manager argument becomes `Operations[T]`; variable-name vectors
become `[]string`, and core zero-rank variable pointers become `[]T`. Consumers
must supply their concrete manager's constructor callbacks. Hop and Cudd owners
have been given this integration contract. Graph reference counts and temporary
C allocations are replaced by Go lifetimes.

Stacks expose both original entry points and idiomatic typed methods. They grow
rather than overflowing a fixed capacity; underflow panics with a diagnostic.

`go test ./misc/parse` passes, covering exhaustive operator truth assignments,
precedence, strict equation rejection, malformed input, diagnostic entry points,
rank callback offsets, and stack growth/lifetime.
