# starting-point implementation and experiment plan

Last updated: 2026-11-11

Status: the bounded Satoko experiment is validated. Token-based drafts now cover
ordinary C declarations/functions, protected preprocessing regions, and a
catalog-driven standard-library conversion pass with native-C differential tests.
Full C/C++ semantics and whole-project compilation remain incomplete; speedup
over an LLM-only baseline is not established.

## Purpose

Build a portable Go library in `starting-point/`, with a thin command-line wrapper in
`starting-point/cmd/starting-point/main.go`, that mechanically converts C and C++ into readable
draft Go. An LLM completes the port using the original source, generated drafts,
explicit representation decisions, and a structured list of unresolved work.

ABC is the first test bed. The primary objective is to reduce the time and number
of semantic decisions needed to produce a validated port. Minutes or hours per
bounded subsystem is a hypothesis to test, not a promise for the entire suite.

The first experiment is a small container utility plus a real consumer, taken
through the entire process. The next experiment exercises a difficult data
representation, such as clause storage or complemented AIG edges.

## Working rules and scope

- Do not use git commands. Repository version control belongs to the user.
- Keep original ABC sources unchanged during translation experiments.
- Every newly generated Go draft file starts with a configurable package declaration.
- Translate function-like macros to ordinary Go functions where possible; leave
  inlining to the Go compiler. Record type and argument-evaluation obligations.
- Lower ternary expressions to if/else, preserving branch laziness and comments.
- Preserve source identifiers, comments, ordering, and recognizable control flow
  wherever practical. Do not preserve license notices in derived artifacts.
- Prefer ordinary Go, but prioritize correct behavior and source correspondence
  over cosmetic renaming or premature redesign.
- Use regex-driven transformations supported by a lossless lexer, delimiter
  matching, and limited structural parsing. Do not require Clang.
- Recognize unsupported constructs explicitly; never silently invent behavior.
- Keep project-specific policies separate from reusable translation machinery.
- Keep generated drafts separate from repaired Go. Regeneration must not
  overwrite completed LLM or human work.
- Support C and C++ incrementally. Full standards coverage, binary compatibility,
  and automatic translation of arbitrary C++ are outside the initial scope.
- A compiler may be needed for reference tests; it is not a translator runtime
  dependency. The initial tool must not call an LLM service itself.

## Success measures

Measure each experiment using a fixed source scope and recorded configuration.

| Measure | What to record |
| --- | --- |
| Time to validated port | Translation, LLM repair, human review, and validation separately |
| Setup cost | Tool development and project-policy work, separate from per-port cost |
| Remaining decisions | Unresolved semantic issues before and after repair |
| Reusable improvement | Fixes promoted into rules or profiles and subsequent sites helped |
| Correctness | Behavioral checks, discrepancies, and unsupported cases |
| Readability | Review of naming, function correspondence, control flow, and helper burden |
| Mechanical coverage | Translated declarations and unsupported constructs, not just lines |
| Performance | Translation duration, input size, and peak memory where available |

Use a matched manual/LLM-only baseline where feasible. A replay after rules have
been tuned to a fixture measures repeatability, not independent generalization.
Reserve an unseen related consumer to test whether rules transfer.

## Proposed architecture

```text
Sources + explicit target configuration
  -> lossless tokenization and delimiter matching
  -> declarations, macros, symbols, and dependency inventory
  -> project-wide representation plan
  -> ordered source-span rewrites
  -> draft Go + provenance + unresolved-work manifest
  -> bounded LLM repair + compilation + behavioral comparison
  -> reviewed reusable rules and updated project decisions
```

### Source and configuration model

Record source file hashes, language mode, include paths, macro definitions, target
integer widths, and relevant build variants. Hashes provide reproducibility
without requiring version-control commands. Preserve inactive conditional text
and record which branches were selected. Unknown preprocessor conditions become
explicit issues; do not merge incompatible branches or silently choose one.

### Lossless lexical layer

Retain exact token spans and intervening text, including comments and whitespace.
Handle escaped strings, character literals, C++ raw strings, line continuations,
and preprocessor lines. Match delimiters outside protected regions. Preserve
malformed or unsupported input with useful diagnostics rather than dropping it.

### Structural and symbol inventory

Recognize functions, parameters, structs, enums, typedefs, globals, macros, and
basic statement boundaries. Track scope, file-local linkage, header definitions,
and symbol collisions. Build enough cross-file information to avoid inconsistent
names and duplicate definitions. C++ names, overloads, and templates may initially
be inventoried without being translated.

Use a small expression parser when precedence or nesting matters. Casts,
ternaries, assignment expressions, comma expressions, and increments used as
values require more than textual substitution. Unsupported ambiguity must remain
visible. A complete C++ parser is not a prerequisite for the first experiment.

### Rewrite engine

Each rule declares its recognized shape, required semantic facts, source edits,
diagnostics, and provenance. Use regex for local recognition within safe regions.
Apply edits against a known source snapshot with deterministic ordering and
overlap detection. Cross-pass dependencies must be explicit.

Start with declarations, member access, function signatures, straightforward
control flow, and known project macros. Expand only as fixtures justify it.
Do not transform an expression unless its evaluation count, ordering constraints,
and side effects are understood well enough for that transformation.

### Representation profiles

Store reviewed decisions once and apply them across all affected declarations
and uses. Profiles cover numeric widths, pointer versus slice versus handle
representations, allocation, ownership, containers, macros, and naming/package
mapping. A pointer declaration alone is insufficient evidence for a slice.

Record the rationale, affected symbols, validation obligations, and dependencies
of each decision. Track proposed versus accepted decisions. A changed decision
invalidates dependent generated artifacts and identifies repaired code to review.

Avoid a large generic C runtime emulation layer. Small, named compatibility
helpers are acceptable when they keep source intent clear and preserve behavior.

### Output and handoff contract

Proposed artifacts, with names finalized during implementation:

- `draft/`: generated output, allowed to contain explicitly unsupported regions.
- `repaired/`: completed Go maintained independently of regeneration.
- `source-map.json`: generated declaration/span to original file/span mapping.
- `issues.json`: structured unresolved work with stable identities where possible.
- `decisions.json`: representation choices and their status.
- `run.json`: source hashes, configuration, rule versions, and run measurements.
- `handoff/`: bounded task bundles with relevant original source and dependencies.

Every issue needs an ID, category, original source span and text, affected symbol,
generated location if available, missing fact or decision, related symbols, and
resolution status. Distinguish syntax repair, semantic decisions, and validation
failures. Do not assign numerical confidence scores without a defensible basis.

Draft mode need not compile. Optional scaffold mode may emit conspicuous panic
stubs for unsupported bodies while retaining original text in the handoff. Stubs
must remain tracked blockers and cannot count as a completed port. Never emit
plausible zero-return bodies to obtain a green build.

Source mappings must remain valid after formatting or explicitly describe which
pre-format artifact they refer to. Prefer declaration identities plus spans over
fragile line-number-only references.

### Library and CLI boundary

The library owns analysis, planning, transformation, and reporting. The CLI owns
argument parsing, filesystem orchestration, and exit status. Keep core operations
usable with in-memory fixtures. Candidate operations are `Analyze`, `Plan`,
`Transform`, and `BuildHandoff`; these are design sketches, not frozen APIs.

Choose the module path, minimum Go version, artifact schema, and initial CLI
surface in the first implementation milestone. Avoid building an extensive
framework before demonstrating one end-to-end port.

## ABC-specific translation risks

| Area | Required treatment |
| --- | --- |
| Iterator macros | Preserve filtering, bound reevaluation, nested bodies, and mutation behavior |
| Allocation macros | Distinguish object allocation, arrays, resize, clearing, and ownership |
| Tagged pointers | Choose a consistent edge/handle representation before rewriting callers |
| Bitfields and unions | Establish accessed semantics and layout requirements before replacing storage |
| Flexible or zero-length arrays | Recover allocation and indexing conventions, not just declarations |
| Integer semantics | Record widths, signedness, promotions, wrapping, and sentinel values |
| C truth values | Separate conditions from integer-valued comparison results |
| Switch and loops | Preserve fallthrough and `continue` behavior in lowered loops |
| Memory operations | Distinguish byte fill/copy from element initialization and assignment |
| Interior pointers | Review aliasing, reallocation, and lifetime assumptions |
| C++ object behavior | Inventory constructors, destructors, references, overloads, and exceptions |
| Templates | Translate a bounded supported subset or selected instantiations explicitly |

Initial inspected examples:

- `../src/sat/satoko/utils/vec/vec_uint.h`: small vector, allocation macros,
  iterator macros, pointer access, and memory operations.
- `../src/sat/satoko/utils/heap.h`: real vector consumer, ternaries, index maps,
  signed/unsigned interactions, and iteration macros.
- `../src/sat/xsat/xsatClause.h`: bitfields and a zero-length union array.
- `../src/aig/aig/aig.h`: complemented pointer operations and iterator macros.
- `../src/sat/glucose2/Alloc.h`: templated region allocator with reference offsets,
  ownership transfer, growth, and exceptions.

## Milestones and progress

Use `not started`, `in progress`, `blocked`, or `complete`. Completion requires
recorded evidence, not merely generated code. Milestone numbering describes the
intended sequence; implementation details can change after experiment results.

| ID | Milestone | Status | Evidence |
| --- | --- | --- | --- |
| M0 | Planning and initial source reconnaissance | complete | This document and inspected candidates above |
| M1 | Freeze first experiment and validation contract | complete | `experiments/satoko/{scope,profile,source-hashes}.json`; C harness |
| M2 | Minimal reusable translation pipeline | in progress | Library + CLI + lexer + conservative drafts; richer expression/configuration handling outstanding |
| M3 | Container plus consumer port, end to end | complete | Bounded 38-function port; 6,136 differential operations pass; see experiment report below |
| M4 | Representation-heavy experiment | not started | — |
| M5 | Bounded C++ end-to-end experiment | not started | — |
| M6 | Generalization and larger ABC pilot | not started | — |

### M1 — Freeze the first experiment

Preferred starting point: Satoko's unsigned vector and heap consumer. The heap
also depends on signed-index and activity-vector helpers; inventory those before
freezing scope. If that closure is too large, choose a smaller actual consumer
and record why. Do not substitute a toy caller and describe it as the real port.

- [x] Inventory selected declarations and transitive dependencies.
- [x] Choose full-file versus bounded-function scope and list exclusions.
- [x] Specify target widths, macro configuration, assertion behavior, and packages.
- [x] Define vector storage, growth, aliasing, and ownership decisions.
- [x] Build a small C reference harness and fixed input/operation corpus.
- [x] Define expected observations and invalid-input boundaries.
- [x] Record baseline repair effort or explain why a baseline is unavailable.
- [x] Choose module path, minimum Go version, and experiment artifact locations.

Acceptance: the source scope, input corpus, observable behavior, and measurement
method are written down and the reference harness runs reproducibly.

Source-reading findings to investigate before making rules: `vec_uint_init`
uses byte-oriented `memset`, not element-wise assignment; `vec_uint_alloc_exact`
sets `cap` to zero; some operations expose backing storage or copy capacity-sized
regions. These are observed source details, not adjudicated bugs. Do not silently
fix them during translation. Separate defined behavior, uninitialized data, and
possible source defects, and document any intentional divergence.

### M2 — Build the minimal reusable pipeline

- [x] Establish the Go library and thin CLI layout.
- [x] Implement lossless tokenization and balanced-region recognition.
- [ ] Inventory the declarations and macros needed by M1.
- [x] Implement ordered, non-overlapping source-span edits.
- [ ] Add the necessary expression handling and first transformation rules.
- [ ] Emit provenance, decisions, issues, and reproducible run metadata.
- [x] Add focused lexical, rewrite, and unsupported-input regression fixtures.
- [ ] Verify deterministic output and safe regeneration boundaries.

Acceptance: the frozen input produces readable drafts and actionable issues;
unsupported regions are preserved; source content cannot disappear silently.
Repeated runs with the same inputs and configuration produce equivalent artifacts
apart from explicitly volatile metadata such as timing.

### M3 — Complete the container and consumer experiment

- [x] Generate drafts through starting-point rather than hand-authoring the initial port.
- [x] Package original source, dependency declarations, and decisions for repair.
- [x] Record LLM/human repairs and elapsed effort separately from tool development.
- [x] Compile the repaired Go and resolve all active stubs in the selected scope.
- [x] Compare vector operations and consumer behavior against the C harness.
- [x] Exercise empty/full states, growth, ordering, repeated updates, and boundaries.
- [x] Check heap index consistency and ordering, with tie behavior specified.
- [x] Review source correspondence and helper complexity.
- [x] Promote recurring fixes into reviewed rules and test a fresh related case.
- [x] Write an experiment report with failures, exclusions, and actual measurements.

Acceptance: the selected real consumer runs against the ported container, passes
the agreed behavioral corpus and invariants, and has no unresolved semantic
blockers within the declared scope. Results include actual repair effort and
limitations, even if the speed hypothesis is not met.

### M4 — Complete a representation-heavy experiment

