package if_

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
	"slices"
	"testing"
)

func TestDsdCanonicalPinProfileMatchesConstructedPaths(t *testing.T) {
	p, cut := ifDelayTestManager(t)
	defer If_ManStop(p)
	p.pIfDsdMan = If_DsdManAlloc(3, 0)
	defer If_DsdManFree(p.pIfDsdMan, 0)
	perm := make([]byte, 3)
	lit := If_DsdManCompute(p.pIfDsdMan, dau.Dau_DsdToTruth("(abc)", 3), 3, perm, "")
	p.vTtDsds[3] = vec.Vec_IntStart(2)
	vec.Vec_IntWriteEntry(p.vTtDsds[3], 1, lit)
	p.vTtPerms[3] = vec.Vec_StrAlloc(12)
	vec.Vec_StrFill(p.vTtPerms[3], 12, 0)
	for i, pin := range perm {
		vec.Vec_StrWriteEntry(p.vTtPerms[3], 6+i, pin)
	}
	for i := int32(0); i < 3; i++ {
		If_ObjCutBest(If_CutLeaf(p, cut, i)).Delay = 0
	}
	aig := vec.Vec_IntAlloc(0)
	defer vec.Vec_IntFree(aig)
	If_CutDsdBalanceEval(p, cut, aig)
	data := vec.Vec_IntArray(aig)
	paths := [][3]int{{0, -1, -1}, {-1, 0, -1}, {-1, -1, 0}}
	for i := 0; i < len(data)-1; i += 2 {
		a, b := paths[data[i]>>1], paths[data[i+1]>>1]
		row := [3]int{-1, -1, -1}
		for pin := 0; pin < 3; pin++ {
			if a[pin] >= 0 {
				row[pin] = a[pin] + 1
			}
			if b[pin] >= 0 {
				row[pin] = max(row[pin], b[pin]+1)
			}
		}
		paths = append(paths, row)
	}
	root := paths[len(paths)-1]
	expected := []byte{byte(root[0]), byte(root[1]), byte(root[2])}
	profile := make([]byte, 3)
	If_CutDsdBalancePinDelays(p, cut, profile)
	if !slices.Equal(profile, expected) {
		t.Fatalf("canonical permutation%v pin profile%v constructed path depths%v", perm, profile, expected)
	}
}
