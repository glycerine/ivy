# Original ABC functionality conversion oracle

Generated with this repository's original `./abc` executable on 2026-11-12:

    ./abc -c 'read_blif abcgo/base/abc/test_vectors/function_off.blif; bdd -s; sop; write_blif /tmp/abc-function-off-oracle.blif'

`function_off_oracle.blif` is that output with only its timestamp comment
removed. The Go test compares the exact reordered fanin names and normalized
SOP rows, beyond exhaustive semantic tests of both phase choices. No Go output
was used to generate this reference. The source executable is not needed when
running Go unit tests.