Choose one after M3: xSAT clause storage or complemented AIG edges. Prefer the
one with a small executable dependency closure and strong observable invariants.

- [ ] Inventory allocation, layout, access, mutation, and identity operations.
- [ ] Compare candidate Go representations and record the selected tradeoff.
- [ ] Encode the accepted representation in the project profile.
- [ ] Translate a real operation chain that both constructs and consumes values.
- [ ] Repair and validate against the reference implementation.
- [ ] Test relevant aliasing, identity, complement, indexing, or growth invariants.
- [ ] Record time saved or added by the shared representation policy.

Acceptance: the representation works across multiple dependent operations, not
only an isolated accessor. No unsafe pointer trick is mechanically retained
without an explicit, justified policy and validation.

### M5 — Demonstrate bounded C++ support

Candidate: a manageable region allocator or utility in the Glucose sources,
with a real consumer. Final selection follows dependency inventory.

- [ ] Specify the supported C++ feature subset for this experiment.
- [ ] Plan namespace, class, method, constructor, and overload mapping.
- [ ] Choose template handling: supported generics or explicit instantiations.
- [ ] Resolve reference aliasing, destruction, ownership transfer, and exceptions.
- [ ] Generate drafts, complete repair, and perform behavioral comparison.
- [ ] Report unsupported constructs and reusable C++ rules.

Acceptance: an actual C++ utility and consumer are ported through the tool with
documented semantics. Recognizing C++ syntax alone does not establish support.

### M6 — Test transfer and choose the next scope

- [ ] Select a new ABC subsystem not used to tune the initial rules.
- [ ] Measure remaining semantic decisions and total repair effort.
- [ ] Group handoffs by dependency closure rather than arbitrary file boundaries.
- [ ] Check package naming, file-local symbols, headers, and dependency cycles.
- [ ] Measure translation throughput and memory on a larger source inventory.
- [ ] Identify whether the next investment is rules, representations, or validation.

Acceptance: publish a measured recommendation on expansion, including what still
requires expert judgment. Do not infer whole-suite feasibility from compilation
or a high percentage of rewritten lines alone.

## Validation strategy

Use focused tests at each layer: lossless token round trips; protected strings
and comments; nested delimiters; precedence-sensitive rewrites; diagnostic and
provenance integrity; and end-to-end behavior. Test semantic traps rather than
only snapshots of the implementation's own output.

For differential tests, feed identical defined inputs to the reference C/C++ and
repaired Go. Compare stable outputs or explicit invariants. Avoid asserting that
uninitialized memory, undefined behavior, addresses, or nonunique solver output
must match. Record platform assumptions and intentional differences.

Compilation is required for completed ports but is insufficient evidence of
correctness. For larger solver work, validate returned models or proof-related
properties where available, rather than assuming textual output equality.

## Decision register

| ID | Decision | Status | Rationale / follow-up |
| --- | --- | --- | --- |
| D01 | Go library with thin CLI under starting-point/ | accepted | User-specified structure |
| D02 | No mandatory Clang dependency | accepted | Portability and source preservation |
| D03 | Regex plus lossless lexical/structural support | accepted direction | Avoid unsafe whole-file substitution |
| D04 | Preserve source naming and flow initially | accepted direction | Reduce review and repair cost |
| D05 | Container + real consumer before harder representation | accepted | User-confirmed experiment sequence |
| D06 | Satoko vector + heap as first candidate | accepted | 17 heap functions + 21 vector helpers; scope.json |
| D07 | Exact numeric, storage, and package policies | accepted for experiment 1 | profile.json: uint32/int32/uint64, explicit capacity, one package |
| D08 | Clause storage versus AIG edges for M4 | open | First reduce recurring repair work identified by experiment 1 |
| D09 | First supported C++ subset | open | Resolve before M5 |

## Progress log

| Date | Work completed | Evidence / next action |
| --- | --- | --- |
| 2026-11-11 | Agreed architecture direction and experiment sequence; inspected representative ABC constructs; wrote initial plan | Begin M1 scope and reference-harness work; no implementation or test results claimed |
| 2026-11-11 | Implemented prototype, preserved comments, generated 95 function drafts, repaired 38 selected functions, validated four seeded C/Go traces | First experiment passes; next investment is structural control-flow and macro/allocation rules, before M4 |

For each future work session, update milestone checkboxes and status, append a
dated log entry, and link concrete artifacts or test results. Record blockers
with the missing information and next action. Keep hypotheses distinct from
measurements; revise the plan when evidence changes the approach.

### Experiment report template

```text
Experiment / date:
Source files, hashes, selected symbols, and exclusions:
Tool/profile versions and target configuration:
Accepted representation decisions:
Reference harness and corpus:
Generated artifacts and repaired output:
Issues before repair / after repair:
Tool development / translation / LLM repair / human review / validation time:
Tests run and outcomes:
Discrepancies and intentional divergences:
Reusable rules learned and fresh-case results:
Readability assessment:
Conclusion and next milestone:
```

## Experiment 1 outcome — 2026-11-11

### Artifacts and reproducibility

- [Prototype library](translate.go), [lexer](lex.go), and [CLI](cmd/starting-point/main.go).
- [Frozen generated drafts and handoff](experiments/satoko/_draft-v1/run.json).
  The four `.original` and `.go.txt` files retain full source, including comments,
  directives, macros, and declarations outside the recognized subset.
- [Selected symbols](experiments/satoko/scope.json),
  [accepted representation profile](experiments/satoko/profile.json), and
  [source hashes](experiments/satoko/source-hashes.json).
- [C reference harness](experiments/satoko/reference/main.c) includes the original
  headers without modifying them.
- [Repaired heap](experiments/satoko/repaired/heap.go),
  [repaired vector helpers](experiments/satoko/repaired/vector.go), and
  [differential tests](experiments/satoko/repaired/differential_test.go).
- [Recorded repair procedure](experiments/satoko/repair.py) consumes generated
  function drafts; [repair dispositions](experiments/satoko/repairs.json) identify
  selected resolved issues. This Python script captures the LLM's fixture-specific
  repair work; the translator itself is Go and does not invoke the script.
- Satoko's original license is copied into the experiment directory.

Run from the repository root:

```sh
go -C starting-point test -v ./...
go -C starting-point run ./cmd/starting-point -out /tmp/starting-point-new-run \
  ../src/sat/satoko/utils/vec/vec_uint.h \
  ../src/sat/satoko/utils/vec/vec_int.h \
  ../src/sat/satoko/utils/vec/vec_sdbl.h \
  ../src/sat/satoko/utils/heap.h
```

Choose a nonexistent output directory. The first saved generation is intentionally
frozen; current code additionally fixes pointer declarations such as `type *p`
and preserves CRLF continued line comments. Current output therefore need not be
byte-identical to `_draft-v1`. Results are deterministic for a fixed version/input
apart from run duration. Originals and function byte spans make the handoff
reviewable without a version-control dependency.

### Scope and semantic decisions

The experiment covers all 17 functions in `heap.h` and 21 vector helpers required
by the heap and harness: unsigned data, signed indices, and encoded activity.
The full vector API, sorting, interior pointers, sdbl arithmetic, and solver
integration are excluded. The initial source exploration preceded the recorded
scope file; future experiments should freeze scope before tool tuning.

The Go vectors retain separate size and capacity fields and backing slices of
length capacity. This makes the growth policy and existing access code easy to
compare. `unsigned` is `uint32`, `int` is `int32`, and activity is `uint64`.
`compare` and `heap_in_heap` become booleans after reviewing their selected uses.
GC replaces explicit deallocation; use-after-free and OOM behavior are excluded.
This is recognizable Go with retained C naming, not a finished public Go API.

Two subtle source behaviors were explicitly addressed:

1. `heap_insert` expands an iterator that reads new, uninitialized index cells
   into an unused variable. The repair omits that read and initializes the cells
   to `-1`. The original reference still contains it: passing traces are evidence
   on this GCC platform, not a portability proof for indeterminate C reads.
2. The reverse loop in `heap_build` performs unsigned subtraction before conversion
   to signed `int`. The port uses `int32(size/2 - 1)` and records the GCC target's
   conversion behavior, including the empty case. Build requires a preallocated
   index map and a distinct, duplicate-free entry vector.

The draft is deliberately not buildable Go. It preserves unsupported text and
marks every recognized function for review. The library has no preprocessor,
full symbol resolver, expression parser, or verified generic macro lowering yet.
The issue report consolidates provenance and unresolved work rather than emitting
all of the separate planned artifact files. No inactive-branch evaluation is
claimed. These are why M2 remains in progress despite the bounded M3 success.

### Measurements and validation

| Observation | Result |
| --- | --- |
| Inputs inventoried | 4 headers, 95 recognized functions |
| Initial review issues | 99: one per function plus four file-level reviews |
| Selected and repaired | 38 functions; all selected function issues resolved in repaired scope |
| Remaining outside scope | 57 unselected function reviews; global/file reviews are not automatically resolved |
| Full body replacements in recorded repair | 17 of 38; remaining 21 use drafts plus local repairs |
| First generation duration | 4,559,646 ns (about 4.56 ms), including input/draft file I/O; excludes Go build/startup and final JSON serialization |
| Differential corpus | Seeds 1, 7, 42, 20261111; 1,534 operations each; 6,136 total |
| Trace comparison | Exact match after every operation, about 1.10 MB of output per seed |
| State compared | Heap contents/order, index maps, vector size/capacity/content, pop results |
| Independent Go invariants | Heap order after repair operations; heap-to-index consistency after every operation |
| Boundaries | Empty rebuild, singleton removal, clear/reuse, growth past 16/32/64, equal weights, high-bit activities, increase/decrease/update |
| Comment checks | Original comments retained in all four drafts; protected literals, signature comments, raw strings, continued comments tested |
| Fresh related case | `vec_char_size` and `vec_char_clear` drafts parse and type-check without repairs; not a behavioral port of vec_char |
| Toolchains | Go 1.27.3 linux/amd64; GCC 13.3.0; module declares Go 1.22 (not separately tested with that compiler) |
| Test command | `go -C starting-point test -v ./...` passed |

Timing was not instrumented separately for reasoning, repair, and review in this
first run. Observed timestamps bound initial inspection at 18:37:12 UTC, reference
setup/draft launch around 18:39:42, and the first passing repaired port before
18:42:30. Documentation and regression work continued afterward. These are wall
clock checkpoints, not measured isolated LLM repair time. Human review time and
peak memory are unmeasured. No matched LLM-only baseline was run; no speedup claim
is justified. Future runs must collect per-phase timing explicitly.

### Failures and lessons

- The first C compile failed because ABC's clock helper needs POSIX feature
  declarations. Adding `_POSIX_C_SOURCE=200809L` fixed the reference configuration.
- The first CLI invocation used an incorrect relative source path and left an
  empty output directory. The CLI currently leaves partial artifacts on error;
  callers must use a fresh directory. Atomic publication is a follow-up.
- Go cache writes were denied by the sandbox for some invocations. Approved
  cache access allowed tests to finish; this was environmental, not a port defect.
- A declaration spelling `heap_t *p` remained C in the first draft. A reusable
  regex fix and regression test now cover it. More complex declarators still need
  explicit structural recognition.
- Allocation, typed resizing, C predicates, ternaries, and iterator macros account
  for substantial repair work. Seventeen replaced bodies is too many to claim
  that this prototype already makes porting mostly mechanical.
- Lossless tokens successfully protect comments and strings while enabling local
  edits. Tests cover real ABC comments, not just synthetic examples. Comment
  preservation is now an explicit continuing requirement.
- Passing compilation alone would have missed the importance of encoded activity
  ordering, exact growth, index sentinels, and target-specific integer conversion.
  Full state traces provided much stronger evidence for this small subsystem.

### Conclusion and next work

The experiment demonstrates a functioning source-to-draft-to-repaired-port
workflow for a real ABC consumer. It supports the source-preserving design and
shows that a bounded port is tractable. It does not establish whole-suite
feasibility, automatic semantic correctness, or an advantage over direct LLM
porting. C++ support remains future work.

Before expanding to M4, prioritize the repeated work exposed here:

1. Implement structural statement recognition and brace insertion, including
   nested `if/else`, loops, and comments attached to those statements. Do not
   promote the repair script's line-based regex into a generic control-flow rule.
2. Move reviewed vector allocation/resizing and iterator semantics into explicit
   Go profile rules; track which semantic preconditions authorize each rewrite.
3. Add limited expression parsing for casts, predicates, and ternaries while
   preserving lazy evaluation and side effects.
4. Improve issue granularity: separate mechanical syntax repair from missing
   semantic decisions, with generated-span provenance and dependency links.
5. Regenerate into a new directory, measure reduction in required repairs, and
   validate the same corpus. Then port a new real consumer with rules frozen to
   measure transfer, recording phase timing before starting.
6. Proceed to the representation-heavy experiment once that replay provides
   evidence that recurring work has actually moved out of the LLM repair pass.

These priorities refine M2 rather than abandoning the M4/M5 sequence. Keep the
frozen first run as a baseline so future improvements can be measured honestly.


