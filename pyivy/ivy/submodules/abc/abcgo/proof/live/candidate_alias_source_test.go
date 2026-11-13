package live

import (
	"slices"
	"testing"
)

func TestSourceCandidateVectorsRetainAliasWhenNoExclusions(t *testing.T) {
	hints := []int32{2, 3}
	result := findRemainingMonotoneCandidates(nil, hints)
	result[0] = 7
	if hints[0] != 7 {
		t.Fatal("source returns the hint vector itself")
	}
	filtered := findRemainingMonotoneCandidates([]int32{7}, hints)
	if !slices.Equal(filtered, []int32{3}) {
		t.Fatalf("filtered %v", filtered)
	}
	vectors := &antecedentConsequentVectorsStruct{attrConsequentCandidates: hints}
	result = updateAnteConseVectors(vectors)
	result[0] = 9
	if hints[0] != 9 {
		t.Fatal("source returns candidate vector itself for empty antecedents")
	}
	vectors.attrAntecedents = []int32{9}
	if got := updateAnteConseVectors(vectors); !slices.Equal(got, []int32{3}) {
		t.Fatalf("filtered %v", got)
	}
}
