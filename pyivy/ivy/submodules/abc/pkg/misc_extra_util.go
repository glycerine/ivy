package abc

// SIS compatibility utilities from src/misc/extra/extraUtilUtil.c.

import (
	"fmt"
	"os"
	"runtime/metrics"
	"strings"
	"time"
)

const ExtraSoftDataLimit = 67_108_864

type ExtraOptionIssue uint8

const (
	ExtraOptionOK ExtraOptionIssue = iota
	ExtraOptionUnknown
	ExtraOptionMissingArgument
)

type ExtraOption struct {
	Option   byte
	Argument string
	Issue    ExtraOptionIssue
	Done     bool
}

type ExtraOptionParser struct {
	arguments []string
	options   string
	index     int
	scan      string
	position  int
}

func NewExtraOptionParser(arguments []string, optionString string) *ExtraOptionParser {
	return &ExtraOptionParser{arguments: append([]string(nil), arguments...), options: optionString}
}

func (parser *ExtraOptionParser) Reset() {
	if parser == nil {
		return
	}
	parser.index, parser.scan, parser.position = 0, "", 0
}

func (parser *ExtraOptionParser) Index() int {
	if parser == nil {
		return 0
	}
	return parser.index
}

func (parser *ExtraOptionParser) Next() ExtraOption {
	if parser == nil {
		return ExtraOption{Done: true}
	}
	if parser.position >= len(parser.scan) {
		parser.scan, parser.position = "", 0
		if parser.index == 0 {
			parser.index++
		}
		if parser.index >= len(parser.arguments) {
			return ExtraOption{Done: true}
		}
		argument := parser.arguments[parser.index]
		if len(argument) < 2 || argument[0] != '-' {
			return ExtraOption{Done: true}
		}
		parser.index++
		if argument == "--" {
			return ExtraOption{Done: true}
		}
		parser.scan = argument[1:]
	}
	option := parser.scan[parser.position]
	parser.position++
	place := strings.IndexByte(parser.options, option)
	if place < 0 || option == ':' {
		return ExtraOption{Option: '?', Issue: ExtraOptionUnknown}
	}
	result := ExtraOption{Option: option}
	if place+1 >= len(parser.options) || parser.options[place+1] != ':' {
		return result
	}
	if parser.position < len(parser.scan) {
		result.Argument = parser.scan[parser.position:]
		parser.scan, parser.position = "", 0
		return result
	}
	parser.scan, parser.position = "", 0
	if parser.index >= len(parser.arguments) {
		return ExtraOption{Option: '?', Issue: ExtraOptionMissingArgument}
	}
	result.Argument = parser.arguments[parser.index]
	parser.index++
	return result
}

func FormatExtraMilliseconds(milliseconds int64) string {
	return fmt.Sprintf("%d.%02d sec", milliseconds/1000, milliseconds%1000/10)
}

func ExtraSaveString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := strings.Clone(*value)
	return &copy
}

func ExtraTildeExpand(name string) string { return strings.Clone(name) }

func normalizeExtraFileMode(mode string) (string, error) {
	mode = strings.ReplaceAll(mode, "b", "")
	switch mode {
	case "x":
		return "r", nil
	case "r", "r+", "w", "w+", "a", "a+":
		return mode, nil
	default:
		return "", ErrDimension
	}
}

func CheckExtraFile(name, mode string) (bool, error) {
	mode, err := normalizeExtraFileMode(mode)
	if err != nil {
		return false, err
	}
	flags := os.O_RDONLY
	switch mode {
	case "r+":
		flags = os.O_RDWR
	case "w":
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case "w+":
		flags = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	case "a":
		flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case "a+":
		flags = os.O_RDWR | os.O_CREATE | os.O_APPEND
	}
	file, err := os.OpenFile(name, flags, 0o666)
	if err != nil {
		return false, nil
	}
	if closeErr := file.Close(); closeErr != nil {
		return false, closeErr
	}
	return true, nil
}

type ExtraFileSearchLimits struct {
	MaxCandidates, MaxPathBytes int
	Stop                        func() bool
}

func normalizeExtraFileSearchLimits(limits ExtraFileSearchLimits) (ExtraFileSearchLimits, error) {
	if limits.MaxCandidates < 0 || limits.MaxPathBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxCandidates == 0 {
		limits.MaxCandidates = 1 << 16
	}
	if limits.MaxPathBytes == 0 {
		limits.MaxPathBytes = 16 << 20
	}
	return limits, nil
}

func SearchExtraFile(file, path, mode string, limits ExtraFileSearchLimits) (string, bool, error) {
	limits, err := normalizeExtraFileSearchLimits(limits)
	if err != nil {
		return "", false, err
	}
	if _, err = normalizeExtraFileMode(mode); err != nil {
		return "", false, err
	}
	if path == "" {
		path = "."
	}
	components := strings.Split(path, ":")
	if len(components) > limits.MaxCandidates {
		return "", false, ErrLimit
	}
	for _, component := range components {
		if limits.Stop != nil && limits.Stop() {
			return "", false, ErrCancelled
		}
		candidate := file
		if component != "." {
			candidate = component + "/" + file
		}
		candidate = ExtraTildeExpand(candidate)
		if len(candidate) > limits.MaxPathBytes {
			return "", false, ErrLimit
		}
		found, checkErr := CheckExtraFile(candidate, mode)
		if checkErr != nil {
			return "", false, checkErr
		}
		if found {
			return candidate, true, nil
		}
	}
	return "", false, nil
}

func ExtraOutOfMemory(size int64) error {
	if size < 0 {
		return ErrDimension
	}
	return fmt.Errorf("%w: out of memory allocating %d bytes", ErrLimit, size)
}

func ExtraCPUTimeSeconds() float64 {
	samples := []metrics.Sample{{Name: "/cpu/classes/user:cpu-seconds"}}
	metrics.Read(samples)
	if samples[0].Value.Kind() != metrics.KindFloat64 {
		return 0
	}
	return samples[0].Value.Float64()
}

func ExtraCPUTime() time.Duration {
	return time.Duration(ExtraCPUTimeSeconds() * float64(time.Second))
}

func ExtraMemoryTest() {}