## Follow-up: package declarations, macro functions, and ternaries

User direction: all Go output files need a package declaration; function-like
`#define`s should become ordinary functions; ternary logic should become if/else.

Implemented in prototype-2:

- `Analyze` defaults to `package translated`; `AnalyzeWithOptions` and the CLI's
  `-package` flag select the name. Invalid names are rejected. Per-function JSON
  entries remain fragments; complete draft files carry the package declaration.
- Expression macros become function drafts. Reviewed types can be supplied with
  `Options.Macros`; absent types become named placeholders, not guessed integers.
  Original definitions remain as comments, and each macro has a review issue.
  Ordinary Go functions allow the compiler to decide whether to inline.
- Function arguments evaluate eagerly and once, unlike macro substitution.
  Call-site review remains mandatory for side effects, argument mutation,
  discarded arguments, and conditional use. Statement/iterator, token-pasting,
  stringizing, variadic, and empty macros retain commented originals plus explicit
  issues rather than being misrepresented as equivalent functions.
- Ternaries occupying an entire return expression, local initializer, or simple
  assignment are lowered to nested if/else. Branches remain lazy. Nested branches
  are handled; ternaries embedded in calls, arithmetic, or other conditions remain
  explicit `ternary-review` issues pending typed expression lowering. Converting
  C integer/pointer conditions into Go booleans still requires semantic review.
- [_draft-v2](experiments/satoko/_draft-v2/run.json) records a fresh four-header run
  with `package satoko`. Frozen v1 artifacts are historical evidence, not current
  generator output, and were not rewritten.

Validation: `go -C starting-point test -v ./...` passes, including typed macro draft
parse/type checks, package-name checks, comment preservation, execution of nested
ternary branches with call counters, and the existing C/Go differential suite.
The four new draft files were checked for their leading package declarations.

Conclusion: these rules remove additional mechanical work while preserving
source correspondence. The original repaired-port measurements are unchanged;
no claim is made yet that v2 reduces the 17 recorded body replacements. The next
replay should measure that directly after macro profiles and expression typing
are available. M2 remains in progress.


### Package default refinement

The default package name now comes from each source file's containing directory,
per user direction. Bare filenames resolve against the current working directory.
`-package` / `Options.Package` still override it for a shared destination package.
Invalid directory-derived Go identifiers require an explicit override. Tests
cover relative and absolute source paths, explicit overrides, and invalid names.
Historical runs retain their recorded package names.


### Macro draft correction: copy replacement text directly

User clarification supersedes the earlier expression-only macro policy. The tool
now emits a function for ordinary named-parameter function macros, including
iterator and statement macros. Parameters default to `int` placeholders. The
replacement text is copied directly inside the function braces, after removing
preprocessor line continuations. No return statement, loop body, iterator API,
or call-site adaptation is invented. Macro bodies intentionally retain C syntax
and ternaries for second-pass repair. Explicit signature overrides remain usable.
Variadic/unsupported parameter syntax still produces a signature-review issue.

[Draft v3](experiments/satoko/_draft-v3/00_vec_uint.h.go.txt) now includes
`func vec_uint_foreach(vec int, entry int, i int)` with the original replacement
text as its body. Other vector iterator macros receive functions as well. Package
names use their source directories (`vec` and `utils`) in this run.

Validation: all tests pass, including exact replacement-text checks for iterator
headers, statement macros, expressions, ternaries, stringizing, token pasting,
and empty replacements. The existing C/Go differential suite also passes. These
macro drafts are deliberately incomplete Go; compiler diagnostics and original
source guide the second pass. Earlier v1/v2 artifacts remain historical records.


### Typedef/struct pairs — prototype-4

Implemented the requested adjacent `typedef struct tag alias;` plus
`struct tag { ... };` transformation as `type alias struct { ... }`. Ordinary
scalar and pointer fields change to Go name/type order; field order, indentation,
line breaks, and comments are retained. Self-referential `struct tag *next` fields
use the alias. Numeric types follow the existing mappings (unsigned → uint32,
int → int32); pointers remain pointers for later representation review.

The pass uses masked tokens, matched delimiters, and edits at original source
spans. It does not cross preprocessor directives or transform nested pairs inside
another declaration. Complex fields such as arrays, bitfields, function pointers,
and comma-separated declarators remain visible with field-review issues. Header
and typedef comments are retained instead of being discarded with the C syntax.

[Draft v4 heap](experiments/satoko/_draft-v4/03_heap.h.go.txt) now contains the
requested heap_t Go struct; all three vector declarations are converted as well.
Previous drafts remain historical artifacts.

Validation: `go -C starting-point test ./...` passes, including parse/type checks of
converted structs, exact heap field lines, comments, self-pointers, complex-field
diagnostics, preprocessor boundaries, and the existing differential suite.


### Macro signature placeholder annotation

Macro-generated functions with placeholder argument types now include
`//TODO: argument types were guessed, go compiler should tell us the correction.`
on the signature line, after the opening brace. Explicitly supplied argument
types and zero-argument macros do not receive the warning. Replacement text is
unchanged. Signature tests cover the inline annotation and both exceptions.


### If-statement braces — prototype-5

Unbraced if bodies now receive braces in ordinary function drafts, including
single-line and multiline bodies. Statement boundaries are recognized using
balanced tokens rather than line substitution. Nested if/else binding, else-if
chains, loop statements used as if bodies, comments, and protected literals are
preserved. Else bodies also receive braces; closing braces are placed on the same
Go line as else without discarding intervening comments. Existing braced bodies
are retained. Bracing runs before ternary lowering so an unbraced conditional
return can also be lowered safely within its new block.

Macro replacement text remains verbatim under the separately requested macro
policy. Unsupported statement syntax remains second-pass work.

[Draft v5](experiments/satoko/_draft-v5/00_vec_uint.h.go.txt) demonstrates braces
around the vector capacity guard and includes the macro argument TODO annotations.
Validation: `go -C starting-point test ./...` passes, covering exact requested output,
nested/dangling else, comments, multiline else, loop bodies, idempotence,
parse/type checking, and existing C/Go differential tests.


### Header include guards — prototype-6

Recognized whole-file `#ifndef NAME` / `#define NAME` / matching final `#endif`
wrappers are omitted from Go drafts, including comments on those directive lines.
Leading license comments, other comments, nested conditionals, and original source
copies remain intact. Matching requires the same guard name, an empty define,
no outer else/elif, and no substantive content outside the wrapper. Unmatched or
ambiguous patterns remain unchanged.

Generated `experiments/satoko/_draft-v6/`; all four Satoko header guards are absent
from their Go drafts. Tests cover nested directives and negative matches.
`go -C starting-point test ./...` passes, including the C/Go differential suite.


### Ternary indentation — prototype-7

Generated if/else blocks now reuse the original statement's leading indentation,
indent branch statements one level deeper (four spaces, or a tab for tab-indented
source), and align closing braces and else with their if. Nested ternaries carry
the indentation recursively. Initializer declarations and their following if
align, and the replacement no longer adds an unnecessary trailing blank line.
Existing expression/comment contents are not globally reformatted.

Generated `experiments/satoko/_draft-v7/`. Exact-output tests cover the requested
allocation example, nested tab-indented branches, and variable initializers.
`go -C starting-point test ./...` passes, including lazy-branch execution and differential
checks. No gofmt invocation on incomplete draft output is required.


### Empty for-loop bodies — prototype-8

Per user direction, a trailing bodyless for-header in a macro replacement now
receives `{}`, and a for-loop with a null-statement body (`;`) receives `{}` in
both macro and ordinary function drafts. This is an explicit exception to the
otherwise verbatim macro replacement policy. Header expressions are retained;
nonempty bodies, comments, and literals remain unchanged.

Generated `experiments/satoko/_draft-v8/`; the vector foreach macro functions now
contain explicit empty loop bodies. Tests cover bodyless headers, null bodies,
comments, existing bodies, protected literals, and repeated application.
`go -C starting-point test ./...` passes, including the C/Go differential suite.


### Includes and standalone markers — prototype-9

Go drafts now prefix include directives with `//`, preserving their text and
spacing. Continued include directives have every physical line commented.
Unknown standalone top-level identifier lines such as ABC_NAMESPACE_HEADER_START
and ABC_NAMESPACE_HEADER_END are also commented. Known type/keyword lines,
function declarations, tokens inside function bodies, and existing comments or
literals are left alone. This is a conservative marker heuristic, not symbol
resolution. Source copies remain unchanged.

Generated `experiments/satoko/_draft-v9/`. Tests cover the exact requested include
block, continued includes, idempotence, and preservation of multiline code.
`go -C starting-point test ./...` passes, including the differential suite.


### Consistent package per output directory — prototype-10

Corrected the CLI's per-file package inference. The package is now resolved once
before output creation: explicit `-package`, otherwise the directory name of the
first source input. That single value is supplied to every translation in the
batch and recorded in `run.json`. The destination directory name is irrelevant.
`ResolvePackage` exposes the same rule for library batch callers; standalone
single-file analysis retains its source-directory default.

Generated `experiments/satoko/_draft-v10/`; all four Go drafts use `package vec`
because vec_uint.h is the first source. Older mixed-package draft directories
remain historical experiment records. CLI tests cover inputs from different
source directories, a differently named destination, an explicit override,
recorded metadata, and invalid-name rejection before output creation.
`go -C starting-point test ./...` passes.


### Nonempty for-loop bodies — prototype-11

Extended structural statement bracing to nonempty, unbraced for bodies. A nested
if/else is treated as the complete loop body, and following statements remain
outside the loop. Multiline loops get an aligned closing brace. Comments,
already-braced bodies, and macro replacement policy are preserved.

Generated `experiments/satoko/_draft-v11/`; vec_uint_find now has braces around
its loop, with the final return 0 outside it. The batch retains a shared vec
package. Tests cover the reported example, nested else/comments, and idempotence.
`go -C starting-point test ./...` passes. C for-header syntax remains draft syntax for a
separate transformation; this change fixes body structure.


### While loops — prototype-12

Ordinary while statements now become condition-only for statements. Their
conditions and comments are preserved; unbraced bodies receive braces using
structural statement boundaries. The trailing condition of a do/while statement
is not independently renamed, since that would change its semantics; full
do/while lowering remains second-pass work. Macro bodies retain their separately
specified copying policy.

Generated `experiments/satoko/_draft-v12/`; both heap percolation loops now use for.
Tests cover braced/unbraced bodies, nested if, comments/literals, do/while
boundaries, idempotence, and parse/type checking of generated Go.
`go -C starting-point test ./...` passes.


### Go output extension — prototype-13

Generated draft filenames now end in `.go` instead of `.go.txt`; source snapshots
and run metadata are unchanged. CLI tests discover the new extension and still
verify a shared output package. Generated `experiments/satoko/_draft-v13/` contains
four `.go` drafts. This saved experiment has a separate go.mod so its incomplete
syntax does not break the translator's module tests; the CLI does not generate
module files automatically. Earlier artifacts retain their historical filenames.

Validation: `go -C starting-point test ./...` passes. This filename change enables direct
Go tooling; it does not assert the full ABC drafts compile or format yet.


### Underscore-prefixed draft directories

Renamed all 13 versioned experiment directories to `_draft-vXX` and updated
links, repair replay paths, and repaired-source provenance comments. Removed the
now-unnecessary experiment go.mod: Go's recursive package discovery skips these
underscore-prefixed directories. Explicit `go test .` from a draft directory
still requests compilation. Future experiment output paths should use the same
underscore prefix; the CLI continues to honor the caller's `-out` value.


### Braces in retained-signature functions — prototype-14

The missing braces in vec_uint_asc_compare/vec_uint_desc_compare were caused by
fallback behavior: unsupported `const void *` parameters made the generator
restore the entire original function, discarding successful body transformations.
The fallback now retains only the original signature and keeps the transformed
body. Signature issues remain explicit. Macro bodies also receive the structural
bracing pass so recognizable if statements there are not exempt from braces.

Generated `experiments/satoko/_draft-v14/`. Added regression tests for unsupported
signatures and macro if/else plus a token-based audit of all four assembled ABC
drafts: all 54 if statements have braced bodies. This audit includes functions
with retained C signatures and excludes comments/literals. It is corpus evidence,
not a claim of complete C++ grammar coverage. `go -C starting-point test ./...` passes.


### C/C++ const qualifiers — prototype-15

Removed const qualifier tokens from assembled C-derived Go drafts and per-function
draft metadata, including retained signatures, casts, fields, and macro-generated
bodies. Comments, literals, longer identifiers, preprocessor text, and original
source snapshots remain unchanged. This is a C-qualifier cleanup pass, not a rule
for removing legitimate Go const declarations from arbitrary Go source.

Generated `experiments/satoko/_draft-v15/`. Tests cover qualifier positions,
comments/strings, retained comparator signatures, macro bodies, and metadata.
`go -C starting-point test ./...` passes.


### For-header parentheses — prototype-16

