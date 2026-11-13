package fraig

import (
	"github.com/glycerine/abcgo/sat/msat"
	"testing"
)

func TestFeedbackCompressionLeavesExistingPartialBits(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
	defer Fraig_ManFree(p)
	a := Fraig_ManReadIthVar(p, 0)
	p.iPatsPerm = 31
	p.iWordPerm = 1
	p.iWordStart = 3
	a.puSimD[0] = uint32(1) << 31 // old nonpermanent pattern in the destination slot
	a.puSimD[2] = 0               // newly retained pattern is zero
	msat.Msat_IntVecPush(p.vPatsReal, 64)
	Fraig_FeedBackCompress(p)
	if !fraigBit(a.puSimD, 31) {
		t.Fatal("C compression leaves the existing partial-word bit set")
	}
}

func TestFeedbackCompressionCopiesOverlappingSources(t *testing.T) {
	p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
	defer Fraig_ManFree(p)
	a := Fraig_ManReadIthVar(p, 0)
	p.iPatsPerm = 31
	p.iWordPerm = 1
	p.iWordStart = 3
	a.puSimD[0] = uint32(1) << 31
	a.puSimD[1] = 0
	msat.Msat_IntVecPush(p.vPatsReal, 32)
	msat.Msat_IntVecPush(p.vPatsReal, 31)
	Fraig_FeedBackCompress(p)
	if !fraigBit(a.puSimD, 31) || !fraigBit(a.puSimD, 32) {
		t.Fatal("overlapping source pattern was overwritten")
	}
}
