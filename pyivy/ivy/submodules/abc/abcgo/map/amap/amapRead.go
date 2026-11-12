package amap

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/map/amap/formula"
)

const (
	AMAP_STRING_GATE    = "GATE"
	AMAP_STRING_PIN     = "PIN"
	AMAP_STRING_NONINV  = "NONINV"
	AMAP_STRING_INV     = "INV"
	AMAP_STRING_UNKNOWN = "UNKNOWN"
)

type Amap_PinPhase_t int32

const (
	AMAP_PHASE_UNKNOWN Amap_PinPhase_t = iota
	AMAP_PHASE_INV
	AMAP_PHASE_NONINV
)

func Amap_ParseGateAlloc(_ *Aig_MmFlex_t, nPins int32) *Amap_Gat_t {
	if nPins < 0 {
		panic("amap: negative pin count")
	}
	return &Amap_Gat_t{Pins: make([]Amap_Pin_t, int(nPins))}
}

func Amap_ParseStrsav(_ *Aig_MmFlex_t, value string) string { return strings.Clone(value) }

func Amap_LoadFile(fileName string) string {
	data, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Cannot open file %q: %v.\n", fileName, err)
		return ""
	}
	return string(data) + "\n.end\n"
}

// Amap_RemoveComments blanks comment text in place in a copied byte buffer,
// joins backslash-continued lines, and reports the source's dot and line counts.
func Amap_RemoveComments(input string, dots, lines *int32) string {
	data := []byte(input)
	var nDots, nLines int32
	for i := 0; i < len(data); i++ {
		if data[i] == '#' {
			for i < len(data) && data[i] != '\n' {
				data[i] = ' '
				i++
			}
			if i >= len(data) {
				break
			}
		}
		if data[i] == '\n' {
			end := i - 1
			if end >= 0 && data[end] == '\r' {
				end--
			}
			if end >= 0 && data[end] == '\\' {
				for j := end; j <= i; j++ {
					data[j] = ' '
				}
			} else {
				nLines++
			}
		} else if data[i] == '.' {
			nDots++
		}
	}
	if dots != nil {
		*dots = nDots
	}
	if lines != nil {
		*lines = nLines
	}
	return string(data)
}

func Amap_DeriveTokens(input string) *Vec_Ptr_t {
	tokens := Vec_PtrAlloc(int32(len(input) / 4))
	fields := strings.FieldsFunc(input, func(r rune) bool {
		return r == ' ' || r == '=' || r == '\t' || r == '\r' || r == '\n'
	})
	for i := 0; i < len(fields); i++ {
		token := fields[i]
		if token == "LATCH" {
			for i+1 < len(fields) && fields[i+1] != AMAP_STRING_GATE {
				i++
			}
			continue
		}
		Vec_PtrPush(tokens, token)
	}
	return tokens
}

func Amap_ParseCountPins(tokens *Vec_Ptr_t, pos int32) int32 {
	var count int32
	for i := pos; i < Vec_PtrSize(tokens); i++ {
		value, ok := Vec_PtrEntry(tokens, i).(string)
		if !ok {
			continue
		}
		if value == AMAP_STRING_GATE {
			return count
		}
		if value == AMAP_STRING_PIN {
			count++
		}
	}
	return count
}

func Amap_GateCollectNames(_ *Aig_MmFlex_t, text string, names []string) int32 {
	e, err := formula.Parse(strings.TrimSuffix(strings.TrimSpace(text), ";"))
	if err != nil {
		return 0
	}
	variables := e.Variables()
	if len(variables) > len(names) {
		return 0
	}
	copy(names, variables)
	return int32(len(variables))
}

func Amap_ParseGateWithSamePins(gate *Amap_Gat_t) *Amap_Gat_t {
	if gate == nil || gate.nPins != 1 || len(gate.Pins) != 1 || gate.Pins[0].pName != "*" || gate.pLib == nil {
		return nil
	}
	vars := make([]string, AMAP_MAXINS)
	n := Amap_GateCollectNames(gate.pLib.pMemGates, gate.pForm, vars)
	if n == 0 {
		return nil
	}
	clone := *gate
	clone.nPins = uint32(n)
	clone.Pins = make([]Amap_Pin_t, n)
	for i := range clone.Pins {
		clone.Pins[i] = gate.Pins[0]
		clone.Pins[i].pName = vars[i]
	}
	return &clone
}

func Amap_CollectFormulaTokens(tokens *Vec_Ptr_t, first string, pos int32) (string, int32, error) {
	parts := []string{first}
	if strings.HasSuffix(first, ";") {
		return strings.TrimSuffix(strings.Join(parts, " "), ";"), pos, nil
	}
	for pos < Vec_PtrSize(tokens) {
		token, ok := Vec_PtrEntry(tokens, pos).(string)
		if !ok {
			return "", pos, fmt.Errorf("formula token %d is not text", pos)
		}
		pos++
		parts = append(parts, token)
		if strings.HasSuffix(token, ";") {
			joined := strings.Join(parts, " ")
			return strings.TrimSuffix(joined, ";"), pos, nil
		}
	}
	return "", pos, fmt.Errorf("formula is missing terminating semicolon")
}

