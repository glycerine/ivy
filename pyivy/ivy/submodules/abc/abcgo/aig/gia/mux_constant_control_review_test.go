package gia

import (
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

func checkRawMuxControlCase(t *testing.T, variable bool, control, reverse, phases int32, hash bool) (int, int) {
	t.Helper()
	g := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	ciCount := 2
	c := control
	if variable {
		c = Gia_ManAppendCi(g) ^ control
		ciCount++
	}
	then, otherwise := a, b
	if reverse != 0 {
		then, otherwise = b, a
	}
	then ^= phases & 1
	otherwise ^= (phases >> 1) & 1
	g.EnableMuxes()
	if hash {
		Gia_ManHashStart(g)
	}
	before := g.nObjs
	root := Gia_ManAppendMuxReal(g, c, then, otherwise)
	if root != before<<1 || g.nObjs != before+1 {
		t.Fatal("raw MUX constructor must allocate exactly one node")
	}
	_, _, _, kind := g.Gate(root >> 1)
	canonicalControl := c
	if reverse == 0 {
		canonicalControl ^= 1
	}
	wantKind := 4
	if canonicalControl == 0 {
		wantKind = 5
	}
	if kind != wantKind || g.nBufs != giaBool(wantKind == 5) || g.nMuxes != giaBool(wantKind == 4) {
		t.Fatalf("control %d reverse %d phases %d: incorrect physical node kind or counts", c, reverse, phases)
	}
	Gia_ManAppendCo(g, root)
	variants := []*Gia_Man_t{g, Gia_ManDup(g), Gia_ManDupNormalize(g, 0), Gia_ManDupNormalize(g, 1), Gia_ManCleanup(g)}
	truth := 0
	for mask := 0; mask < 1<<uint(ciCount); mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		condition := control != 0
		if variable {
			inputs = append(inputs, mask&4 != 0)
			condition = inputs[2] != (control != 0)
		}
		tv, ev := inputs[0], inputs[1]
		if reverse != 0 {
			tv, ev = ev, tv
		}
		tv = tv != (phases&1 != 0)
		ev = ev != (phases&2 != 0)
		want := ev
		if condition {
			want = tv
		}
		for _, variant := range variants {
			checkQuantificationOutputs(t, variant, inputs, []bool{want})
		}
		if want {
			truth |= 1 << uint(mask)
		}
	}
	return kind, truth
}

func TestRawMuxConstantAndVariableControlMatrix(t *testing.T) {
	for _, variable := range []bool{false, true} {
		for control := int32(0); control <= 1; control++ {
			for reverse := int32(0); reverse <= 1; reverse++ {
				for phases := int32(0); phases < 4; phases++ {
					for _, hash := range []bool{false, true} {
						if hash && phases&1 != 0 {
							continue // Source's active-hash precondition: regular then argument.
						}
						checkRawMuxControlCase(t, variable, control, reverse, phases, hash)
					}
				}
			}
		}
	}
}

func TestRawMuxDfsWithConstantMappedControl(t *testing.T) {
	for control := int32(0); control <= 1; control++ {
		for reverse := int32(0); reverse <= 1; reverse++ {
			for phases := int32(0); phases < 4; phases++ {
				for _, hash := range []bool{false, true} {
					for _, remap := range []bool{false, true} {
						src, dst := Gia_ManStart(1), Gia_ManStart(1)
						a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
						then, otherwise := a, b
						if reverse != 0 {
							then, otherwise = b, a
						}
						src.EnableMuxes()
						dst.EnableMuxes()
						root := Gia_ManAppendMuxReal(src, c, then^(phases&1), otherwise^((phases>>1)&1))
						Gia_ManAppendCo(src, root)
						Gia_ManFillValue(src)
						src.pObjs[0].Value = 0
						x, y := Gia_ManAppendCi(dst), Gia_ManAppendCi(dst)
						if remap {
							x, y = y, x
						}
						Gia_ManCi(src, 0).Value = uint32(x)
						Gia_ManCi(src, 1).Value = uint32(y)
						Gia_ManCi(src, 2).Value = uint32(control)
						if hash {
							Gia_ManHashStart(dst)
						}
						Gia_ManDupOrderDfs_rec(dst, src, Gia_ManCo(src, 0))
						if dst.nObjs != 5 || dst.nMuxes+dst.nBufs != 1 || Gia_ManCiNum(dst) != 2 || Gia_ManCoNum(dst) != 1 {
							t.Fatal("constant-mapped DFS changed node or terminal count")
						}
						for mask := 0; mask < 4; mask++ {
							inputs := []bool{mask&1 != 0, mask&2 != 0}
							x, y := inputs[0], inputs[1]
							if remap {
								x, y = y, x
							}
							want, err := src.Evaluate([]bool{x, y, control != 0})
							if err != nil {
								t.Fatal(err)
							}
							checkQuantificationOutputs(t, dst, inputs, want)
						}
					}
				}
			}
		}
	}
}

