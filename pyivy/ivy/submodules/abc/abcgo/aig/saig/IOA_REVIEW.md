# SAIG BLIF input/output review

## Readability and representation

BLIF parsing and printing now operate on `aig.Graph` and `aig.Edge`, with
`ReadBLIF` and `WriteBLIF` returning standard Go errors. The legacy
`Saig_ManReadBlif` and `Saig_ManDumpBlif` names remain as thin compatibility
wrappers. Signal names are returned as ordinary strings rather than a shared
static buffer. The reader intentionally supports the dialect emitted by this
writer and rejects more general multi-row covers instead of guessing.

## Bug review

The reader validates model/end records, duplicate signal definitions, missing
fanins, malformed truth rows, unsupported directives and output/latch counts.
Round-trip tests exhaust every assignment of a small sequential graph with
complemented logic and a latch feedback input. Regression tests verify
unsupported multi-row tables and unknown directives fail explicitly. Writer
errors and close errors are returned to callers.

## Verification and scope

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./aig/saig/saigIoa.go ./aig/saig/saigIoa_test.go -count=1
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./aig/saig/saigIoa.go ./aig/saig/saigIoa_test.go
```

The full SAIG package remains blocked by unrelated C-syntax drafts. General
BLIF covers, nonzero latch initialization and external BLIF dialects are outside
the source writer's own restricted format.
