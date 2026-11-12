# External plugin protocol

## Readability pass

The executable plugin protocol stays distinct from command parsing and uses
Frame's registered commands and data bridge. Binary metadata remains attached
to the frame. Public protocol readers retain their original names; their values
use Go slices and the dependency-neutral IOA AIGER model. Process arguments are
passed directly to os/exec. Every temporary file belongs to one invocation;
`!` explicitly retains its directory.

## Separate bug and fidelity pass

Preserved command discovery (`-abc -list-commands`), execution input/output
arguments, result/depth/abstraction/witness parsing and six-bit AIG encoding.
The protocol uses Frame bridge formats `gia-aiger`, `gia-flop-classes` and
`gia-counterexample`. A bridge must implement them; missing integration returns
an error and is not silently treated as successful conversion.

Tests exercise a real discovered plugin executable, status and depth parsing,
compressed AIG roundtrip including padding, input expansion and witness
verification. Status parsing now respects token length and whitespace instead
of hardcoding an eight-byte offset. Truncated encodings return errors rather
than reading beyond the buffer. Byte-count rounding avoids integer overflow.
Zero-live-input witness expansion returns an explicit error rather than looping
forever. Unlike the source's unchecked witness conversion, malformed witness
length or a nonfailing trace returns failure. Initial register state is zeroed
when expanding omitted unused primary inputs, matching the source algorithm.

Race tests and vet pass for selected command core, external adapters and plugin
files. The full package now passes race tests and vet; AUTO_REVIEW.md records
the exact MF evaluator integration dependency.