Removed the matched outer parentheses from C-style for headers after body
bracing, in both ordinary function drafts and macro-generated functions. Nested
expression parentheses, comments, and literals are retained. Condition-only for
loops produced from while are cleaned up too. The pass only removes a pair whose
closing delimiter immediately precedes the body, avoiding partial-expression
rewrites in already-unwrapped headers.

Generated `experiments/satoko/_draft-v16/`. Tests cover three-clause and
condition-only loops, nested grouping, comments, and protected strings. A
vec_uint_find-shaped generated function now passes go/format directly.
`go -C starting-point test ./...` passes. Other unresolved C syntax can still prevent
formatting an entire ABC draft file.


### Qualified function signatures — prototype-17

Fixed parameter parsing to remove const qualifiers before recognizing the
signature, rather than only after fallback to the C declaration. Functions such
as vec_uint_duplicate and vec_uint_copy now use func, name/type parameter order,
no static/inline/void return keyword, and an opening brace on the signature line.
Void-return signatures also no longer contain a redundant extra space.

C void-pointer parameters now use any as a draft representation, with an explicit
void-pointer-review issue to recover concrete types and repair casts/dereferences
from call sites. This avoids generating an invalid bare pointer type while
allowing comparator signatures to become Go functions too.

Generated `experiments/satoko/_draft-v17/`. A corpus regression confirms all 95
inventoried functions across the four headers now have func signatures with
same-line opening braces. Targeted signatures pass Go parse/type checks.
`go -C starting-point test ./...` passes, including prior bracing and differential checks.


## Project rename and recursive translation — prototype-18

Renamed the project directory, CLI, module, and Go library:

- Project: `starting-point/`.
- Command: `starting-point/cmd/starting-point/main.go`.
- Module: `github.com/glycerine/abc/starting-point`.
- Go package: `startingpoint` (the hyphen cannot appear in a Go identifier).

`starting-point abc` now creates `starting-point-cpp2go-abc` in the current
working directory and mirrors source-containing directories for every `.c`,
`.cpp`, and `.h` file. Output retains the input extension before `.go`, avoiding
same-stem collisions. Tree packages derive from their destination directory;
invalid identifier characters are normalized to underscores, leading digits
receive a prefix, and keywords receive a suffix. `-package` overrides the entire
tree. Flat file-list mode retains its existing shared-package behavior.

`TranslateTree` is the library entry point. It inventories before creating output,
refuses existing destinations, skips symlinks/version-control metadata/prior
output trees, writes a run report plus per-file reports and source snapshots,
and emits a commented-source fallback for analysis failures. Filesystem errors
still stop a run. The output folder is caller-owned; no module is generated.

Validation:

- `go -C starting-point test ./...` passes, including recursive mirroring,
  extension collisions, per-directory packages, package overrides, fallback
  preservation, existing-output refusal, and destination-inside-source behavior.
- Installed `/home/jaten/go/bin/starting-point` and verified that the exact
  `starting-point abc` invocation creates the expected output tree in a temporary
  fixture, with all three requested source extensions.
- Full ABC run: `/tmp/starting-point .` produced 1,630 `.go` outputs in
  `../starting-point-cpp2go-abc/` in 15.85 seconds. The inventory comprised 1,292
  `.c`, 319 `.h`, and 19 `.cpp` files, including the experiment reference harness.
- Verified every generated file's package declaration against its run record.
  All files in each output directory use the same deterministic package name.
- Seven files required commented-source fallbacks; other outputs remain drafts
  with unresolved constructs. This is not a whole-ABC semantic port or compile
  success claim.

Fallback source files and diagnostics:

| Source | Diagnostic |
| --- | --- |
| src/bdd/cudd/cuddExport.c | mismatched delimiter at byte 10985 |
| src/bdd/cudd/cuddInt.h | unmatched closing parenthesis at byte 4965 |
| src/bdd/cudd/cuddZddMisc.c | mismatched delimiter at byte 7207 |
| src/bdd/mtr/mtrBasic.c | mismatched delimiter at byte 12622 |
| src/misc/zlib/inflate.c | mismatched delimiter at byte 43923 |
| src/misc/zlib/trees.c | unclosed delimiter at byte 33407 |
| src/proof/int2/int2Bmc.c | mismatched delimiter at byte 9267 |

These are analyzer diagnostics, not claims that the original C is malformed;
conditional compilation is not evaluated by the current delimiter matcher.

Four original files contain non-UTF-8 bytes, retained in the generated drafts:

- `src/aig/gia/giaEmbed.c`
- `src/opt/dau/dauCanon.c`
- `src/proof/int/int.h`
- `src/proof/int2/int2.h`

Conclusion: tree traversal, project naming, and output/package organization are
validated on the complete ABC inventory. Next translation improvements should
address conditional-compilation-aware structural analysis and an explicit input
encoding policy. Existing per-function semantic limitations remain; this run
measures batch coverage and generation, not correctness of the full port.


### Invalid UTF-8 bytes in comments — prototype-19

Added token-aware cleanup that replaces each invalid UTF-8 byte in an output
comment with `_`. Valid Unicode (including an actual U+FFFD character), code,
string/character literals, and original snapshots are unchanged. One-byte
substitutions preserve byte offsets. Cleanup applies to full drafts, per-function
draft metadata, and commented-source fallback output.

All four previously identified source files had their invalid bytes in comments.
Refreshed their existing mirrored `.go` files and metadata after verifying that
regeneration differed only by this cleanup, and updated the installed command.
All 1,630 generated `.go` files now decode as UTF-8. Verified original snapshot
hashes for the four affected files remain unchanged. Tests cover valid Unicode,
invalid bytes inside and outside comments, macro/function metadata, and all four
real ABC files. `go -C starting-point test ./...` passes.


### Project-name invocation inside the checkout — prototype-20

Reproduced the reported `starting-point abc` usage error from inside the ABC
checkout: `abc` there is the bundled executable, not a source subdirectory. The
CLI previously fell through to file-list mode and hid the reason behind usage.

The resolver now prefers an actual directory argument, then accepts a bare name
matching the current checkout directory as shorthand for that checkout (including
when a same-named executable exists). Other missing paths and file/directory
mismatches receive specific diagnostics. Existing destinations remain protected.
Tests reproduce both parent-directory and inside-checkout invocation and the
same-named executable collision. Updated the installed command; all tests pass.

Verified the installed command with the exact `starting-point abc` invocation
from inside this checkout: it successfully generated all 1,630 outputs with the
same seven explicitly reported analysis fallbacks.


### Idiomatic output filenames — prototype-21

Unique source stems now produce `stem.go` rather than `stem.c.go`/`stem.h.go`.
Names are planned for the complete destination inventory before writing. Shared
stems retain source extensions; promotion is checked repeatedly for secondary
collisions (for example x.c, x.h, x.c.cpp). Destination-directory names are also
reserved. Flat batches use the same rule, with numeric disambiguation when
identical basenames come from different source directories.

Updated the installed command and renamed the existing full ABC output using
filesystem renames, preserving contents and updating run.json. Of 1,630 outputs,
1,456 now use short names and 174 retain collision-qualified names. Verified all
manifest paths exist and are distinct; src/aig/gia/giaSim.go is present. Detailed
metadata and original snapshots retain source-based names and remain unchanged.
Tests cover unique/shared stems, independent directories, secondary collisions,
directory collisions, mirroring, and flat batches. All tests pass.


### Integer object-macro constants — prototype-22

Simple integer `#define NAME literal` directives now become Go const declarations.
C suffixes are removed from the literal but their width/signedness is represented
in the declaration under the existing ABC LP64 assumptions: int=32, long=64,
long long=64. For example NUMBER1 3716960521u becomes
`const NUMBER1 uint32 = 3716960521`, rather than architecture-dependent Go uint.

Plain U selects uint32 when the value fits, otherwise uint64. UL/ULL select
uint64; L/LL select int64, with large nondecimal literals selecting uint64 when
required by C's candidate types. Unary minus on unsigned literals wraps at the
selected width. Unsuffixed integers remain untyped Go constants. Unsupported
replacement expressions and values outside the supported literal range remain
unchanged. Comments and surrounding spacing are preserved. Conversion occurs
after C qualifier cleanup so generated Go const keywords survive, including
per-function draft metadata.

Updated the installed command. `go -C starting-point test ./...` passes, including
Go parse/type checks, suffix-width boundaries, unsigned negation, comment
preservation, and retained unsupported macros. Generated examples in
`experiments/satoko/_draft-v22/{giaUtil,aigUtil}.go` confirm NUMBER1 and NUMBER2
are uint32 constants. This draft contains those two utility sources, rather than
the earlier four-header Satoko fixture. Full-tree outputs will adopt the rule on
the next generation run. LLP64/other target integer-width profiles remain future
work; suffix interpretation is explicitly LP64, not inferred from Go's uint.


### Visible fallback diagnostics — prototype-23

The CLI now lists every source-preserving fallback after the run summary, with
source-relative path, generated output path, and the analysis diagnostic. The
heading explicitly says these are commented original sources whose analysis
failed. TreeFile/run.json also record FallbackReason, so callers do not need to
open per-file reports to identify the cause. Zero-fallback runs omit the list.
Installed the updated command. Tests verify the printed paths/reason, persisted
manifest diagnostic, and zero-fallback behavior; all tests pass.


### Investigated causes of the seven fallbacks

Read-only inspection distinguishes three causes (no ABC sources changed):

- Five files fail because delimiter matching reads both alternatives of a
  preprocessor conditional as one token stream. cuddExport.c:281–286,
  cuddZddMisc.c:198–207, and mtrBasic.c:427–438 each open fprintf once but contain
  an alternative argument list/closing parenthesis in each SIZEOF_VOID_P branch.
  inflate.c:630–634 has alternative if-condition openings under GUNZIP with a
  shared continuation; the unmatched opening is reported much later at line 1196.
  trees.c:954–974 similarly has alternative block openings under FORCE_STORED
  and FORCE_STATIC. These are analyzer/preprocessor limitations, not evidence
  that those five C files are malformed.
- cuddInt.h:138–139 exposes a lexer bug: a multiline block comment starts on a
  #define line, but directive scanning stops at the physical newline. The next
  comment line is incorrectly scanned as code, including its closing parenthesis.
- int2Bmc.c:270 contains an actual extra closing parenthesis in the visible source:
  `int iObjF = Gia_ObjFaninId0(p->pGia, iObj) );`. Similar spellings occur later.
  This case requires source-level repair/review rather than solely preprocessor
  handling; no original file has been edited.

Next targeted fixes: track block-comment state across directive newlines, then
make structural matching aware of conditional branches instead of concatenating
mutually exclusive code. The final case should retain a precise repair diagnostic.


### Multiline directive-comment lexer fix — prototype-24

Directive scanning now tracks block comments across physical newlines, keeping
comment text inside the directive token until its closing delimiter. Quoted
literals and line comments protect comment-like text; continued lines and CRLF
are handled. Unterminated block comments receive a precise lexical diagnostic.
Token round-trip text and offsets remain unchanged.

`cuddInt.h` now analyzes successfully, including its multiline #define comments
and integer constant conversions. Added LF/CRLF, comment-marker-in-literal,
line-comment, unterminated-comment, and real cuddInt.h regressions. All tests pass;
the installed command is updated. Rechecked the original seven fallback sources:
cuddInt.h succeeds and the other six retain their prior diagnostics.

Earlier full-tree outputs were renamed externally to .0/.1/.2 and were left
untouched. A fresh single-file result is saved in
`experiments/satoko/_draft-v24/cuddInt.go`. This is a generated draft, not a claim
that the entire header is now valid Go. The next full-tree generation will no
longer need the cuddInt.h source-preserving fallback.


### Source line/column diagnostics — prototype-25

Lexer/delimiter errors now carry structured byte offsets. Initial analysis maps
those offsets against the unchanged original input, reporting 1-based source
line and byte-column numbers plus the zero-based byte offset. Tabs count as one
byte; CRLF and UTF-8 byte widths are handled consistently. Tree run.json records
FallbackLocation as structured fields alongside FallbackReason. CLI output and
per-file diagnostic messages include the coordinates. Errors arising after
rewriting are explicitly labelled transformed-draft coordinates rather than
being falsely mapped to original source.

Read-only recheck of current fallback locations:

| Source | Line | Column | Byte offset |
| --- | --- | --- | --- |
| src/bdd/cudd/cuddExport.c | 285 | 80 | 10985 |
| src/bdd/cudd/cuddZddMisc.c | 206 | 39 | 7207 |
| src/bdd/mtr/mtrBasic.c | 437 | 64 | 12622 |
| src/misc/zlib/inflate.c | 1196 | 9 | 43923 |
| src/misc/zlib/trees.c | 911 | 1 | 33407 |
| src/proof/int2/int2Bmc.c | 270 | 56 | 9267 |

Updated the installed command. All tests pass, including UTF-8/CRLF/tab location
checks, unterminated multiline comments, printed fallback messages, and structured
location persistence. Source locations must be looked up in original C/C++ files,
not in transformed .go files whose inserted/removed lines change coordinates.


