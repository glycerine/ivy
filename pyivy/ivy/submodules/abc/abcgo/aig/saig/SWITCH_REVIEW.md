# SAIG switching and compact AIG review

The generated `saigSwitch.go` draft has been replaced with native graph
simulation and a Go slice backed bit-parallel simulator. The C `Vec_Int_t`
result still stores `float32` values as their IEEE bit patterns for source ABI
compatibility.

Review checks:

- Primary inputs receive independent 32-bit random patterns; latch outputs
  begin at zero and transfer their prior frame's latch-input word thereafter.
- Complemented edges are applied before AND/XOR evaluation, and output nodes
  copy their driver's value.
- Prefix frames are excluded from the probability denominator; invalid/empty
  frame requests return an empty vector instead of dividing by zero.
- Compact AIG varints use Go-owned byte storage and detect truncated or
  out-of-range values. Unsupported XOR node kinds fail explicitly because the
  source compact format stores only AND nodes.

Focused tests cover switching formulas, output probability consistency, and a
compact input/output round trip. Run the race+vet commands with the explicit
file lists from this review's sibling `saigSwitch_test.go` and `saig.go`.
