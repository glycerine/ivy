package abc

// Stand-alone replacement for src/opt/rwr/rwrTemp.c.

import (
	"slices"
	"strconv"
)

const (
	RewriteNPNVariableCount = 5
	RewriteNPNClassCount    = 13719
	RewriteNPNDefaultCount  = 100
)

// RewriteNPNClass is one line of nnclass_stats5.txt after parsing. Truth uses
// ABC's five-variable bit order: bit k is the value at assignment k.
type RewriteNPNClass struct {
	Truth     uint32
	Frequency int
}

type RewriteNPNLimits struct {
	MaxInputBytes, MaxClasses, MaxEntries, MaxOutputBytes int
	Stop                                                  func() bool
}

func normalizeRewriteNPNLimits(limits RewriteNPNLimits) (RewriteNPNLimits, error) {
	if limits.MaxInputBytes < 0 || limits.MaxClasses < 0 || limits.MaxEntries < 0 || limits.MaxOutputBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxInputBytes == 0 {
		limits.MaxInputBytes = 16 << 20
	}
	if limits.MaxClasses == 0 {
		limits.MaxClasses = RewriteNPNClassCount
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = RewriteNPNClassCount
	}
	if limits.MaxOutputBytes == 0 {
		limits.MaxOutputBytes = 64 << 20
	}
	return limits, nil
}

func rewriteNPNCancelled(limits RewriteNPNLimits) bool {
	return limits.Stop != nil && limits.Stop()
}

func rewriteNPNNextField(data []byte, position int) ([]byte, int) {
	for position < len(data) && (data[position] == ' ' || data[position] == '\t' || data[position] == '\r' || data[position] == '\n') {
		position++
	}
	begin := position
	for position < len(data) && data[position] != ' ' && data[position] != '\t' && data[position] != '\r' && data[position] != '\n' {
		position++
	}
	return data[begin:position], position
}

// ParseRewriteNPNClasses parses the two-column hexadecimal/frequency dataset
// consumed by Rwr_Temp. It accepts arbitrary ASCII whitespace but requires the
// complete 0xXXXXXXXX truth spelling used for five-variable functions.
func ParseRewriteNPNClasses(data []byte, limits RewriteNPNLimits) ([]RewriteNPNClass, error) {
	limits, err := normalizeRewriteNPNLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxInputBytes {
		return nil, ErrLimit
	}
	if rewriteNPNCancelled(limits) {
		return nil, ErrCancelled
	}
	classes := make([]RewriteNPNClass, 0, min(limits.MaxClasses, len(data)/13))
	position := 0
	for {
		hex, next := rewriteNPNNextField(data, position)
		position = next
		if len(hex) == 0 {
			break
		}
		frequencyText, next := rewriteNPNNextField(data, position)
		position = next
		if len(frequencyText) == 0 {
			return nil, ErrFormat
		}
		if len(classes) == limits.MaxClasses {
			return nil, ErrLimit
		}
		index := len(classes)
		if index&255 == 0 && rewriteNPNCancelled(limits) {
			return nil, ErrCancelled
		}
		if len(hex) != 10 || hex[0] != '0' || (hex[1] != 'x' && hex[1] != 'X') {
			return nil, ErrFormat
		}
		truth, parseErr := strconv.ParseUint(string(hex[2:]), 16, 32)
		if parseErr != nil {
			return nil, ErrFormat
		}
		frequency64, parseErr := strconv.ParseInt(string(frequencyText), 10, 32)
		if parseErr != nil {
			return nil, ErrFormat
		}
		classes = append(classes, RewriteNPNClass{Truth: uint32(truth), Frequency: int(frequency64)})
	}
	return classes, nil
}

func rewriteNPNSort(classes []RewriteNPNClass, limits RewriteNPNLimits) ([]RewriteNPNClass, error) {
	ordered := slices.Clone(classes)
	if len(ordered) < 2 {
		return ordered, nil
	}
	scratch := make([]RewriteNPNClass, len(ordered))
	source, destination := ordered, scratch
	comparisons := 0
	for width := 1; width < len(source); {
		if rewriteNPNCancelled(limits) {
			return nil, ErrCancelled
		}
		for begin := 0; begin < len(source); begin += 2 * width {
			middle := min(begin+width, len(source))
			end := min(begin+2*width, len(source))
			left, right, output := begin, middle, begin
			for left < middle && right < end {
				comparisons++
				if comparisons&1023 == 0 && rewriteNPNCancelled(limits) {
					return nil, ErrCancelled
				}
				// Select the left record on a tie to preserve deterministic
				// input order where the source qsort left the order unspecified.
				if source[left].Frequency >= source[right].Frequency {
					destination[output] = source[left]
					left++
				} else {
					destination[output] = source[right]
					right++
				}
				output++
			}
			output += copy(destination[output:end], source[left:middle])
			copy(destination[output:end], source[right:end])
		}
		source, destination = destination, source
		if width > len(source)/2 {
			width = len(source)
		} else {
			width *= 2
		}
	}
	if len(source) > 0 && &source[0] != &ordered[0] {
		copy(ordered, source)
	}
	return ordered, nil
}

