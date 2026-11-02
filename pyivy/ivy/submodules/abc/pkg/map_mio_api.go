package abc

// Port of the accessor functions of src/map/mio/mioApi.c that the skeleton had left as stubs.
// A NULL char* argument or result is the empty string.

// abcInfinity is ABC_INFINITY (misc/util/abc_global.h).
const abcInfinity = 1000000000

// MioLibraryReadDelayNand2Rise is Mio_LibraryReadDelayNand2Rise.
// C: float Mio_LibraryReadDelayNand2Rise( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibraryReadDelayNand2Rise() float32 {
	if pLib.pGateNand2 != nil {
		return float32(pLib.pGateNand2.pPins.dDelayBlockRise)
	}
	return 0
}

// MioLibraryReadDelayNand2Fall is Mio_LibraryReadDelayNand2Fall.
// C: float Mio_LibraryReadDelayNand2Fall( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibraryReadDelayNand2Fall() float32 {
	if pLib.pGateNand2 != nil {
		return float32(pLib.pGateNand2.pPins.dDelayBlockFall)
	}
	return 0
}

// MioLibraryReadGateNameMax is Mio_LibraryReadGateNameMax: the longest gate name.
// C: int Mio_LibraryReadGateNameMax( Mio_Library_t * pLib )
func (pLib *MioLibrary) MioLibraryReadGateNameMax() int {
	LenMax := 0
	for pGate := pLib.pGates; pGate != nil; pGate = pGate.pNext {
		if LenCur := len(pGate.pName); LenMax < LenCur {
			LenMax = LenCur
		}
	}
	return LenMax
}

// MioLibraryReadGateByName is Mio_LibraryReadGateByName (pOutName == "" is NULL).
// C: Mio_Gate_t * Mio_LibraryReadGateByName( Mio_Library_t * pLib, char * pName, char * pOutName )
func (pLib *MioLibrary) MioLibraryReadGateByName(pName string, pOutName string) *MioGate {
	pGate, found, err := pLib.tName2Gate.Get(pName)
	if err != nil || !found {
		return nil
	}
	if pOutName == "" {
		return pGate
	}
	if pGate.pOutName == pOutName {
		return pGate
	}
	if pGate.pTwin != nil && pGate.pTwin.pOutName == pOutName {
		return pGate.pTwin
	}
	return nil
}

// MioLibraryReadGateByTruth is Mio_LibraryReadGateByTruth: the first gate with the given truth table.
// C: Mio_Gate_t * Mio_LibraryReadGateByTruth( Mio_Library_t * pLib, word t )
func (pLib *MioLibrary) MioLibraryReadGateByTruth(t uint64) *MioGate {
	for pGate := pLib.pGates; pGate != nil; pGate = pGate.pNext {
		if pGate.nInputs <= 6 && pGate.uTruth == t {
			return pGate
		}
	}
	return nil
}

// MioLibraryReadSopByName is Mio_LibraryReadSopByName ("" when the gate is not found).
// C: char * Mio_LibraryReadSopByName( Mio_Library_t * pLib, char * pName )
func (pLib *MioLibrary) MioLibraryReadSopByName(pName string) string {
	if pGate, found, err := pLib.tName2Gate.Get(pName); err == nil && found {
		return pGate.pSop
	}
	return ""
}

// MioGateAddToProfile2 is Mio_GateAddToProfile2.
// C: void Mio_GateAddToProfile2( Mio_Gate_t * pGate, int Prof )
func (pGate *MioGate) MioGateAddToProfile2(Prof int) { pGate.Profile2 += Prof }

// MioPinReadDelayFanoutRise is Mio_PinReadDelayFanoutRise.
// C: double Mio_PinReadDelayFanoutRise( Mio_Pin_t * pPin )
func (pPin *MioPin) MioPinReadDelayFanoutRise() float64 { return pPin.dDelayFanoutRise }

// MioPinReadDelayFanoutFall is Mio_PinReadDelayFanoutFall.
// C: double Mio_PinReadDelayFanoutFall( Mio_Pin_t * pPin )
func (pPin *MioPin) MioPinReadDelayFanoutFall() float64 { return pPin.dDelayFanoutFall }

// MioGateReadPinName is Mio_GateReadPinName ("" when there is no such pin).
// C: char * Mio_GateReadPinName( Mio_Gate_t * pGate, int iPin )
func (pGate *MioGate) MioGateReadPinName(iPin int) string {
	i := 0
	for pPin := pGate.pPins; pPin != nil; pPin = pPin.pNext {
		if i == iPin {
			return pPin.pName
		}
		i++
	}
	return ""
}

// MioGateReadPinDelay is Mio_GateReadPinDelay (ABC_INFINITY when there is no such pin).
// C: float Mio_GateReadPinDelay( Mio_Gate_t * pGate, int iPin )
func (pGate *MioGate) MioGateReadPinDelay(iPin int) float32 {
	i := 0
	for pPin := pGate.pPins; pPin != nil; pPin = pPin.pNext {
		if i == iPin {
			return float32(0.5*pPin.dDelayBlockRise + 0.5*pPin.dDelayBlockFall)
		}
		i++
	}
	return abcInfinity
}
