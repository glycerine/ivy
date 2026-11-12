package gia

import (
	"bufio"
	"fmt"
	"os"

	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManSimReadFile(filename string) *Vec_Int_t {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("Cannot open input file.")
		return nil
	}
	defer f.Close()
	result := vec.Vec_IntAlloc(1000)
	r := bufio.NewReader(f)
	for {
		c, err := r.ReadByte()
		if err != nil {
			break
		}
		if c == '0' || c == '1' {
			vec.Vec_IntPush(result, int32(c-'0'))
		}
	}
	return result
}
func Gia_ManSimWriteFile(filename string, patterns *Vec_Int_t, outputs int32) int32 {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Cannot open output file.")
		return 0
	}
	defer f.Close()
	if outputs <= 0 || int32(vec.Vec_IntSize(patterns))%outputs != 0 {
		panic("gia: invalid output pattern dimensions")
	}
	w := bufio.NewWriter(f)
	for i, bit := range vec.Vec_IntArray(patterns) {
		if bit != 0 && bit != 1 {
			panic("gia: invalid output bit")
		}
		if err = w.WriteByte(byte('0' + bit)); err != nil {
			return 0
		}
		if int32(i)%outputs == outputs-1 {
			if err = w.WriteByte('\n'); err != nil {
				return 0
			}
		}
	}
	if err = w.Flush(); err != nil {
		return 0
	}
	if err = f.Close(); err != nil {
		return 0
	}
	return 1
}

// Gia_ManSimSimulateOne evaluates a stream of input patterns from zero state.
func Gia_ManSimSimulateOne(p *Gia_Man_t, patterns *Vec_Int_t) *Vec_Int_t {
	nPi := Gia_ManPiNum(p)
	data := vec.Vec_IntArray(patterns)
	if nPi <= 0 || len(data)%int(nPi) != 0 {
		panic("gia: invalid input pattern dimensions")
	}
	Gia_ManConst0(p).fMark1 = 0
	for i := int32(0); i < Gia_ManRegNum(p); i++ {
		Gia_ManRo(p, i).fMark1 = 0
	}
	defer func() {
		for _, obj := range p.pObjs {
			obj.fMark1 = 0
		}
	}()
	result := vec.Vec_IntAlloc(1000)
	for start := 0; start < len(data); start += int(nPi) {
		for i := int32(0); i < nPi; i++ {
			value := data[start+int(i)]
			if value != 0 && value != 1 {
				panic("gia: invalid input bit")
			}
			Gia_ManPi(p, i).fMark1 = uint32(value)
		}
		for _, obj := range p.pObjs {
			if Gia_ObjIsAnd(obj) != 0 {
				obj.fMark1 = (Gia_ObjFanin0(obj).fMark1 ^ obj.fCompl0) & (Gia_ObjFanin1(obj).fMark1 ^ obj.fCompl1)
			} else if Gia_ObjIsCo(obj) != 0 {
				obj.fMark1 = Gia_ObjFanin0(obj).fMark1 ^ obj.fCompl0
			}
		}
		for i := int32(0); i < Gia_ManPoNum(p); i++ {
			vec.Vec_IntPush(result, int32(Gia_ManPo(p, i).fMark1))
		}
		for i := int32(0); i < Gia_ManRegNum(p); i++ {
			Gia_ManRo(p, i).fMark1 = Gia_ManRi(p, i).fMark1
		}
	}
	return result
}
func Gia_ManSimSimulatePattern(p *Gia_Man_t, input, output string) {
	patterns := Gia_ManSimReadFile(input)
	if patterns == nil {
		return
	}
	nPi := Gia_ManPiNum(p)
	if nPi <= 0 || int32(vec.Vec_IntSize(patterns))%nPi != 0 {
		fmt.Printf("The number of 0s and 1s in the input file (%d) does not evenly divide by the number of primary inputs (%d).\n", vec.Vec_IntSize(patterns), nPi)
		return
	}
	out := Gia_ManSimSimulateOne(p, patterns)
	if Gia_ManSimWriteFile(output, out, Gia_ManPoNum(p)) != 0 {
		fmt.Printf("Output patterns are written into file %q.\n", output)
	}
}
