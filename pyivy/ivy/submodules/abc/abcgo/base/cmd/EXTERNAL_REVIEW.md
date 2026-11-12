# External synthesis tools

## Readability pass

`external.go` shares executable lookup and SIS/MVSIS BLIF roundtrip logic.
Transfers call registered codec handlers directly, so internal serialization
neither creates history records nor triggers duplicate network backups. Child
arguments go through os/exec rather than shell concatenation. Temporary BLIFs
are private to each invocation and always removed.

## Separate fidelity and bug pass

SIS/MVSIS retain the original `-x -c` read/transform/write protocol; Capo retains
`-f`, `-log` and optional plotting. Resource-file executable overrides remain
supported. Registered `write_blif` must serialize the current network and
`read_blif` must replace it; absent codecs return an explicit failure. This
avoids importing the unfinished ABC package into command infrastructure.

Tests run a real subprocess with spaces in its filename, transfer the generated
BLIF through both registered handlers, and verify that a failed subprocess
never invokes the reader. Race tests pass for the extended command-core scope.
Native external tool integration still depends on installing those tools and
registering the network codecs; these tests validate the process/file protocol.