func Amap_ParseTokens(tokens *Vec_Ptr_t, verbose int32) *Amap_Lib_t {
	if tokens == nil || Vec_PtrSize(tokens) == 0 {
		return nil
	}
	first, _ := Vec_PtrEntry(tokens, 0).(string)
	if first != AMAP_STRING_GATE {
		fmt.Printf("The first line should begin with %s.\n", AMAP_STRING_GATE)
		return nil
	}
	lib := Amap_LibAlloc()
	lib.fVerbose = verbose
	var previous *Amap_Gat_t
	var multiName string
	var multiCount int32
	for pos := int32(0); pos < Vec_PtrSize(tokens); {
		token, ok := Vec_PtrEntry(tokens, pos).(string)
		if !ok || token == ".end" {
			break
		}
		if token != AMAP_STRING_GATE {
			Amap_LibFree(lib)
			fmt.Printf("Expected GATE at token %d, got %q.\n", pos, token)
			return nil
		}
		pos++
		if pos+3 >= Vec_PtrSize(tokens) {
			Amap_LibFree(lib)
			fmt.Println("Incomplete GATE declaration.")
			return nil
		}
		name, _ := Vec_PtrEntry(tokens, pos).(string)
		pos++
		areaText, _ := Vec_PtrEntry(tokens, pos).(string)
		area, err := strconv.ParseFloat(areaText, 64)
		if err != nil {
			Amap_LibFree(lib)
			fmt.Printf("Invalid area %q for gate %q.\n", areaText, name)
			return nil
		}
		pos++
		outputName, _ := Vec_PtrEntry(tokens, pos).(string)
		pos++
		formulaStart, _ := Vec_PtrEntry(tokens, pos).(string)
		pos++
		form, next, err := Amap_CollectFormulaTokens(tokens, formulaStart, pos)
		if err != nil {
			Amap_LibFree(lib)
			fmt.Printf("Cannot read formula of gate %q: %v.\n", name, err)
			return nil
		}
		pos = next
		nPins := Amap_ParseCountPins(tokens, pos)
		gate := Amap_ParseGateAlloc(lib.pMemGates, nPins)
		gate.Id = uint32(Vec_PtrSize(lib.vGates))
		gate.pLib, gate.nPins = lib, uint32(nPins)
		gate.pName, gate.pOutName, gate.dArea, gate.pForm = strings.Clone(name), strings.Clone(outputName), area, form
		Vec_PtrPush(lib.vGates, gate)
		for pin := int32(0); pin < nPins; pin++ {
			if pos+8 >= Vec_PtrSize(tokens) {
				Amap_LibFree(lib)
				fmt.Printf("Incomplete pin %d of gate %q.\n", pin, name)
				return nil
			}
			if got, _ := Vec_PtrEntry(tokens, pos).(string); got != AMAP_STRING_PIN {
				Amap_LibFree(lib)
				fmt.Printf("Cannot parse gate %q: expected PIN, got %q.\n", name, got)
				return nil
			}
			pos++
			pinName, _ := Vec_PtrEntry(tokens, pos).(string)
			pos++
			phaseText, _ := Vec_PtrEntry(tokens, pos).(string)
			pos++
			phase := AMAP_PHASE_UNKNOWN
			switch phaseText {
			case AMAP_STRING_UNKNOWN:
				phase = AMAP_PHASE_UNKNOWN
			case AMAP_STRING_INV:
				phase = AMAP_PHASE_INV
			case AMAP_STRING_NONINV:
				phase = AMAP_PHASE_NONINV
			default:
				Amap_LibFree(lib)
				fmt.Printf("Cannot read phase %q for pin %q of gate %q.\n", phaseText, pinName, name)
				return nil
			}
			values := make([]float64, 6)
			for i := range values {
				text, _ := Vec_PtrEntry(tokens, pos).(string)
				pos++
				values[i], err = strconv.ParseFloat(text, 64)
				if err != nil {
					Amap_LibFree(lib)
					fmt.Printf("Invalid pin timing/load value %q in gate %q.\n", text, name)
					return nil
				}
			}
			gate.Pins[pin] = Amap_Pin_t{pName: strings.Clone(pinName), Phase: int32(phase), dLoadInput: values[0], dLoadMax: values[1], dDelayBlockRise: values[2], dDelayFanoutRise: values[3], dDelayBlockFall: values[4], dDelayFanoutFall: values[5], dDelayBlockMax: max(values[2], values[4])}
		}
		if gate.nPins == 1 && gate.Pins[0].pName == "*" {
			clone := Amap_ParseGateWithSamePins(gate)
			if clone == nil {
				Amap_LibFree(lib)
				fmt.Printf("Cannot expand wildcard pin of gate %q.\n", name)
				return nil
			}
			clone.Id = gate.Id
			Vec_PtrWriteEntry(lib.vGates, int(Vec_PtrSize(lib.vGates)-1), clone)
			gate = clone
		}
		if previous != nil && previous.pName == gate.pName {
			previous.pTwin, gate.pTwin = gate, previous
			if multiName == "" {
				multiName = gate.pName
			}
			multiCount++
		}
		previous = gate
	}
	if multiCount > 0 {
		fmt.Printf("Warning: Detected %d multi-output gates (for example, %q).\n", multiCount, multiName)
	}
	return lib
}

func Amap_LibReadBuffer(buffer string, verbose int32) *Amap_Lib_t {
	buffer = Amap_RemoveComments(buffer, nil, nil)
	tokens := Amap_DeriveTokens(buffer)
	defer Vec_PtrFree(tokens)
	return Amap_ParseTokens(tokens, verbose)
}

func Amap_LibReadFile(fileName string, verbose int32) *Amap_Lib_t {
	buffer := Amap_LoadFile(fileName)
	if buffer == "" {
		return nil
	}
	lib := Amap_LibReadBuffer(buffer, verbose)
	if lib != nil {
		lib.pName = strings.Clone(fileName)
	}
	return lib
}