## Opaque conditional-preprocessor regions — prototype-26

Changed direction per user instruction: do not select or evaluate preprocessor
branches. Preserve #if/#ifdef/#ifndef regions through their matching #endif,
including #elif/#else, byte-for-byte in the draft. Recognized outer include guards
retain their existing removal behavior. Nested conditional regions are tracked.

For structural analysis only, replace each opaque region with temporary markers
and its common external delimiter interface. Balanced internal syntax cancels;
shared closings/openings allow partial constructs such as fprintf calls and if
headers to match their surrounding code. Every branch must have the same
interface, including the implicit empty branch when no else exists. This is not
branch selection: all original alternatives are restored after every rewrite.
No expression/type/control-flow transformations run inside a preserved island.

When branch interfaces disagree or a conditional is incomplete, preserve a larger
span from the current top-level declaration through EOF instead of guessing.
This conservative recovery can leave more code untranslated; its exact original
span is reported. Functions whose boundaries intersect synthetic delimiters are
retained rather than guessed. Conditional-review issues record original spans;
function metadata, source hashes, and diagnostics map back to original input.

Validation and outcomes:

- Added tests for nested regions, elif/else, shared call closings, shared condition
  openings, alternative brace openings, conservative recovery, preserved original
  text, translation before/after islands, source mapping, and marker removal.
- All five former preprocessor failures analyze successfully. Saved focused
  examples in `experiments/satoko/_draft-v26/`. Their preserved regions occupy
  5.4–22.0% of their source bytes; no claim that exactly 99% is translated.
- `go -C starting-point test ./...` passes; updated the installed command.
- Full 1,630-file run in `/tmp/starting-point-conditional-check-v26` completed in
  22.64 seconds with only int2Bmc.c still falling back. Audited every output: all
  1,575 reported conditional spans are preserved verbatim and no temporary
  structural markers leaked into generated files.
- User had removed the original extra parenthesis at int2Bmc.c:270. Inspection
  found two more at lines 284 and 286. With explicit user authorization, removed
  those two extra parentheses. A fresh single-file translation succeeds, saved
  in `experiments/satoko/_draft-v26-int2/int2Bmc.go`.
- Repeated the full-tree translation after those corrections in
  `/tmp/starting-point-conditional-check-v26-fixed`: all 1,630 files translated
  with zero whole-file source-preserving fallbacks.

These changes eliminate conditional-structure whole-file failures without
resolving build configurations. Lexically malformed text and other unsupported
source can still require review. Draft generation is not a compilation or
semantic-equivalence claim.

## Ordinary function coverage and control braces — prototype-27

The int2Bmc.go review exposed a coverage gap: function recognition required
`static inline`, so only two of this file's eight functions received body
rewrites. This was not caused by conditional isolation: int2Bmc.c contains no
conditional-preprocessor regions.

Implemented:

- Recognize ordinary top-level C function definitions, with optional static,
  inline, and extern qualifiers. Retain the conservative parameter-declarator
  rules and opaque-region boundary checks.
- Compose local macro/struct edits into function-body rewrites, avoiding
  overlapping edits. A full-tree trial exposed this in bzlib/compress.c.
- Move existing opening braces onto if/for/while/switch/else header lines,
  preserving intervening comments; perform this before removing for parentheses.
- Descend into iterator-macro blocks when locating statements, so their nested
  if statements also acquire braces.
- Add regressions for ordinary functions, qualifier combinations, comments,
  nested macro definitions, iterator blocks, and all eight int2Bmc functions.
  The real-file regression also checks every if body has a same-line brace.

Validation: `go -C starting-point test ./...` passes. The intermediate full-tree
run in `starting-point-cpp2go-abc.7` translated 1,630 files with zero fallbacks;
`starting-point-cpp2go-abc.8` exposed an overlapping-edit error in acecPo.c
when a recursive statement parse failed after accumulating edits. Added local
rollback for failed parses inside blocks and a real-file regression.
`starting-point-cpp2go-abc.9` is the final validation run with that correction.
The installed command includes these fixes.

Remaining int2Bmc repair work: comma-separated declarations, C casts, iterator
macro semantics, assignment expressions in arguments, integer truth values,
and a ternary assignment used as an iterator-macro body. Recognizing all eight
functions is a coverage improvement, not evidence that the draft compiles.

Final prototype-27 validation: all 1630 files in `.9` translated with zero whole-file fallbacks. Audited all 1575 conditional-region spans: each remains verbatim, with no leaked projection markers.

## Actionable preserved-conditional reporting — prototype-28

The CLI now lists every preserved conditional opener with its original source
path, one-based line and byte column, zero-based byte offset, directive text,
and output path. Nested #if/#ifdef/#ifndef directives are listed individually;
removed header guards are excluded. Locations come from original directive
tokens, including when conservative recovery preserved a larger surrounding
region. No branch is selected automatically.

Both recursive and file-batch commands print these entries. Library Result and
tree-file metadata expose Conditionals in run.json and per-file JSON. A zero
fallback count no longer suppresses this review list.

Validation: all Go tests pass, including nested directives, indentation/column
accuracy, output with zero fallbacks, and JSON persistence. The installed command
was exercised on cuddExport.c: it reports all 15 conditional openers, including
`src/bdd/cudd/cuddExport.c:282:1` for `#if SIZEOF_VOID_P == 8`.

## Named local declarations — prototype-29

Local declaration conversion previously recognized only a short built-in list
and names ending in `_t`, leaving `sat_solver * pSat;` unchanged. It now also
learns named types from ordinary top-level function return/parameter types and
simple typedefs in the current file. Candidates must begin in a declaration
statement position; unknown names in multiplication expressions remain intact.
Comments within declarators are retained for review rather than discarded.

All Go tests pass, including named pointers, initializers, parameter-derived
types, multiplication preservation, comments, and the real int2Bmc regression.
Installed the command and generated
`experiments/satoko/_draft-v29-int2/int2Bmc.go`: both pSat declarations (lines 94
and 170) now read `var pSat *sat_solver;`. Existing full-tree snapshots were not
regenerated for this focused change. Types known only through external headers,
comma-separated declarations, and complex declarators still need further work.

## Comma-separated locals and zero values — prototype-30

Simple local declaration lists now become grouped Go var declarations when the
types agree and all initializers are absent or literal zero. The int2Bmc line
`int i, Lit, RetValue = 0;` becomes `var i, Lit, RetValue int32;`, retaining the
existing C int -> int32 mapping. Single literal-zero initializers are omitted
too; subsequent assignments to zero are preserved.

Mixed pointer/value types and nonzero initializers become separate declarations
in original order, preserving dependencies and calls (including argument commas).
Commented or complex declarators remain for later repair; for-header declaration
lowering remains separate work.

All tests pass, including a real int2Bmc assertion and coverage of grouped zeros,
nonzero initializers, mixed pointer types, comments, and ordinary assignments.
Installed prototype-30 and saved fresh output in
`experiments/satoko/_draft-v30-int2/int2Bmc.go` (the requested declaration is at
line 99). Full-tree snapshots were not regenerated for this focused change.

## Optional preprocessor listings and standalone structs — prototype-31

- Added `-pre`, default false, to both directory and file-batch CLI modes.
  `starting-point -pre abc` prints preserved conditional locations. JSON metadata
  always retains them; analysis-failure reporting remains independent.
- The aigMem.c struct bodies lacked adjacent typedefs: their aliases live in
  aig.h. Standalone named structs now become `type TAG struct { ... }`; unpaired
  `typedef struct TAG ALIAS;` declarations become `type ALIAS = TAG`, connecting
  header aliases to definitions without guessing from trailing underscores or
  requiring include expansion. Existing adjacent-pair conversion still works.
- Reused scalar/pointer field conversion and review issues. Original comments
  remain intact, multiple pointer levels survive, and standalone opening braces
  join the type header line. Struct definitions inside opaque #if regions stay
  untouched; aliases outside those regions may be translated independently.

All tests pass, including standalone/nonadjacent definitions, header aliases,
real aigMem.c fields and comment preservation, configuration boundaries, and
preprocessor output suppression. Installed the command, exercised both flag
settings, and saved focused aig.h/aigMem.c output under
`experiments/satoko/_draft-v31-aig/`. This was a focused validation, not a new
full-tree run. C char remains the current draft type `char`; representation
selection and complex field declarators remain later work.

## Formatted output calls — prototype-32

Direct printf/fprintf calls in executable blocks now become fmt.Printf/fmt.Fprintf.
Standalone sprintf calls with a simple named destination become assignments,
`buffer = fmt.Sprintf(format, args...)`, since Go returns the formatted string.
Expression uses of sprintf's C character-count result and complex destinations
remain for repair. C format-specifier differences, buffer representation, and
FILE-pointer-to-io.Writer conversion are still semantic review tasks.

The file receives one fmt import when conversion requires it. Existing single,
grouped, aliased, dot, and blank fmt imports are handled without adding duplicate
paths. Comments, literals, qualified methods, and opaque preprocessor regions
are protected. Function draft metadata receives the call rewrites too.

All tests pass, covering all three calls, count-result preservation, imports,
idempotence, comments, and conditional isolation. Installed prototype-32 and
saved `experiments/satoko/_draft-v32-int2/int2Bmc.go`: one fmt import and both
printf calls converted. Full-tree snapshots were not regenerated.

## If header parentheses — prototype-33

The header-unwrapping pass previously handled only for loops. It now removes
the outer C parentheses from if and else-if headers after bodies are braced,
including macro function drafts. Nested expression/call parentheses remain,
and whitespace just inside the removed if delimiters is trimmed. Opaque
preprocessor regions remain unchanged.

Updated brace audits to inspect Go-style conditions rather than requiring C
parentheses. All tests pass, including nested grouping, else-if, multiline
headers, comments/literals, and the real Aig_ManPackAddPatternTry condition.
Installed prototype-33 and generated
`experiments/satoko/_draft-v33-aig/aigPack.go`; the example now reads
`if Aig_ManPackAddPatternTry( p, k, vLits ) {`.

## Macro hash lexer regression and diagnostic coordinates — prototype-34

The full-tree run reported new failures in casCore.c, casDec.c, and fretFlow.c.
The formatted-call pass introduced in prototype-32 re-lexed assembled drafts;
it exposed an existing lexer bug that treated any # as a directive start.
Generated macro functions contained `printf(#f ...)` or `xe##_dist`, so the
lexer hid the remainder of those lines, including closing parentheses. The
prototype-33 if-parenthesis removal was not the cause.

Directives now require first-non-trivia position on a logical line. Mid-line
stringification/token-pasting hashes remain punctuation in drafts, preserving
balanced delimiters without pretending to implement their macro semantics.
Added regression coverage for all three real source files, hash operators,
comments before directives, and transformed diagnostic coordinates.

Errors from the formatted-call pass now include line and column as well as byte
offset. They are explicitly labeled transformed-draft coordinates (and function
name for a function-fragment failure), rather than incorrectly reporting those
positions as original C source locations. Original-source syntax diagnostics
retain their existing original line/column/offset reporting.

All Go tests pass and the updated command is installed. Full src-tree validation
completed in `/tmp/starting-point-src-v34`: all 1,625 files translated with zero
source-preserving fallbacks. Existing user output is untouched.

## Paired struct brace placement — prototype-35

Adjacent typedef/struct pairs now join the opening brace onto the Go type
header, as standalone structs already did. Both paths share the same helper;
intervening multiline/header comments are preserved, with the brace moved ahead
of them when necessary. Regression tests parse/type-check simple examples and
check both Xyz structs from src/template.c. All tests pass. Installed the command
and generated `experiments/satoko/_draft-v35-template/template.go`, containing
`type Xyz_ParTry_t struct {` and `type Xyz_ManTry_t struct {`.

## Package-level variables — prototype-36

starter.c exposed a scope gap: storage-class removal and local declaration/zero
initialization rules never ran on file-scope variables. A new package-variable
pass now routes simple top-level declarations through the shared lowering logic.
It removes static/const qualifiers on recognized definitions, emits Go var/type
ordering, omits literal-zero initializers, and retains nonzero initializers.

Verified starter output:
- `static int nThreadsRunning = 0;` -> `var nThreadsRunning int32;`
- `pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;` ->
  `var mutex pthread_mutex_t = PTHREAD_MUTEX_INITIALIZER;`

The scanner skips nested groups (including multiline function parameter lists),
struct bodies and declarations intersecting opaque preprocessing regions.
Comments within unsupported declarations remain available for repair. Existing
local behavior is unchanged, and initializer constants still need their own
representation decisions.

All Go tests pass, including package zeros, nonzero values, declaration lists,
prototype preservation, conditional isolation, and the real starter.c examples.
Installed the command and generated `experiments/satoko/_draft-v36-starter/`.
Full src-tree validation completed in `/tmp/starting-point-src-v36`: all 1,625
files translated with zero source-preserving fallbacks.

## String pointer defaults and strlen — prototype-37

