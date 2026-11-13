package cnf

import "testing"

func TestMsopTableTruths(t *testing.T) {
	covers := Cnf_ReadMsops()
	masks := [4][2]uint16{{0x5555, 0xaaaa}, {0x3333, 0xcccc}, {0x0f0f, 0xf0f0}, {0x00ff, 0xff00}}
	for truth, cover := range covers {
		var got uint16
		for _, encoded := range cover {
			cube := uint16(0xffff)
			lit := int(encoded)
			for b := 3; b >= 0; b-- {
				digit := lit % 3
				lit /= 3
				if digit < 2 {
					cube &= masks[b][digit]
				}
			}
			got |= cube
		}
		if got != uint16(truth) {
			t.Fatalf("truth %04x got %04x", truth, got)
		}
	}
}
