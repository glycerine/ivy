# Downsizing and command review

## Compactness/readability pass

The original downsizing algorithm is implemented against the single SC_Man and
TimingNetwork model. Native node-ID slices retain the area-priority candidate
order across postponed rounds; local maps replace temporary C traversal marks.
No separate network mapping or duplicate timing representation is introduced.
The window builder, candidate evaluator, priority collection, overlap test,
progress report, internal driver and public driver retain original API names.
All timing/mapping errors propagate to callers. Library and constraint commands
use the common abcmain registry and real model/MIO operations. Registration rows
mark mapping-changing commands so the shared dispatcher owns undo preparation.

## Independent source/bug pass

Compared every downsizing function against sclDnsize.c. Candidate trials restore
both fanin loads and cone timing, including error paths. Reverse ring traversal,
notch limiting, strict gain selection, overlap deferral, incremental updates,
iteration/timeout behavior and final MIO conversion are retained. Native manager
sorting supports arbitrary IDs; barriers remain topology-transparent rather than
requiring a duplicated pointer graph. Source assertion failures (invalid pivot,
no observable boundary) become errors. Progress avoids division by zero for
zero initial area/delay. The runtime statistics use duration fields.

An observed two-point delay-tangent source bug is fixed in the shared library
utility as documented in REVIEW_reader.md. `read_constr -n` implements the
advertised new-format selection; C accidentally toggles verbosity instead.
Library installation transfers IDs using the supplied frame's actual MIO library
rather than accidentally reading unrelated global frame state.

## Validation and remaining work

Combined reader/model/constraint/time/manager/load/util/command/downsizing selected
files pass race tests. A real two-node mapped chain downsizes both cells, reduces
area4 to2 and preserves delay; zero-notch rejection restores gate/timing state.
An independent original ABC executable oracle reads the generated Liberty and
GENLIB, downsizes the same chain and selects the same two small cells.
The reader original-C lookup and byte-for-byte SCL serialization oracles also pass.

Full package race tests and vet now pass after the buffer sources landed.
All21 original registered command handlers plus the unregistered old-buffer entry
point are wired. Topological clone and buffer mutation are explicit host graph
interfaces; absent capabilities produce errors. A concrete ABC network adapter is
owned by base/abc. Buffer algorithms/tests/review have separate ownership.
Numeric positivity restrictions are command-specific: zero timing gap remains
valid even though read_lib's explicitly supplied slew/gain must be positive.
