package abc

// Simple LUT-library construction from src/map/mpm/mpmLib.c.

const (
	MPMMaximumVariables = 12
	MPMUnitArea         = 20
	MPMUnitTime         = 1
)

type MPMLUTLibrary struct {
	Name              string
	MaximumLUTSize    int
	VariablePinDelays bool
	Areas             [MPMMaximumVariables + 1]int
	Delays            [MPMMaximumVariables + 1][MPMMaximumVariables + 1]int
}

// NewSimpleMPMLUTLibrary implements Mpm_LibLutSetSimple. Index zero and delay
// entries at or above each LUT's input count retain their native zero values.
func NewSimpleMPMLUTLibrary(size int) (*MPMLUTLibrary, error) {
	if size < 0 || size > MPMMaximumVariables {
		return nil, ErrDimension
	}
	library := &MPMLUTLibrary{MaximumLUTSize: size}
	for inputs := 1; inputs <= size; inputs++ {
		library.Areas[inputs] = MPMUnitArea
		for pin := 0; pin < inputs; pin++ {
			library.Delays[inputs][pin] = MPMUnitTime
		}
	}
	return library, nil
}
