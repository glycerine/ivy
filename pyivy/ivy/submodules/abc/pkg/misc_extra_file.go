package abc

// File, name, binary, hexadecimal, and string utilities from src/misc/extra/extraUtilFile.c.

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type ExtraFileUtilityLimits struct {
	MaxBytes, MaxLines, MaxOutputBytes int
	Stop                               func() bool
}

func normalizeExtraFileUtilityLimits(limits ExtraFileUtilityLimits) (ExtraFileUtilityLimits, error) {
	if limits.MaxBytes < 0 || limits.MaxLines < 0 || limits.MaxOutputBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 64 << 20
	}
	if limits.MaxLines == 0 {
		limits.MaxLines = 1 << 20
	}
	if limits.MaxOutputBytes == 0 {
		limits.MaxOutputBytes = 64 << 20
	}
	return limits, nil
}

func ExtraFileExtension(name string) string {
	if dot := strings.LastIndexByte(name, '.'); dot >= 0 {
		return name[dot+1:]
	}
	return name
}

func ExtraFileNameAppend(base, suffix string) string { return base + suffix }

func ExtraFileNameGeneric(name string) string {
	if dot := strings.LastIndexByte(name, '.'); dot >= 0 {
		return name[:dot]
	}
	return strings.Clone(name)
}

func ExtraFileNameGenericAppend(base, suffix string) string {
	return ExtraFileNameGeneric(base) + suffix
}

func CorrectExtraFilePath(name string) string {
	return strings.NewReplacer(">", "/", "\\", "/").Replace(name)
}

func ExtraFileNameWithoutPath(name string) string {
	if slash := strings.LastIndexAny(name, "/\\"); slash >= 0 {
		return name[slash+1:]
	}
	return name
}

func ExtraFilePathWithoutName(name string) (string, bool) {
	if slash := strings.LastIndexAny(name, "/\\"); slash >= 0 {
		return CorrectExtraFilePath(name[:slash]), true
	}
	return "", false
}

func ExtraFileDesignName(name string) string {
	end := strings.IndexByte(name, '.')
	if end < 0 {
		end = len(name)
	}
	begin := end
	for begin > 0 {
		character := name[begin-1]
		if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_') {
			break
		}
		begin--
	}
	return strings.Clone(name[begin:end])
}

func ExtraSimilarFileName(wrong string, suffixes ...string) (string, bool, error) {
	if len(suffixes) == 0 {
		return "", false, nil
	}
	base := ExtraFileNameGeneric(wrong)
	for _, suffix := range suffixes {
		candidate := base + suffix
		file, e := os.Open(candidate)
		if e != nil {
			if os.IsNotExist(e) {
				continue
			}
			return "", false, e
		}
		if e = file.Close(); e != nil {
			return "", false, e
		}
		return candidate, true, nil
	}
	return "", false, nil
}

type ExtraFileInspection struct {
	Size  int64
	Empty bool
}

func InspectExtraFile(name string) (ExtraFileInspection, error) {
	file, e := os.Open(name)
	if e != nil {
		return ExtraFileInspection{}, e
	}
	info, e := file.Stat()
	closeErr := file.Close()
	if e != nil {
		return ExtraFileInspection{}, e
	}
	if closeErr != nil {
		return ExtraFileInspection{}, closeErr
	}
	return ExtraFileInspection{Size: info.Size(), Empty: info.Size() == 0}, nil
}

func ExtraFileSize(name string) (int64, error) {
	inspection, e := InspectExtraFile(name)
	return inspection.Size, e
}