func TestRawMuxControlsOriginalSourceWitness(t *testing.T) {
	exe := compileGiaSourceOracle(t, "mux_constant_control.c")
	data, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C raw MUX controls: %v: %s", err, data)
	}
	rows := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(rows) != 32 {
		t.Fatalf("C emitted %d cases, want 32", len(rows))
	}
	witnesses := 0
	for _, row := range rows {
		fields := strings.Fields(row)
		if len(fields) != 6 {
			t.Fatalf("invalid C row %q", row)
		}
		values := make([]int, 6)
		for i, field := range fields {
			values[i], err = strconv.Atoi(field)
			if err != nil {
				t.Fatal(err)
			}
		}
		variable, control, reverse, phases, sourceMux, sourceTruth := values[0] != 0, int32(values[1]), int32(values[2]), int32(values[3]), values[4], values[5]
		kind, truth := checkRawMuxControlCase(t, variable, control, reverse, phases, false)
		if kind == 5 {
			// Original C silently recognizes this as AND rather than MUX.
			a, b := 0xa, 0xc
			if reverse != 0 {
				a, b = b, a
			}
			if phases&1 != 0 {
				a ^= 0xf
			}
			if phases&2 != 0 {
				b ^= 0xf
			}
			if sourceMux != 0 || sourceTruth != a&b || sourceTruth == truth {
				t.Fatalf("invalid original-C defect witness: %s; repaired truth %x", row, truth)
			}
			witnesses++
		} else if sourceMux != 1 || sourceTruth != truth {
			t.Fatalf("ordinary representable MUX changed: %s; Go truth %x", row, truth)
		}
	}
	if witnesses != 8 {
		t.Fatalf("C reproduced %d unrepresentable controls, want 8", witnesses)
	}
}

func TestRawMuxFalseControlSelectsLowerDataInput(t *testing.T) {
	g := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	g.EnableMuxes()
	Gia_ManAppendCo(g, Gia_ManAppendMuxReal(g, 0, b, a))
	checkQuantificationOutputs(t, g, []bool{true, false}, []bool{true})
}

func TestRawMuxConstantControlsKeepConstructorPreconditions(t *testing.T) {
	for _, name := range []string{"missing_mux_storage", "same_data_object", "invalid_then", "hash_complemented_then", "control_overlaps_data"} {
		t.Run(name, func(t *testing.T) {
			g := Gia_ManStart(1)
			a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			then, otherwise := b, a
			if name != "missing_mux_storage" {
				g.EnableMuxes()
			}
			switch name {
			case "same_data_object":
				then, otherwise = a, a^1
			case "invalid_then":
				then = g.nObjs << 1
			case "hash_complemented_then":
				Gia_ManHashStart(g)
				then ^= 1
			case "control_overlaps_data":
				otherwise = 0
			}
			before := g.nObjs
			defer func() {
				if recover() == nil {
					t.Fatal("raw MUX constructor accepted an invalid input")
				}
				if g.nObjs != before {
					t.Fatal("invalid raw MUX append changed the graph")
				}
			}()
			Gia_ManAppendMuxReal(g, 0, then, otherwise)
		})
	}
}
