package abc

// Owned log encoding and parsing from src/base/abci/abcLog.c.

import (
	"strconv"
	"strings"
)

type ABCLogStatus int8

const (
	ABCLogUnknown ABCLogStatus = -1
	ABCLogSAT     ABCLogStatus = 0
	ABCLogUNSAT   ABCLogStatus = 1
)

type ABCLogRecord struct {
	Status         ABCLogStatus
	BugFreeDepth   int
	Engine         string
	Counterexample *Counterexample
}

type ABCLogOptions struct {
	MaxBytes, MaxFrames, MaxBits, MaxEngineBytes int
	Stop                                         func() bool
}

func normalizeABCLogOptions(options ABCLogOptions) (ABCLogOptions, error) {
	if options.MaxBytes < 0 || options.MaxFrames < 0 || options.MaxBits < 0 || options.MaxEngineBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 64 << 20
	}
	if options.MaxFrames == 0 {
		options.MaxFrames = 1 << 20
	}
	if options.MaxBits == 0 {
		options.MaxBits = 64 << 20
	}
	if options.MaxEngineBytes == 0 {
		options.MaxEngineBytes = 1 << 20
	}
	return options, nil
}

func abcLogStopped(options ABCLogOptions) error {
	if options.Stop != nil && options.Stop() {
		return ErrCancelled
	}
	return nil
}

func abcLogStatusToken(status ABCLogStatus) (string, error) {
	switch status {
	case ABCLogSAT:
		return "snl_SAT", nil
	case ABCLogUNSAT:
		return "snl_UNSAT", nil
	case ABCLogUnknown:
		return "snl_UNK", nil
	default:
		return "", ErrDimension
	}
}

func abcLogBits(trace *Counterexample, begin, end int, stop func() bool) (string, error) {
	if begin < 0 || end < begin || trace == nil || end > trace.BitCount() {
		return "", ErrDimension
	}
	bits := make([]byte, end-begin)
	for index := begin; index < end; index++ {
		if stop != nil && stop() {
			return "", ErrCancelled
		}
		bits[index-begin] = '0'
		if trace.bit(index) {
			bits[index-begin] = '1'
		}
	}
	return string(bits), nil
}

// MarshalABCLog returns the exact active three-line format without filesystem
// access. An empty engine is written as "unknown", matching a nil C string.
func MarshalABCLog(record ABCLogRecord, options ABCLogOptions) ([]byte, error) {
	options, err := normalizeABCLogOptions(options)
	if err != nil {
		return nil, err
	}
	if record.BugFreeDepth < 0 || record.BugFreeDepth > options.MaxFrames {
		return nil, ErrDimension
	}
	if err = abcLogStopped(options); err != nil {
		return nil, err
	}
	token, err := abcLogStatusToken(record.Status)
	if err != nil {
		return nil, err
	}
	engine := record.Engine
	if engine == "" {
		engine = "unknown"
	}
	if len(engine) > options.MaxEngineBytes || len(strings.Fields(engine)) != 1 || strings.IndexByte(engine, 0) >= 0 {
		return nil, ErrDimension
	}
	var state, inputs string
	if record.Counterexample == nil {
		state, inputs = "NULL", "NULL"
	} else {
		trace := record.Counterexample
		if err = trace.Validate(); err != nil {
			return nil, err
		}
		if trace.Frames() > options.MaxFrames || trace.BitCount() > options.MaxBits {
			return nil, ErrLimit
		}
		state, err = abcLogBits(trace, 0, trace.Registers(), options.Stop)
		if err != nil {
			return nil, err
		}
		inputs, err = abcLogBits(trace, trace.Registers(), trace.BitCount(), options.Stop)
		if err != nil {
			return nil, err
		}
	}
	var builder strings.Builder
	builder.WriteString(token)
	builder.WriteByte(' ')
	builder.WriteString(strconv.Itoa(record.BugFreeDepth))
	builder.WriteByte(' ')
	builder.WriteString(engine)
	if record.Counterexample != nil && record.Status == ABCLogSAT {
		builder.WriteByte(' ')
		builder.WriteString(strconv.Itoa(record.Counterexample.Property))
		if record.Counterexample.LastFrame() != record.BugFreeDepth {
			builder.WriteByte(' ')
			builder.WriteString(strconv.Itoa(record.Counterexample.LastFrame()))
		}
	}
	builder.WriteByte('\n')
	builder.WriteString(state)
	builder.WriteByte('\n')
	builder.WriteString(inputs)
	builder.WriteByte('\n')
	if builder.Len() > options.MaxBytes {
		return nil, ErrLimit
	}
	return []byte(builder.String()), nil
}

