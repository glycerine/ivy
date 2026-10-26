package abc

// Observable no-op stubs from place_legalize.c, acbSets.c, wlcJson.c,
// giaSim4.c, giaSim5.c, extraUtilCfs.c, extraUtilGen.c, plaMerge.c, mpm.c,
// abcRunGen.c, abcEco.c, cbaCba.c, giaIiff.c and sfmMit.c. They remain concrete
// compatibility spellings until their native subsystems acquire behavior.

// LegalizePlacement implements the pinned legalize function, whose body is an
// explicit unimplemented no-op.
func LegalizePlacement() {}

// AcbDataReadTest implements the pinned empty Acb_DataReadTest hook.
func AcbDataReadTest() {}

// WlcJSONTest implements the pinned empty Wlc_JsonTest hook.
func WlcJSONTest() {}

// GiaSim4Try implements the pinned Gia_Sim4Try stub. The native body ignores
// every argument and returns zero.
func GiaSim4Try(file0, file1, file2 string, timeout, words, beam, levelLow, levelHigh int, order, fancy, useBuffer, skipMFFC, verbose bool) int {
	return 0
}

// SimInit and SimEnd preserve the pinned empty simulation package hooks. The
// ABC frame is opaque because this port does not expose a fabricated manager.
func SimInit(frame uintptr) {}
func SimEnd(frame uintptr)  {}

// GiaDatFree preserves the pinned empty simulation-data release hook. Go-owned
// GIA data uses ordinary lifetime management elsewhere.
func GiaDatFree(data uintptr) {}

// ExtraCommandCfs preserves the pinned empty command body and does not mutate
// the supplied concrete GIA manager.
func ExtraCommandCfs(manager *GiaManager, limit, repetitions, unseenUse, rareUse int, replaceRare, constantSimulation, dagNodes bool, flipProbability float32, verbose bool) {
}

// ExtraCommandGen preserves the pinned generator stub's unconditional null
// result for every parameter combination.
func ExtraCommandGen(algorithm, lutSize, lutCount, levels, limit, bestTries, multi int, xor, verbose bool) *GiaManager {
	return nil
}

// PlaManDist1Merge preserves the pinned placeholder's unconditional zero
// result. The PLA manager remains opaque until its surrounding subsystem lands.
func PlaManDist1Merge(manager uintptr) int { return 0 }

// MpmManTest preserves the pinned empty mapper test hook.
func MpmManTest() {}

// TestInit and TestEnd preserve the pinned empty template-package hooks. The
// ABC frame is opaque because neither body inspects it.
func TestInit(frame uintptr) {}
func TestEnd(frame uintptr)  {}

// AcbNtkRunGen and AbcNtkEco preserve pinned experimental hooks whose bodies
// ignore their filename arrays and flags.
func AcbNtkRunGen(fileNames [2]string, verbose bool) {}
func AbcNtkEco(fileNames [3]string)                  {}

// CbaManReadCba and CbaManWriteCba preserve the pinned binary-CBA placeholders.
// The native manager is still opaque, so its always-null result and ignored
// input pointer are represented by uintptr rather than a fabricated Go type.
func CbaManReadCba(fileName string) uintptr           { return 0 }
func CbaManWriteCba(fileName string, manager uintptr) {}

// GiaManDeriveMatches preserves the pinned matcher placeholder. Every pointer
// argument is ignored and the result is always zero.
func GiaManDeriveMatches(names, truths, matches4, configurations uintptr, truthMemories, matches2 [3]uintptr) int {
	return 0
}

// GiaManIiffTest preserves the pinned IIFF placeholder's unconditional null
// result. The concrete GIA manager argument is not inspected or mutated.
func GiaManIiffTest(fileName string, manager *GiaManager, lutSize, cutCount int, useGates, useCells, useLUTs, verbose bool) *GiaManager {
	return nil
}

// SFMMit names the timing-mitigation placeholder declared by sfmMit.c. The
// pinned constructor always returns nil, so no live value is fabricated.
type SFMMit struct{}

// SFMMitStart preserves the pinned constructor's unconditional null result.
// Its surrounding libraries and network remain opaque until those subsystems
// have concrete Go representations.
func SFMMitStart(cellLibrary, standardCellLibrary, externalConstraints, network uintptr, criticalDelta int) *SFMMit {
	return nil
}

// The remaining SFMMit functions preserve explicit empty/zero placeholder
// bodies. Slice arguments are concrete and must remain unchanged.
func SFMMitStop(manager *SFMMit)                                       {}
func SFMMitNetworkDelay(manager *SFMMit) int                           { return 0 }
func SFMMitNetworkMinimumSlack(manager *SFMMit) int                    { return 0 }
func SFMMitObjectDelay(manager *SFMMit, object int) int                { return 0 }
func SFMMitTransferLoad(manager *SFMMit, newObject, oldObject uintptr) {}
func SFMMitTimingGrow(manager *SFMMit)                                 {}
func SFMMitUpdateLoad(manager *SFMMit, timingNodes []int, add bool)    {}
func SFMMitUpdateTiming(manager *SFMMit, timingNodes []int)            {}
func SFMMitSortByArrival(manager *SFMMit, nodes []int, pivot int) int  { return 0 }
func SFMMitPriorityNodes(manager *SFMMit, candidates []int, window int) int {
	return 0
}
func SFMMitNodeIsNoncritical(manager *SFMMit, pivot, node uintptr) int { return 0 }
func SFMMitEvaluateRemapping(manager *SFMMit, mffc []int, object uintptr, fanins, mapping []int, gate1 uintptr, gate1Fanins []byte, gate2 uintptr, gate2Fanins []byte) int {
	return 0
}
