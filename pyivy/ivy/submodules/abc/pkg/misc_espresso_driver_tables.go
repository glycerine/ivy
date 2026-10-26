package abc

// Immutable driver tables from src/misc/espresso/main.h.

type EspressoCommandKind uint8

const (
	EspressoCommandMinimize EspressoCommandKind = iota
	EspressoCommandMany
	EspressoCommandExact
	EspressoCommandQuineMcCluskey
	EspressoCommandSingleOutput
	EspressoCommandSingleOutputBoth
	EspressoCommandSimplify
	EspressoCommandEcho
	EspressoCommandOutputPhase
	EspressoCommandAllOutputPhases
	EspressoCommandPair
	EspressoCommandAllPairs
	EspressoCommandCheck
	EspressoCommandStats
	EspressoCommandVerify
	EspressoCommandPLAVerify
	EspressoCommandEquivalentOutputs
	EspressoCommandMap
	EspressoCommandMapDontCare
	EspressoCommandFSM
	EspressoCommandContain
	EspressoCommandDistanceOneMerge
	EspressoCommandDistanceOneMergeInputs
	EspressoCommandDisjoint
	EspressoCommandDisjointSharp
	EspressoCommandIntersect
	EspressoCommandMinterms
	EspressoCommandPrimes
	EspressoCommandSeparate
	EspressoCommandSharp
	EspressoCommandUnion
	EspressoCommandXOR
	EspressoCommandEssentials
	EspressoCommandExpand
	EspressoCommandGasp
	EspressoCommandIrredundant
	EspressoCommandMakeSparse
	EspressoCommandReduce
	EspressoCommandTautology
	EspressoCommandSuperGasp
	EspressoCommandLexicalSort
	EspressoCommandTest
)

type EspressoCommandDescriptor struct {
	Name                    string
	Kind                    EspressoCommandKind
	InputPLAs               int
	NeedsOff, NeedsDontCare bool
}

var espressoCommandDescriptors = [...]EspressoCommandDescriptor{
	{"ESPRESSO", EspressoCommandMinimize, 1, true, true},
	{"many", EspressoCommandMany, 1, true, true},
	{"exact", EspressoCommandExact, 1, true, true},
	{"qm", EspressoCommandQuineMcCluskey, 1, true, true},
	{"single_output", EspressoCommandSingleOutput, 1, true, true},
	{"so", EspressoCommandSingleOutput, 1, true, true},
	{"so_both", EspressoCommandSingleOutputBoth, 1, true, true},
	{"simplify", EspressoCommandSimplify, 1, false, false},
	{"echo", EspressoCommandEcho, 1, false, false},
	{"opo", EspressoCommandOutputPhase, 1, true, true},
	{"opoall", EspressoCommandAllOutputPhases, 1, true, true},
	{"pair", EspressoCommandPair, 1, true, true},
	{"pairall", EspressoCommandAllPairs, 1, true, true},
	{"check", EspressoCommandCheck, 1, true, true},
	{"stats", EspressoCommandStats, 1, false, false},
	{"verify", EspressoCommandVerify, 2, false, true},
	{"PLAverify", EspressoCommandPLAVerify, 2, false, true},
	{"equiv", EspressoCommandEquivalentOutputs, 1, true, true},
	{"map", EspressoCommandMap, 1, false, false},
	{"mapdc", EspressoCommandMapDontCare, 1, false, false},
	{"fsm", EspressoCommandFSM, 1, false, true},
	{"contain", EspressoCommandContain, 1, false, false},
	{"d1merge", EspressoCommandDistanceOneMerge, 1, false, false},
	{"d1merge_in", EspressoCommandDistanceOneMergeInputs, 1, false, false},
	{"disjoint", EspressoCommandDisjoint, 1, true, false},
	{"dsharp", EspressoCommandDisjointSharp, 2, false, false},
	{"intersect", EspressoCommandIntersect, 2, false, false},
	{"minterms", EspressoCommandMinterms, 1, false, false},
	{"primes", EspressoCommandPrimes, 1, false, true},
	{"separate", EspressoCommandSeparate, 1, true, true},
	{"sharp", EspressoCommandSharp, 2, false, false},
	{"union", EspressoCommandUnion, 2, false, false},
	{"xor", EspressoCommandXOR, 2, true, true},
	{"essen", EspressoCommandEssentials, 1, false, true},
	{"expand", EspressoCommandExpand, 1, true, false},
	{"gasp", EspressoCommandGasp, 1, true, true},
	{"irred", EspressoCommandIrredundant, 1, false, true},
	{"make_sparse", EspressoCommandMakeSparse, 1, true, true},
	{"reduce", EspressoCommandReduce, 1, false, true},
	{"taut", EspressoCommandTautology, 1, false, false},
	{"super_gasp", EspressoCommandSuperGasp, 1, true, true},
	{"lexsort", EspressoCommandLexicalSort, 1, false, false},
	{"test", EspressoCommandTest, 1, true, true},
}

