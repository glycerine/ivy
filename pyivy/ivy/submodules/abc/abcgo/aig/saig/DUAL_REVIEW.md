# Dual rail port review

`saigDual.go` now implements `Saig_ManDupDual` and `Saig_ManBlockPo` over the
native AIG graph. The dual-rail representation stores false and true rails in
that order. Tests cover the unknown-output case (both rails equal) and the
register/output interface added by output blocking.

Review checks:

- Complemented inputs swap rails while preserving the unknown state.
- Primary inputs use complementary rails unless explicitly selected as dual
  inputs; latch rails use either per-latch DC polarity or the global option.
- AND and XOR nodes compute their false and true rails independently.
- The output miter checks the requested zero, one, or unknown condition.
- `Saig_ManBlockPo` appends the countdown state and patches only primary output
  drivers; existing latch inputs remain in order.

Focused race and vet commands:

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race \
  ./aig/saig/saig.go ./aig/saig/saigDual.go ./aig/saig/saigDual_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo vet \
  ./aig/saig/saig.go ./aig/saig/saigDual.go ./aig/saig/saigDual_test.go
```
