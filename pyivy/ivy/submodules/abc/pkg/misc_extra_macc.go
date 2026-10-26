package abc

// Constant multiplier-accumulator generation from src/misc/extra/extraUtilMacc.c.

import (
	"math"
	"strconv"
	"strings"
)

type MaccLimits struct {
	MaxEntries, MaxOperations, MaxBytes int
	Stop                                func() bool
}

type MaccPlan struct {
	bits    int
	entries []uint32
	levels  [][]int
}

type MaccGeneratedFile struct {
	Name, Text string
}

func normalizeMaccLimits(limits MaccLimits) (MaccLimits, error) {
	if limits.MaxEntries < 0 || limits.MaxOperations < 0 || limits.MaxBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = 1 << 20
	}
	if limits.MaxOperations == 0 {
		limits.MaxOperations = 1 << 26
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 64 << 20
	}
	return limits, nil
}

func checkMaccBits(bits int) error {
	if bits < 1 || bits > 15 {
		return ErrDimension
	}
	return nil
}

func maccMask(bits int) uint32 { return uint32(1<<uint(bits)) - 1 }

func maccBound(bits int) int { return 1 << uint(bits-1) }

func maccIndex(value, bits int) int { return int(uint32(value) & maccMask(bits)) }

func GenerateMaccPlan(bits int, limits MaccLimits) (*MaccPlan, error) {
	if err := checkMaccBits(bits); err != nil {
		return nil, err
	}
	limits, err := normalizeMaccLimits(limits)
	if err != nil {
		return nil, err
	}
	entryCount := 1 << uint(bits)
	if entryCount > limits.MaxEntries {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	plan := &MaccPlan{bits: bits, entries: make([]uint32, entryCount), levels: make([][]int, 2*bits)}
	sentinel := ^uint32(0)
	mask, bound := int(maccMask(bits)), maccBound(bits)
	plan.entries[0], plan.entries[1], plan.entries[mask] = sentinel, sentinel, sentinel
	for value := 2; value < entryCount; value++ {
		if value&1 == 0 {
			plan.entries[value] = sentinel
		}
	}
	for shift := 0; shift < bits; shift++ {
		plan.levels[0] = append(plan.levels[0], 1<<uint(shift))
	}
	operations := 0
	try := func(value, one, two int, subtract bool, level *[]int) error {
		operations++
		if operations > limits.MaxOperations {
			return ErrLimit
		}
		if limits.Stop != nil && limits.Stop() {
			return ErrCancelled
		}
		if value < -bound || value >= bound || plan.entries[maccIndex(value, bits)] != 0 {
			return nil
		}
		if value > 0 {
			*level = append(*level, value)
		}
		encoded := uint32(one)<<16 | uint32(two)
		if subtract {
			encoded |= 1 << 15
		}
		plan.entries[maccIndex(value, bits)] = encoded
		return nil
	}
	for adders := 1; adders < bits; adders++ {
		complete := true
		for value := 1; value < entryCount; value++ {
			if plan.entries[value] == 0 {
				complete = false
				break
			}
		}
		if complete {
			break
		}
		level := &plan.levels[adders]
		for firstLevel := 0; firstLevel < bits; firstLevel++ {
			for secondLevel := 0; secondLevel < bits; secondLevel++ {
				if firstLevel+secondLevel != adders-1 {
					continue
				}
				for _, one := range plan.levels[firstLevel] {
					for _, two := range plan.levels[secondLevel] {
						if err := try(one+two, one, two, false, level); err != nil {
							return nil, err
						}
						if err := try(one-two, one, two, true, level); err != nil {
							return nil, err
						}
						if err := try(two-one, two, one, true, level); err != nil {
							return nil, err
						}
					}
				}
				for _, one := range *level {
					operations++
					if operations > limits.MaxOperations {
						return nil, ErrLimit
					}
					negative := -one
					index := maccIndex(negative, bits)
					if plan.entries[index] == 0 {
						plan.entries[index] = uint32(one) << 16
					}
				}
			}
		}
	}
	for index, entry := range plan.entries {
		if entry == 0 {
			return nil, ErrInvariant
		}
		_ = index
	}
	remaining := limits.MaxOperations - operations
	if remaining <= 0 {
		return nil, ErrLimit
	}
	validationLimits := limits
	validationLimits.MaxOperations = remaining
	if err := plan.Validate(validationLimits); err != nil {
		return nil, err
	}
	return plan, nil
}

func (plan *MaccPlan) check() error {
	if plan == nil || checkMaccBits(plan.bits) != nil || len(plan.entries) != 1<<uint(plan.bits) || len(plan.levels) != 2*plan.bits {
		return ErrInvariant
	}
	return nil
}

func (plan *MaccPlan) Bits() (int, error) {
	if err := plan.check(); err != nil {
		return 0, err
	}
	return plan.bits, nil
}

func (plan *MaccPlan) Entries() ([]uint32, error) {
	if err := plan.check(); err != nil {
		return nil, err
	}
	return append([]uint32(nil), plan.entries...), nil
}

func (plan *MaccPlan) evaluate(value int, state []uint8, operations *int, limits MaccLimits) (int, error) {
	bound := maccBound(plan.bits)
	if value < -bound || value > bound {
		return 0, ErrDimension
	}
	index := maccIndex(value, plan.bits)
	*operations++
	if *operations > limits.MaxOperations {
		return 0, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return 0, ErrCancelled
	}
	if state[index] == 1 {
		return 0, ErrInvariant
	}
	entry := plan.entries[index]
	if entry == ^uint32(0) {
		return value, nil
	}
	if entry == 0 {
		return 0, ErrInvariant
	}
	state[index] = 1
	one, two := int(entry>>16), int(entry&0x7fff)
	left, err := plan.evaluate(one, state, operations, limits)
	if err != nil {
		return 0, err
	}
	result := -left
	if two != 0 {
		right, rightErr := plan.evaluate(two, state, operations, limits)
		if rightErr != nil {
			return 0, rightErr
		}
		if entry&(1<<15) != 0 {
			result = left - right
		} else {
			result = left + right
		}
	}
	state[index] = 2
	return result, nil
}

func (plan *MaccPlan) Coefficient(value int, limits MaccLimits) (int, error) {
	if err := plan.check(); err != nil {
		return 0, err
	}
	bound := maccBound(plan.bits)
	if value < -bound || value >= bound {
		return 0, ErrDimension
	}
	limits, err := normalizeMaccLimits(limits)
	if err != nil {
		return 0, err
	}
	operations := 0
	return plan.evaluate(value, make([]uint8, len(plan.entries)), &operations, limits)
}

func (plan *MaccPlan) Validate(limits MaccLimits) error {
	if err := plan.check(); err != nil {
		return err
	}
	limits, err := normalizeMaccLimits(limits)
	if err != nil {
		return err
	}
	bound := maccBound(plan.bits)
	operations := 0
	for value := -bound; value < bound; value++ {
		coefficient, coefficientErr := plan.evaluate(value, make([]uint8, len(plan.entries)), &operations, limits)
		if coefficientErr != nil {
			return coefficientErr
		}
		if coefficient != value {
			return ErrInvariant
		}
	}
	return nil
}

type maccText struct {
	builder strings.Builder
	limit   int
}

func (text *maccText) write(parts ...string) error {
	length := 0
	for _, part := range parts {
		if len(part) > text.limit-text.builder.Len()-length {
			return ErrLimit
		}
		length += len(part)
	}
	for _, part := range parts {
		text.builder.WriteString(part)
	}
	return nil
}

func maccDecimal(value int) string { return strconv.Itoa(value) }

func maccAbsolute(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func maccPadded(value, width int) string {
	result := strconv.Itoa(value)
	if len(result) < width {
		result = strings.Repeat("0", width-len(result)) + result
	}
	return result
}

func checkMaccFormat(constant, bits, width int, limits MaccLimits) (MaccLimits, error) {
	if err := checkMaccBits(bits); err != nil || width <= 0 || width > math.MaxInt32-bits {
		return limits, ErrDimension
	}
	bound := maccBound(bits)
	if constant < -bound || constant >= bound {
		return limits, ErrDimension
	}
	return normalizeMaccLimits(limits)
}

func maccModuleName(prefix string, constant int) string {
	suffix := "_pos"
	if constant < 0 {
		suffix = "_neg"
	}
	return prefix + maccPadded(maccAbsolute(constant), 3) + suffix
}

func FormatDirectConstMultiplier(constant, bits, width int, limits MaccLimits) (string, error) {
	limits, err := checkMaccFormat(constant, bits, width, limits)
	if err != nil {
		return "", err
	}
	total := width + bits
	text := maccText{limit: limits.MaxBytes}
	minus := " "
	if constant < 0 {
		minus = "-"
	}
	lines := [][]string{
		{"// ", maccDecimal(width), "-bit multiplier by ", maccDecimal(bits), "-bit constant ", maccDecimal(constant), " generated by ABC\n"},
		{"module ", maccModuleName("mul", constant), " (\n"},
		{"    input  [", maccDecimal(width - 1), ":0] i,\n"},
		{"    output [", maccDecimal(width - 1), ":0] o\n"},
		{");\n"},
		{"    wire [", maccDecimal(bits - 1), ":0] c = ", maccDecimal(bits), "'h", strconv.FormatInt(int64(maccAbsolute(constant)), 16), ";\n"},
		{"    wire [", maccDecimal(total - 1), ":0] I = {{", maccDecimal(bits), "{i[", maccDecimal(width - 1), "]}}, i};\n"},
		{"    wire [", maccDecimal(total - 1), ":0] m = I * c;\n"},
		{"    wire [", maccDecimal(total - 1), ":0] t = ", minus, "m;\n"},
		{"    assign o = t[", maccDecimal(total - 1), ":", maccDecimal(bits), "];\n"},
		{"endmodule\n\n"},
	}
	for _, line := range lines {
		if err = text.write(line...); err != nil {
			return "", err
		}
	}
	return text.builder.String(), nil
}

func FormatDirectConstMacc(constant, bits, width int, timestamp string, limits MaccLimits) (string, error) {
	limits, err := checkMaccFormat(constant, bits, width, limits)
	if err != nil || strings.ContainsAny(timestamp, "\r\n") {
		if err != nil {
			return "", err
		}
		return "", ErrDimension
	}
	total := width + bits
	text := maccText{limit: limits.MaxBytes}
	operator := "+"
	if constant < 0 {
		operator = "-"
	}
	lines := [][]string{
		{"// ", maccDecimal(total), "-bit multiplier-accumulator with constant ", maccDecimal(constant), " generated by ABC on ", timestamp, "\n"},
		{"module ", maccModuleName("mulacc", constant), " (\n"},
		{"    input  [", maccDecimal(total - 1), ":0] i,\n"},
		{"    input  [", maccDecimal(total - 1), ":0] s,\n"},
		{"    output [", maccDecimal(total - 1), ":0] o\n"},
		{");\n"},
		{"    wire [", maccDecimal(total - 1), ":0] c = ", maccDecimal(total), "'h", strconv.FormatInt(int64(maccAbsolute(constant)), 16), ";\n"},
		{"    wire [", maccDecimal(total - 1), ":0] m = i * c;\n"},
		{"    assign o = s ", operator, " m;\n"},
		{"endmodule\n\n"},
	}
	for _, line := range lines {
		if err = text.write(line...); err != nil {
			return "", err
		}
	}
	return text.builder.String(), nil
}

func (plan *MaccPlan) renderCoefficient(text *maccText, constant, width int, state []uint8, operations *int, limits MaccLimits) error {
	index := maccIndex(constant, plan.bits)
	*operations++
	if *operations > limits.MaxOperations {
		return ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return ErrCancelled
	}
	if state[index] == 1 {
		return ErrInvariant
	}
	if state[index] == 2 {
		return nil
	}
	state[index] = 1
	total := width + plan.bits
	sign := "n"
	if constant < 0 {
		sign = "N"
	}
	entry := plan.entries[index]
	if entry == ^uint32(0) {
		zeros, reduced := 0, constant
		for zeros < plan.bits && reduced&1 == 0 {
			zeros++
			reduced >>= 1
		}
		if reduced == -1 {
			if err := text.write("    wire [", maccDecimal(total-1), ":0] N1 = -n1;\n"); err != nil {
				return err
			}
		}
		if maccAbsolute(reduced) != 1 {
			if err := plan.renderCoefficient(text, reduced, width, state, operations, limits); err != nil {
				return err
			}
		}
		if zeros > 0 {
			reducedSign := "n"
			if reduced < 0 {
				reducedSign = "N"
			}
			if err := text.write("    wire [", maccDecimal(total-1), ":0] ", sign, maccDecimal(maccAbsolute(constant)), " = ", reducedSign, maccDecimal(maccAbsolute(reduced)), " << ", maccDecimal(zeros), ";\n"); err != nil {
				return err
			}
		}
	} else {
		one, two := int(entry>>16), int(entry&0x7fff)
		if one == 0 {
			return ErrInvariant
		}
		if err := plan.renderCoefficient(text, one, width, state, operations, limits); err != nil {
			return err
		}
		if two != 0 {
			if err := plan.renderCoefficient(text, two, width, state, operations, limits); err != nil {
				return err
			}
			operator := "+"
			if entry&(1<<15) != 0 {
				operator = "-"
			}
			if err := text.write("    wire [", maccDecimal(total-1), ":0] ", sign, maccDecimal(maccAbsolute(constant)), " = n", maccDecimal(one), " ", operator, " n", maccDecimal(two), ";\n"); err != nil {
				return err
			}
		} else if err := text.write("    wire [", maccDecimal(total-1), ":0] N", maccDecimal(one), " = -n", maccDecimal(one), ";\n"); err != nil {
			return err
		}
	}
	state[index] = 2
	return nil
}

type maccSynthKind uint8

const (
	maccSynthMultiplier maccSynthKind = iota
	maccSynthAccumulator
	maccSynthWideAccumulator
)

func (plan *MaccPlan) formatSynth(constant, width int, timestamp string, kind maccSynthKind, limits MaccLimits) (string, error) {
	if err := plan.check(); err != nil {
		return "", err
	}
	limits, err := checkMaccFormat(constant, plan.bits, width, limits)
	if err != nil || strings.ContainsAny(timestamp, "\r\n") {
		if err != nil {
			return "", err
		}
		return "", ErrDimension
	}
	total := width + plan.bits
	prefix, description := "mul", "-bit multiplier by "
	if kind != maccSynthMultiplier {
		prefix, description = "macc", "-bit multiplier-accumulator by "
	}
	text := maccText{limit: limits.MaxBytes}
	comment := []string{"// ", maccDecimal(width), description, maccDecimal(plan.bits), "-bit constant ", maccDecimal(constant), " generated by ABC\n"}
	if kind == maccSynthWideAccumulator {
		comment = []string{"// ", maccDecimal(total), "-bit multiplier-accumulator by constant ", maccDecimal(constant), " generated by ABC on ", timestamp, "\n"}
	}
	if err = text.write(comment...); err != nil {
		return "", err
	}
	if err = text.write("module ", maccModuleName(prefix, constant), " (\n"); err != nil {
		return "", err
	}
	portWidth := width
	if kind == maccSynthWideAccumulator {
		portWidth = total
	}
	if err = text.write("    input  [", maccDecimal(portWidth-1), ":0] i,\n"); err != nil {
		return "", err
	}
	if kind != maccSynthMultiplier {
		name := "c"
		if kind == maccSynthWideAccumulator {
			name = "s"
		}
		if err = text.write("    input  [", maccDecimal(portWidth-1), ":0] ", name, ",\n"); err != nil {
			return "", err
		}
	}
	if err = text.write("    output [", maccDecimal(portWidth-1), ":0] o\n);\n"); err != nil {
		return "", err
	}
	if constant == 0 {
		assignment := "    assign o = " + maccDecimal(width) + "'h0;\n"
		if kind != maccSynthMultiplier {
			assignment = "    assign o = c;\n"
			if kind == maccSynthWideAccumulator {
				assignment = "    assign o = s;\n"
			}
		}
		if err = text.write(assignment); err != nil {
			return "", err
		}
	} else {
		if kind == maccSynthWideAccumulator {
			err = text.write("    wire [", maccDecimal(total-1), ":0] n1 = i;\n")
		} else {
			err = text.write("    wire [", maccDecimal(total-1), ":0] n1 = {{", maccDecimal(plan.bits), "{i[", maccDecimal(width-1), "]}}, i};\n")
		}
		if err != nil {
			return "", err
		}
		operations := 0
		if err = plan.renderCoefficient(&text, constant, width, make([]uint8, len(plan.entries)), &operations, limits); err != nil {
			return "", err
		}
		sign := "n"
		if constant < 0 {
			sign = "N"
		}
		absolute := maccDecimal(maccAbsolute(constant))
		switch kind {
		case maccSynthMultiplier:
			err = text.write("    assign o = ", sign, absolute, "[", maccDecimal(total-1), ":", maccDecimal(plan.bits), "];\n")
		case maccSynthAccumulator:
			err = text.write("    wire [", maccDecimal(width-1), ":0] s = ", sign, absolute, "[", maccDecimal(total-1), ":", maccDecimal(plan.bits), "];\n", "    assign o = s + c;\n")
		case maccSynthWideAccumulator:
			err = text.write("    assign o = s + ", sign, absolute, ";\n")
		}
		if err != nil {
			return "", err
		}
	}
	if err = text.write("endmodule\n\n"); err != nil {
		return "", err
	}
	return text.builder.String(), nil
}

func (plan *MaccPlan) FormatMultiplier(constant, width int, limits MaccLimits) (string, error) {
	return plan.formatSynth(constant, width, "", maccSynthMultiplier, limits)
}

func (plan *MaccPlan) FormatAccumulator(constant, width int, limits MaccLimits) (string, error) {
	return plan.formatSynth(constant, width, "", maccSynthAccumulator, limits)
}

func (plan *MaccPlan) FormatWideAccumulator(constant, width int, timestamp string, limits MaccLimits) (string, error) {
	return plan.formatSynth(constant, width, timestamp, maccSynthWideAccumulator, limits)
}

func GenerateDirectMaccSpecifications(bits, width int, timestamp string, limits MaccLimits) ([]MaccGeneratedFile, error) {
	limits, err := normalizeMaccLimits(limits)
	if err != nil || checkMaccBits(bits) != nil || width <= 0 {
		return nil, ErrDimension
	}
	bound := maccBound(bits)
	if 2*bound > limits.MaxEntries {
		return nil, ErrLimit
	}
	result := make([]MaccGeneratedFile, 0, 2*bound)
	for constant := -bound; constant < bound; constant++ {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		text, formatErr := FormatDirectConstMacc(constant, bits, width, timestamp, limits)
		if formatErr != nil {
			return nil, formatErr
		}
		result = append(result, MaccGeneratedFile{Name: "macc_spec_" + maccPadded(maccIndex(constant, bits), 3) + ".v", Text: text})
	}
	return result, nil
}

func (plan *MaccPlan) GenerateWideAccumulatorSet(width int, timestamp string, limits MaccLimits) ([]MaccGeneratedFile, error) {
	if err := plan.check(); err != nil {
		return nil, err
	}
	limits, err := normalizeMaccLimits(limits)
	if err != nil || width <= 0 {
		return nil, ErrDimension
	}
	bound := maccBound(plan.bits)
	if 2*bound > limits.MaxEntries {
		return nil, ErrLimit
	}
	result := make([]MaccGeneratedFile, 0, 2*bound)
	for constant := -bound; constant < bound; constant++ {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		text, formatErr := plan.FormatWideAccumulator(constant, width, timestamp, limits)
		if formatErr != nil {
			return nil, formatErr
		}
		result = append(result, MaccGeneratedFile{Name: "macc" + maccPadded(maccIndex(constant, plan.bits), 3) + ".v", Text: text})
	}
	return result, nil
}