func EspressoCommandDescriptors() []EspressoCommandDescriptor {
	result := make([]EspressoCommandDescriptor, len(espressoCommandDescriptors))
	copy(result, espressoCommandDescriptors[:])
	return result
}

func LookupEspressoCommand(name string) (EspressoCommandDescriptor, bool) {
	for _, descriptor := range espressoCommandDescriptors {
		if descriptor.Name == name {
			return descriptor, true
		}
	}
	return EspressoCommandDescriptor{}, false
}

const (
	EspressoDebugComplement uint32 = 1 << iota
	EspressoDebugEssentials
	EspressoDebugExpand
	EspressoDebugExpandDetail
	EspressoDebugGasp
	EspressoDebugIrredundant
	EspressoDebugReduce
	EspressoDebugReduceDetail
	EspressoDebugSparse
	EspressoDebugTautology
	EspressoDebugExact
	EspressoDebugMinimumCover
	EspressoDebugMinimumCoverDetail
	EspressoDebugSharp
	EspressoDebugIrredundantDetail
)

type EspressoDebugDescriptor struct {
	Name string
	Mask uint32
}

var espressoDebugDescriptors = [...]EspressoDebugDescriptor{
	{"", EspressoDebugExpand | EspressoDebugEssentials | EspressoDebugIrredundant | EspressoDebugReduce | EspressoDebugSparse | EspressoDebugGasp | EspressoDebugSharp | EspressoDebugMinimumCover},
	{"compl", EspressoDebugComplement},
	{"essen", EspressoDebugEssentials},
	{"expand", EspressoDebugExpand},
	{"expand1", EspressoDebugExpandDetail | EspressoDebugExpand},
	{"irred", EspressoDebugIrredundant},
	{"irred1", EspressoDebugIrredundantDetail | EspressoDebugIrredundant},
	{"reduce", EspressoDebugReduce},
	{"reduce1", EspressoDebugReduceDetail | EspressoDebugReduce},
	{"mincov", EspressoDebugMinimumCover},
	{"mincov1", EspressoDebugMinimumCoverDetail | EspressoDebugMinimumCover},
	{"sparse", EspressoDebugSparse},
	{"sharp", EspressoDebugSharp},
	{"taut", EspressoDebugTautology},
	{"gasp", EspressoDebugGasp},
	{"exact", EspressoDebugExact},
}

func EspressoDebugDescriptors() []EspressoDebugDescriptor {
	result := make([]EspressoDebugDescriptor, len(espressoDebugDescriptors))
	copy(result, espressoDebugDescriptors[:])
	return result
}

func LookupEspressoDebug(name string) (uint32, bool) {
	for _, descriptor := range espressoDebugDescriptors {
		if descriptor.Name == name {
			return descriptor.Mask, true
		}
	}
	return 0, false
}

// ApplyEspressoBehaviorOption applies one -e option to an owned value.
func ApplyEspressoBehaviorOption(options EspressoRuntimeOptions, name string) (EspressoRuntimeOptions, error) {
	switch name {
	case "eat":
		options.EchoComments = false
	case "eatdots":
		options.EchoUnknownCommands = false
	case "fast":
		options.SingleExpand = true
	case "kiss":
		options.KISS = true
	case "ness":
		options.RemoveEssential = false
	case "nirr":
		options.ForceIrredundant = false
	case "nunwrap":
		options.UnwrapOnset = false
	case "onset":
		options.RecomputeOnset = true
	case "pos":
		options.Positive = true
	case "random":
		options.UseRandomOrder = true
	case "strong":
		options.UseSuperGasp = true
	default:
		return EspressoRuntimeOptions{}, ErrUnsupported
	}
	return options, nil
}