func ReadExtraFile(file *os.File, limits ExtraFileUtilityLimits) ([]byte, error) {
	limits, e := normalizeExtraFileUtilityLimits(limits)
	if e != nil {
		return nil, e
	}
	if file == nil {
		return nil, ErrDimension
	}
	info, e := file.Stat()
	if e != nil {
		return nil, e
	}
	if info.Size() < 0 || info.Size() > int64(limits.MaxBytes) {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	if _, e = file.Seek(0, 0); e != nil {
		return nil, e
	}
	data := make([]byte, int(info.Size()))
	offset := 0
	for offset < len(data) {
		count, readErr := file.Read(data[offset:])
		offset += count
		if readErr != nil {
			return nil, readErr
		}
		if count == 0 {
			return nil, ErrFormat
		}
	}
	after, e := file.Stat()
	if e != nil {
		return nil, e
	}
	if after.Size() != info.Size() {
		return nil, ErrInvariant
	}
	if len(data)+1 > limits.MaxOutputBytes {
		return nil, ErrLimit
	}
	return append(data, '\n'), nil
}

func ReadExtraTwoFiles(first, second *os.File, limits ExtraFileUtilityLimits) ([]byte, error) {
	limits, e := normalizeExtraFileUtilityLimits(limits)
	if e != nil {
		return nil, e
	}
	if first == nil || second == nil {
		return nil, ErrDimension
	}
	firstInfo, e := first.Stat()
	if e != nil {
		return nil, e
	}
	secondInfo, e := second.Stat()
	if e != nil {
		return nil, e
	}
	if firstInfo.Size() < 0 || secondInfo.Size() < 0 || firstInfo.Size() > int64(limits.MaxBytes) || secondInfo.Size() > int64(limits.MaxBytes)-firstInfo.Size() {
		return nil, ErrLimit
	}
	combinedLimit := limits
	combinedLimit.MaxOutputBytes--
	firstData, e := ReadExtraFile(first, combinedLimit)
	if e != nil {
		return nil, e
	}
	secondData, e := ReadExtraFile(second, combinedLimit)
	if e != nil {
		return nil, e
	}
	firstData = firstData[:len(firstData)-1]
	secondData = secondData[:len(secondData)-1]
	if len(firstData)+len(secondData)+1 > limits.MaxOutputBytes {
		return nil, ErrLimit
	}
	result := make([]byte, 0, len(firstData)+len(secondData)+1)
	result = append(result, firstData...)
	result = append(result, secondData...)
	return append(result, '\n'), nil
}

func ReadExtraFileContents(name string, limits ExtraFileUtilityLimits) ([]byte, error) {
	file, e := os.Open(name)
	if e != nil {
		return nil, e
	}
	data, readErr := ReadExtraFile(file, limits)
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return data, nil
}

func ReadExtraFileContents2(firstName, secondName string, limits ExtraFileUtilityLimits) ([]byte, error) {
	first, e := os.Open(firstName)
	if e != nil {
		return nil, e
	}
	defer first.Close()
	second, e := os.Open(secondName)
	if e != nil {
		return nil, e
	}
	data, readErr := ReadExtraTwoFiles(first, second, limits)
	closeErr := second.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return data, nil
}

func ExtraFileIsType(name string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if suffix != "" && len(name) > len(suffix) && strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

func FormatExtraTimeStamp(value time.Time) string {
	return value.Local().Format("Mon Jan _2 15:04:05 2006")
}

func ExtraTimeStamp() string { return FormatExtraTimeStamp(time.Now()) }

func ReadExtraBinary(text string) (uint32, error) {
	var result uint32
	for _, character := range text {
		if character != '0' && character != '1' {
			return 0, ErrFormat
		}
		result = result*2 + uint32(character-'0')
	}
	return result, nil
}

func FormatExtraBinary(words []uint32, bitCount int, lowToHigh bool) (string, error) {
	if bitCount < 0 || len(words) < (bitCount+31)/32 {
		return "", ErrDimension
	}
	result := make([]byte, bitCount)
	for position := 0; position < bitCount; position++ {
		bit := position
		if !lowToHigh {
			bit = bitCount - 1 - position
		}
		result[position] = '0' + byte(words[bit/32]>>uint(bit&31)&1)
	}
	return string(result), nil
}

func ReadExtraHex(text string, digits, maxWords int) ([]uint32, error) {
	if digits < 0 || maxWords < 0 || len(text) < digits {
		return nil, ErrDimension
	}
	wordCount := (digits + 7) / 8
	if maxWords == 0 {
		maxWords = 1 << 20
	}
	if wordCount > maxWords {
		return nil, ErrLimit
	}
	result := make([]uint32, wordCount)
	for index := 0; index < digits; index++ {
		character := text[digits-1-index]
		var digit byte
		switch {
		case character >= '0' && character <= '9':
			digit = character - '0'
		case character >= 'A' && character <= 'F':
			digit = character - 'A' + 10
		case character >= 'a' && character <= 'f':
			digit = character - 'a' + 10
		default:
			return nil, ErrFormat
		}
		result[index/8] |= uint32(digit) << uint((index%8)*4)
	}
	return result, nil
}

func extraHexDigitCount(variables int, atLeastOne bool) (int, error) {
	if variables < 0 {
		return 0, ErrDimension
	}
	if variables > 30 {
		return 0, ErrLimit
	}
	digits := (1 << uint(variables)) / 4
	if atLeastOne && digits == 0 {
		digits = 1
	}
	return digits, nil
}

func ReadExtraHexadecimal(text string, variables, maxWords int) ([]uint32, error) {
	digits, e := extraHexDigitCount(variables, true)
	if e != nil {
		return nil, e
	}
	return ReadExtraHex(text, digits, maxWords)
}

func formatExtraHexDigits(words []uint32, digits int, upper, reverse bool) (string, error) {
	if digits < 0 || len(words) < (digits+7)/8 {
		return "", ErrDimension
	}
	result := make([]byte, digits)
	for position := 0; position < digits; position++ {
		digitIndex := digits - 1 - position
		if reverse {
			digitIndex = position
		}
		digit := byte(words[digitIndex/8] >> uint((digitIndex%8)*4) & 15)
		base := byte('a')
		if upper {
			base = 'A'
		}
		if digit < 10 {
			result[position] = '0' + digit
		} else {
			result[position] = base + digit - 10
		}
	}
	return string(result), nil
}

func FormatExtraHexadecimal(words []uint32, variables int) (string, error) {
	digits, e := extraHexDigitCount(variables, false)
	if e != nil {
		return "", e
	}
	return formatExtraHexDigits(words, digits, false, false)
}

func FormatExtraHexadecimalString(words []uint32, variables int) (string, error) {
	if variables < 0 || len(words) == 0 {
		return "", ErrDimension
	}
	if variables == 0 {
		if words[0]&1 == 0 {
			return "0", nil
		}
		return "1", nil
	}
	if variables == 1 {
		if words[0]&1 != 0 {
			return "1", nil
		}
		return "2", nil
	}
	return FormatExtraHexadecimal(words, variables)
}

func FormatExtraHex(words []uint32, variables int, prefix, reverse bool) (string, error) {
	if variables < 0 || variables > 5 {
		return "", ErrDimension
	}
	minterms := 1 << uint(variables)
	digits := (minterms + 3) / 4
	text, e := formatExtraHexDigits(words, digits, true, reverse)
	if e != nil {
		return "", e
	}
	if prefix {
		text = "0x" + text
	}
	return text, nil
}

func ExtraPrintSymbols(character byte, count int, newline bool) (string, error) {
	if count < 0 {
		return "", ErrDimension
	}
	result := strings.Repeat(string(character), count)
	if newline {
		result += "\n"
	}
	return result, nil
}

func ExtraStringAppend(given, addition string) string { return given + addition }

func ExtraStringClean(given, keep string) string {
	var builder strings.Builder
	for _, character := range given {
		if strings.ContainsRune(keep, character) {
			builder.WriteRune(character)
		}
	}
	return builder.String()
}

func SortExtraFileLines(data []byte, limits ExtraFileUtilityLimits) ([]byte, error) {
	limits, e := normalizeExtraFileUtilityLimits(limits)
	if e != nil {
		return nil, e
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	if strings.IndexByte(string(data), 0) >= 0 {
		return nil, ErrFormat
	}
	text := string(data) + "\n"
	lines := strings.Split(text, "\n")
	lines = lines[:len(lines)-1]
	if len(lines) > limits.MaxLines {
		return nil, ErrLimit
	}
	kept := lines[:0]
	for _, line := range lines {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		if line != "" {
			kept = append(kept, line)
		}
	}
	sort.Strings(kept)
	var builder strings.Builder
	for _, line := range kept {
		if builder.Len()+len(line)+1 > limits.MaxOutputBytes {
			return nil, ErrLimit
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	return []byte(builder.String()), nil
}

func SortExtraFile(inputName, outputName string, limits ExtraFileUtilityLimits) error {
	data, e := os.ReadFile(inputName)
	if e != nil {
		return e
	}
	result, e := SortExtraFileLines(data, limits)
	if e != nil {
		return e
	}
	return os.WriteFile(outputName, result, 0o666)
}

func AddExtraLineNumbers(data []byte, limits ExtraFileUtilityLimits) ([]byte, error) {
	limits, e := normalizeExtraFileUtilityLimits(limits)
	if e != nil {
		return nil, e
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	if strings.IndexByte(string(data), 0) >= 0 {
		return nil, ErrFormat
	}
	result := make([]byte, 0, min(limits.MaxOutputBytes, len(data)+64))
	for offset, line := 0, 0; offset < len(data); line++ {
		if line >= limits.MaxLines {
			return nil, ErrLimit
		}
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		end := min(offset+999, len(data))
		if newline := strings.IndexByte(string(data[offset:end]), '\n'); newline >= 0 {
			end = offset + newline + 1
		}
		chunk := data[offset:end]
		if len(chunk) < 2 {
			return nil, ErrFormat
		}
		numbered := fmt.Sprintf("%03d\n", line)
		if len(result)+len(chunk)-2+len(numbered) > limits.MaxOutputBytes {
			return nil, ErrLimit
		}
		result = append(result, chunk[:len(chunk)-2]...)
		result = append(result, numbered...)
		offset = end
	}
	return result, nil
}

func AddExtraFileLineNumbers(inputName, outputName string, limits ExtraFileUtilityLimits) error {
	data, e := os.ReadFile(inputName)
	if e != nil {
		return e
	}
	result, e := AddExtraLineNumbers(data, limits)
	if e != nil {
		return e
	}
	return os.WriteFile(outputName, result, 0o666)
}
