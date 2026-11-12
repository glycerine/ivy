# NM review

Readability/idiomatic pass: reviewed the API, entry model and map/ring storage.
The package has no C allocation/hash-bin dependence; strings are owned Go values.
Compatibility Nm_* functions retain int32 object IDs, while Manager.Lookup can
represent empty names distinctly from missing IDs. ID enumeration is explicitly
sorted and returns []int32. Kept namesake rings because their anchor-selection
semantics match the original C lookup behavior.

Separate bug pass: exercised removal of anchor, middle and final namesakes;
duplicate IDs, missing deletion, reset/rebuild, unique-name collisions, empty
names and 10,000 entries. Found a zero-value Manager panic on first insertion;
added a failing test, lazily initialized maps in table insertion, and documented
the now-usable zero value. Package tests and race-enabled package tests pass.

This package is fully ported. Consumers of the old vector-returning enumeration
API must use the documented []int32 result.