func abcLogParseNonnegative(text string) (int, error) {
	value, err := strconv.Atoi(text)
	if err != nil || value < 0 {
		return 0, ErrDimension
	}
	return value, nil
}

func abcLogParseBitLine(line string, maximum int) ([]bool, error) {
	if line == "" {
		return []bool{}, nil
	}
	if len(line) > maximum {
		return nil, ErrLimit
	}
	bits := make([]bool, len(line))
	for index := range line {
		switch line[index] {
		case '0':
		case '1':
			bits[index] = true
		default:
			return nil, ErrUnsupported
		}
	}
	return bits, nil
}

// ParseABCLog parses one complete three-line record. Unlike the native reader,
// NULL remains a nil counterexample and unsupported characters cannot silently
// shift bit positions.
func ParseABCLog(data []byte, options ABCLogOptions) (ABCLogRecord, error) {
	var result ABCLogRecord
	options, err := normalizeABCLogOptions(options)
	if err != nil {
		return result, err
	}
	if len(data) == 0 || len(data) > options.MaxBytes || strings.IndexByte(string(data), 0) >= 0 {
		if len(data) > options.MaxBytes {
			return result, ErrLimit
		}
		return result, ErrDimension
	}
	if err = abcLogStopped(options); err != nil {
		return result, err
	}
	text := strings.TrimSuffix(string(data), "\n")
	lines := strings.Split(text, "\n")
	if len(lines) != 3 {
		return result, ErrDimension
	}
	header := strings.Fields(lines[0])
	if len(header) < 3 {
		return result, ErrDimension
	}
	switch header[0] {
	case "snl_SAT":
		result.Status = ABCLogSAT
		if len(header) > 5 {
			return ABCLogRecord{}, ErrDimension
		}
	case "snl_UNSAT":
		result.Status = ABCLogUNSAT
		if len(header) != 3 {
			return ABCLogRecord{}, ErrDimension
		}
	case "snl_UNK":
		result.Status = ABCLogUnknown
		if len(header) != 3 {
			return ABCLogRecord{}, ErrDimension
		}
	default:
		return result, ErrDimension
	}
	result.BugFreeDepth, err = abcLogParseNonnegative(header[1])
	if err != nil || result.BugFreeDepth > options.MaxFrames {
		if result.BugFreeDepth > options.MaxFrames {
			return ABCLogRecord{}, ErrLimit
		}
		return ABCLogRecord{}, ErrDimension
	}
	result.Engine = header[2]
	if len(result.Engine) > options.MaxEngineBytes {
		return ABCLogRecord{}, ErrLimit
	}
	property, cexFrame := -1, result.BugFreeDepth
	if result.Status == ABCLogSAT && len(header) >= 4 {
		property, err = strconv.Atoi(header[3])
		if err != nil || property < -1 {
			return ABCLogRecord{}, ErrDimension
		}
		if len(header) == 5 {
			cexFrame, err = abcLogParseNonnegative(header[4])
			if err != nil {
				return ABCLogRecord{}, err
			}
		}
	}
	if cexFrame >= options.MaxFrames {
		return ABCLogRecord{}, ErrLimit
	}
	stateNull, inputNull := lines[1] == "NULL", lines[2] == "NULL"
	if stateNull || inputNull {
		if !stateNull || !inputNull {
			return ABCLogRecord{}, ErrDimension
		}
		return result, nil
	}
	state, err := abcLogParseBitLine(lines[1], options.MaxBits)
	if err != nil {
		return ABCLogRecord{}, err
	}
	inputs, err := abcLogParseBitLine(lines[2], options.MaxBits-len(state))
	if err != nil {
		return ABCLogRecord{}, err
	}
	frames := cexFrame + 1
	if frames < 1 || len(inputs)%frames != 0 || len(state) > options.MaxBits-len(inputs) {
		return ABCLogRecord{}, ErrDimension
	}
	trace, err := NewCounterexample(len(state), len(inputs)/frames, frames)
	if err != nil {
		return ABCLogRecord{}, err
	}
	trace.Property = property
	for index, value := range append(state, inputs...) {
		trace.set(index, value)
		if options.Stop != nil && options.Stop() {
			return ABCLogRecord{}, ErrCancelled
		}
	}
	result.Counterexample = trace
	return result, nil
}