func rewriteNPNAppend(output []byte, limit int, text string) ([]byte, error) {
	if len(text) > limit-len(output) {
		return nil, ErrLimit
	}
	return append(output, text...), nil
}

func rewriteNPNAppendIndex(output []byte, limit, index int) ([]byte, error) {
	if index < 10 {
		var err error
		output, err = rewriteNPNAppend(output, limit, "0")
		if err != nil {
			return nil, err
		}
	}
	digits := strconv.AppendInt(nil, int64(index), 10)
	if len(digits) > limit-len(output) {
		return nil, ErrLimit
	}
	return append(output, digits...), nil
}

// FormatRewriteNPNBLIF sorts by decreasing frequency and returns the BLIF that
// Rwr_Temp wrote to 5npn_100.blif. Equal-frequency classes retain input order;
// the C qsort did not specify a tie order. Input records are never modified.
func FormatRewriteNPNBLIF(classes []RewriteNPNClass, entries int, limits RewriteNPNLimits) ([]byte, error) {
	limits, err := normalizeRewriteNPNLimits(limits)
	if err != nil {
		return nil, err
	}
	if entries < 0 || entries > len(classes) {
		return nil, ErrDimension
	}
	if len(classes) > limits.MaxClasses || entries > limits.MaxEntries {
		return nil, ErrLimit
	}
	for _, class := range classes {
		if class.Frequency < -1<<31 || class.Frequency > 1<<31-1 {
			return nil, ErrDimension
		}
	}
	if rewriteNPNCancelled(limits) {
		return nil, ErrCancelled
	}
	ordered, err := rewriteNPNSort(classes, limits)
	if err != nil {
		return nil, err
	}
	output := make([]byte, 0, min(limits.MaxOutputBytes, 4096))
	for _, text := range []string{
		"# Most frequent NPN classes of 5 vars.\n",
		".model 5npn\n",
		".inputs a b c d e\n",
		".outputs",
	} {
		output, err = rewriteNPNAppend(output, limits.MaxOutputBytes, text)
		if err != nil {
			return nil, err
		}
	}
	for index := 0; index < entries; index++ {
		if index&255 == 0 && rewriteNPNCancelled(limits) {
			return nil, ErrCancelled
		}
		output, err = rewriteNPNAppend(output, limits.MaxOutputBytes, " ")
		if err == nil {
			output, err = rewriteNPNAppendIndex(output, limits.MaxOutputBytes, index)
		}
		if err != nil {
			return nil, err
		}
	}
	output, err = rewriteNPNAppend(output, limits.MaxOutputBytes, "\n")
	if err != nil {
		return nil, err
	}
	for index := 0; index < entries; index++ {
		if rewriteNPNCancelled(limits) {
			return nil, ErrCancelled
		}
		output, err = rewriteNPNAppend(output, limits.MaxOutputBytes, ".names a b c d e ")
		if err == nil {
			output, err = rewriteNPNAppendIndex(output, limits.MaxOutputBytes, index)
		}
		if err == nil {
			output, err = rewriteNPNAppend(output, limits.MaxOutputBytes, "\n")
		}
		if err != nil {
			return nil, err
		}
		truth := ordered[index].Truth
		for assignment := 0; assignment < 32; assignment++ {
			if truth&(uint32(1)<<uint(assignment)) == 0 {
				continue
			}
			line := [7]byte{'0', '0', '0', '0', '0', ' ', '1'}
			for variable := 0; variable < RewriteNPNVariableCount; variable++ {
				line[RewriteNPNVariableCount-1-variable] += byte((assignment >> uint(variable)) & 1)
			}
			if len(line)+1 > limits.MaxOutputBytes-len(output) {
				return nil, ErrLimit
			}
			output = append(output, line[:]...)
			output = append(output, '\n')
		}
	}
	output, err = rewriteNPNAppend(output, limits.MaxOutputBytes, ".end\n")
	if err != nil {
		return nil, err
	}
	return output, nil
}

// BuildRewriteNPNBLIF composes the source routine's parsing, frequency sort,
// and BLIF generation without fixed filenames or process-global arrays.
func BuildRewriteNPNBLIF(stats []byte, entries int, limits RewriteNPNLimits) ([]byte, error) {
	classes, err := ParseRewriteNPNClasses(stats, limits)
	if err != nil {
		return nil, err
	}
	return FormatRewriteNPNBLIF(classes, entries, limits)
}

// BuildDefaultRewriteNPNBLIF preserves Rwr_Temp's exact 13,719-record input
// contract and selection of the 100 most frequent classes.
func BuildDefaultRewriteNPNBLIF(stats []byte, limits RewriteNPNLimits) ([]byte, error) {
	classes, err := ParseRewriteNPNClasses(stats, limits)
	if err != nil {
		return nil, err
	}
	if len(classes) != RewriteNPNClassCount {
		return nil, ErrFormat
	}
	return FormatRewriteNPNBLIF(classes, RewriteNPNDefaultCount, limits)
}