C char-pointer types now default to Go string in recognized function parameters,
returns, package/local variables, and struct fields, including compact `char*`
spelling. Further pointer levels use the same mapping recursively (`char **`
becomes `*string` pending representation review). The local declaration matcher
also accepts an adjacent star without requiring whitespace after the base type.

Direct strlen calls become len calls in code and function metadata, without
adding fmt imports. Comments, literals, qualified methods, and opaque #if regions
remain intact. The shared library-call pass still handles formatted output.

All tests pass, including real starter.c coverage and a strlen-only file with no
fmt import. Installed the command and generated
`experiments/satoko/_draft-v37-starter/starter.go`, with
`func Abc_UtilStrsav(s string) string` and `malloc(len(s)+1)` in its draft body.
This does not yet repair strcpy/malloc, C truth-value conditions, nil string
returns, mutable byte buffers, or integer-width differences in len results.

## Array parameters and variables; mixed FILE declarations — prototype-38

The main signature was retained because `char *argv[]` was outside the recognized
parameter grammar. Parameter parsing now accepts pointer levels and array
suffixes. The outer array parameter dimension becomes a Go slice (C parameters
decay); inner dimensions retain their bounds. Ordinary local/package arrays keep
fixed bounds and dimension order. Plain char now maps to byte, so char buffers
become byte arrays while char pointers retain the string default.

Added FILE to known declaration types so existing mixed-declarator splitting
handles `FILE *pFile, *pOutput = stdout;` without losing the initializer. The
starter draft now contains:

```go
func main(argc int32, argv []string) int32 {
    var pFile *FILE;
    var pOutput *FILE = stdout;
    var ThreadIds [MAX_COMM_NUM]pthread_t;
    var pBufferCopy string;
    var Buffer [MAX_COMM_NUM]byte;
```

C main's arguments/result are intentionally retained in this mechanical draft;
a proper Go entry point, FILE/stdout representation, string/byte-array interop,
complex array initializers, and array struct fields still need further work.
All Go tests pass, including multidimensional arrays, sized/unsized array
parameters, package arrays, mixed initializers, and real starter.c coverage.
Installed prototype-38 and generated `experiments/satoko/_draft-v38-starter/`.
Full src-tree validation completed in `/tmp/starting-point-src-v38`: all 1,625
files translated with zero source-preserving fallbacks.

## Go main signature and command-line arguments — prototype-39

Conventional zero- and two-parameter C main definitions now emit `func main()`
with no result. Within that body, the actual argument-count parameter name maps
to `len(os.Args)` and the argument-vector name maps to `os.Args`, including
nonstandard spellings. Other functions and comments retain their names/text.
An os import is added when needed; ordinary/grouped existing imports are reused.

Main's `return 0;` becomes a bare return; nonzero or computed return expressions
become `os.Exit(int(expression))` to preserve the process exit status. Generated
nested macro functions are skipped because they have independent return scopes.
Package naming still follows the existing directory/override policy; users can
select `-package main` for a standalone executable draft.

All tests pass, including renamed arguments, helper-function isolation, import
deduplication, zero/nonzero returns, and parsing the resulting Go fixture.
Installed prototype-39 and generated `experiments/satoko/_draft-v39-starter/`.
The real starter draft has `func main()`, os.Args references, and exit status 1
on its usage path. Remaining C constructs in its body still require repair.

## FILE pointers — prototype-40

Recognized C FILE pointers now map to *os.File in parameters, return types,
local/package variables, and struct fields; additional pointer levels are
preserved. Generated os.File references trigger an os import even in files
without main, without duplicating main's import. Comments, strings, and opaque
conditional regions do not trigger imports.

All Go tests pass, including declarations across those contexts and import
protection/deduplication. Installed prototype-40 and generated
`experiments/satoko/_draft-v40-starter/starter.go`: pFile and pOutput are now
*os.File. stdout and C file-operation calls remain separate conversion work.

## Systematic standard-library facility — prototype-41

Milestones completed:

- [x] Survey C11 header families and document Go directions and incompatibilities.
  [STDLIB.md](STDLIB.md) contains primary-source links, the full mapping inventory,
  explicit representation policies, and follow-up priorities. Some standards
  sites were unavailable; accessible implementation documentation plus native-C
  differential tests were used rather than claiming unavailable pages were read.
- [x] Expose a reusable token-driven API, `RewriteStandardLibrary`, and an
  executable `LibraryRules()` catalog. The catalog covers 301 names: 117 with
  implemented rule paths, 23 conditional on call shapes/known representations,
  and 161 explicitly retained for review. These are candidate mappings, not
  unconditional semantic-equivalence claims.
- [x] Integrate the pass after structural/type rewriting and before restoring
  opaque preprocessor regions; update function metadata as well as file drafts.
- [x] Support nested arguments, evaluate arguments once, preserve comments and
  literals, protect source-defined/shadowed names and qualified calls, and manage
  deterministic imports (including aliases, conflicts and blank imports).
- [x] Add per-site `Result.Library` decisions with messages, source-line excerpts,
  and line/column/byte locations explicitly labeled as pre-library Go-draft
  coordinates. Add tree JSON per-file counts and a concise CLI aggregate summary.
  This is independent of `-pre`; thousands of review locations are not dumped to
  the terminal by default.
- [x] Compile and run generated Go fixtures, compare a complete translated C
  consumer with native C, run the repository tests, and validate the whole ABC
  src tree for analysis regressions.

Implemented families include literal-mode fopen/OpenFile, fclose, bounded fgets
for byte buffers, byte I/O, binary element-count I/O for known byte arrays/slices,
file positioning and standard streams; atoi/atol/atoll decimal-prefix conversion;
integer abs; ASCII/C-locale character classification/case conversion; common
float/double math; string comparison/value operations; bounded byte memory
operations; format normalization and count/status handling; process-status
constants, exit, and explicitly qualified getenv/assert drafts.

Important decisions and conclusions:

1. Name-only replacement is inadequate for the main examples. atoi must accept
   numeric prefixes, and fgets must preserve n-1, newline, NUL and partial-EOF
   behavior. A fresh buffered reader on each fgets call would lose prefetched
   bytes; Scanner would change record semantics. The initial bounded File.Read
   loop is explicit and correct for the tested single-stream cases, but should
   later be optimized through coordinated buffering/locking.
2. Small immediately invoked functions preserve C one-result status conventions
   while exposing Go value/error operations, without introducing a hidden C
   runtime or colliding locals around goto-heavy code. They increase draft size;
   the repair pass can promote them into ordinary statements or shared helpers.
3. Known byte buffers get NUL-aware strlen/string handling. Raw pointers,
   arbitrary slices and struct layouts are not silently treated as byte slices.
   Negative C int-to-byte conversions happen at runtime, avoiding Go's rejection
   of a negative constant cast directly to byte.
4. Locale, long-double precision, ferror/feof/errno state, raw allocation and
   ownership, pointer-returning string searches, atof/strto* end pointers,
   qsort/bsearch, random sequences, CPU clocks, shell execution and threading
   remain explicit repair work. C double/float types now map to float64/float32
   to make the supported math conversions usable.
5. Package/type context is reused for function metadata and import transactions
   avoid copying the full source identifier set per call. This reduced the
   initial full-tree prototype run from 227 seconds to about one minute.

Validation evidence:

- `go -C starting-point test ./...` passes, including generated-program execution.
- Native-C versus fully translated-Go consumer outputs agree for empty inputs,
  long lines, newline-only input, a final line without newline, and embedded NUL.
- Generated runtime fixtures cover numeric prefixes and widths, nested calls and
  argument side effects, math, ASCII EOF handling, format verbs and counts,
  byte memory, negative byte values, repeated bounded reads, append, EOF,
  seeking, missing-file behavior, rename/remove and import aliasing.
- Full src validation: 1,625 files, zero source-preserving fallbacks; 33,530
  converted library/standard-object sites and 7,908 retained review sites.
  The converted count includes semantic review obligations and is not a count
  of proven-equivalent calls. The largest retained families are assertion truth
  values, raw memset/memcpy, sprintf result/destination shapes, clocks, strtok,
  allocation/free, qsort, binary buffers and atof.
- All 1,563 reported opaque conditional spans in the release run were preserved
  verbatim, with no temporary projection markers leaked into generated files.
- Coverage details are saved in
  `experiments/stdlib/prototype-41-summary.json`; the focused starter output is
  `experiments/satoko/_draft-v41-starter-final/starter.go`.

Starter's focused result has 45 converted sites and seven retained review sites
(assert truth values, clock, free, malloc, strcpy, system). Its requested fopen,
atoi and fgets paths now have Go drafts, but pthread code, C truth values and
remaining representation mismatches still prevent treating the entire program
as a finished port. Whole-ABC compilation and an LLM speedup benchmark remain
outside this validation claim.

## Go time package — prototype-42

Added time(NULL)/time(0)/time(nil) -> time.Now().Unix(), and an expression-safe
output-pointer form that evaluates/writes once and returns the same timestamp.
C time_t now maps to int64. difftime computes the signed difference of Unix
seconds without signed-overflow or time.Duration-range problems.

Per the user's request, clock() is no longer retained as an unresolved call:
it becomes time.Now().UnixNano(), clock_t maps to int64, and CLOCKS_PER_SEC maps
to int64(time.Second). The per-call record explicitly says this changes C CPU
time to wall time and recommends time.Now/time.Since for later monotonic elapsed
measurement. Source overrides of CLOCKS_PER_SEC retain clock for unit review.
Calendar/struct tm operations remain separate coordinated conversion work.

The import planner does not mistake the C time() callee for a source variable
named time; genuine variable/import name collisions still receive an alias.
All tests pass, including actual generated-Go execution, pointer output, positive
and negative difftime, extreme signed endpoints, consistent clock units,
import collisions, conditional preservation, and real starter.c coverage.
Installed prototype-42 and generated `experiments/satoko/_draft-v42-starter/`:
`var clk int64 = time.Now().UnixNano()` and the final clock()-clk expression now
use the same nanosecond policy. Updated the 302-name catalog and its references.


## Prototype 43: no generated closures

User direction: never insert a new func() closure to adapt C library semantics;
leave compiler errors and structural repair to downstream users and LLMs.
The library pass now rejects any conversion that introduces a function literal,
including adapters nested in arguments, and rolls back its unused imports.
Printf/fprintf always emit direct fmt calls, including expression contexts and
ABC_PRT's comma expression. Fopen emits direct os.Open/OpenFile; atoi/atol/atoll
emit direct strconv calls. Multiple Go results, error handling, and C numeric
prefix differences remain downstream work. Other adapter-dependent calls remain
in the draft with explicit review records. Existing opaque source is preserved.

Replaced obsolete adapter-runtime tests with no-closure regression fixtures for
numeric, memory, file, nested-call and timing cases. Kept runtime checks for direct
formatting and math conversions. Historical prototype-41/42 runtime-equivalence
claims above do not apply to this deliberately simpler draft policy. Retained
review counts can rise as adapters are withdrawn.

Validation: `go test ./...` passes. Regenerated starter.c in
`/tmp/starting-point-v43-starter`: ABC_PRT contains direct fmt.Printf calls,
without anonymous functions. Installed the updated starting-point executable.
Existing user output directories were not overwritten.


## Prototype 44: forward declarations and named variables

Comment out file-scope direct function prototypes, including static/extern and
multiline declarations, preserving their original text and comments. Force a
newline after the commented prototype so a following same-line definition or
variable remains active. Function definitions and function-pointer variables are
not classified as forward declarations. Opaque preprocessor regions remain intact.
Strip extern alongside static/const before package-variable lowering:
`extern int nTotal, nGood, nEqual;` becomes
`var nTotal, nGood, nEqual int32;`.

Recognize otherwise unknown UpperCamelCase type spellings such as ProgressBar
and MyWidget at declaration boundaries, in addition to learned types and _t
names. This is a draft heuristic for types from headers, not full name resolution;
ambiguous multiplication using an UpperCamelCase operand can still need repair.
Lowercase operands and ALL_CAPS constants are excluded from this heuristic.
`ProgressBar * pProgress;` now becomes `var pProgress *ProgressBar;`.

Validation: all Go tests pass, with real abcCut.c regression coverage and focused
cases for multiline prototypes, same-line definitions, extern variables,
function-pointer exclusion, opaque regions, and named local declarations.
Generated `/tmp/starting-point-v44-cut/abcCut.go` confirms all reported fixes.
Installed prototype-44; existing user output directories are not overwritten.


## Prototype 45: floating-point casts

Interpret the user's float632 spelling as float32. Convert scalar C `(double)`
and `(float)` casts to Go float64(expr) and float32(expr), respectively. The token
pass identifies a unary operand including calls, subscripts, member access,
numeric literals and nested casts, preserving binary-expression precedence.
It runs on whole projected drafts and function metadata, so macros and global
initializers receive the same conversion. Comments/strings, pointer casts,
sizeof(type), and opaque conditional text are preserved. No closures are added.

Validation: all Go tests pass, including nested casts, literal exponents, operand
precedence, comments, macros, globals, and opaque-region protection. Installed
prototype-45; existing output directories remain unchanged.


