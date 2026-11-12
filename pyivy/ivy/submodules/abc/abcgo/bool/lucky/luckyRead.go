package lucky

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func Abc_TruthGetBit(table []uint64, index int32) int32 {
	return int32(table[index>>6] >> uint(index&63) & 1)
}
func Abc_TruthSetBit(table []uint64, index int32) { table[index>>6] |= 1 << uint(index&63) }
func Abc_TruthXorBit(table []uint64, index int32) { table[index>>6] ^= 1 << uint(index&63) }
func Abc_TruthGetHex(table []uint64, index int32) int32 {
	return int32(table[index>>4] >> uint((index<<2)&63) & 15)
}
func Abc_TruthSetHex(table []uint64, index, value int32) {
	table[index>>4] |= uint64(value) << uint((index<<2)&63)
}
func Abc_TruthXorHex(table []uint64, index, value int32) {
	table[index>>4] ^= uint64(value) << uint((index<<2)&63)
}
func Abc_TruthReadHexDigit(c byte) int32 {
	switch {
	case c >= '0' && c <= '9':
		return int32(c - '0')
	case c >= 'a' && c <= 'f':
		return int32(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return int32(c - 'A' + 10)
	}
	return -1
}
func Abc_TruthWriteHexDigit(w io.Writer, digit int32) error {
	if digit < 0 || digit >= 16 {
		return fmt.Errorf("invalid hexadecimal digit %d", digit)
	}
	_, err := fmt.Fprintf(w, "%c", "0123456789ABCDEF"[digit])
	return err
}
func truthHexDigits(vars int32) int32 {
	Kit_TruthWordNum_64bit(vars)
	if vars < 2 {
		return 1
	}
	return 1 << uint(vars-2)
}
func Abc_TruthReadHex(table []uint64, text string, vars int32) error {
	fields := strings.Fields(strings.TrimRight(text, "\x00"))
	if len(fields) == 0 {
		return fmt.Errorf("missing truth table")
	}
	text = strings.TrimPrefix(strings.TrimPrefix(fields[0], "0x"), "0X")
	digits := truthHexDigits(vars)
	if int32(len(text)) != digits {
		return fmt.Errorf("truth table has %d hex digits, expected %d", len(text), digits)
	}
	clear(table[:Kit_TruthWordNum_64bit(vars)])
	for i := int32(0); i < digits; i++ {
		digit := Abc_TruthReadHexDigit(text[digits-1-i])
		if digit < 0 {
			return fmt.Errorf("invalid hexadecimal symbol %q", text[digits-1-i])
		}
		Abc_TruthSetHex(table, i, digit)
	}
	return nil
}
func Abc_TruthWriteHex(w io.Writer, table []uint64, vars int32) error {
	digits := truthHexDigits(vars)
	text := make([]byte, digits)
	for i := int32(0); i < digits; i++ {
		text[i] = "0123456789ABCDEF"[Abc_TruthGetHex(table, digits-1-i)]
	}
	n, err := w.Write(text)
	if err == nil && n != len(text) {
		return io.ErrShortWrite
	}
	return err
}
func Abc_TruthStoreAlloc(vars, count int32) *Abc_TtStore_t {
	words := Kit_TruthWordNum_64bit(vars)
	if count < 0 {
		panic("negative truth-table count")
	}
	p := &Abc_TtStore_t{NVars: vars, NWords: words, NFuncs: count, Funcs: make([][]uint64, count)}
	for i := range p.Funcs {
		p.Funcs[i] = make([]uint64, words)
	}
	return p
}
func Abc_TruthStoreFree(p *Abc_TtStore_t) {
	if p != nil {
		p.Funcs = nil
		p.NFuncs = 0
	}
}
func Abc_FileRead(name string) (string, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(data) + "\n\n\x00", nil
}
func readTruthLines(name string) ([]string, int32, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 1<<20)
	var lines []string
	var vars int32
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			break
		}
		fields := strings.Fields(text)
		digits := strings.TrimPrefix(strings.TrimPrefix(fields[0], "0x"), "0X")
		if len(lines) == 0 {
			for vars = 2; vars <= 16; vars++ {
				if len(digits) == int(truthHexDigits(vars)) {
					break
				}
			}
			if vars > 16 {
				return nil, 0, fmt.Errorf("invalid truth-table width %d", len(digits))
			}
		} else if len(digits) != int(truthHexDigits(vars)) {
			return nil, 0, fmt.Errorf("truth-table width differs on line %d", len(lines)+1)
		}
		lines = append(lines, text)
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}
	return lines, vars, nil
}
func Abc_TruthGetParams(name string, vars, count *int32) error {
	lines, n, err := readTruthLines(name)
	if vars != nil {
		*vars = n
	}
	if count != nil {
		*count = int32(len(lines))
	}
	return err
}
func Abc_Create_TtSpore(name string) (*Abc_TtStore_t, error) {
	lines, vars, err := readTruthLines(name)
	if err != nil {
		return nil, err
	}
	return Abc_TruthStoreAlloc(vars, int32(len(lines))), nil
}
func Abc_TruthStoreRead(name string, store *Abc_TtStore_t) error {
	lines, vars, err := readTruthLines(name)
	if err != nil {
		return err
	}
	if len(lines) > len(store.Funcs) || len(lines) > 0 && vars != store.NVars {
		return fmt.Errorf("truth store dimensions mismatch")
	}
	for i, line := range lines {
		if err := Abc_TruthReadHex(store.Funcs[i], line, vars); err != nil {
			return fmt.Errorf("line %d: %w", i+1, err)
		}
	}
	store.NFuncs = int32(len(lines))
	return nil
}
func writeTruthStore(name string, store *Abc_TtStore_t, prefixed bool) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	for _, table := range store.Funcs[:store.NFuncs] {
		if prefixed {
			if _, err = writer.WriteString("0x"); err != nil {
				file.Close()
				return err
			}
		}
		if err = Abc_TruthWriteHex(writer, table, store.NVars); err != nil {
			file.Close()
			return err
		}
		if err = writer.WriteByte('\n'); err != nil {
			file.Close()
			return err
		}
	}
	if prefixed {
		if err = writer.WriteByte('\n'); err != nil {
			file.Close()
			return err
		}
	}
	if err = writer.Flush(); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}
func Abc_TruthStoreWrite(name string, store *Abc_TtStore_t) error {
	return writeTruthStore(name, store, false)
}
func WriteToFile(name string, store *Abc_TtStore_t, data []uint64) error {
	tables := make([][]uint64, store.NFuncs)
	for i := range tables {
		start := i * int(store.NWords)
		tables[i] = data[start : start+int(store.NWords)]
	}
	copyStore := *store
	copyStore.Funcs = tables
	return Abc_TruthStoreWrite(name, &copyStore)
}
func WriteToFile1(name string, store *Abc_TtStore_t, data [][]uint64) error {
	copyStore := *store
	copyStore.Funcs = data
	return writeTruthStore(name, &copyStore, true)
}
func WriteToFile2(name string, store *Abc_TtStore_t, data []uint64) error {
	tables := make([][]uint64, store.NFuncs)
	for i := range tables {
		start := i * int(store.NWords)
		tables[i] = data[start : start+int(store.NWords)]
	}
	return WriteToFile1(name, store, tables)
}
func setTtStore(name string) (*Abc_TtStore_t, error) {
	store, err := Abc_Create_TtSpore(name)
	if err != nil {
		return nil, err
	}
	if err = Abc_TruthStoreRead(name, store); err != nil {
		return nil, err
	}
	return store, nil
}
func ReadTruthStore(name string) (*Abc_TtStore_t, error) { return setTtStore(name) }
