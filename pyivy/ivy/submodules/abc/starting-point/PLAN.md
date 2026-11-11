# starting-point implementation and experiment plan

Last updated: 2026-11-11

Status: first bounded C experiment validated. A source-preserving prototype and
repaired Satoko heap/vector port exist. Broader structural parsing and C++ support
remain unimplemented; speedup over an LLM-only baseline is not established.

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
