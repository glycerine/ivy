package mio

import "fmt"

// MatchingProvider is implemented by the consuming mapper (GIA/NF), whose
// original algorithms depend on MIO. It keeps that dependency one-directional.
type MatchingProvider interface {
	DeriveMioMatches(*Mio_Library_t, int32, int32, int32) (MatchData, error)
	DeriveMioMatches2(*Mio_Library_t) (MatchData2, error)
}
type MatchData struct {
	Truths  *Vec_Mem_t
	Matches *Vec_Wec_t
	Cells   []Mio_Cell2_t
}
type MatchData2 struct {
	Names             *Vec_Ptr_t
	Truths            *Vec_Wrd_t
	Matches4, Configs *Vec_Int_t
	TruthMemory       [3]*Vec_Mem_t
	Matches2          [3]*Vec_Int_t
}

func (lib *Mio_Library_t) SetMatchingProvider(provider MatchingProvider) {
	Mio_LibraryMatchesStop(lib)
	Mio_LibraryMatches2Stop(lib)
	lib.matchingProvider = provider
}
func Mio_LibraryMatchesStop(lib *Mio_Library_t) {
	lib.vTtMem = nil
	lib.vTt2Match = nil
	lib.pCells = nil
	lib.nCells = 0
}
func Mio_LibraryMatchesStart(lib *Mio_Library_t, filter, permutation, quick int32) error {
	if lib.vTtMem != nil && lib.fPinFilter == filter && lib.fPinPerm == permutation && lib.fPinQuick == quick {
		return nil
	}
	if lib.matchingProvider == nil {
		return fmt.Errorf("MIO: NF matching provider is not registered")
	}
	data, err := lib.matchingProvider.DeriveMioMatches(lib, filter, permutation, quick)
	if err != nil {
		return err
	}
	if data.Truths == nil || data.Matches == nil {
		return fmt.Errorf("MIO: matching provider returned incomplete data")
	}
	Mio_LibraryMatchesStop(lib)
	lib.fPinFilter, lib.fPinPerm, lib.fPinQuick = filter, permutation, quick
	lib.vTtMem, lib.vTt2Match, lib.pCells = data.Truths, data.Matches, data.Cells
	lib.nCells = int32(len(data.Cells))
	return nil
}
func Mio_LibraryMatchesFetch(lib *Mio_Library_t, truths **Vec_Mem_t, matches **Vec_Wec_t, cells *[]Mio_Cell2_t, count *int32, filter, permutation, quick int32) error {
	if err := Mio_LibraryMatchesStart(lib, filter, permutation, quick); err != nil {
		return err
	}
	*truths, *matches, *cells, *count = lib.vTtMem, lib.vTt2Match, lib.pCells, lib.nCells
	return nil
}
func Mio_LibraryMatches2Stop(lib *Mio_Library_t) {
	lib.vNames = nil
	lib.vTruths = nil
	lib.vTt2Match4 = nil
	lib.vConfigs = nil
	lib.vTtMem2 = [3]*Vec_Mem_t{}
	lib.vTt2Match2 = [3]*Vec_Int_t{}
}
func Mio_LibraryMatches2Start(lib *Mio_Library_t) error {
	if lib.vNames != nil {
		return nil
	}
	if lib.matchingProvider == nil {
		return fmt.Errorf("MIO: GIA matching provider is not registered")
	}
	data, err := lib.matchingProvider.DeriveMioMatches2(lib)
	if err != nil {
		return err
	}
	if data.Names == nil || data.Truths == nil || data.Matches4 == nil || data.Configs == nil {
		return fmt.Errorf("MIO: GIA matching provider returned incomplete data")
	}
	for i := 0; i < 3; i++ {
		if data.TruthMemory[i] == nil || data.Matches2[i] == nil {
			return fmt.Errorf("MIO: GIA matching provider returned incomplete table %d", i)
		}
	}
	lib.vNames, lib.vTruths, lib.vTt2Match4, lib.vConfigs = data.Names, data.Truths, data.Matches4, data.Configs
	lib.vTtMem2, lib.vTt2Match2 = data.TruthMemory, data.Matches2
	return nil
}
func Mio_LibraryMatches2Fetch(lib *Mio_Library_t, names **Vec_Ptr_t, truths **Vec_Wrd_t, matches4, configs **Vec_Int_t, memory []*Vec_Mem_t, matches2 []*Vec_Int_t) error {
	if len(memory) < 3 || len(matches2) < 3 {
		return fmt.Errorf("MIO: matching output arrays require three entries")
	}
	if err := Mio_LibraryMatches2Start(lib); err != nil {
		return err
	}
	*names, *truths, *matches4, *configs = lib.vNames, lib.vTruths, lib.vTt2Match4, lib.vConfigs
	copy(memory, lib.vTtMem2[:])
	copy(matches2, lib.vTt2Match2[:])
	return nil
}
func Mio_LibraryTransferCellIds() error {
	frame := abcMioFrame()
	lib := mioCurrentLibrary()
	if lib == nil {
		return fmt.Errorf("MIO: genlib library is unavailable")
	}
	finder, ok := frame.Resource("pLibScl").(interface{ MioCellID(string) (int32, bool) })
	if !ok {
		return fmt.Errorf("MIO: standard-cell library ID lookup is unavailable")
	}
	staged := make(map[*Mio_Gate_t]int32)
	for g := lib.pGates; g != nil; g = g.pNext {
		if g.nInputs == 0 {
			continue
		}
		id, ok := finder.MioCellID(g.pName)
		if !ok {
			return fmt.Errorf("MIO: cannot find standard-cell ID for %s", g.pName)
		}
		staged[g] = id
	}
	for g, id := range staged {
		g.Cell = id
	}
	return nil
}
