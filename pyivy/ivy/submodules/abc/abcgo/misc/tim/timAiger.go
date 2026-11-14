package tim

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// MarshalAigerTiming uses a private deterministic extension payload for the
// Go TIM manager. The surrounding AIGER extension remains ABC's standard h
// chunk; readers of unrelated hierarchy formats continue to use their normal
// fallback path.
func (p *Tim_Man_t) MarshalAigerTiming(_ bool) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("tim: nil manager")
	}
	var b bytes.Buffer
	b.WriteString("TIM1")
	putI := func(v int32) { _ = binary.Write(&b, binary.LittleEndian, v) }
	putF := func(v float32) { putI(int32(math.Float32bits(v))) }
	putI(int32(p.nCis))
	putI(int32(p.nCos))
	putI(int32(p.nTravIds))
	putI(boolInt(p.fUseTravId))
	putI(int32(len(p.vDelayTables)))
	for _, table := range p.vDelayTables {
		putI(int32(len(table)))
		for _, v := range table {
			putF(v)
		}
	}
	putI(int32(len(p.vBoxes)))
	for _, x := range p.vBoxes {
		for _, v := range []int32{int32(x.iBox), int32(x.TravId), int32(x.nInputs), int32(x.nOutputs), int32(x.iDelayTable), int32(x.iCopy), int32(x.firstIn), int32(x.firstOut), boolInt(x.fBlack), int32(len(x.Inouts))} {
			putI(v)
		}
		for _, v := range x.Inouts {
			putI(int32(v))
		}
	}
	for _, group := range [][]Tim_Obj_t{p.pCis, p.pCos} {
		putI(int32(len(group)))
		for _, x := range group {
			for _, v := range []int32{int32(x.Id), int32(x.TravId), int32(x.iObj2Box), int32(x.iObj2Num)} {
				putI(v)
			}
			putF(x.timeArr)
			putF(x.timeReq)
		}
	}
	return b.Bytes(), nil
}

func boolInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

// UnmarshalAigerTiming recognizes only this bridge's payload; unknown ABC h
// extensions are left to Gia's generic timing data container.
func UnmarshalAigerTiming(data []byte) (*Tim_Man_t, error) {
	if len(data) < 4 || string(data[:4]) != "TIM1" {
		return nil, fmt.Errorf("tim: unknown timing extension")
	}
	r := bytes.NewReader(data[4:])
	readI := func() (int32, error) { var v int32; e := binary.Read(r, binary.LittleEndian, &v); return v, e }
	readF := func() (float32, error) { v, e := readI(); return math.Float32frombits(uint32(v)), e }
	readCount := func() (int, error) {
		v, e := readI()
		if e != nil {
			return 0, e
		}
		if v < 0 || int64(v) > int64(len(data)) {
			return 0, fmt.Errorf("tim: invalid serialized count %d", v)
		}
		return int(v), nil
	}
	nCis, e := readCount()
	if e != nil {
		return nil, e
	}
	nCos, e := readCount()
	if e != nil {
		return nil, e
	}
	trav, e := readCount()
	if e != nil {
		return nil, e
	}
	use, e := readI()
	if e != nil {
		return nil, e
	}
	p := Tim_ManStart(nCis, nCos)
	p.nTravIds = trav
	p.fUseTravId = use != 0
	n, e := readCount()
	if e != nil {
		return nil, e
	}
	p.vDelayTables = make([][]float32, n)
	for i := range p.vDelayTables {
		k, err := readCount()
		if err != nil {
			return nil, err
		}
		p.vDelayTables[i] = make([]float32, k)
		for j := range p.vDelayTables[i] {
			p.vDelayTables[i][j], e = readF()
			if e != nil {
				return nil, e
			}
		}
	}
	n, e = readCount()
	if e != nil {
		return nil, e
	}
	p.vBoxes = make([]*Tim_Box_t, n)
	for i := range p.vBoxes {
		v := make([]int32, 10)
		for j := range v {
			v[j], e = readI()
			if e != nil {
				return nil, e
			}
		}
		if v[9] < 0 || int64(v[9]) > int64(len(data)) {
			return nil, fmt.Errorf("tim: invalid box inout count")
		}
		x := &Tim_Box_t{iBox: int(v[0]), TravId: int(v[1]), nInputs: int(v[2]), nOutputs: int(v[3]), iDelayTable: int(v[4]), iCopy: int(v[5]), firstIn: int(v[6]), firstOut: int(v[7]), fBlack: v[8] != 0, Inouts: make([]int, int(v[9]))}
		for j := range x.Inouts {
			z, err := readI()
			if err != nil {
				return nil, err
			}
			x.Inouts[j] = int(z)
		}
		p.vBoxes[i] = x
	}
	for gi := 0; gi < 2; gi++ {
		n, e = readCount()
		if e != nil {
			return nil, e
		}
		want := nCis
		if gi == 1 {
			want = nCos
		}
		if n != want {
			return nil, fmt.Errorf("tim: terminal count mismatch")
		}
		group := p.pCis
		if gi == 1 {
			group = p.pCos
		}
		for i := range group {
			var v [4]int32
			for j := range v {
				v[j], e = readI()
				if e != nil {
					return nil, e
				}
			}
			group[i].Id = int(v[0])
			group[i].TravId = int(v[1])
			group[i].iObj2Box = int(v[2])
			group[i].iObj2Num = int(v[3])
			group[i].timeArr, e = readF()
			if e != nil {
				return nil, e
			}
			group[i].timeReq, e = readF()
			if e != nil {
				return nil, e
			}
		}
	}
	if r.Len() != 0 {
		return nil, fmt.Errorf("tim: trailing timing extension bytes")
	}
	return p, nil
}

var _ io.Reader = (*bytes.Reader)(nil)
