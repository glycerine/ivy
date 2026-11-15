package gia

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// giaScanPatterns validates binary input/output rows while reusing one row
// buffer. Parameter-only scans do not retain or decode the dataset. A visit
// receives a temporary row view; callers retaining it must copy it.
func giaScanPatterns(path string, visit func(int, []byte)) (width, count int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	scan.Buffer(make([]byte, 4096), 16<<20)
	width = -1
	var bits []byte
	for scan.Scan() {
		line := scan.Bytes()
		if bytes.IndexByte(line, '.') >= 0 {
			continue
		}
		columns := 0
		bits = bits[:0]
		for _, c := range line {
			if c == '0' || c == '1' {
				columns++
				if visit != nil {
					bits = append(bits, c-'0')
				}
			}
		}
		if width < 0 {
			width = columns
		}
		if columns != width {
			return 0, 0, fmt.Errorf("inconsistent row width %d, expected %d", columns, width)
		}
		if visit != nil {
			visit(count, bits)
		}
		count++
	}
	if err = scan.Err(); err != nil {
		return 0, 0, err
	}
	if count == 0 || width < 1 || count%64 != 0 {
		return 0, 0, fmt.Errorf("expected nonempty rows in groups of 64 with one output")
	}
	return width, count, nil
}

// giaReadPatterns retains rows for callers that explicitly request them.
func giaReadPatterns(path string) ([][]byte, error) {
	var rows [][]byte
	_, _, err := giaScanPatterns(path, func(_ int, row []byte) { rows = append(rows, append([]byte(nil), row...)) })
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func Gia_ManSimParamRead(path string, nIns, nWords *int32) int32 {
	width, count, err := giaScanPatterns(path, nil)
	if err != nil {
		fmt.Printf("Cannot read simulation file %q: %v.\n", path, err)
		return 0
	}
	*nIns, *nWords = int32(width-1), int32(count/64)
	return 1
}

func Gia_ManSimFileRead(path string, nIns, nWords int32, sims *Vec_Wrd_t, values *Vec_Int_t) {
	// Validate before changing caller storage, preserving the existing Go error
	// contract without retaining every decoded byte row. The second pass packs
	// directly into the destination, as in the C reader.
	width, count, err := giaScanPatterns(path, nil)
	if err != nil {
		panic(fmt.Sprintf("gia: cannot read simulation file: %v", err))
	}
	if nIns < 0 || nWords < 0 || count != int(nWords)*64 || width != int(nIns)+1 || Vec_WrdSize(sims) != int(nIns*nWords) {
		panic("gia: simulation file dimensions do not match")
	}
	_, _, err = giaScanPatterns(path, func(pattern int, row []byte) {
		for i := int32(0); i < nIns; i++ {
			if row[i] != 0 {
				Vec_WrdArray(sims)[int(i*nWords)+pattern/64] |= uint64(1) << uint(pattern%64)
			}
		}
		Vec_IntPush(values, int32(row[nIns]))
	})
	if err != nil {
		panic(fmt.Sprintf("gia: cannot read simulation file: %v", err))
	}
	fmt.Printf("Finished reading %d simulation patterns for %d inputs. Probability of 1 at the output is %6.2f %%.\n", count, nIns, 100*float64(Vec_IntSum(values))/float64(count))
}

func Gia_ManSimLogStats(p *Gia_Man_t, path string, total, correct, guess int32) {
	stats := struct {
		Name    string `json:"name"`
		Input   int32  `json:"input"`
		Output  int32  `json:"output"`
		And     int32  `json:"and"`
		Level   int32  `json:"level"`
		Total   int32  `json:"total"`
		Correct int32  `json:"correct"`
		Guess   int32  `json:"guess"`
	}{p.pName, Gia_ManCiNum(p), Gia_ManCoNum(p), Gia_ManAndNum(p), Gia_ManLevelNum(p), total, correct, guess}
	data, err := json.MarshalIndent(stats, "", "    ")
	if err == nil {
		err = os.WriteFile(path, append(data, '\n'), 0666)
	}
	if err != nil {
		panic(fmt.Sprintf("gia: cannot write simulation statistics: %v", err))
	}
}

func Gia_ManCompareValues(p *Gia_Man_t, inputs *Vec_Wrd_t, values *Vec_Int_t, path string) {
	if Gia_ManCiNum(p) == 0 || Gia_ManCoNum(p) == 0 {
		panic("gia: comparison requires inputs and outputs")
	}
	nWords := Vec_WrdSize(inputs) / int(Gia_ManCiNum(p))
	if Vec_IntSize(values) != nWords*64 || nWords == 0 {
		panic("gia: invalid comparison values")
	}
	Gia_ManSimulateWordsInit(p, inputs)
	out := Gia_ManObjSim(p, Gia_ManCo(p, 0).id)
	var correct int32
	for i, value := range Vec_IntArray(values) {
		if int32(out[i/64]>>uint(i%64)&1) == value {
			correct++
		}
	}
	total, ones := int32(Vec_IntSize(values)), Vec_IntSum(values)
	guess := max(ones, total-ones)
	fmt.Printf("Total = %6d.  Errors = %6d.  Correct = %6d.  (%6.2f %%)   Naive guess = %6d.  (%6.2f %%)\n", total, total-correct, correct, 100*float64(correct)/float64(total), guess, 100*float64(guess)/float64(total))
	if path != "" {
		Gia_ManSimLogStats(p, path, total, correct, guess)
		fmt.Printf("Finished dumping statistics into file %q.\n", path)
	}
}

func Gia_ManTestOneFile(p *Gia_Man_t, path, stats string) {
	var inputs, words int32
	if Gia_ManSimParamRead(path, &inputs, &words) == 0 {
		return
	}
	if inputs != Gia_ManCiNum(p) {
		fmt.Printf("The number of inputs in the file %q (%d) does not match the AIG (%d).\n", path, inputs, Gia_ManCiNum(p))
		return
	}
	sims := Vec_WrdStart(int(inputs * words))
	values := Vec_IntAlloc(int(words * 64))
	Gia_ManSimFileRead(path, inputs, words, sims, values)
	Gia_ManCompareValues(p, sims, values, stats)
}
