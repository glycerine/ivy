# Audit: Missing Go Functionality vs Python Ivy

**Created:** 2026-05-02 05:45 UTC

This audit compares `~/ivy/pyivy/ivy/ivy/` (Python, source of truth) against `~/ivy/goivy/` (Go port). Items are ordered by functional severity. Each item can seed a separate planning cycle.

---

## 1. UPDR/PDR Verification Engine (Placeholder)

**Python:** `ivy_updr.py` (full PDR implementation using Z3 solver)  
**Go:** `updr/updr.go:42-74` — `CheckModule()` is a placeholder that creates a trivial safe system (`x=false` init, `x'=x` trans, `false` bad) and returns "always safe" without doing real verification.

The Python UPDR (Unbounded Property-Directed Reachability) is the core verification algorithm. It uses the Z3 solver to iteratively strengthen inductive invariants until either proving a property or finding a counterexample trace. The Go stub demonstrates the structural scaffolding (init/trans/bad triple, frame sequence) but skips the actual solver interaction, clause generalization, predecessor computation, and frame propagation that make PDR work. All interactive verification features in the webui that invoke PDR (the "PDR step" button, induction checks, bounded checks) are dead because this backend is hollow.

**Dependencies:** Requires z3bridge to be fully wired, plus actions/transrel (already ported) for forward/reverse image computation.

---

## 2. WebUI Session Solver Wiring (8 Operations)

**Python:** `ivy_ui.py`, `ivy_ui_cti.py`, `ivy_graph_ui.py` — all have working implementations that call into `ivy_interp`, `ivy_transrel`, `ivy_art`, `ivy_solver`  
**Go:** `webui/session.go:287-378` — 8 case branches emit "not yet wired" status messages and return without computation:

| Line | Operation | What Python Does |
|------|-----------|-----------------|
| 289 | PDR step | Calls `ag.bmc_conjecture()` or `ag.check_bounded_safety()` via UPDR |
| 292 | Concrete | Computes concrete model via Z3 `get_model_clauses()` |
| 295 | Reverse | Computes reverse image via `transrel.reverse_image()` + Z3 |
| 298 | Reach | Reachability analysis via `art` + Z3 |
| 301 | Weaken | Weakens invariant by dropping clauses |
| 304 | Save abstraction | Exports concept spaces to file |
| 337 | Add relation | Parses formula string and adds to concept domain |
| 377 | Action execution | Dispatches to `interp.eval_action()` engine |

These are the user-facing interactive verification operations in the web UI. Each one requires the session to hold an `AnalysisGraph` with a live Z3 solver context, call into the interp/transrel/art pipeline, and return structured results (models, counterexample traces, updated ARG states) back to the frontend. The Go side has the HTTP handlers and SSE event plumbing built, but the core computation is missing.

**Dependencies:** Items 1 (UPDR), interp pipeline, z3bridge integration.

---

## 3. WebUI CTI (Counter-To-Induction) Stubs

**Python:** `ivy_ui_cti.py`, `ivy_graph_ui.py`  
**Go:** `webui/ui_cti.go` — 10 functions are stubs:

| Line | Function | What Python Does |
|------|----------|-----------------|
| 96 | `AutodetectTransitive()` | Scans axioms for transitive+symmetry relations, enables checkboxes |
| 119 | `BoundedCheck()` | Creates AnalysisGraph, executes BMC steps, checks condition at each bound |
| 142 | `CheckInductiveness()` | Creates AnalysisGraph, computes dual_clauses, checks induction |
| 165 | `Diagram()` | Computes reverse image, gets Z3 model, extracts state diagram |
| 223 | `ShowUsedRelations()` | Parses clause string, finds used symbols, enables checkboxes |
| 291 | `GatherFacts()` | Collects node facts and edge facts from concept graph, filters duplicates |
| 319 | `GetSelectedConjecture()` | Negates gathered facts, substitutes skolem constants with universally-quantified variables |
| 328 | `MinimizeConjecture()` | Uses BMC + unsat core to minimze conjecture clause set |
| 336 | `IsSufficient()` | Creates AnalysisGraph, checks pre-state implies post-state under conjecture |
| 344 | `IsInductive()` | Self-induction check: checks conjecture is inductive relative to itself |

The CTI workflow is the interactive invariant strengthening loop: the user sees a counterexample to induction, selects facts from the concept graph that distinguish good from bad states, gathers them into a conjecture, minimizes it, checks sufficiency, and adds it to the invariant. Every step in this loop is stubbed. Without these, the web UI can display graphs but cannot actually do interactive verification.

**Dependencies:** Items 1, 2, concept session Z3 integration.

---

## 4. WebUI ARG (Analysis Reachability Graph) Navigation Stubs

**Python:** `ivy_ui.py`, `ivy_graph_ui.py`  
**Go:** `webui/ui_main.go` — 15 functions are stubs:

