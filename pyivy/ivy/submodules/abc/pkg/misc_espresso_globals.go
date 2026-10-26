package abc

// Owned replacements for process-global state from src/misc/espresso/globals.c.

import (
	"math/bits"
	"strconv"
	"strings"
)

// EspressoRuntimeOptions owns the driver switches that the source kept in
// writable package globals. Algorithm APIs continue to accept their narrower
// concrete option structs; this type is the driver-level configuration.
type EspressoRuntimeOptions struct {
	Debug               uint32
	VerboseDebug        bool
	EchoComments        bool
	EchoUnknownCommands bool
	ForceIrredundant    bool
	SkipMakeSparse      bool
	KISS                bool
	Positive            bool
	PrintSolution       bool
	RecomputeOnset      bool
	RemoveEssential     bool
	SingleExpand        bool
	Summary             bool
	Trace               bool
	UnwrapOnset         bool
	UseRandomOrder      bool
	UseSuperGasp        bool
}

func DefaultEspressoRuntimeOptions() EspressoRuntimeOptions {
	return EspressoRuntimeOptions{
		EchoComments:        true,
		EchoUnknownCommands: true,
		ForceIrredundant:    true,
		PrintSolution:       true,
		RemoveEssential:     true,
		UnwrapOnset:         true,
	}
}

// MinimizeOptions projects driver switches into the concrete minimizer. Limits
// and cancellation remain caller-owned because globals.c did not define them.
func (options EspressoRuntimeOptions) MinimizeOptions() (EspressoMinimizeOptions, error) {
	return EspressoMinimizeOptions{
		RecomputeOnset:   options.RecomputeOnset,
		UnwrapOnset:      options.UnwrapOnset,
		RemoveEssential:  options.RemoveEssential,
		SingleExpand:     options.SingleExpand,
		UseSuperGasp:     options.UseSuperGasp,
		SkipMakeSparse:   options.SkipMakeSparse,
		ForceIrredundant: options.ForceIrredundant,
		UseRandomOrder:   options.UseRandomOrder,
		RandomSeed:       314973,
		ReduceOrder:      EspressoReduceDistanceOrder,
	}, nil
}

type EspressoPLATypeDescriptor struct {
	Name string
	Type EspressoPLAOutputType
}

var espressoPLATypeDescriptors = [...]EspressoPLATypeDescriptor{
	{"-f", EspressoPLAOutputOn},
	{"-r", EspressoPLAOutputOff},
	{"-d", EspressoPLAOutputDontCare},
	{"-fd", EspressoPLAOutputOn | EspressoPLAOutputDontCare},
	{"-fr", EspressoPLAOutputOn | EspressoPLAOutputOff},
	{"-dr", EspressoPLAOutputDontCare | EspressoPLAOutputOff},
	{"-fdr", EspressoPLAOutputOn | EspressoPLAOutputDontCare | EspressoPLAOutputOff},
	{"-fc", EspressoPLAOutputOn | EspressoPLAOutputConstraints},
	{"-rc", EspressoPLAOutputOff | EspressoPLAOutputConstraints},
	{"-dc", EspressoPLAOutputDontCare | EspressoPLAOutputConstraints},
	{"-fdc", EspressoPLAOutputOn | EspressoPLAOutputDontCare | EspressoPLAOutputConstraints},
	{"-frc", EspressoPLAOutputOn | EspressoPLAOutputOff | EspressoPLAOutputConstraints},
	{"-drc", EspressoPLAOutputDontCare | EspressoPLAOutputOff | EspressoPLAOutputConstraints},
	{"-fdrc", EspressoPLAOutputOn | EspressoPLAOutputDontCare | EspressoPLAOutputOff | EspressoPLAOutputConstraints},
	{"-pleasure", EspressoPLAOutputPleasure},
	{"-eqn", EspressoPLAOutputEquation},
	{"-eqntott", EspressoPLAOutputEquation},
	{"-kiss", EspressoPLAOutputKISS},
	{"-cons", EspressoPLAOutputConstraints},
	{"-scons", EspressoPLAOutputSymbolicConstraints},
}

func EspressoPLATypeDescriptors() []EspressoPLATypeDescriptor {
	result := make([]EspressoPLATypeDescriptor, len(espressoPLATypeDescriptors))
	copy(result, espressoPLATypeDescriptors[:])
	return result
}

// LookupEspressoPLAType accepts the source spellings with or without their
// leading dash. Matching remains case-sensitive.
func LookupEspressoPLAType(name string) (EspressoPLAOutputType, bool) {
	if !strings.HasPrefix(name, "-") {
		name = "-" + name
	}
	for _, descriptor := range espressoPLATypeDescriptors {
		if descriptor.Name == name {
			return descriptor.Type, true
		}
	}
	return 0, false
}

var espressoRuntimeNames = [...]string{
	"READ       ", "COMPL      ", "ONSET      ", "ESSEN      ",
	"EXPAND     ", "IRRED      ", "REDUCE     ", "EXPAND_GASP",
	"IRRED_GASP ", "REDUCE_GASP", "PRIMES     ", "MINCOV     ",
	"MV_REDUCE  ", "RAISE_IN   ", "VERIFY     ", "WRITE      ",
}

func NewEspressoRuntimeTotals() *EspressoCoverTotals {
	names := make([]string, len(espressoRuntimeNames))
	copy(names, espressoRuntimeNames[:])
	return NewEspressoCoverTotals(names)
}

func espressoBytePopulation(value byte) int { return bits.OnesCount8(value) }

// FormatRuntime returns the former runtime() report from caller-owned totals.
func (totals *EspressoCoverTotals) FormatRuntime(maxBytes int) (string, error) {
	if totals == nil || len(totals.Names) != len(totals.Milliseconds) || len(totals.Names) != len(totals.Calls) || maxBytes < 0 {
		return "", ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = 1 << 20
	}
	total := int64(1)
	for index, elapsed := range totals.Milliseconds {
		if elapsed < 0 || totals.Calls[index] < 0 || elapsed > int64(^uint64(0)>>1)-total {
			return "", ErrInvariant
		}
		total += elapsed
	}
	var builder strings.Builder
	for index, calls := range totals.Calls {
		if calls == 0 {
			continue
		}
		timeText, _ := formatEspressoMilliseconds(totals.Milliseconds[index])
		high, low := bits.Mul64(uint64(totals.Milliseconds[index]), 1000)
		percentageTenths, _ := bits.Div64(high, low, uint64(total))
		line := "# " + totals.Names[index] + "\t" + strconv.Itoa(calls) + " call(s) for " + timeText + " (" + strconv.FormatUint(percentageTenths/10, 10) + "." + strconv.FormatUint(percentageTenths%10, 10) + "%)\n"
		if len(line) > maxBytes-builder.Len() {
			return "", ErrLimit
		}
		builder.WriteString(line)
	}
	return builder.String(), nil
}
