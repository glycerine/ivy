package aig

import "testing"

func TestCreateMiterDoesNotInheritOutputRoles(t *testing.T) {
	for _, role := range []string{"constraint", "assertion"} {
		t.Run(role, func(t *testing.T) {
			a, b := NewGraph(), NewGraph()
			a.AddOutput(a.NewInput())
			b.AddOutput(b.NewInput())
			if role == "constraint" {
				Aig_ManSetConstrNum(a, 1)
			} else {
				a.Assertions = 1
			}
			result := Aig_ManCreateMiter(a, b, 0)
			if result.nConstrs != 0 || result.Assertions != 0 {
				t.Fatalf("new XOR miter inherited operand output roles: constraints=%d assertions=%d", result.nConstrs, result.Assertions)
			}
			if err := result.Check(); err != nil {
				t.Fatal(err)
			}
			for _, value := range []bool{false, true} {
				outputs, err := result.Evaluate([]bool{value})
				if err != nil || len(outputs) != 1 || outputs[0] {
					t.Fatalf("input=%v outputs=%v err=%v", value, outputs, err)
				}
			}
		})
	}
}
