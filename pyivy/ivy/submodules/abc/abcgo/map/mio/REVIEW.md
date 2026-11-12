# MIO port review

The actual `map/mio` package builds and passes its ordinary and race suites.
The parser, expression encoding, SOP operations, library/gate/pin accessors,
root selection, formatting, timing/area transforms, profiles and frame commands
are production Go. No generated source is hidden by build tags or renaming.

## Separate readability and compactness pass

The port retains ABC's compact AND-expression representation and original
function names. A small precedence parser replaces the mutable strtok/stack
parser while preserving explicit/implicit AND, XOR, OR, prefix/postfix negation
and precedence. Topological evaluation and root-first expression construction
remain distinct. SOP derivation retains both polarities and subsumption-free
cube operations. Maps and slices replace raw linked hash records and realloc.
Library commands use the existing `abcmain.Frame` registry and resource lifetime
rather than another global session implementation.

Historical `mioForm.c` is an obsolete alternative implementation of the same
functions in `mioFunc.c`, absent from the original module.make. `mioForm.go`
records this source relationship; the active implementation is in mioFunc.go.
Inactive CUDD truth evaluation is superseded by the source's active compact
expression evaluator. Internal token-scanning helpers are replaced by the Go
parser, rather than exposed as mutable char-pointer APIs.

## Separate bug review and corrections

- Name matching uses whole tokens, preventing `a` from matching `aa` or `aaa`.
- Explicit pins must cover formula names exactly; duplicates/unused pins and
  missing declarations return errors rather than reaching uninitialized arrays.
- Malformed operators/parentheses, missing semicolons/numeric fields, invalid
  phases, nonfinite numbers and negative areas are rejected with source lines.
- Wildcard pin expansion duplicates timing records without aliasing linked nodes.
- The reader rejects third outputs and repeated output names instead of leaving
  inconsistent twin links; original ordering and name ordering remain separate.
- The final comment needs no trailing newline and cannot run past its buffer.
- Ordinary and extended timing records retain all intended fields.
- Numeric writer precision retains fractional capacitances and timing values;
  optional SOP output is commented so the result remains valid genlib.
- Verilog combines twin outputs into one module and escapes non-simple identifiers.
- Timing/area/name changes invalidate both matching caches.
- Writers/profile readers propagate I/O/parse errors. Profile changes are staged
  until parsing succeeds. Selected-gate printing preserves existing cell IDs.
- Negative pin indices are rejected; missing pin delay returns positive infinity.
- Delay transfer recomputes the gate maximum after copying pin delays.
- Truth-table composition uses the actual expression for all arities instead of
  the old alternate SOP evaluator's missing OR accumulation for small functions.
- NPN enumeration and scheduled generation produce independently checked orbits.
- Replacing a frame library invalidates dependent mapper resources and unmaps
  networks through that frame, preserving the framework's ownership rules.

## Integration boundaries and limits

Liberty parsing belongs to `map/scl`, as in the original source. SCL owns the
Liberty-to-genlib bridge and implements `LibertyReader`; a frame may register it
as `mio.libertyReader`. The secondary AMAP genlib representation similarly uses
`GenlibPreparer` at `mio.genlibPreparer`. Missing adapters return explicit errors;
MIO does not pretend that a secondary representation was prepared.

NF/GIA matching algorithms live in `giaNf.c`/`giaIiff.c` and depend on MIO. Their
owner confirmed these algorithms remain unported GIA drafts. `MatchingProvider`
keeps the original dependency direction and implements validated per-library
cache lifecycle/fetching. Cache creation returns an explicit error without a
provider and preserves caller outputs. **Actual NF/GIA matching-engine
integration remains outstanding outside this package.** No fake match tables or
success stubs are supplied. SCL cell-ID lookup has a separate typed interface
implemented by the SCL owner.

The source's 32-bit SOP cubes hold at most 16 input literals. The port reports
that limit instead of silently truncating higher-input cubes. Generic expression
truth generation is bounded to 20 variables. `Mio_GateReadTruthP` exposes a
read-only snapshot for <=6 variables and the stored slice for larger truth
arrays; callers should treat returned truth data as read-only. Root collection
using fixed six-delay cell records rejects limits above six inputs.

Library/frame mutations require caller serialization. Race tests verify tested
operations, not arbitrary concurrent mutation of the same library.

## Validation

`GOCACHE=/tmp/abcgo-cache go -C abcgo test -race -count=1 ./map/mio`
passes against the actual package. Coverage includes Boolean/SOP evaluation,
eight-variable truth tables, malformed formulas/genlib, wildcard/explicit pin
ordering, standard/extended timings, exclusions, twin outputs, genlib precision
round trips, Verilog output, profile transactions, root selection, frame command
read/write/print workflow, timing composition, NPN orbits and explicit missing
matching-provider errors.

`TestOriginalCExpressionTruth` compiles the evaluator extracted directly from
original `src/map/mio/exp.h` with a minimal vector scaffold and compares 100
random compact expressions. It passed with the C compiler available here; the
oracle test skips when original source or compiler is absent. C is used only by
this optional differential test, never by the production package.
