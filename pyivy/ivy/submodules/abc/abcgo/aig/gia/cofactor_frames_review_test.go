package gia

import (
	"slices"
	"testing"
)

func TestCofactoringFramesMatchSequentialSimulation(t *testing.T) {
	cases, rows := 0, 0
	for signs := 0; signs < 8; signs++ {
		for typed := 0; typed < 2; typed++ {
			for layout := 0; layout < 4; layout++ {
				for save := int32(0); save < 2; save++ {
					for init := int32(0); init < 2; init++ {
						for disable := int32(0); disable < 2; disable++ {
							src := windowReviewGraph(signs, typed, layout)
							Gia_ManSetRegNum(src, 2)
							params := &Gia_ParFra_t{NFrames: 3, fSaveLastLit: save, FInit: init, FDisableSt: disable}
							raw := Gia_ManUnrollStart(src, params)
							for count := int32(1); count <= 3; count++ {
								graph := Gia_ManUnrollAdd(raw, count).(*Gia_Man_t)
								for assignment := 0; assignment < 1<<Gia_ManCiNum(graph); assignment++ {
									inputs := make([]bool, Gia_ManCiNum(graph))
									for i := range inputs {
										inputs[i] = assignment>>i&1 != 0
									}
									state := []bool{false, false}
									prefix := 0
									if init == 0 {
										prefix = 2
										offset := 0
										if save == 0 {
											offset = 6
										}
										copy(state, inputs[offset:offset+2])
									}
									var expected []bool
									last := false
									for frame := 0; frame < int(count); frame++ {
										base := frame * 2
										if save != 0 {
											base += prefix
										}
										outputs, err := src.Evaluate([]bool{inputs[base], inputs[base+1], state[0], state[1]})
										if err != nil {
											t.Fatal(err)
										}
										if save == 0 {
											expected = append(expected, outputs[:2]...)
										}
										last = outputs[Gia_ManPoNum(src)-1]
										copy(state, outputs[2:])
									}
									values, outputs, err := graph.evaluateObjects(inputs)
									if err != nil {
										t.Fatal(err)
									}
									if !slices.Equal(outputs, expected) {
										t.Fatalf("phase%d typed%d layout%d save%d init%d disable%d frames%d assignment%d: outputs%v want%v", signs, typed, layout, save, init, disable, count, assignment, outputs, expected)
									}
									literal := Gia_ManUnrollLastLit(raw)
									if save != 0 {
										if got := values[literal>>1] != (literal&1 != 0); got != last {
											t.Fatal("last-property literal disagrees with sequential simulation")
										}
									} else if literal != 0 {
										t.Fatal("fixed mode unexpectedly saved a property literal")
									}
									rows++
								}
								cases++
							}
							Gia_ManUnrollStop(raw)
						}
					}
				}
			}
		}
	}
	t.Logf("%d ordinary/typed/layout/phase/save/init/hash frame cases and %d assignments", cases, rows)
}
