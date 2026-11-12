# Progress indicator review

Readability and idiomatic-Go pass: uses `io.Writer`, `strings.Builder`, synchronized
per-indicator state, atomic global enable/batch settings, and typed update labels.
Command/frame settings enter through `Bar_SetEnabled` and `Bar_SetBatchMode`,
avoiding a dependency cycle with the command package.

Separate bug pass checked throttle thresholds, final position, nil handling,
stopped indicators, long labels and writer failures. Fixed short writes with a
nil underlying error being silently accepted; `Err` now reports `io.ErrShortWrite`.

Regression tests cover initial/updated rendering, throttling, completion, clearing,
idempotent stop, disabled/batch mode, long labels, ordinary errors and short writes.

Validation: `GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./misc/bar` passed.
