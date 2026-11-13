package abci

// ABC verification log reader and writer, ported from abcLog.c.
import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/misc/util"
)

func Abc_NtkWriteLogFile(name string, cex *util.Abc_Cex_t, status, frames int32, command string) {
	f, err := os.Create(name)
	if err != nil {
		fmt.Printf("Cannot open log file for writing %q.\n", name)
		return
	}
	defer f.Close()
	switch status {
	case 1:
		fmt.Fprint(f, "snl_UNSAT")
	case 0:
		fmt.Fprint(f, "snl_SAT")
	case -1:
		fmt.Fprint(f, "snl_UNK")
	default:
		fmt.Println("Abc_NtkWriteLogFile(): Cannot recognize solving status.")
	}
	fmt.Fprintf(f, " %d %s", frames, command)
	if cex != nil && status == 0 {
		fmt.Fprintf(f, " %d", cex.IPo)
	}
	if cex != nil && cex.IFrame != frames {
		fmt.Fprintf(f, " %d", cex.IFrame)
	}
	fmt.Fprintln(f)
	if cex == nil {
		fmt.Fprint(f, "NULL")
	} else {
		for i := int32(0); i < cex.NRegs; i++ {
			if util.Abc_InfoHasBit(cex.PData, i) {
				fmt.Fprint(f, "1")
			} else {
				fmt.Fprint(f, "0")
			}
		}
	}
	fmt.Fprintln(f)
	if cex == nil {
		fmt.Fprint(f, "NULL")
	} else {
		if cex.NBits-cex.NRegs != cex.NPis*(cex.IFrame+1) {
			panic("abci: inconsistent counterexample size")
		}
		for i := cex.NRegs; i < cex.NBits; i++ {
			if util.Abc_InfoHasBit(cex.PData, i) {
				fmt.Fprint(f, "1")
			} else {
				fmt.Fprint(f, "0")
			}
		}
	}
	fmt.Fprintln(f)
}

// abcLogAtoi follows atoi's whitespace, sign, and digit-prefix parsing.
func abcLogAtoi(s string) int32 {
	s = strings.TrimLeft(s, " \t\n\r\v\f")
	i := 0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	start := i
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == start {
		return 0
	}
	n, _ := strconv.ParseInt(s[:i], 10, 32)
	return int32(n)
}

func Abc_NtkReadLogFile(name string, result **util.Abc_Cex_t, framesOut *int32) int32 {
	f, err := os.Open(name)
	if err != nil {
		fmt.Printf("Cannot open log file for reading %q.\n", name)
		return -1
	}
	defer f.Close()
	r := bufio.NewReader(f)
	// fgets(Buffer, 1000, ...) consumes at most 999 bytes including newline.
	line := make([]byte, 0, 999)
	for len(line) < 999 {
		b, e := r.ReadByte()
		if e != nil {
			break
		}
		line = append(line, b)
		if b == '\n' {
			break
		}
	}
	header := string(line)
	status, frames, po, cexFrame := int32(-1), int32(-1), int32(-1), int32(-1)
	switch {
	case strings.HasPrefix(header, "snl_UNSAT"):
		status = 1
		frames = abcLogAtoi(header[len("snl_UNSAT"):])
	case strings.HasPrefix(header, "snl_SAT"):
		status = 0
		tokens := strings.FieldsFunc(header[len("snl_SAT"):], func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' })
		if len(tokens) == 0 {
			panic("abci: SAT log missing frame count")
		}
		frames = abcLogAtoi(tokens[0])
		if len(tokens) > 2 {
			po = abcLogAtoi(tokens[2])
			if len(tokens) > 3 {
				cexFrame = abcLogAtoi(tokens[3])
			}
		}
	case strings.HasPrefix(header, "snl_UNK"):
		frames = abcLogAtoi(header[len("snl_UNK"):])
	default:
		fmt.Println("Unrecognized status.")
	}
	bits := make([]byte, 0, 100)
	for {
		b, e := r.ReadByte()
		if e != nil {
			break
		}
		if b == '\n' {
			break
		}
		if b == '0' || b == '1' {
			bits = append(bits, b-'0')
		}
	}
	regs := len(bits)
	tail, _ := io.ReadAll(r)
	for _, b := range tail {
		if b == '0' || b == '1' {
			bits = append(bits, b-'0')
		}
	}
	frame := cexFrame
	if frame == -1 {
		frame = frames
	}
	var cex *util.Abc_Cex_t
	if len(bits) > 0 {
		trace := len(bits) - regs
		if trace == 0 {
			fmt.Println("Cannot read counter example.")
			return -1
		}
		if trace%int(frame+1) != 0 {
			fmt.Println("Incorrect number of bits.")
			return -1
		}
		cex = util.Abc_CexAlloc(int32(regs), int32(trace/int(frame+1)), frame+1)
		cex.IPo, cex.IFrame = po, frame
		for i, v := range bits {
			if v != 0 {
				util.Abc_InfoSetBit(cex.PData, int32(i))
			}
		}
	} else {
		cex = util.Abc_CexAlloc(0, 0, frame+1)
		cex.IPo, cex.IFrame = po, frame
	}
	if result != nil {
		*result = cex
	}
	if framesOut != nil {
		*framesOut = frames
	}
	return status
}
