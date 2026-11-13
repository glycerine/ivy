package acec

import "testing"

// The C oracle supports width six and records the same four statuses used by
// testSourceMultiplierAcec, including the discarded-output false equivalence.
func TestSourceMultiplier6Truth(t *testing.T) { testSourceMultiplierTruth(t, 6) }
func TestSourceMultiplier6Acec(t *testing.T)  { testSourceMultiplierAcec(t, 6) }
