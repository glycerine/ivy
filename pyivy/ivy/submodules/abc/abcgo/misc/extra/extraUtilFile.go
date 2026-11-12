package extra

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

/**CFile****************************************************************

  FileName    [extraUtilFile.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [File management utilities.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilFile.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

// File names are owned immutable Go strings. Buffer-mutating C helpers return
// their new string instead of relying on shared static or caller-owned buffers.
func Extra_FileNameWithoutPath(name string) string {
	if i := strings.LastIndexAny(name, "/\\"); i >= 0 {
		return name[i+1:]
	}
	return name
}
func Extra_FileNameExtension(name string) string {
	base := Extra_FileNameWithoutPath(name)
	if i := strings.LastIndexByte(base, '.'); i >= 0 {
		return base[i+1:]
	}
	return base
}
func Extra_FileNameAppend(base, suffix string) string { return base + suffix }
func Extra_FileNameGeneric(name string) string {
	slash := strings.LastIndexAny(name, "/\\")
	if dot := strings.LastIndexByte(name, '.'); dot > slash {
		return name[:dot]
	}
	return name
}
func Extra_FileNameGenericAppend(base, suffix string) string {
	return Extra_FileNameGeneric(base) + suffix
}
func Extra_FileNameCorrectPath(name string) string {
	return strings.NewReplacer("\\", "/", ">", "/").Replace(name)
}
func Extra_FilePathWithoutName(name string) string {
	if i := strings.LastIndexAny(name, "/\\"); i >= 0 {
		return Extra_FileNameCorrectPath(name[:i])
	}
	return ""
}
func Extra_FileDesignName(name string) string {
	end := strings.IndexByte(name, '.')
	if end < 0 {
		end = len(name)
	}
	start := end
	for start > 0 {
		ch := name[start-1]
		if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '_') {
			break
		}
		start--
	}
	return name[start:end]
}
func Extra_FileGetSimilarName(wrong, s1, s2, s3, s4, s5 string) string {
	base := Extra_FileNameGeneric(wrong)
	for _, suffix := range []string{s1, s2, s3, s4, s5} {
		if suffix == "" {
			break
		}
		candidate := base + suffix
		if file, err := os.Open(candidate); err == nil {
			file.Close()
			return candidate
		}
	}
	return ""
}
func Extra_FileCheck(name string) int32 {
	info, err := os.Stat(name)
	if err != nil || info.IsDir() {
		return 0
	}
	file, err := os.Open(name)
	if err != nil {
		return 0
	}
	file.Close()
	return 1
}
func Extra_FileSize(name string) int64 {
	info, err := os.Stat(name)
	if err != nil || info.IsDir() {
		return 0
	}
	return info.Size()
}
func Extra_FileRead(file *os.File) (string, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}
func Extra_FileRead2(first, second *os.File) (string, error) {
	a, err := Extra_FileRead(first)
	if err != nil {
		return "", err
	}
	b, err := Extra_FileRead(second)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(a, "\n") + b, nil
}
func Extra_FileReadContents(name string) (string, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}
func Extra_FileReadContents2(first, second string) (string, error) {
	a, err := os.ReadFile(first)
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(second)
	if err != nil {
		return "", err
	}
	return string(a) + string(b) + "\n", nil
}
func Extra_FileIsType(name, s1, s2, s3 string) int32 {
	for _, suffix := range []string{s1, s2, s3} {
		if suffix != "" && len(name) > len(suffix) && strings.HasSuffix(name, suffix) {
			return 1
		}
	}
	return 0
}
func Extra_TimeStamp() string { return time.Now().Format("Mon Jan _2 15:04:05 2006") }
func Extra_ReadBinary(text string) uint32 {
	var n uint32
	for i := 0; i < len(text); i++ {
		if text[i] != '0' && text[i] != '1' {
			panic("invalid binary digit")
		}
		n = n*2 + uint32(text[i]-'0')
	}
	return n
}
func filePrintBinary(out io.Writer, words []uint32, nBits int32, reverse bool) {
	if nBits < 0 || int64(nBits) > int64(len(words))*32 {
		panic("binary width outside word storage")
	}
	var text strings.Builder
	for pos := int32(0); pos < nBits; pos++ {
		bit := nBits - 1 - pos
		if reverse {
			bit = pos
		}
		text.WriteByte('0' + byte(words[bit/32]>>uint(bit%32)&1))
	}
	fmt.Fprint(out, text.String())
}
func Extra_PrintBinary(out io.Writer, words []uint32, nBits int32) {
	filePrintBinary(out, words, nBits, false)
}
func Extra_PrintBinary2(out io.Writer, words []uint32, nBits int32) {
	filePrintBinary(out, words, nBits, true)
}
func fileHexDigit(ch byte) (uint32, bool) {
	switch {
	case ch >= '0' && ch <= '9':
		return uint32(ch - '0'), true
	case ch >= 'a' && ch <= 'f':
		return uint32(ch - 'a' + 10), true
	case ch >= 'A' && ch <= 'F':
		return uint32(ch - 'A' + 10), true
	}
	return 0, false
}

// Extra_ReadHex ORs decoded digits into the caller's words, matching ABC.
func Extra_ReadHex(words []uint32, text string, nDigits int32) int32 {
	if nDigits < 0 || int64(nDigits) > int64(len(words))*8 || int(nDigits) > len(text) {
		return 0
	}
	// Validate before modifying output, so failure never leaves a partial value.
	for i := int32(0); i < nDigits; i++ {
		if _, ok := fileHexDigit(text[i]); !ok {
			return 0
		}
	}
	for k := int32(0); k < nDigits; k++ {
		digit, _ := fileHexDigit(text[nDigits-1-k])
		words[k/8] |= digit << uint(k%8*4)
	}
	return 1
}
func fileTruthBits(nVars int32) int {
	if nVars < 0 || nVars > 30 {
		panic("truth variable count outside supported range")
	}
	return 1 << uint(nVars)
}
func Extra_ReadHexadecimal(words []uint32, text string, nVars int32) int32 {
	count := (fileTruthBits(nVars) + 31) / 32
	if len(words) < count {
		return 0
	}
	clear(words[:count])
	digits := (fileTruthBits(nVars) + 3) / 4
	return Extra_ReadHex(words, text, int32(digits))
}
func fileHexString(words []uint32, digits int, upper, reverse bool) string {
	if digits < 0 || int64(digits) > int64(len(words))*8 {
		panic("hex width outside word storage")
	}
	alphabet := "0123456789abcdef"
	if upper {
		alphabet = "0123456789ABCDEF"
	}
	text := make([]byte, digits)
	for i := range text {
		k := digits - 1 - i
		if reverse {
			k = i
		}
		text[i] = alphabet[words[k/8]>>uint(k%8*4)&15]
	}
	return string(text)
}
func Extra_PrintHexadecimal(out io.Writer, words []uint32, nVars int32) {
	fmt.Fprint(out, fileHexString(words, fileTruthBits(nVars)/4, false, false))
}
func Extra_PrintHexadecimalString(words []uint32, nVars int32) string {
	if len(words) == 0 {
		panic("missing truth words")
	}
	if nVars == 0 {
		if words[0]&1 != 0 {
			return "1"
		}
		return "0"
	}
	if nVars == 1 {
		if words[0]&1 != 0 {
			return "1"
		}
		return "2"
	}
	return fileHexString(words, fileTruthBits(nVars)/4, false, false)
}
func Extra_PrintHex(out io.Writer, words []uint32, nVars int32) {
	fmt.Fprint(out, "0x", fileHexString(words, (fileTruthBits(nVars)+3)/4, true, false))
}
func Extra_PrintHex2(out io.Writer, words []uint32, nVars int32) {
	fmt.Fprint(out, fileHexString(words, (fileTruthBits(nVars)+3)/4, true, false))
}
func Extra_PrintHexReverse(out io.Writer, words []uint32, nVars int32) {
	fmt.Fprint(out, "0x", fileHexString(words, (fileTruthBits(nVars)+3)/4, true, true))
}
func Extra_PrintSymbols(out io.Writer, ch byte, nTimes int32, newline int32) {
	if nTimes < 0 {
		panic("negative symbol count")
	}
	fmt.Fprint(out, strings.Repeat(string([]byte{ch}), int(nTimes)))
	if newline != 0 {
		fmt.Fprintln(out)
	}
}
func Extra_StringAppend(given, added string) string { return given + added }
func Extra_StringClean(given, keep string) string {
	var allowed [256]bool
	for i := 0; i < len(keep); i++ {
		allowed[keep[i]] = true
	}
	var out strings.Builder
	for i := 0; i < len(given); i++ {
		if allowed[given[i]] {
			out.WriteByte(given[i])
		}
	}
	return out.String()
}
func Extra_StringCompare(first, second string) int32 { return int32(strings.Compare(first, second)) }
func Extra_FileSort(input, output string) error {
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	sort.Strings(lines)
	var result strings.Builder
	for _, line := range lines {
		if line != "" {
			result.WriteString(line)
			result.WriteByte('\n')
		}
	}
	return os.WriteFile(output, []byte(result.String()), 0666)
}

// Read before writing so input and output may name the same file safely.
func Extra_FileLineNumAdd(input, output string) error {
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	var result strings.Builder
	if len(data) > 0 {
		lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
		for i, line := range lines {
			fmt.Fprintf(&result, "%s%03d\n", strings.TrimSuffix(line, "\r"), i)
		}
	}
	return os.WriteFile(output, []byte(result.String()), 0666)
}
