# SAIG cone-of-influence review

## Readability and representation

`Saig_ManSupport` walks an `aig.Graph` with explicit `aig.Edge` values and
returns unique latch-input output edges in first-visit order. Go maps replace
the source traversal-ID mutation, so support queries do not alter the graph's
scratch state. Primary inputs and constants are naturally excluded.

## Bug review

Traversal regularizes complemented edges before checking node kind or ID. It
deduplicates both visited nodes and latch-input outputs, including when multiple
roots overlap. Tests cover duplicate roots, shared latch support, primary-only
cones, constants, and self-fed latch input support. The printer stops when a
frontier adds no new state, preventing feedback loops from making it run
forever.

## Verification and scope

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./aig/saig/saigCone.go ./aig/saig/saigCone_test.go -count=1
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./aig/saig/saigCone.go ./aig/saig/saigCone_test.go
```

The whole SAIG package remains blocked by generated C-syntax drafts outside
this cone group.
