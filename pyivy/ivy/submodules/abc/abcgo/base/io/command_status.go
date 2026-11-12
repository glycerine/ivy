package io

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	abcmain "github.com/glycerine/abcgo/base/main"
)

func readVerificationLog(r io.Reader) (int32, int32, *abcmain.Counterexample, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 16<<20)
	if !scanner.Scan() {
		return -1, -1, nil, fmt.Errorf("empty verification log")
	}
	fields := strings.Fields(scanner.Text())
	if len(fields) < 2 {
		return -1, -1, nil, fmt.Errorf("invalid verification header")
	}
	status := int32(-1)
	switch fields[0] {
	case "snl_UNSAT":
		status = 1
	case "snl_SAT":
		status = 0
	case "snl_UNK":
	default:
		return -1, -1, nil, fmt.Errorf("unrecognized verification status")
	}
	depth, err := strconv.ParseInt(fields[1], 10, 32)
	if err != nil {
		return -1, -1, nil, fmt.Errorf("invalid verification depth: %w", err)
	}
	output, frame := int64(-1), depth
	if status == 0 && len(fields) > 3 {
		output, err = strconv.ParseInt(fields[3], 10, 32)
		if err != nil {
			return -1, -1, nil, err
		}
	}
	if status == 0 && len(fields) > 4 {
		frame, err = strconv.ParseInt(fields[4], 10, 32)
		if err != nil {
			return -1, -1, nil, err
		}
	}
	var bits []byte
	regs := 0
	for line := 0; scanner.Scan(); line++ {
		for _, c := range []byte(scanner.Text()) {
			if c == '0' || c == '1' {
				bits = append(bits, c-'0')
			}
		}
		if line == 0 {
			regs = len(bits)
		}
	}
	if err := scanner.Err(); err != nil {
		return -1, -1, nil, err
	}
	if frame < 0 {
		if len(bits) == 0 {
			return status, int32(depth), nil, nil
		}
		return -1, -1, nil, fmt.Errorf("invalid counterexample frame")
	}
	trace := len(bits) - regs
	if len(bits) != 0 && trace == 0 {
		return -1, -1, nil, fmt.Errorf("cannot read counterexample trace")
	}
	if trace != 0 && trace%int(frame+1) != 0 {
		return -1, -1, nil, fmt.Errorf("incorrect number of trace bits")
	}
	cex := &abcmain.Counterexample{PIs: int32(trace / int(frame+1)), Registers: int32(regs), Output: int32(output), Frame: int32(frame), Bits: make([]uint64, (len(bits)+63)/64)}
	for i, bit := range bits {
		if bit != 0 {
			cex.Bits[i/64] |= uint64(1) << uint(i%64)
		}
	}
	return status, int32(depth), cex, nil
}

func writeVerificationLog(w io.Writer, status, depth int32, cex *abcmain.Counterexample) error {
	label := map[int32]string{1: "snl_UNSAT", 0: "snl_SAT", -1: "snl_UNK"}[status]
	if label == "" {
		return fmt.Errorf("unrecognized verification status")
	}
	if _, err := fmt.Fprintf(w, "%s %d unknown", label, depth); err != nil {
		return err
	}
	if cex != nil && status == 0 {
		if _, err := fmt.Fprintf(w, " %d", cex.Output); err != nil {
			return err
		}
	}
	if cex != nil && cex.Frame != depth {
		if _, err := fmt.Fprintf(w, " %d", cex.Frame); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if cex == nil {
		_, err := fmt.Fprint(w, "NULL\nNULL\n")
		return err
	}
	if cex.Registers < 0 || cex.PIs < 0 || cex.Frame < 0 {
		return fmt.Errorf("invalid counterexample dimensions")
	}
	total := int64(cex.Registers) + int64(cex.PIs)*(int64(cex.Frame)+1)
	if total > int64(len(cex.Bits))*64 {
		return fmt.Errorf("short counterexample bit vector")
	}
	for i := int64(0); i < total; i++ {
		if i == int64(cex.Registers) {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "%d", cex.Bits[i/64]>>uint(i%64)&1); err != nil {
			return err
		}
	}
	if total == int64(cex.Registers) {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

func IoCommandReadStatus(frame *abcmain.Frame, args []string) int32 {
	if len(args) != 2 || strings.HasPrefix(args[1], "-") {
		fmt.Fprintln(frame.Err, "usage: read_status [-h] <file>")
		return 1
	}
	f, err := os.Open(args[1])
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	defer f.Close()
	status, depth, cex, err := readVerificationLog(f)
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 1
	}
	frame.SetVerificationResult(status, depth, cex)
	return 0
}

func IoCommandWriteStatus(frame *abcmain.Frame, args []string) int32 {
	if len(args) != 2 || strings.HasPrefix(args[1], "-") {
		fmt.Fprintln(frame.Err, "usage: write_status [-h] <file>")
		return 1
	}
	f, err := os.Create(args[1])
	if err != nil {
		fmt.Fprintln(frame.Err, err)
		return 0
	}
	status, depth, cex := frame.VerificationResult()
	err = writeVerificationLog(f, status, depth, cex)
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		fmt.Fprintln(frame.Err, err)
	}
	return 0
}
