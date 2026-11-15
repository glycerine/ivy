package gia

import (
	"fmt"
	"testing"
)

func TestCofactorAndQuantificationRetainOutputRoles(t *testing.T) {
	for _, op := range []string{"cofactor0", "cofactor1", "exists2", "universal", "exists"} {
		for _, regs := range []bool{false, true} {
			if regs && (op == "universal" || op == "exists") {
				continue
			}
			t.Run(fmt.Sprintf("%s_regs%v", op, regs), func(t *testing.T) {
				src := quantificationReviewGraph(0, regs, op == "exists")
				constr := int32(1)
				if op == "exists" {
					constr = 0
				}
				if err := src.SetAigerProperties(1, constr); err != nil {
					t.Fatal(err)
				}
				var out *Gia_Man_t
				switch op {
				case "cofactor0":
					out = Gia_ManDupCofactorVar(src, 0, 0)
				case "cofactor1":
					out = Gia_ManDupCofactorVar(src, 0, 1)
				case "exists2":
					out = Gia_ManDupExist2(src, 0)
				case "universal":
					out = Gia_ManDupUniv(src, 0)
				case "exists":
					out = Gia_ManDupExist(src, 0)
				}
				if out.aigerBadOutputs != 1 || out.nConstrs != constr {
					t.Fatalf("lost roles bad%d constraints%d, want1/%d", out.aigerBadOutputs, out.nConstrs, constr)
				}
			})
		}
	}
}

func quantPartitionApply(src *Gia_Man_t, op string, variable int32) *Gia_Man_t {
	switch op {
	case "cofactor0":
		return Gia_ManDupCofactorVar(src, variable, 0)
	case "cofactor1":
		return Gia_ManDupCofactorVar(src, variable, 1)
	case "exists2":
		return Gia_ManDupExist2(src, variable)
	case "universal":
		return Gia_ManDupUniv(src, variable)
	case "exists":
		return Gia_ManDupExist(src, variable)
	default:
		panic("unknown quantified partition test operation")
	}
}

func quantPartitionGraph(signs int32, regs, single, typed, layout bool) *Gia_Man_t {
	src := quantificationReviewGraph(signs, regs, single)
	if !typed {
		src = Gia_ManDupNoMuxes(src, 1)
	}
	if layout {
		src = Gia_ManDupOrderDfsReverse(src)
	}
	return src
}

func TestQuantificationPartitionAndTruthMatrix(t *testing.T) {
	for _, op := range []string{"cofactor0", "cofactor1", "exists2", "universal", "exists"} {
		t.Run(op, func(t *testing.T) {
			for _, regs := range []bool{false, true} {
				if regs && (op == "universal" || op == "exists") {
					continue
				}
				for _, typed := range []bool{false, true} {
					for _, layout := range []bool{false, true} {
						for _, signs := range []int32{0, 2, 5, 7} {
							template := quantPartitionGraph(signs, regs, op == "exists", typed, layout)
							for bad := int32(0); bad <= Gia_ManPoNum(template); bad++ {
								for constr := int32(0); constr <= Gia_ManPoNum(template)-bad; constr++ {
									for variable := int32(0); variable < Gia_ManPiNum(template); variable++ {
										src := quantPartitionGraph(signs, regs, op == "exists", typed, layout)
										if err := src.SetAigerProperties(bad, constr); err != nil {
											t.Fatal(err)
										}
										out := quantPartitionApply(src, op, variable)
										if out.nRegs != src.nRegs || Gia_ManCiNum(out) != 4 || Gia_ManCoNum(out) != Gia_ManCoNum(src) {
											t.Fatal("quantification changed complete output/state interface")
										}
										if src.nConstrs != constr || src.aigerBadOutputs != bad {
											t.Fatal("quantification changed source role counts")
										}
										if name, spec := out.Identity(); name != "quantification" || spec != "quantification.aig" {
											t.Fatal("quantification identity lost")
										}
										parsed := partitionReviewRoundTrip(t, out, bad, constr)
										for mask := 0; mask < 16; mask++ {
											original := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
											negative, positive := append([]bool(nil), original...), append([]bool(nil), original...)
											negative[variable] = false
											positive[variable] = true
											zero, err := src.Evaluate(negative)
											if err != nil {
												t.Fatal(err)
											}
											one, err := src.Evaluate(positive)
											if err != nil {
												t.Fatal(err)
											}
											want := make([]bool, len(zero))
											for i := range want {
												switch op {
												case "cofactor0":
													want[i] = zero[i]
												case "cofactor1":
													want[i] = one[i]
												case "exists", "exists2":
													want[i] = zero[i] || one[i]
												case "universal":
													// The C routine quantifies only the first output universally;
													// auxiliary outputs use their positive cofactor.
													want[i] = one[i]
													if i == 0 {
														want[i] = zero[i] && one[i]
													}
												}
											}
											checkQuantificationOutputs(t, out, original, want)
											checkQuantificationOutputs(t, parsed, original, want)
										}
									}
								}
							}
						}
					}
				}
			}
		})
	}
}

func TestDerivedCofactorOperationsKeepUnclassifiedOutputs(t *testing.T) {
	for _, kind := range []string{"guarded_input0", "guarded_input1", "guarded_node0", "guarded_node1", "block", "mux", "mux_alias"} {
		t.Run(kind, func(t *testing.T) {
			single := kind == "mux" || kind == "mux_alias"
			src := quantPartitionGraph(2, false, single, true, true)
			constr := int32(1)
			if single {
				constr = 0
			}
			if err := src.SetAigerProperties(1, constr); err != nil {
				t.Fatal(err)
			}
			var out, negative *Gia_Man_t
			id, value := int32(0), int32(0)
			switch kind {
			case "guarded_input0", "guarded_input1":
				id = Gia_ManCi(src, 0).id
			case "guarded_node0", "guarded_node1":
				id = Gia_ObjFaninId0p(src, Gia_ManPo(src, 0))
			}
			if kind == "guarded_input1" || kind == "guarded_node1" {
				value = 1
			}
			switch kind {
			case "block":
				out = Gia_ManDupBlock(src, 1)
			case "mux", "mux_alias":
				negative = src
				if kind == "mux" {
					negative = quantPartitionGraph(5, false, true, true, false)
					if err := negative.SetAigerProperties(0, 1); err != nil {
						t.Fatal(err)
					}
				}
				out = Gia_ManDupMux(0, src, negative)
			default:
				out = Gia_ManDupCofactorObj(src, id, value)
			}
			parsed := partitionReviewRoundTrip(t, out, 0, 0)
			for mask := 0; mask < 16; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				values, original, err := src.evaluateObjects(inputs)
				if err != nil {
					t.Fatal(err)
				}
				want := append([]bool(nil), original...)
				switch kind {
				case "mux", "mux_alias":
					if !inputs[0] {
						want, err = negative.Evaluate(inputs)
						if err != nil {
							t.Fatal(err)
						}
					}
				case "block": // Block size one retains each input/output but derives roles.
				default:
					// Guarded substitution is original output AND (selected signal == value).
					// On the guard's domain, substitution leaves every output unchanged.
					guard := values[id] == (value != 0)
					for i := range want {
						want[i] = want[i] && guard
					}
				}
				checkQuantificationOutputs(t, out, inputs, want)
				checkQuantificationOutputs(t, parsed, inputs, want)
			}
		})
	}
}