| Line | Function | What Python Does |
|------|----------|-----------------|
| 194 | `NodeExecuteCommands()` | Returns available actions for a state node |
| 257 | `CheckLocalSafety()` | Checks safety at a specific ARG node via Z3 |
| 263 | `CheckBoundedSafety()` | Bounded safety check at node |
| 269 | `FindExtension()` | Finds successor states from node via action execution |
| 275 | `ExecuteAction()` | Executes named action at node, creates new ARG state |
| 281 | `RecalculateAll()` | Recalculates all transition edges in ARG |
| 286 | `RecalculateEdge()` | Recalculates a single transition edge |
| 292 | `DecomposeEdge()` | Decomposes transition into sub-steps |
| 298 | `ViewSourceEdge()` | Locates source code for transition action |
| 318 | `CoverNode()` | Marks node as covered by another (subsumption) |
| 329 | `JoinNode()` | Joins node with another state |
| 335 | `TryConjecture()` | Tests conjecture at node via BMC or concept graph |
| 351 | `TryRememberedGraph()` | Sets up remembered concept graph for node |
| 376 | `BMC()` | Bounded model checking at node |
| 409 | `TryProperty()` | Checks property and launches BMC |

These functions implement the ARG exploration workflow: building the state graph by executing actions, checking safety, finding extensions, decomposing transitions, covering nodes. Together with item 3, they constitute the entire interactive verification experience.

**Dependencies:** Items 1, 2, art/AnalysisGraph operations.

---

## 5. WebUI Concept Graph Operations Stubs

**Python:** `ivy_graph_ui.py`, `concept_interactive_session.py`  
**Go stubs in multiple files:**

| File | Line | Function | What Python Does |
|------|------|----------|-----------------|
| `graph_widget.go` | 284 | `Splatter()` | Splits concept node into finer concepts |
| `graph_widget.go` | 310 | `Recalculate()` | Recomputes concept graph from current state |
| `graph_model.go` | 485 | `MaterializeEdge()` | Invokes Z3 to find witness for edge relation |
| `graph_model.go` | 498 | `AddConstraints()` | Adds constraints to concept session |
| `graph_model.go` | 505 | `SetFacts()` | Sets facts in concept session |
| `graph_model.go` | 510 | `GetFacts()` | Queries concept session for current facts |
| `concept_session.go` | 113 | `Materialize()` | Invokes Z3 to find witness constant for concept |
| `ext_api.go` | 429 | `RegisterArgNewGoal()` | Pushes new proof goal from ARG node |
| `ext_api.go` | 437 | `RegisterArgRecalculate()` | Recalculates facts for ARG node |
| `ext_api.go` | 445 | `RegisterArgCheckCover()` | Checks covering relation between ARG nodes |
| `ext_api.go` | 453 | `RegisterArgRemoveFacts()` | Removes selected facts from ARG node |
| `ext_api.go` | 461 | `RegisterArgJoin()` | Joins two ARG nodes |

The concept graph is the abstract view of system state: nodes represent sorts, edges represent relations, and the user interacts by materializing witnesses, adding constraints, and gathering facts. These operations all require Z3 solver calls to compute the abstract interpretation. The Go types and UI plumbing exist, but the solver invocations are missing.

**Dependencies:** z3bridge, concept domain solver integration.

---

## 6. Proof Tactics: Skolem Witness (propertyTactic)

**Python:** `ivy_proof.py` — `property_tactic()` handles `optskolem` in proof syntax, creates Skolem function witnesses for existential quantifiers in property proofs  
**Go:** `proof/tactics.go:519-522` — returns error "property tactic: Skolem function witness not implemented"

In Python, when a proof uses `property [propname] with optskolem`, the tactic creates fresh Skolem function symbols that witness the existential quantifiers in the property being proved. This allows the prover to refer to specific witnesses in subsequent proof steps. The Go grammar does not yet parse `optskolem`, so the tactic cannot receive the Skolem specification. This blocks proofs that rely on existential witnesses.

**Dependencies:** Parser grammar change to parse `optskolem`, then implementation of Skolem function creation in the proof context.

---

## 7. BMC Loop Unrolling

**Python:** `ivy_bmc.py` — unrolls `While` loops in actions before bounded model checking  
**Go:** `bmc/bmc.go:244-248` — `UnrollAction()` is a stub that returns the action unchanged

Bounded model checking requires finite unrolling of loops. The Python implementation recursively traverses the action AST, finds `While` nodes, and replaces them with `n` copies of the loop body guarded by the loop condition (with a final assume-not-condition). The Go stub means BMC silently produces wrong results on any action containing loops — it checks the action as if the loop body executes zero times.

**Also:** `check/vmt.go:49` notes "For fsmc method, loops would be unrolled first — not yet implemented."

**Dependencies:** None beyond existing AST/action infrastructure.

---

## 8. Z3 Enumerated Types as Native Sorts

