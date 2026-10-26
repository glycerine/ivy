package abc

// Numeric and representation contracts from src/misc/espresso/espresso.h.

const EspressoVersion = "UC Berkeley, Espresso Version #2.3, Release date 01/31/88"

type EspressoBoolean uint8

const (
	EspressoFalse EspressoBoolean = iota
	EspressoTrue
	EspressoMaybe
)

func (value EspressoBoolean) String() string {
	switch value {
	case EspressoFalse:
		return "FALSE"
	case EspressoTrue:
		return "TRUE"
	case EspressoMaybe:
		return "MAYBE"
	default:
		return ""
	}
}

type EspressoCubeInputValue uint8

const (
	EspressoCubeInputInvalid EspressoCubeInputValue = iota
	EspressoCubeInputZero
	EspressoCubeInputOne
	EspressoCubeInputDash
)

type EspressoRuntimeSlot uint8

const (
	EspressoRuntimeRead EspressoRuntimeSlot = iota
	EspressoRuntimeComplement
	EspressoRuntimeOnset
	EspressoRuntimeEssentials
	EspressoRuntimeExpand
	EspressoRuntimeIrredundant
	EspressoRuntimeReduce
	EspressoRuntimeGaspExpand
	EspressoRuntimeGaspIrredundant
	EspressoRuntimeGaspReduce
	EspressoRuntimePrimes
	EspressoRuntimeMinimumCover
	EspressoRuntimeMVReduce
	EspressoRuntimeRaiseInput
	EspressoRuntimeVerify
	EspressoRuntimeWrite
	EspressoRuntimeSlotCount
)
