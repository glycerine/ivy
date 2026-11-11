# cpp2go implementation and experiment plan

Last updated: 2026-11-11

Status: first bounded C experiment validated. A source-preserving prototype and
repaired Satoko heap/vector port exist. Broader structural parsing and C++ support
remain unimplemented; speedup over an LLM-only baseline is not established.

## Purpose

Build a portable Go library in `cpp2go/`, with a thin command-line wrapper in
`cpp2go/cmd/cpp2go/main.go`, that mechanically converts C and C++ into readable
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

- [x] Generate drafts through cpp2go rather than hand-authoring the initial port.
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
| D01 | Go library with thin CLI under cpp2go/ | accepted | User-specified structure |
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

- [Prototype library](translate.go), [lexer](lex.go), and [CLI](cmd/cpp2go/main.go).
- [Frozen generated drafts and handoff](experiments/satoko/draft-v1/run.json).
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
go -C cpp2go test -v ./...
go -C cpp2go run ./cmd/cpp2go -out /tmp/cpp2go-new-run \
  ../src/sat/satoko/utils/vec/vec_uint.h \
  ../src/sat/satoko/utils/vec/vec_int.h \
  ../src/sat/satoko/utils/vec/vec_sdbl.h \
  ../src/sat/satoko/utils/heap.h
```

Choose a nonexistent output directory. The first saved generation is intentionally
frozen; current code additionally fixes pointer declarations such as `type *p`
and preserves CRLF continued line comments. Current output therefore need not be
byte-identical to `draft-v1`. Results are deterministic for a fixed version/input
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
| Test command | `go -C cpp2go test -v ./...` passed |

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
- [draft-v2](experiments/satoko/draft-v2/run.json) records a fresh four-header run
  with `package satoko`. Frozen v1 artifacts are historical evidence, not current
  generator output, and were not rewritten.

Validation: `go -C cpp2go test -v ./...` passes, including typed macro draft
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

[Draft v3](experiments/satoko/draft-v3/00_vec_uint.h.go.txt) now includes
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

[Draft v4 heap](experiments/satoko/draft-v4/03_heap.h.go.txt) now contains the
requested heap_t Go struct; all three vector declarations are converted as well.
Previous drafts remain historical artifacts.

Validation: `go -C cpp2go test ./...` passes, including parse/type checks of
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

[Draft v5](experiments/satoko/draft-v5/00_vec_uint.h.go.txt) demonstrates braces
around the vector capacity guard and includes the macro argument TODO annotations.
Validation: `go -C cpp2go test ./...` passes, covering exact requested output,
nested/dangling else, comments, multiline else, loop bodies, idempotence,
parse/type checking, and existing C/Go differential tests.


### Header include guards — prototype-6

Recognized whole-file `#ifndef NAME` / `#define NAME` / matching final `#endif`
wrappers are omitted from Go drafts, including comments on those directive lines.
Leading license comments, other comments, nested conditionals, and original source
copies remain intact. Matching requires the same guard name, an empty define,
no outer else/elif, and no substantive content outside the wrapper. Unmatched or
ambiguous patterns remain unchanged.

Generated `experiments/satoko/draft-v6/`; all four Satoko header guards are absent
from their Go drafts. Tests cover nested directives and negative matches.
`go -C cpp2go test ./...` passes, including the C/Go differential suite.


### Ternary indentation — prototype-7

Generated if/else blocks now reuse the original statement's leading indentation,
indent branch statements one level deeper (four spaces, or a tab for tab-indented
source), and align closing braces and else with their if. Nested ternaries carry
the indentation recursively. Initializer declarations and their following if
align, and the replacement no longer adds an unnecessary trailing blank line.
Existing expression/comment contents are not globally reformatted.

Generated `experiments/satoko/draft-v7/`. Exact-output tests cover the requested
allocation example, nested tab-indented branches, and variable initializers.
`go -C cpp2go test ./...` passes, including lazy-branch execution and differential
checks. No gofmt invocation on incomplete draft output is required.


### Empty for-loop bodies — prototype-8

Per user direction, a trailing bodyless for-header in a macro replacement now
receives `{}`, and a for-loop with a null-statement body (`;`) receives `{}` in
both macro and ordinary function drafts. This is an explicit exception to the
otherwise verbatim macro replacement policy. Header expressions are retained;
nonempty bodies, comments, and literals remain unchanged.

Generated `experiments/satoko/draft-v8/`; the vector foreach macro functions now
contain explicit empty loop bodies. Tests cover bodyless headers, null bodies,
comments, existing bodies, protected literals, and repeated application.
`go -C cpp2go test ./...` passes, including the C/Go differential suite.


### Includes and standalone markers — prototype-9

Go drafts now prefix include directives with `//`, preserving their text and
spacing. Continued include directives have every physical line commented.
Unknown standalone top-level identifier lines such as ABC_NAMESPACE_HEADER_START
and ABC_NAMESPACE_HEADER_END are also commented. Known type/keyword lines,
function declarations, tokens inside function bodies, and existing comments or
literals are left alone. This is a conservative marker heuristic, not symbol
resolution. Source copies remain unchanged.

Generated `experiments/satoko/draft-v9/`. Tests cover the exact requested include
block, continued includes, idempotence, and preservation of multiline code.
`go -C cpp2go test ./...` passes, including the differential suite.


### Consistent package per output directory — prototype-10

Corrected the CLI's per-file package inference. The package is now resolved once
before output creation: explicit `-package`, otherwise the directory name of the
first source input. That single value is supplied to every translation in the
batch and recorded in `run.json`. The destination directory name is irrelevant.
`ResolvePackage` exposes the same rule for library batch callers; standalone
single-file analysis retains its source-directory default.

Generated `experiments/satoko/draft-v10/`; all four Go drafts use `package vec`
because vec_uint.h is the first source. Older mixed-package draft directories
remain historical experiment records. CLI tests cover inputs from different
source directories, a differently named destination, an explicit override,
recorded metadata, and invalid-name rejection before output creation.
`go -C cpp2go test ./...` passes.
