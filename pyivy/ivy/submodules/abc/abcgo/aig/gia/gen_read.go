package gia

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// giaReadPatterns reads binary input/output rows and ignores PLA directives.
func giaReadPatterns(path string) ([][]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var rows [][]byte
	scan := bufio.NewScanner(f)
	scan.Buffer(make([]byte, 4096), 16<<20)
	width := -1
	for scan.Scan() {
		line := scan.Text()
		if strings.ContainsRune(line, '.') {
			continue
		}
		var bits []byte
		for i := 0; i < len(line); i++ {
			if line[i] == '0' || line[i] == '1' {
				bits = append(bits, line[i]-'0')
			}
		}
		if width < 0 {
			width = len(bits)
		}
		if len(bits) != width {
			return nil, fmt.Errorf("inconsistent row width %d, expected %d", len(bits), width)
		}
		rows = append(rows, bits)
	}
	if err := scan.Err(); err != nil {
		return nil, err
	}
	if len(rows) == 0 || width < 1 || len(rows)%64 != 0 {
		return nil, fmt.Errorf("expected nonempty rows in groups of 64 with one output")
	}
	return rows, nil
}

func Gia_ManSimParamRead(path string, nIns, nWords *int32) int32 {
	rows, err := giaReadPatterns(path)
	if err != nil {
		fmt.Printf("Cannot read simulation file %q: %v.\n", path, err)
		return 0
	}
	*nIns, *nWords = int32(len(rows[0])-1), int32(len(rows)/64)
	return 1
}

func Gia_ManSimFileRead(path string, nIns, nWords int32, sims *Vec_Wrd_t, values *Vec_Int_t) {
	rows, err := giaReadPatterns(path)
	if err != nil {
		panic(fmt.Sprintf("gia: cannot read simulation file: %v", err))
	}
	if nIns < 0 || nWords < 0 || len(rows) != int(nWords)*64 || len(rows[0]) != int(nIns)+1 || Vec_WrdSize(sims) != int(nIns*nWords) {
		panic("gia: simulation file dimensions do not match")
	}
	for pattern, row := range rows {
		for i := int32(0); i < nIns; i++ {
			if row[i] != 0 {
				Vec_WrdArray(sims)[int(i*nWords)+pattern/64] |= uint64(1) << uint(pattern%64)
			}
		}
		Vec_IntPush(values, int32(row[nIns]))
	}
	fmt.Printf("Finished reading %d simulation patterns for %d inputs. Probability of 1 at the output is %6.2f %%.\n", len(rows), nIns, 100*float64(Vec_IntSum(values))/float64(len(rows)))
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
