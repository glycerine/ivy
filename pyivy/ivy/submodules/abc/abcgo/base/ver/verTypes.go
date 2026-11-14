// Signal and gate kinds from verCore.c.
package ver

type Ver_SignalType_t int32

const (
	VER_SIG_NONE Ver_SignalType_t = 0

	VER_SIG_INPUT Ver_SignalType_t = VER_SIG_NONE + 1

	VER_SIG_OUTPUT Ver_SignalType_t = VER_SIG_INPUT + 1

	VER_SIG_INOUT Ver_SignalType_t = VER_SIG_OUTPUT + 1

	VER_SIG_REG Ver_SignalType_t = VER_SIG_INOUT + 1

	VER_SIG_WIRE Ver_SignalType_t = VER_SIG_REG + 1
)

// types of verilog gates
type Ver_GateType_t int32

const (
	VER_GATE_AND Ver_GateType_t = 0

	VER_GATE_OR Ver_GateType_t = VER_GATE_AND + 1

	VER_GATE_XOR Ver_GateType_t = VER_GATE_OR + 1

	VER_GATE_BUF Ver_GateType_t = VER_GATE_XOR + 1

	VER_GATE_NAND Ver_GateType_t = VER_GATE_BUF + 1

	VER_GATE_NOR Ver_GateType_t = VER_GATE_NAND + 1

	VER_GATE_XNOR Ver_GateType_t = VER_GATE_NOR + 1

	VER_GATE_NOT Ver_GateType_t = VER_GATE_XNOR + 1
)