## Prototype 46: shared package assertion helper

User-selected design: keep call sites clean as assert(x, "x"), quoting and
escaping the condition text. The initial inline if/panic design was withdrawn
before installation. Generate one additional assert.go per directory using its
resolved package; try assert2.go, assert3.go, ... if a translated filename is
already occupied. Both recursive tree and flat batch output support this.
Tree reports list HelperFiles; library users have Result.NeedsAssertHelper()
and WriteAssertHelper(directory, packageName) for coordinated output.

The helper has signature func assert(x bool, message string), and on false calls
panic("assertion failed, we do not see: " + message), exactly as requested.
Numeric/pointer truth-value mismatches remain for downstream repair. No closures
or per-call inline panic blocks are introduced. Comment preservation now counts
actual comment tokens, so a quoted copy of a comment in the diagnostic does not
mistakenly block translation. Original shadowed names and opaque text remain
protected. Nested library changes affect the condition but not its message text.

Validation: all Go tests pass. Tests cover exact condition quoting, comments,
nested calls, source-name protection, default/overridden packages, one helper
for multiple files, assert.go/assert2.go collisions, no helper in unused packages,
flat CLI batches, and executing translated assertions with the generated helper.
Regenerated abcCut.c in /tmp/starting-point-v46-helper-cut and verified the helper
and message-bearing calls. Installed prototype-46; existing outputs unchanged.


## Prototype 47: assertion caller location

Generate the user's requested helper template with fmt/runtime/path imports,
an early return on true, runtime.Caller(2), and a basename:line prefix when
caller information is available. Keep the chosen directory package and filename
collision policy. Caller(2) intentionally reports the caller of the function
containing assert, as requested; it is not the direct assert line (Caller(1)).
Updated the execution regression to verify the exact main.go:6 location prefix
through a translated check function, including the successful assertion path.

Validation: all Go tests pass, including the generated helper runtime check.
Installed prototype-47. Existing generated outputs are unchanged until regeneration.


## Prototype 48: immediate assertion location

Correct the helper to runtime.Caller(1), per user clarification. The reported
location now identifies the assert call itself, not its caller's caller.
Update the runtime regression to expect assert.go:3 in the translated check
function instead of main.go:6. This supersedes prototype-47's frame choice.

Validation: all Go tests pass, including the exact immediate call-site check.
Installed prototype-48; regenerate outputs to update existing helper files.


## Prototype 49: attached namespace terminators

Split top-level }ABC_NAMESPACE_IMPL_END (and HEADER_END) into a closing brace,
newline, and commented marker. Use token identity so comments and string literals
remain unchanged. If code follows the marker on the same line, add another newline
so it is not swallowed by the generated comment. Standalone marker handling
remains idempotent. Added real aigTest.c coverage and marker/collision regressions.

Validation: all Go tests pass, including the actual aigTest.c input.
Installed prototype-49; existing output files are unchanged until regeneration.


## Prototype 50: POSIX mutex declarations and methods

Translate pthread_mutex_t to sync.Mutex with automatic deduplicated imports and
remove PTHREAD_MUTEX_INITIALIZER on declarations. Translate pthread_mutex_lock
and pthread_mutex_unlock to Lock/Unlock methods, stripping address-of from a
receiver such as &mutex. Pointer receivers, grouped arguments and member access
work as well. Apply the pass consistently to full drafts and function metadata.
No closure/status adapters: integer result checks remain downstream repair.
Record init/destroy/trylock/timedlock as review sites rather than guessing
attribute, lifetime, timeout or return-value semantics.

Validation: all Go tests pass, including actual starter.c, import alias/name
collisions, source-defined function protection, nested assertions and an
executed generated program using four concurrent workers to increment a shared
counter under the translated mutex. Installed prototype-50; regenerate outputs.


## Prototype 51: inspected-output backlog, configuration and enums

Created root IDEAS.md from the existing out/ tree with concrete path/line examples,
relative priorities, rough pattern counts, implementation boundaries and semantic
pitfalls. The original user output remains unchanged.

Added ParseConfig/ReadConfig and Options.Defines. The CLI reads ./starting.cfg or
-cfg PATH. An absent default is optional; explicit missing files and malformed
entries are errors. Configured zero disables symbols for #if as well as
#ifdef/defined, matching the user's platform-choice intent. This intentionally
is not C's defined-but-zero distinction. Unlisted symbols and unsupported
expressions stay unknown; source macro expansion is not attempted.

Choose decidable conditional chains (including elif/elseif, else, nested selected
branches and choices inside header guards). Blank discarded bytes internally so
source coordinates remain stable. Preserve original source hashes and snapshots.
Unknown enclosing blocks stay opaque. Comment pthread.h includes even after
opaque regions are restored. Record active configuration in analysis/tree JSON.

Convert standalone enum definitions into Go const groups. Implicit sequences use
iota; explicit resets use their expressions and subsequent names use Previous+1.
Named enums use the existing int32 draft policy, retaining comments and tag aliases.
The later prototype-52 work extends local enum/tag-reference handling.

Validation before the larger backlog work: library/CLI tests passed and all 1625
ABC source files translated with zero source-preserving fallbacks using WIN32=0
in /tmp/starting-point-v51-src. That snapshot reported 37787 library conversions
and 3878 retained review sites. It is a draft translation, not an ABC build.

## Prototype 52: mechanical backlog and configuration-template usability

User explicitly requested implementation of IDEAS.md with unit tests preventing
feature cross-talk. IDEAS.md now records completed bounded rules and unresolved
representation work; the implementation does not invent ownership or union layout.

Implemented scalar casts/suffixes; constant conditions; switch headers, grouped
cases and fallthrough; simple prefix increments in Go if initializers; braced
do/while without continue; ordered comma-expression statements and independent
for-header assignments; consistent macro member access; numeric/pointer truth
conditions using conservative known types/return types; single-object ABC_ALLOC
and ABC_CALLOC; bitfield expansion with layout/truncation comments; array/comma
struct fields; typed aggregate literals and zero values; persistent local-static
storage with collision-free names; local enums, tag references and aliases;
reviewed-result expression-macro returns; direct mutex-status assertion removal;
and whole known struct memset-to-zero conversion. Add targeted review metadata
for retained unions, casts, layouts, control-flow and ownership/lifecycle decisions.

A shared typedef/enum-tag inventory supplies names from headers to consumers.
Function metadata now takes actual emitted function text, avoiding disagreements
when whole-file passes choose collision-safe names for lifted static variables.
Regression tests caught and fixed same-line block separators, enum separators,
retained do/while macro formatting, increment operand boundaries, and assertion
status-check spacing. Existing tests were updated only where behavior intentionally
improved; compiled/runtime fixtures verify the resulting semantics.

Config parsing now accepts // comments to end of line anywhere. If -cfg is absent
and ./starting.cfg does not exist, a successful CLI run writes ./_starting.cfg:
sorted/deduplicated SYMBOL = 0 placeholders plus observed conditional directives
and source locations. The file is inactive until renamed; existing templates are
never overwritten. Explicit -cfg suppresses generation. CLI tests run in an
isolated working directory so generated templates do not pollute the repository.

Testing instruction: never run tests under experiments/. All subsequent checks
use `go -C starting-point test . ./cmd/starting-point`, including generated-Go
runtime fixtures in temporary directories. No experiment drafts are compiled.

The first full prototype-52 inventory translated 1625 files with zero fallbacks
in 109.08 seconds (/tmp/starting-point-v52-src), including 102 assertion helpers.
It reported 38011 conversions and 9202 review sites; review coverage expanded to
ABC allocation/free and pthread lifecycle calls, so counts are not comparable to
the narrower old library catalog as a measure of regressions. A final full-tree
check follows the last cross-talk fixes.


Final validation and boundaries:
- The configured whole-tree check in /tmp/starting-point-v52-final translated all
  1625 files with zero fallbacks in 112.10 seconds, producing 102 assert helpers.
  It reported 38328 converted and 8885 retained library sites after the expanded
  allocation/POSIX catalog. This validates translation coverage, not an ABC build.
- Library/CLI tests pass with `go -C starting-point test . ./cmd/starting-point`.
  Experiments were excluded after the user's instruction and remain excluded.
- A final targeted regression tightened static lifting: literal/zero/address
  initializers can lift, but C++ lazy calls, automatic-variable values/addresses,
  and ambiguous shadowing remain local for review. Tests verify both persistence
  and non-hoisting of dynamic initialization. No new anonymous functions are used.
- Installed prototype-52. The user's out/ tree was inspected, not overwritten;
  regenerate to use the new transformations and configuration support.


## Prototype 53: omit include guards from configuration templates

The source-transformation guard detector already required a complete outer
#ifndef NAME / #define NAME / #endif wrapper. Configuration discovery now uses
the user's simpler near-top .h pattern: adjacent matching #ifndef NAME and empty
#define NAME directives, ignoring comments/whitespace and allowing leading
preamble directives. Stop at ordinary code or after 32 significant directives;
do not classify nested platform-condition pairs as file guards. No suffix/name
heuristic is used. Omit guard directives from template comments and guard names
from assignments, including names encountered in other inventoried files.

Tests cover the four reported Glucose headers, comments between the pair,
preamble directives, code outside the guard, genuine feature names ending in _h,
valued defines, mismatches, late/nested pairs. Library and CLI tests pass; no
experiments were tested. Installed prototype-53. Existing templates retain the
non-overwrite policy: move an existing _starting.cfg aside before regenerating.


## Prototype 54: irregular ABC header guards in configuration templates

Inspected the actual missed headers. espresso.h defines its ABC__...__espresso_h
marker at line 21 before testing it at line 58. place_base.h tests
!defined(PLACE_BASE_H_) at line 10, defines a differently named ABC__...__place_base_h
marker at line 11, and tests that marker again at line 24. Neither matches the
ordinary adjacent same-name ifndef/define pattern.

Extend template-only detection to near-top empty definitions whose names identify
the current header (basename-derived names or ABC__...__basename_h), including
later reversed tests. Recognize #if !defined(NAME) as well. A mismatched outer
pair is accepted only when both names identify this header. Valued definitions,
unrelated header names and ordinary feature flags remain configurable. This does
not change source branch selection or remove these unconventional source blocks.

Regression tests use both real headers and verify IBM_WATC, IBMPC and BPI remain
in the template. Existing Glucose tests and guard-boundary tests continue to pass.
Library/CLI tests pass; experiments excluded. Installed prototype-54. Existing
templates are still not overwritten automatically.


## Prototype 55: decimal and scientific literal defines

Extend the existing numeric #define pass to complete decimal/scientific literals:
1.15, .5, 1., 1e-3, 2.5E+12, and signed forms become untyped Go constants while
preserving their textual values and comments. C f/F suffixes become explicit
float32 constants. Long-double suffixes, malformed literals, expressions and
function macros remain for their existing handling; no float64 rounding step is
introduced into the translator. Existing integer suffix behavior is unchanged.

Tests type-check the generated constants (including exact rational values and
float32 typing), cover malformed/expression/string forms, preserve opaque regions,
check function metadata, and verify the actual REO_REORDER_LIMIT in reo.h.
Library/CLI tests pass; experiments excluded. Installed prototype-55. Regenerate
output files to apply the new literal conversion.


## Prototype 56: short declarations use int16

Add short -> int16 to the shared C-to-Go type mapping. This fixes reo_unit's
lev, TopRef, TopRefNew and n fields and consistently handles short parameters,
returns, pointers, globals, locals and arrays. Scalar casts already used int16.
Regression tests cover the actual reo.h output and the shared declaration paths.
Library/CLI tests pass; experiments excluded. Installed prototype-56.


## Prototype 57: macro results, condition-only loops and string defines

Per the user's updated macro policy, every generated macro function has a result
(default int, or the explicitly configured result type). Expression replacements
lose only parentheses enclosing their whole value and receive a return statement.
C cast/grouping parentheses inside the value remain. Original macro comments are
unchanged. Leading comments remain before return so embedded newlines cannot
insert a Go semicolon between return and its operand. Comma expressions evaluate
earlier operands as ordered statements and return their final value. Statement-
style/empty macros retain their statements and append a clearly marked placeholder
return 0; no invalid return-for/return-if or anonymous-function adapter is emitted.

For headers with empty init and post clauses, remove both semicolons and extra
whitespace: for ; condition; becomes for condition, and for ;; becomes for {}.
Keep comments and preserve either nonempty init or post clause. The rule applies
to ordinary functions and generated macros; actual reoSift.c is regression-tested.

Object-like quoted string defines become const declarations with literal spelling
and comments preserved. Adjacent C string literals become Go concatenation. Do not
misclassify char/wide literals, function macros or general expressions as ordinary
string constants. Unknown preprocessor branches remain opaque.

Tests cover the actual Unit_Regular macro, typed/default results, statement macros,
comment placement, comma evaluation, numeric/string constant cross-talk, literals
and escapes, loop runtime behavior, function metadata and the actual REO sources.
Only library and CLI tests are run; experiments remain excluded.

