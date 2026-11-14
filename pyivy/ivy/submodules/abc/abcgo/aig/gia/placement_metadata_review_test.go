package gia

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unsafe"
)

func TestPlacementMetadataBoundaryRoundTrips(t *testing.T) {
	t.Logf("placement record: %d bytes", unsafe.Sizeof(Gia_Plc_t{}))
	for _, x := range []uint32{0, 1, 0x7fff} {
		for _, y := range []uint32{0, 1, 0x7fff} {
			for flags := uint32(0); flags < 4; flags++ {
				g := Gia_ManStart(1)
				Gia_ManAppendCo(g, Gia_ManAppendCi(g))
				word := flags&1 | x<<1 | (flags>>1)<<16 | y<<17
				var payload []byte
				for id := int32(0); id < g.nObjs; id++ {
					payload = binary.LittleEndian.AppendUint32(payload, word)
				}
				if err := g.SetAigerExtension('p', payload); err != nil {
					t.Fatal(err)
				}
				data, err := MarshalAiger(g, AigerWriteOptions{})
				if err != nil {
					t.Fatal(err)
				}
				round, err := ParseAiger(data, AigerReadOptions{SkipStrash: true})
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(round.AigerExtension('p'), payload) {
					t.Fatalf("x=%d y=%d flags=%d: placement wire bytes changed", x, y, flags)
				}
				for _, record := range round.pPlacement {
					if uint32(record.xCoord) != x || uint32(record.yCoord) != y || uint32(record.fFixed) != flags&1 || uint32(record.fUndef) != flags>>1 {
						t.Fatalf("x=%d y=%d flags=%d: fields changed: %+v", x, y, flags, record)
					}
				}
			}
		}
	}
}

func TestPlacementMetadataRejectsOutOfRangeFields(t *testing.T) {
	for _, field := range []string{"x", "y", "fixed", "undefined"} {
		t.Run(field, func(t *testing.T) {
			g := Gia_ManStart(1)
			Gia_ManAppendCo(g, Gia_ManAppendCi(g))
			g.pPlacement = make([]Gia_Plc_t, g.nObjs)
			switch field {
			case "x":
				g.pPlacement[0].xCoord = 0x8000
			case "y":
				g.pPlacement[0].yCoord = 0x8000
			case "fixed":
				g.pPlacement[0].fFixed = 2
			case "undefined":
				g.pPlacement[0].fUndef = 2
			}
			if _, err := MarshalAiger(g, AigerWriteOptions{}); err == nil {
				t.Fatal("out-of-range placement field was serialized")
			}
		})
	}
}
