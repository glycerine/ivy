package gia

import (
	"bufio"
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"os"
	"strconv"
	"strings"
)

// Gia_ManStgPrint writes rows in the original one-based state format.
func Gia_ManStgPrint(out io.Writer, lines *Vec_Int_t, ins, outs, states int32) {
	data := vec.Vec_IntArray(lines)
	if len(data)%4 != 0 || ins < 1 || outs < 1 || states < 1 {
		panic("gia: invalid state table")
	}
	digits := len(strconv.FormatInt(int64(states), 10))
	for i := 0; i < len(data); i += 4 {
		a, cur, next, b := data[i], data[i+1], data[i+2], data[i+3]
		if a < 0 || uint64(a) >= uint64(1)<<uint(ins) || b < 0 || uint64(b) >= uint64(1)<<uint(outs) || cur < 1 || cur > states || next < 1 || next > states {
			panic("gia: invalid state row")
		}
		fmt.Fprintf(out, "%0*b %*d %*d %0*b\n", ins, a, digits, cur, digits, next, outs, b)
	}
}
func Gia_ManStgReadLines(path string, ins, outs, states *int32) *Vec_Int_t {
	if strings.HasSuffix(path, "aig") {
		fmt.Printf("Input file %q has extension %q.\n", path, "aig")
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Cannot open file %q.\n", path)
		return nil
	}
	defer f.Close()
	rows := vec.Vec_IntAlloc(1000)
	inputBits, outputBits, nStates := -1, -1, int32(1)
	scan := bufio.NewScanner(f)
	line := 0
	for scan.Scan() {
		line++
		text := strings.TrimSpace(scan.Text())
		if text == "" || text[0] == '.' || text[0] == '#' {
			continue
		}
		fields := strings.Fields(text)
		if len(fields) != 4 {
			fmt.Printf("Invalid state row at line %d.\n", line)
			return nil
		}
		if inputBits < 0 {
			inputBits = len(fields[0])
			outputBits = len(fields[3])
		}
		if len(fields[0]) != inputBits || len(fields[3]) != outputBits {
			fmt.Printf("Inconsistent state row at line %d.\n", line)
			return nil
		}
		a, e1 := strconv.ParseUint(fields[0], 2, 31)
		cur, e2 := strconv.ParseInt(fields[1], 10, 32)
		next, e3 := strconv.ParseInt(fields[2], 10, 32)
		b, e4 := strconv.ParseUint(fields[3], 2, 31)
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil || cur < 0 || next < 0 {
			fmt.Printf("Invalid state row at line %d.\n", line)
			return nil
		}
		vec.Vec_IntPushFour(rows, int32(a), int32(cur), int32(next), int32(b))
		nStates = max(nStates, int32(cur)+1)
	}
	if scan.Err() != nil || inputBits < 0 {
		return nil
	}
	if ins != nil {
		*ins = int32(inputBits)
	}
	if outs != nil {
		*outs = int32(outputBits)
	}
	if states != nil {
		*states = nStates
	}
	return rows
}
func Gia_ManStgRead(path string, hot, verbose int32) *Gia_Man_t {
	var ins, outs, states int32
	lines := Gia_ManStgReadLines(path, &ins, &outs, &states)
	if lines == nil {
		return nil
	}
	return Gia_ManStgKHot(lines, ins, outs, states, hot, verbose)
}
