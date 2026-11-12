package gia

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

func giaDumpBase(p *Gia_Man_t, name string) string {
	if name != "" {
		return name
	}
	return p.pName
}
func giaSeedSimulation(seed int32) {
	Gia_ManRandomW(1)
	for i := int32(0); i < seed; i++ {
		Gia_ManRandomW(0)
	}
}
func giaSimulationBit(p *Gia_Man_t, obj *Gia_Obj_t, pattern int32) byte {
	return byte(Gia_ManObjSim(p, obj.id)[pattern/64] >> uint(pattern%64) & 1)
}
func giaWriteDump(path string, data []byte) {
	if err := os.WriteFile(path, data, 0666); err != nil {
		panic(fmt.Sprintf("gia: write %q: %v", path, err))
	}
}

// Gia_ManDumpFiles writes packed binary training/test input and output data.
func Gia_ManDumpFiles(p *Gia_Man_t, train, test, seed int32, name string) {
	if train < 0 || test < 0 {
		panic("gia: negative pattern count")
	}
	base := giaDumpBase(p, name)
	files := []string{base + "_x.train.data", base + "_y.train.data", base + "_x.test.data", base + "_y.test.data"}
	giaSeedSimulation(seed)
	for partition, units := range []int32{train, test} {
		patterns := units * 64
		// The C routine asks for patterns words (not patterns/64), which also
		// determines the random sequence consumed before the next partition.
		Gia_ManSimulateWords(p, patterns)
		for kind := 0; kind < 2; kind++ {
			width := Gia_ManCiNum(p)
			if kind == 1 {
				width = Gia_ManCoNum(p)
			}
			var buf bytes.Buffer
			for _, value := range []int32{2, patterns, width} {
				if err := binary.Write(&buf, binary.NativeEndian, value); err != nil {
					panic(err)
				}
			}
			bits := make([]byte, int(patterns)*int(width)/8)
			var bit int
			for k := int32(0); k < patterns; k++ {
				for i := int32(0); i < width; i++ {
					var obj *Gia_Obj_t
					if kind == 0 {
						obj = Gia_ManCi(p, i)
					} else {
						obj = Gia_ManCo(p, i)
					}
					bits[bit/8] |= giaSimulationBit(p, obj, k) << uint(bit%8)
					bit++
				}
			}
			buf.Write(bits)
			giaWriteDump(files[2*partition+kind], buf.Bytes())
		}
	}
	giaWriteDump(base+".flist", []byte(strings.Join(files, "\n")+"\n"))
	fmt.Printf("Finished dumping files %q and %q.\n", files[0], files[1])
	fmt.Printf("Finished dumping files %q and %q.\n", files[2], files[3])
	fmt.Printf("Finished dumping file list %q.\n", base+".flist")
}

// Gia_ManDumpPlaFiles writes randomly simulated training, validation, and test covers.
func Gia_ManDumpPlaFiles(p *Gia_Man_t, train, valid, seed int32, name string) {
	if train < 0 || valid < 0 {
		panic("gia: negative pattern count")
	}
	base := giaDumpBase(p, name)
	giaSeedSimulation(seed)
	for i, units := range []int32{train, valid, valid} {
		Gia_ManSimulateWords(p, units)
		var buf bytes.Buffer
		writer := bufio.NewWriter(&buf)
		fmt.Fprintf(writer, ".i %d\n.o %d\n.p %d\n.type fr\n", Gia_ManCiNum(p), Gia_ManCoNum(p), units*64)
		for k := int32(0); k < units*64; k++ {
			for n := int32(0); n < Gia_ManCiNum(p); n++ {
				writer.WriteByte('0' + giaSimulationBit(p, Gia_ManCi(p, n), k))
			}
			writer.WriteByte(' ')
			for n := int32(0); n < Gia_ManCoNum(p); n++ {
				writer.WriteByte('0' + giaSimulationBit(p, Gia_ManCo(p, n), k))
			}
			writer.WriteByte('\n')
		}
		fmt.Fprintln(writer, ".e")
		if err := writer.Flush(); err != nil {
			panic(err)
		}
		giaWriteDump(base+[]string{".train.pla", ".valid.pla", ".test.pla"}[i], buf.Bytes())
	}
	fmt.Printf("Finished dumping files: %q.\n", base+".{train, valid, test}.pla")
}
