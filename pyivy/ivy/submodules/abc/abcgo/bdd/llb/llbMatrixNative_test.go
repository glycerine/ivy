package llb

import "testing"

func TestMatrixVerificationAndVarOrder(t *testing.T) {
	m := Llb_MtrAlloc(1, 1, 3, 3)
	m.pMatrix[0][0], m.pMatrix[2][0] = true, true
	m.pMatrix[1][1] = true
	m.pColSums = []int32{1, 1, 1}
	m.pRowSums = []int32{2, 1, 0}
	Llb_MtrVerifyMatrix(m)
	got := Llb_MtrFindVarOrder(m)
	want := []int32{2, 1, 0}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("variable order = %v, want %v", got, want)
		}
	}
	if Llb_MtrVarName(m, 0) != "pi" || Llb_MtrVarName(m, 1) != "CS" || Llb_MtrVarName(m, 2) != "NS" {
		t.Fatalf("unexpected variable labels")
	}
}

func TestScheduleColumnChoiceAndPartialProduct(t *testing.T) {
	m := Llb_MtrAlloc(0, 0, 4, 2)
	m.pMatrix[0][0], m.pMatrix[1][0] = true, true
	m.pMatrix[2][1] = true
	m.pRowSums = []int32{2, 1}
	m.pColSums = []int32{1, 1, 1, 0}
	m.pProdVars[0], m.pProdNums[0] = true, 1
	if got := Llb_MtrFindBestColumn(m, 1); got != 1 {
		t.Fatalf("best quantification column = %d, want 1", got)
	}
	Llb_MtrUseSelectedColumn(m, 1)
	if m.pProdVars[0] || m.pProdNums[0] != 0 {
		t.Fatalf("last occurrence should quantify variable: %v / %v", m.pProdVars, m.pProdNums)
	}
	if m.pProdVars[1] || m.pProdNums[1] != 0 {
		t.Fatalf("unselected variable changed: %v / %v", m.pProdVars, m.pProdNums)
	}
}

func TestMatrixSwapKeepsColumnStateTogether(t *testing.T) {
	m := Llb_MtrAlloc(0, 0, 3, 2)
	m.pMatrix[0][1], m.pMatrix[2][0] = true, true
	m.pColSums[0], m.pColSums[2] = 1, 1
	m.pColGrps[0] = &Llb_Grp_t_{Id: 10}
	m.pColGrps[2] = &Llb_Grp_t_{Id: 20}
	Llb_MtrSwapColumns(m, 0, 2)
	if !m.pMatrix[0][0] || !m.pMatrix[2][1] || m.pColSums[0] != 1 || m.pColSums[2] != 1 {
		t.Fatal("column payload and cached counts did not move together")
	}
	if m.pColGrps[0].Id != 20 || m.pColGrps[2].Id != 10 {
		t.Fatal("column group order did not move with matrix columns")
	}
}
