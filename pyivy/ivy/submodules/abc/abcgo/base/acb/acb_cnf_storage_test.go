package acb

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

// Vec_WecInit and Vec_WecPushLevel preserve previously allocated level
// entries. ACB uses those entries as byte vectors for per-node CNF storage.
func TestCNFStorageResizePreservesEntries(t *testing.T) {
	p := &Acb_Ntk_t{}
	acbResizeCnfs(p, 2)
	vec.Vec_StrPush(&p.vCnfs[0], 7)
	vec.Vec_StrPush(&p.vCnfs[1], 9)
	acbResizeCnfs(p, 1)
	acbResizeCnfs(p, 2)
	if got := vec.Vec_StrEntry(&p.vCnfs[1], 0); got != 9 {
		t.Fatalf("retained CNF byte = %d, want 9", got)
	}
	acbResizeCnfs(p, cap(p.vCnfs)+1)
	if got := vec.Vec_StrEntry(&p.vCnfs[0], 0); got != 7 {
		t.Fatalf("CNF byte after allocation growth = %d, want 7", got)
	}
	if vec.Vec_StrSize(&p.vCnfs[len(p.vCnfs)-1]) != 0 {
		t.Fatal("new CNF level is not empty")
	}
}
