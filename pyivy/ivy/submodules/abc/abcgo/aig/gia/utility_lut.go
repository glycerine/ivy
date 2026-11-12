package gia

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"os"
)

const MAX_LUT_SIZE = 8

// Gia_MapLut_t follows the original native binary record layout.
type Gia_MapLut_t struct {
	Type, Out, StartId, NFans int32
	Delay                     float32
	PFans                     [MAX_LUT_SIZE]int32
	PTruth                    [1 << (MAX_LUT_SIZE - 5)]uint32
}

func Gia_AigerWriteLut(p *Gia_Man_t, path string) {
	maxSize := Gia_ManLutSizeMax(p)
	if maxSize > 6 {
		panic("gia: LUT dump supports at most six inputs")
	}
	words := util.Abc_TruthWordNum(maxSize)
	records := make([]Gia_MapLut_t, 0, 1+Gia_ManCiNum(p)+Gia_ManCoNum(p)+Gia_ManLutNum(p))
	Gia_ManFillValue(p)
	constant := Gia_MapLut_t{Type: 3}
	for i := int32(0); i < words; i++ {
		constant.PTruth[i] = ^uint32(0)
	}
	records = append(records, constant)
	Gia_ManConst0(p).Value = 0
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		r := Gia_MapLut_t{Type: 1, Out: int32(len(records)) << 1}
		for w := int32(0); w < words; w++ {
			r.PTruth[w] = 0xaaaaaaaa
		}
		Gia_ManCi(p, i).Value = uint32(r.Out)
		records = append(records, r)
	}
	temp := Vec_WrdStart(int(Gia_ManObjNum(p)))
	for _, obj := range p.pObjs[1:] {
		if Gia_ObjIsLut(p, obj.id) == 0 {
			continue
		}
		r := Gia_MapLut_t{Type: 3, Out: int32(len(records)) << 1}
		for i, id := range Gia_ObjLutFanins(p, obj.id) {
			r.PFans[i] = int32(Gia_ManObj(p, id).Value)
			r.NFans++
		}
		truth := Gia_LutComputeTruth6(p, obj.id, temp)
		r.PTruth[0], r.PTruth[1] = uint32(truth), uint32(truth>>32)
		obj.Value = uint32(r.Out)
		records = append(records, r)
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		fan := Gia_ObjFanin0(obj)
		r := Gia_MapLut_t{Type: 2, Out: int32(len(records)) << 1, NFans: 1}
		r.PFans[0] = int32(fan.Value)
		pattern := uint32(0xaaaaaaaa)
		if Gia_ObjFaninC0(obj)^Gia_ObjIsConst0(fan) != 0 {
			pattern = 0x55555555
		}
		for w := int32(0); w < words; w++ {
			r.PTruth[w] = pattern
		}
		obj.Value = uint32(r.Out)
		records = append(records, r)
	}
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("Cannot open file %q for writing: %v\n", path, err)
		return
	}
	if err = binary.Write(f, binary.NativeEndian, records); err != nil {
		fmt.Printf("Cannot write file %q: %v\n", path, err)
	} else {
		fmt.Printf("Successfully dumped %d bytes of binary data.\n", len(records)*binary.Size(Gia_MapLut_t{}))
	}
	if err = f.Close(); err != nil {
		fmt.Printf("Cannot close file %q: %v\n", path, err)
	}
}
func Gia_DumpLutSizeDistrib(p *Gia_Man_t, path string) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Cannot open file %q: %v\n", path, err)
		return
	}
	defer f.Close()
	var counts [10]int
	for _, obj := range p.pObjs {
		if Gia_ObjIsLut(p, obj.id) != 0 {
			size := Gia_ObjLutSize(p, obj.id)
			if size > 0 && size < 10 {
				counts[size]++
			}
		}
	}
	fmt.Fprint(f, p.pName)
	for i := 1; i < 10; i++ {
		fmt.Fprintf(f, " %d", counts[i])
	}
	fmt.Fprintln(f)
}
