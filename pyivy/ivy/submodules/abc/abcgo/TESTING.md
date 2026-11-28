# Testing abcgo

`go test ./...` needs only this module. It does not need the original C
sources (`../src`), a native `abc`, `libabc.a`, or any file outside the
module.

Many tests compare the port with the original ABC: they compile C fixtures
against `../src`, run the native `abc`, or read C sources. They do so through
`internal/testoracle`, which replays what the C side produced from recordings
in each package's `test_vectors/oracle/` (`<TestName>.json`, gzipped when
large), and C files the tests read from `test_vectors/oracle/_files/` at the
module root. In replay only the C side is recorded: the Go side always runs.

## Re-recording

Re-record after changing a test's C-side commands, scripts or inputs, or the
C sources. Recording needs the module inside the ABC tree (`../src` next to
it), a native release `abc`, and `../libabc.a` with assertions on:

    (cd .. && make ABC_USE_STDINT_H=1 ABC_USE_NO_READLINE=1 libabc.a)
    ABCGO_ORACLE_RECORD=1 ABC_BINARY=$PWD/.scratch/abc-release/abc go test ./pkg -run TestX

Two tests are recorded separately:

    # ASan build of giaEdge.c against a release library (C asserts compiled out)
    ABCGO_ORACLE_RECORD=1 ABC_LIBABC=$PWD/.scratch/abc-release/libabc.a \
      ABC_BINARY=$PWD/.scratch/abc-release/abc go test ./base/abci -run 'TestNativeEdgeStaysInBounds$'
    # opt-in (about three minutes of Go time): runs with ABC_MFS3_FULL=1
    ABC_MFS3_FULL=1 ABCGO_ORACLE_RECORD=1 \
      ABC_BINARY=$PWD/.scratch/abc-release/abc go test ./base/abci -run 'TestMfs3DefaultLikeNative$'

A test whose command has no recording fails with "oracle: nothing recorded
for ...", naming the command.

Not recorded: `sat/minisatp` TestAgainstOriginalMiniSat and
TestIncrementalAgainstOriginalMiniSat compare with MiniSat 1.14p (not part of
ABC) when MINISATP_ORACLE / MINISATP_INCREMENTAL_ORACLE name its binaries.
