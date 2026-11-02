package abc

// Hand-checked skeleton of src/map/mio/mioApi.c.

// Mio_LibraryReadDelayNand2Rise
// C: float Mio_LibraryReadDelayNand2Rise( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibraryReadDelayNand2Rise() float32 {
	panic("TODO, stubbed")
}

// Mio_LibraryReadDelayNand2Fall
// C: float Mio_LibraryReadDelayNand2Fall( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibraryReadDelayNand2Fall() float32 {
	panic("TODO, stubbed")
}

// Mio_LibraryReadGateNameMax
// C: int Mio_LibraryReadGateNameMax( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibraryReadGateNameMax() int {
	panic("TODO, stubbed")
}

// Mio_LibraryReadGateByName
// C: Mio_Gate_t * Mio_LibraryReadGateByName( Mio_Library_t * pLib, char * pName, char * pOutName )
func (pLib *MioLibrary) MioLibraryReadGateByName(pName string, pOutName string) *MioGate {
	panic("TODO, stubbed")
}

// Mio_LibraryReadGateByTruth
// C: Mio_Gate_t * Mio_LibraryReadGateByTruth( Mio_Library_t * pLib, word t )
func (pLib *MioLibrary) MioLibraryReadGateByTruth(t uint64) *MioGate {
	panic("TODO, stubbed")
}

// Mio_LibraryReadSopByName
// C: char * Mio_LibraryReadSopByName( Mio_Library_t * pLib, char * pName )
func (pLib *MioLibrary) MioLibraryReadSopByName(pName string) string {
	panic("TODO, stubbed")
}

// Mio_GateAddToProfile2
// C: void Mio_GateAddToProfile2( Mio_Gate_t * pGate, int Prof )
func (pGate *MioGate) MioGateAddToProfile2(Prof int) {
	panic("TODO, stubbed")
}

// Mio_PinReadDelayFanoutRise
// C: double Mio_PinReadDelayFanoutRise( Mio_Pin_t * pPin )
func (pPin *MioPin) MioPinReadDelayFanoutRise() float64 {
	panic("TODO, stubbed")
}

// Mio_PinReadDelayFanoutFall
// C: double Mio_PinReadDelayFanoutFall( Mio_Pin_t * pPin )
func (pPin *MioPin) MioPinReadDelayFanoutFall() float64 {
	panic("TODO, stubbed")
}

// Mio_GateReadPinName
// C: char * Mio_GateReadPinName( Mio_Gate_t * pGate, int iPin )
func (pGate *MioGate) MioGateReadPinName(iPin int) string {
	panic("TODO, stubbed")
}

// Mio_GateReadPinDelay
// C: float Mio_GateReadPinDelay( Mio_Gate_t * pGate, int iPin )
func (pGate *MioGate) MioGateReadPinDelay(iPin int) float32 {
	panic("TODO, stubbed")
}
