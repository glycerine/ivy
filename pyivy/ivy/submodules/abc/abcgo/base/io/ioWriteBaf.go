package io

import (
	"bytes"
	"encoding/binary"
	"fmt"
	stdio "io"
	"os"
	"strings"

	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/base/abc"
)

func WriteBAF(w stdio.Writer, n *abc.Network) error {
	model, err := aigerModelFromNetwork(n)
	if err != nil {
		return err
	}
	if len(model.Bad)+len(model.Constraints)+len(model.Justice)+len(model.Fairness) > 0 {
		return fmt.Errorf("BAF cannot represent AIGER verification properties")
	}
	model, err = model.Normalize()
	if err != nil {
		return err
	}
	var out bytes.Buffer
	field := func(value string) error {
		if strings.IndexByte(value, 0) >= 0 {
			return fmt.Errorf("BAF name contains NUL")
		}
		out.WriteString(value)
		out.WriteByte(0)
		return nil
	}
	if err = field(n.Name); err != nil {
		return err
	}
	for _, count := range []int{len(model.Inputs), len(model.Outputs), len(model.Latches), len(model.ANDs)} {
		fmt.Fprintf(&out, "%d%c", count, 0)
	}
	for _, group := range []struct {
		kind  byte
		count int
	}{{'i', len(model.Inputs)}, {'o', len(model.Outputs)}} {
		for i := 0; i < group.count; i++ {
			name := model.Symbols[ioa.Symbol{Kind: group.kind, Index: uint32(i)}]
			if name == "" {
				name = fmt.Sprintf("%c%d", group.kind, i)
			}
			if err = field(name); err != nil {
				return err
			}
		}
	}
	latches := n.ObjectsOfKind(abc.ObjectLatch)
	if len(latches) != len(model.Latches) {
		return fmt.Errorf("BAF register metadata mismatch")
	}
	for _, latch := range latches {
		names := [3]string{blifObjectName(latch), blifObjectName(latch.Fanins()[0].Object), blifObjectName(latch.Fanouts()[0])}
		for _, name := range names {
			if err = field(name); err != nil {
				return err
			}
		}
	}
	encode := func(lit uint32) uint32 {
		if lit < 2 {
			return lit ^ 1
		}
		return lit
	}
	write := func(word uint32) { var data [4]byte; binary.LittleEndian.PutUint32(data[:], word); out.Write(data[:]) }
	for _, gate := range model.ANDs {
		write(encode(gate.Left))
		write(encode(gate.Right))
	}
	for _, literal := range model.Outputs {
		write(encode(literal))
	}
	for i, latch := range model.Latches {
		next := encode(latch.Next)
		if next > 0x3fffffff {
			return fmt.Errorf("BAF latch literal exceeds packed 30-bit range")
		}
		init := abc.ABC_INIT_DC
		if latch.Init == 0 {
			init = abc.ABC_INIT_ZERO
		} else if latch.Init == 1 {
			init = abc.ABC_INIT_ONE
		}
		if abc.Abc_LatchIsInitNone(latches[i]) != 0 {
			init = abc.ABC_INIT_NONE
		}
		write(next<<2 | uint32(init))
	}
	count, err := w.Write(out.Bytes())
	if err == nil && count != out.Len() {
		return stdio.ErrShortWrite
	}
	return err
}
func WriteBAFFile(n *abc.Network, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = WriteBAF(file, n)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func Io_WriteBaf(n *abc.Network, filename string) error { return WriteBAFFile(n, filename) }
func init()                                             { RegisterFormat(IO_FILE_BAF, nil, WriteBAFFile) }
