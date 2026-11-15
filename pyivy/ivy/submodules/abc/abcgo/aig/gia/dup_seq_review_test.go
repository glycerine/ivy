package gia

import "testing"

func TestRegisterFlipRetainsDfsPortOrder(t *testing.T) {
	for initial := int32(0); initial < 4; initial++ {
		t.Run(string(rune('0'+initial)), func(t *testing.T) {
			src := partitionReviewGraph(5, false, true)
			out := Gia_ManDupFlip(src, []int32{initial})
			for mask := 0; mask < 32; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
				old := append([]bool(nil), inputs...)
				for reg := int32(0); reg < 2; reg++ {
					if initial>>reg&1 != 0 {
						old[3+reg] = !old[3+reg]
					}
				}
				want, err := src.Evaluate(old)
				if err != nil {
					t.Fatal(err)
				}
				for reg := int32(0); reg < 2; reg++ {
					if initial>>reg&1 != 0 {
						want[5+reg] = !want[5+reg]
					}
				}
				checkQuantificationOutputs(t, out, inputs, want)
			}
		})
	}
}

func stateFlipReviewExpected(t *testing.T, src *Gia_Man_t, inputs []bool, initial []int32) []bool {
	t.Helper()
	old := append([]bool(nil), inputs...)
	bit := func(reg int32) bool { return uint32(initial[reg>>5])>>uint(reg&31)&1 != 0 }
	for reg := int32(0); reg < src.nRegs; reg++ {
		if bit(reg) {
			old[Gia_ManPiNum(src)+reg] = !old[Gia_ManPiNum(src)+reg]
		}
	}
	want, err := src.Evaluate(old)
	if err != nil {
		t.Fatal(err)
	}
	for reg := int32(0); reg < src.nRegs; reg++ {
		if bit(reg) {
			want[Gia_ManPoNum(src)+reg] = !want[Gia_ManPoNum(src)+reg]
		}
	}
	return want
}

func TestStateFlipRepresentationMatrix(t *testing.T) {
	cases, rows := 0, 0
	for signs := 0; signs < 8; signs++ {
		for _, typed := range []bool{false, true} {
			for _, layout := range []bool{false, true} {
				for regs := int32(0); regs <= 5; regs++ {
					for mask := int32(0); mask < 1<<regs; mask++ {
						src := partitionReviewGraph(signs, typed, layout)
						src.nRegs = regs // Reinterpret the fixture's existing terminal suffix.
						if err := src.SetAigerProperties(1, 1); err != nil {
							t.Fatal(err)
						}
						initial := []int32{mask}
						out := Gia_ManDupFlip(src, initial)
						if out.nRegs != regs || out.nObjs != src.nObjs || Gia_ManIsNormalized(out) == 0 {
							t.Fatal("state flip changed graph interface")
						}
						for i := int32(0); i < Gia_ManCiNum(src); i++ {
							expected := uint32((i + 1) << 1)
							if i >= Gia_ManPiNum(src) {
								expected ^= uint32(mask >> uint(i-Gia_ManPiNum(src)) & 1)
							}
							if Gia_ManCi(src, i).Value != expected {
								t.Fatal("complemented source CI copy literal changed")
							}
						}
						restored := partitionReviewRoundTrip(t, out, 1, 1)
						twice := Gia_ManDupFlip(out, initial)
						for assignment := 0; assignment < 32; assignment++ {
							inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0, assignment&16 != 0}
							want := stateFlipReviewExpected(t, src, inputs, initial)
							checkQuantificationOutputs(t, out, inputs, want)
							checkQuantificationOutputs(t, restored, inputs, want)
							original, err := src.Evaluate(inputs)
							if err != nil {
								t.Fatal(err)
							}
							checkQuantificationOutputs(t, twice, inputs, original)
							rows++
						}
						cases++
					}
				}
			}
		}
	}
	t.Logf("%d flips, %d input/state rows with serialization and double-flip checks", cases, rows)
}

func TestStateFlipPackedWordBoundaries(t *testing.T) {
	for _, regs := range []int{31, 32, 33, 63, 64, 65} {
		for _, layout := range []bool{false, true} {
			for pattern := 0; pattern < 4; pattern++ {
				src := Gia_ManStart(1)
				a := Gia_ManAppendCi(src)
				states := make([]int32, regs)
				for i := range states {
					states[i] = Gia_ManAppendCi(src)
				}
				Gia_ManAppendCo(src, a^1)
				Gia_ManAppendCo(src, states[regs-1])
				for i := range states {
					Gia_ManAppendCo(src, states[(i+1)%regs]^int32(i&1))
				}
				Gia_ManSetRegNum(src, int32(regs))
				if layout {
					src = Gia_ManDupOrderDfsReverse(src)
				}
				initial := make([]int32, (regs+31)/32)
				for i := 0; i < regs; i++ {
					flip := pattern == 1 || pattern == 2 && i%2 == 0 || pattern == 3 && (i == 0 || i%32 == 31 || i%32 == 0 || i == regs-1)
					if flip {
						initial[i>>5] |= int32(uint32(1) << uint(i&31))
					}
				}
				out := Gia_ManDupFlip(src, initial)
				seed := uint64(73)
				for row := 0; row < 128; row++ {
					inputs := make([]bool, regs+1)
					for i := range inputs {
						seed = seed*6364136223846793005 + 1442695040888963407
						inputs[i] = seed>>63 != 0
					}
					checkQuantificationOutputs(t, out, inputs, stateFlipReviewExpected(t, src, inputs, initial))
				}
			}
		}
	}
}