Validation: library and CLI tests pass (`go -C starting-point test . ./cmd/starting-point`).
The configured full ABC translation processed 1625 files with 0 source-preserving
fallbacks in 118.35 seconds, in /tmp/starting-point-v57-src. Verified the emitted
Unit_Regular return and reoSift condition-only loop at the reported locations.
Installed prototype-57; existing user output is unchanged until regeneration.

### Prototype 58: control statements after labels

Reproduced the missing braces at reoShuffle.c's finish label with a minimal
failing unit test before changing the implementation. The statement walker had
treated the label and following if as one ordinary statement through its first
semicolon. Recognize a label prefix and recursively process its statement,
preserving labels and comments while adding the required control-flow braces.

Validation: the minimal test now passes, as do regression checks for stacked
labels with nested if/for/else and translation of the actual reoShuffle.c cleanup.
Library and CLI tests pass; experiments are excluded. Existing out/ is left
unchanged and must be regenerated to receive the fix.

### Prototype 59: preserve other transformations after named labels

The bracing fix alone was insufficient: red comparison tests showed that named
labels also prevented local declarations (including arrays and initializers),
whole-statement ternaries and discarded comma expressions from being lowered.
Added shared token-based label-boundary recognition to these passes and the
statement walker. Labels and comments remain in place; ternary colons, bitfields
and C++ scope operators are not treated as label boundaries.

Regression coverage compares translation with and without stacked labels for
21 statement forms: declarations, pointers, arrays, aggregates, local statics,
ternary assignments/returns/initializers, comma expressions, control flow,
assertions, printf and casts. This checks that labels do not change the existing
transformation result, without claiming unsupported C constructs now compile.
Boundary tests also cover labels after control headers and non-label colons.
Validation uses library and CLI tests only; experiments and existing out/ are
untouched.
Library and CLI tests pass; installed prototype-59. Regenerate output to apply.

### Prototype 60: direct sprintf assignment for known byte buffers

Added and observed a minimal failing regression with char Suffix[2000] and
sprintf under an if. Known byte-buffer destinations selected an old closure
adapter, which the no-anonymous-functions guard rejected, leaving sprintf
unchanged. Emit destination = fmt.Sprintf(...) directly for standalone calls
with simple named destinations, even when the destination is a known byte array.
Keep any array/string type mismatch visible for downstream repair and note it
in conversion metadata. Count-result uses and complex destinations retain their
existing review behavior.

Validation: minimal regression and actual abcHie.c translation pass, including
fmt import deduplication and function metadata. All library and CLI tests pass;
experiments excluded. Existing out/ remains unchanged until regeneration.

### Prototype 61: string declarations for sprintf destinations

Extended the char-array regression and observed it fail before implementing
promotion of the declaration bound to an emitted fmt.Sprintf assignment.
Uninitialized one-dimensional byte arrays (including C {0} initializers already
lowered to zero-value declarations) now become string declarations, annotated
with `// was [N]byte in C`. Already-string variables remain unchanged. Nonzero
aggregate initializers remain for further representation work.

Resolve destination names by lexical scope so other functions and nested
shadowing declarations are not changed. Preserve comments and terminate the
annotation with a newline even when subsequent source code shares the line.
Retained sprintf calls (count-result use or unsupported formats) do not trigger
promotion. Repeated writes produce a single annotation; fmt aliases work.

Validation: minimal red-to-green test and actual abcHie.c output, function
metadata, unrelated/nested buffers, retained calls, aliases, and a compiled
runtime test pass. Library and CLI suites pass; experiments are excluded.
Existing out/ is unchanged until regenerated.

### Prototype 62: opt-in function macro expansion

Added CLI -inlinemacros (default false) and library Options.InlineMacros. The
first basic regression was observed red before expansion was implemented.
Macro definitions remain source comments; supported invocations expand before
ordinary translation so declarations, control-flow bracing, ternaries, loops,
assertions and standard-library rewrites continue to apply. Default macro-to-
function behavior is unchanged.

Token substitution protects strings/comments and preserves C replacement
parentheses and repeated argument evaluation. Support fixed-arity/empty calls,
nested arguments and macro references, line continuations, source-order
redefinitions/undef, local included headers, standard include guards and pragma
once. Tree mode searches the source root and its src directory in addition to
each including directory; library callers can set IncludePaths. Unknown #if
regions remain opaque, and branch-dependent macros are blocked rather than
selecting a definition. No anonymous functions or compiler dependency.

Preserve unsupported variadic/stringification/token-pasting calls and recursive,
wrong-arity or bounded-expansion cases with macro-inline-review issues. Keep
expansion diagnostics anchored to original invocations; map downstream function
spans, conditional coordinates and source errors back to original input bytes.
Source hashes and Function.Original remain original, and metadata records the
inline policy. This is intentionally a bounded draft expander, not a complete
preprocessor. Object-like macros retain existing constant processing.

Validation: library and CLI suites pass, excluding experiments. Coverage includes
default-off behavior; both flat and tree CLI modes; lexical substitution and
comments; nested/multiline/CRLF/empty/statement macros; side effects/precedence;
include cycles, header guards, include roots, redefinition and undef; configured
and unknown branches; recursive and unsupported forms; depth/size limits;
original coordinates/hash; metadata and real reoTransfer.c header expansion.
A compiled runtime test combines macro expansion, loops, do/while and printf.
A smoke translation of all 10 REO inputs into
/tmp/starting-point-inline-reo-test completed with zero source-preserving
fallbacks (99 converted standard-library sites, 78 retained library review sites).
No experiment tests were run and existing out/ was not modified.
Installed prototype-62; regenerate with -inlinemacros to use the new mode.

### Prototype 63: unconditional expansion and visible macro annotations

Supersedes prototype 62's opt-in policy: macro expansion is now unconditional in
Analyze, AnalyzeWithOptions, tree translation, and CLI file batches. Removed the
-inlinemacros flag and Options.InlineMacros switch, along with the obsolete
macro-to-function generator and MacroTypes signature overrides. Function-like
macros can contribute only part of a statement (notably a for header), so their
caller's body must remain attached after expansion. Unsupported expansions
remain explicit review sites; they never fall back to invented Go functions.

Add standalone `// inlining of macro NAME` comments above expanded code lines.
Block markers carry names through the transformation passes, then move above
the containing line to avoid splitting return expressions or argument lists.
Nested macro names are included; repeated names on one line are deduplicated.
Original definitions and Function.Original remain unchanged, and function draft
metadata includes the final annotations. Retained/unexpanded calls receive no
misleading inlining annotation.

Validation: observed the default-mode regression red before making the policy
change. Replaced obsolete macro-function tests with definition-comment and
call-site expansion checks. Added default CLI flat/tree checks, removed-flag
rejection, named/nested annotations, expression safety and a compiled loop-header
macro test whose caller-owned body contains break. Existing macro and library
coverage remains active. Library and CLI suites pass; experiments are excluded.
Default-mode translation of the 10 REO files to
/tmp/starting-point-default-inline-v63 produced zero source-preserving fallbacks
(99 converted library sites, 78 retained library review sites). Existing out/
was not changed. README documents the new unconditional policy.
Installed prototype-63. Regenerate normally; no macro flag is needed or accepted.

### Prototype 64: postfix indices in simple assignments

Observed a minimal regression fail for the aigDoms.c idiom before adding a
mechanical pass: `dst = src[i++];` becomes `dst = src[i]; i++;` (and similarly
for postfix decrement). Simple member/index paths are accepted; preserve
comments and place the update immediately after the assignment. Bracing happens
first, so the update stays inside any single-statement if/loop body. Macro
expansions receive the same conversion and function metadata matches the draft.

Do not hoist from calls, returns, compound/lazy expressions, control-header
clauses, or assignments that mention the incremented name elsewhere. These
require additional evaluation-order or representation decisions.

Validation: red-to-green minimal case, actual aigDoms.c i++/k++ assignments,
function metadata, label and comment handling, decrement, conservative boundary
cases and a compiled runtime test combining arrays, macro expansion and an if
body. Library and CLI tests run without experiments. Existing out/ is untouched.
Library and CLI suites pass. Installed prototype-64; regenerate output to apply.

### Prototype 65: local feature switches and C++ function signatures

The AbcGlucose2.cpp case was a local feature switch, not an include guard:
#define USE_SIMP_SOLVER 1 precedes #ifdef USE_SIMP_SOLVER. Guard stripping already
runs before constant conversion. The selector previously considered only config
values, leaving the entire function branch opaque. Added source-order tracking
of active local definitions/undefs before projection and constant conversion.
Local presence and numeric truth are distinct (#define X 0 is defined); explicit
configuration retains its documented override/zero-disabled semantics. Unknown
branch modifications and include boundaries invalidate affected local knowledge.
No source-coordinate shifts: removed branches are still blanked. Unsupported
expressions and unresolved branches remain opaque. Existing incomplete-guard
and shared-delimiter recovery tests remain intact.

Comment C++ using-namespace directives, including qualified, multiline, local and
opaque-region cases. Preserve original text and following same-line code; leave
aliases, using declarations, strings and existing comments alone. This pass uses
lexing without delimiter matching, since restored opaque branches may each have
unbalanced delimiters.

Resolving the branch exposed additional retained signatures. Added qualified
parameter types, reference-to-pointer drafts, balanced argument splitting and
ordinary function-pointer parameter conversion to Go func types. Qualified
spellings map :: to .; no anonymous adapters are introduced. C++ namespace
organization, reference call-site adaptation and unsupported body syntax remain
second-pass work.

Validation: observed minimal feature-switch and using-namespace regressions fail
before the fix, then a real-source test exposed 17 retained C++ signatures.
All 61 functions in AbcGlucose2.cpp now have Go func signatures, regression-
checked using the real source. Tests also cover local zero/empty/function defines,
undef/redefinition, config precedence, unknown branches/includes, coordinate
preservation, namespace protections/idempotence, qualified/reference/callback
signatures and compiled callback behavior. Library and CLI suites exclude
experiments. A fresh flat output is in /tmp/starting-point-glucose2-v65; existing
out/ was not changed. README documents branch selection and C++ draft mappings.
Library and CLI suites pass. Installed prototype-65; regenerate output to apply.

### Prototype 66: elapsed runtime as the final CLI line

Time the whole CLI run from main entry, including input setup, translation and
reporting. Print `elapsed: Nm Ssec` last, using whole elapsed seconds and total
minutes. Print after any error diagnostic too, retaining the existing nonzero
exit status. Tree and flat-file modes share the same timing path.

Validation: CLI tests cover subsecond/minute/hour formatting, and execute the
built binary for tree, flat and error cases to verify exactly one elapsed line
at the end of the combined output, including after -pre details. No experiment
tests were run.
CLI tests pass. Installed prototype-66.

### Prototype 67: typed single-object calloc expressions

Observed the requested explicit-allocation regression red, then added a shared
mechanical rule after macro expansion: ((T *) calloc((size_t)(1), sizeof(T)))
becomes &T{}. This also handles the actual ABC_CALLOC expansion in template.c,
without special-casing the macro before expansion or losing its inlining comment.
Recognize redundant enclosing parentheses, literal-one suffixes, whitespace and
comments; require exactly one pointer level and a matching sizeof type. Known
primitive types use new(mappedGoType), since scalar composite literals are not
valid Go. Preserve required grouping before member/index access and leave other
parts of larger expressions intact.

Array/dynamic/zero counts, mismatched types, multiple pointer levels, malloc,
untyped calloc, and malformed arities remain unchanged. Preserve comments and
function draft metadata; no anonymous functions or allocation adapters.

Validation: minimal red-to-green regression, explicit-form boundary tests,
comments/literals/grouping, macro annotation and metadata checks, a compiled
zero-initialization runtime test, and the real src/template.c allocation.
Library and CLI tests exclude experiments; existing out/ is untouched.
Library and CLI suites pass. Installed prototype-67; regenerate output to apply.

### Prototype 68: combine local declarations with first allocation

Observed a minimal declaration-plus-calloc regression fail, then added a pass
after typed allocation lowering. An uninitialized local `var p *T;` followed by
its first use as `p = &T{};` becomes `p := &T{};`. The declaration can be on the
previous line or earlier among unrelated straight-line statements. Matching
already-present `new(T)` assignments are supported too. Preserve declaration
comments and the macro inlining annotation; function metadata matches the draft.

Require the allocation's inferred pointer type to match the declared type.
Leave package variables, initialized/grouped declarations, prior reads/writes,
address escapes, type mismatches, labels, directives, control headers and
cross-block cases unchanged. Multiple independent local allocations can combine;
empty composite literals in intervening assignments are not treated as scopes.
The operation is idempotent and does not hoist allocation or introduce closures.

Validation: minimal red-to-green test, adjacent/earlier declarations, multiple
locals, comment retention, scope/control-flow and first-use negative cases,
metadata, real template.c, and a compiled runtime test combining earlier local
declaration, macro expansion, allocation, field writes and printf. Library and
CLI tests exclude experiments; existing out/ is unchanged.
Library and CLI suites pass. Installed prototype-68; regenerate output to apply.