**Python:** `ivy_solver.py` — `enumerated_to_numeral()` converts enumerated types to Z3 native enumeration sorts  
**Go:** `z3bridge/translate.go:1141` — returns error "cannot interpret enumerated type as a native sort (not yet supported)"

When Ivy declares `type color = {red, green, blue}`, Python can represent this directly as a Z3 enumeration sort with three constructors. The Go bridge falls back to uninterpreted sorts with distinctness axioms, which is sound but less efficient for Z3 and may cause issues with interpreted operations on enumerated types. This matters for model checking where enumeration cardinality bounds are used.

**Dependencies:** z3bridge API for creating enumeration sorts.

---

## 9. Art Analysis Graph Panic Stubs

**Go:** `art/art.go`

| Line | Function | Status |
|------|----------|--------|
| 1130 | `MakeConcreteTrace()` | `panic("TODO: AnalysisGraph.MakeConcreteTrace() is stubbed.")` |
| 1414 | `CheckConstraints()` | `panic("TODO: implement if needed")` |
| 1421 | `StratifyGoals()` | `panic("TODO: implement if needed")` |

**Python status:** `ivy_art.py:404-406` — `MakeConcreteTrace` is *also* stubbed in Python (`# TODO\nreturn`). `CheckConstraints` and `StratifyGoals` do not appear in the Python source at all — these were Go-side placeholders that may never be needed.

`MakeConcreteTrace` would construct a concrete execution trace from an abstract counterexample path in the ARG. This is used for counterexample presentation after a safety violation is found. Since Python also has this stubbed, this is a known limitation of the upstream Ivy project itself, not a porting gap. However, the Go version panics instead of silently returning, which is worse — any code path that reaches it will crash.

**Recommendation:** Change panics to log-and-return-nil to match Python behavior. Implementing real concrete trace extraction is a separate, larger project that isn't done in Python either.

---

## 10. WebUI ShowVerification Entry Point

**Python:** `ivy_show.py` — reads parameters, loads source file, creates isolate, launches UI main loop  
**Go:** `webui/ui_show.go:26-30` — `ShowVerification()` is a stub with TODO comments

This is the entry point that connects "load an .ivy file" to "show its verification state in the web UI." The Go stub creates a session but doesn't parse the file, create the isolate, or populate the analysis graph. Without this, the web UI can only work with programmatically-constructed modules, not with .ivy source files loaded from disk.

**Dependencies:** Parser (already ported), isolate creation (already ported), compiler (already ported). This is mainly wiring work.

---

## 11. LSP Server

**Python:** No Python equivalent (Python Ivy uses Jupyter notebooks and Tkinter)  
**Go:** `ivylsp/ivylsp.go` — `Complete()` returns nil, `StartIO()` prints "not yet implemented"

The Language Server Protocol server would provide IDE integration (VS Code, etc.) with Ivy language features: completion, diagnostics, go-to-definition, hover info. This is a Go-only feature with no Python counterpart to port from. It would need to use the existing parser and compiler infrastructure to provide real-time feedback.

**Dependencies:** Parser, compiler, module system (all ported). Design decision needed on scope.

---

## 12. WebUI Session Proof Stack

**Go:** `webui/session.go:390` — `ProofStackData()` comment says "will be wired to the proof/ package"

The proof stack shows the current proof state: active goals, applied tactics, remaining subgoals. The Go proof/ package has the full ProofChecker, GoalStack, and tactic infrastructure already ported. This function just needs to serialize the proof state into the JSON format that the frontend expects. The frontend JS (`ivyweb_graph.js`) already has proof stack rendering support.

**Dependencies:** Mostly wiring — proof/ package is functional.

---

## Summary Table

| # | Item | Stub Count | Severity | Python Also Stubbed? |
|---|------|-----------|----------|---------------------|
| 1 | UPDR/PDR engine | 1 | Critical | No — Python works |
| 2 | Session solver wiring | 8 | Critical | No |
| 3 | CTI stubs | 10 | Critical | No |
| 4 | ARG navigation stubs | 15 | Critical | No |
| 5 | Concept graph ops | 12 | High | No |
| 6 | Skolem witness tactic | 1 | Medium | No |
| 7 | BMC loop unrolling | 1+1 | Medium | No |
| 8 | Z3 enum native sorts | 1 | Medium | No |
| 9 | Art panic stubs | 3 | Low | Yes (Python also stubbed) |
| 10 | ShowVerification entry | 1 | Medium | No |
| 11 | LSP server | 2 | Low | N/A (Go-only feature) |
| 12 | Proof stack wiring | 1 | Low | No |

**Total genuine stubs/gaps: ~57 functions**

Items 1-5 are all interconnected: the UPDR engine (1) feeds the session operations (2), which feed the CTI workflow (3), ARG navigation (4), and concept graph operations (5). Fixing them in order would progressively light up the interactive verification UI.
