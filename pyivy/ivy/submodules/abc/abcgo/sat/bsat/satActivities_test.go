package bsat

import (
	"math"
	"testing"
)

func TestSolverIntegerActivityRescale(t *testing.T) {
	s := sat_solver_new()
	sat_solver_setnvars(s, 3)
	s.activity[0] = 0x7fffffff
	act_var_bump(s, 0)
	if s.var_inc < 16 || s.activity[0] >= 0x80000000 {
		t.Fatal(s.var_inc, s.activity)
	}
	s.pGlobalVars = []int32{0, 1, 0}
	before := s.activity[1]
	act_var_bump_global(s, 1)
	if s.activity[1] != before+3*s.var_inc {
		t.Fatal("global scaling")
	}
	s.factors = []float64{0, 0, 2.5}
	before = s.activity[2]
	act_var_bump_factor(s, 2)
	if s.activity[2] != before+uint64(int32(float32(s.var_inc)*2.5)) {
		t.Fatal("factor scaling")
	}
	old := s.var_inc
	act_var_decay(s)
	if s.var_inc != old+(old>>4) {
		t.Fatal("decay")
	}
}
func TestSolverDoubleAndExtendedActivities(t *testing.T) {
	s := sat_solver_new()
	sat_solver_setnvars(s, 2)
	s.VarActType = 1
	s.var_inc = math.Float64bits(1)
	s.var_decay = math.Float64bits(1 / 0.95)
	s.activity[0] = math.Float64bits(2)
	act_var_bump(s, 0)
	if math.Float64frombits(s.activity[0]) != 3 {
		t.Fatal("double bump")
	}
	act_var_decay(s)
	if math.Float64frombits(s.var_inc) <= 1 {
		t.Fatal("double decay")
	}
	s.VarActType = 2
	s.var_inc = Xdbl_Const1()
	s.var_decay = Xdbl_FromDouble(2)
	s.activity[0] = Xdbl_FromDouble(2)
	act_var_bump(s, 0)
	if Xdbl_ToDouble(s.activity[0]) != 3 {
		t.Fatal("extended bump")
	}
	act_var_decay(s)
	if Xdbl_ToDouble(s.var_inc) != 2 {
		t.Fatal("extended decay")
	}
}
func TestSolverClauseActivityRepresentations(t *testing.T) {
	s := sat_solver_new()
	sat_solver_setnvars(s, 2)
	h := sat_solver_clause_new(s, []lit{0, 2}, 1)
	c := Sat_MemClauseHand(&s.Mem, h)
	s.act_clas.ptr[0] = 0x7fffffff
	act_clause_bump(s, c)
	if s.cla_inc < 1024 || uint32(s.act_clas.ptr[0]) >= 0x80000000 {
		t.Fatal("integer clause rescale")
	}
	s.ClaActType = 1
	s.cla_inc = 1
	s.act_clas.ptr[0] = int32(math.Float32bits(2.5))
	act_clause_bump(s, c)
	if math.Float32frombits(uint32(s.act_clas.ptr[0])) != 3.5 {
		t.Fatal("float clause bump")
	}
}
func TestExtendedActivityArithmetic(t *testing.T) {
	for _, a := range []float64{1, 1.5, 3, 12, 123.5, 1e100} {
		for _, b := range []float64{1, 2.5, 19} {
			x, y := Xdbl_FromDouble(a), Xdbl_FromDouble(b)
			sum := Xdbl_ToDouble(Xdbl_Add(x, y))
			product := Xdbl_ToDouble(Xdbl_Mul(x, y))
			if math.Abs(sum-(a+b))/(a+b) > 1e-12 || math.Abs(product-a*b)/(a*b) > 1e-12 {
				t.Fatal(a, b, sum, product)
			}
		}
	}
	if Xdbl_Add(Xdbl_Maximum(), Xdbl_Maximum()) != Xdbl_Maximum() || Xdbl_Mul(Xdbl_Maximum(), Xdbl_FromDouble(2)) != Xdbl_Maximum() || Xdbl_Div(Xdbl_Const1(), 200) != Xdbl_Const1() {
		t.Fatal("extended limits")
	}
}
