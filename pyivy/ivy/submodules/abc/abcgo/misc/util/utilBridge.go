package util

// ABC bridge framing and AIG payload encoding. BridgeNetlist separates the wire
// protocol from GIA's manager implementation, avoiding a util<->gia import cycle.
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
)

const (
	BRIDGE_TEXT_MESSAGE = 999996
	BRIDGE_ABORT        = 5
	BRIDGE_PROGRESS     = 3
	BRIDGE_RESULTS      = 101
	BRIDGE_BAD_ABS      = 105
	BRIDGE_VALUE_X      = 0
	BRIDGE_VALUE_0      = 2
	BRIDGE_VALUE_1      = 3
)

type BridgeNetlist struct {
	Inputs, Registers       int32
	Ands                    [][2]int32
	Outputs, RegisterInputs []int32
}

func bridgeAppendUnsigned(dst []byte, value uint32) []byte {
	return binary.AppendUvarint(dst, uint64(value))
}
func aigerNumSize(value uint32) int32 {
	var buffer [binary.MaxVarintLen32]byte
	return int32(binary.PutUvarint(buffer[:], uint64(value)))
}
func bridgeWireLit(lit int32) uint32 {
	if lit < 0 {
		panic("negative AIG literal")
	}
	if lit < 2 {
		return uint32(lit ^ 3)
	}
	return uint32(lit + 2)
}
func Gia_ManToBridgeVec(p *BridgeNetlist) []byte {
	assert(p.Inputs >= 0 && p.Registers >= 0 && int(p.Registers) == len(p.RegisterInputs) && len(p.Outputs) > 0, "bridge graph dimensions")
	data := make([]byte, 0, 128)
	data = bridgeAppendUnsigned(data, uint32(p.Inputs))
	data = bridgeAppendUnsigned(data, uint32(p.Registers))
	data = bridgeAppendUnsigned(data, uint32(len(p.Ands)))
	for i, and := range p.Ands {
		limit := 2 * (p.Inputs + p.Registers + int32(i) + 1)
		assert(and[0] >= 0 && and[0] < limit && and[1] >= 0 && and[1] < limit && and[0] != and[1], "bridge topological AND")
		data = bridgeAppendUnsigned(data, bridgeWireLit(and[0])<<1)
		data = bridgeAppendUnsigned(data, bridgeWireLit(and[1]))
	}
	for _, lit := range p.RegisterInputs {
		data = bridgeAppendUnsigned(data, bridgeWireLit(lit)<<2|BRIDGE_VALUE_0)
	}
	data = bridgeAppendUnsigned(data, uint32(len(p.Outputs)))
	for _, lit := range p.Outputs {
		data = bridgeAppendUnsigned(data, bridgeWireLit(lit)^1)
	}
	return data
}
func bridgeWriteAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if n > 0 {
			data = data[n:]
		}
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}
func WriteBridgePackage(w io.Writer, typ int32, payload []byte) error {
	if typ < 0 || typ > 999999 {
		return fmt.Errorf("bridge message type out of range: %d", typ)
	}
	header := fmt.Sprintf("%06d %016d ", typ, len(payload))
	if err := bridgeWriteAll(w, []byte(header)); err != nil {
		return err
	}
	return bridgeWriteAll(w, payload)
}
func ReadBridgePackage(r io.Reader) (typ int32, payload []byte, err error) {
	var header [24]byte
	if _, err = io.ReadFull(r, header[:]); err != nil {
		return
	}
	if header[6] != ' ' || header[23] != ' ' {
		return 0, nil, fmt.Errorf("invalid bridge header separators")
	}
	kind, e := strconv.ParseUint(string(header[:6]), 10, 20)
	if e != nil {
		return 0, nil, e
	}
	size, e := strconv.ParseUint(string(header[7:23]), 10, 64)
	if e != nil {
		return 0, nil, e
	}
	// A bound protects readers from an untrusted length triggering an allocation.
	if size > 1<<30 {
		return 0, nil, fmt.Errorf("bridge payload exceeds 1 GiB")
	}
	payload = make([]byte, int(size))
	_, err = io.ReadFull(r, payload)
	return int32(kind), payload, err
}
func Gia_CreateHeader(w io.Writer, typ, size int32, data []byte) error {
	if size < 0 || int(size) > len(data) {
		return fmt.Errorf("bridge size exceeds buffer")
	}
	return WriteBridgePackage(w, typ, data[:size])
}
func bridgeSend(w io.Writer, typ, size int32, data []byte) int32 {
	if Gia_CreateHeader(w, typ, size, data) != nil {
		return 0
	}
	return 1
}
func Gia_ManToBridgeText(w io.Writer, size int32, data []byte) int32 {
	return bridgeSend(w, BRIDGE_TEXT_MESSAGE, size, data)
}
func Gia_ManToBridgeAbort(w io.Writer, size int32, data []byte) int32 {
	return bridgeSend(w, BRIDGE_ABORT, size, data)
}
func Gia_ManToBridgeProgress(w io.Writer, size int32, data []byte) int32 {
	return bridgeSend(w, BRIDGE_PROGRESS, size, data)
}
func Gia_ManToBridgeAbsNetlist(w io.Writer, p *BridgeNetlist, typ int32) int32 {
	data := Gia_ManToBridgeVec(p)
	return bridgeSend(w, typ, int32(len(data)), data)
}
func Gia_ManToBridgeBadAbs(w io.Writer) int32 { return bridgeSend(w, BRIDGE_BAD_ABS, 0, nil) }
func Gia_ManFromBridgeHolds(w io.Writer, po int32) error {
	data := []byte{BRIDGE_VALUE_1, 1}
	data = bridgeAppendUnsigned(data, uint32(po))
	data = append(data, 0)
	return WriteBridgePackage(w, BRIDGE_RESULTS, data)
}
func Gia_ManFromBridgeUnknown(w io.Writer, po int32) error {
	data := bridgeAppendUnsigned([]byte{BRIDGE_VALUE_X, 1}, uint32(po))
	return WriteBridgePackage(w, BRIDGE_RESULTS, data)
}
func Gia_ManFromBridgeCex(w io.Writer, p *Abc_Cex_t) error {
	if p == nil || p.IFrame < 0 || p.IPo < 0 {
		return fmt.Errorf("missing counterexample")
	}
	data := bridgeAppendUnsigned([]byte{BRIDGE_VALUE_0, 1}, uint32(p.IPo))
	data = append(data, 1)
	data = bridgeAppendUnsigned(data, uint32(p.IFrame))
	data = bridgeAppendUnsigned(data, 1)
	data = bridgeAppendUnsigned(data, uint32(p.IFrame+1))
	index := p.NRegs
	for f := int32(0); f <= p.IFrame; f++ {
		data = bridgeAppendUnsigned(data, uint32(p.NPis))
		for i := int32(0); i < p.NPis; i++ {
			value := byte(BRIDGE_VALUE_0)
			if Abc_InfoHasBit(p.PData, index) {
				value = BRIDGE_VALUE_1
			}
			data = append(data, value)
			index++
		}
	}
	assert(index == p.NBits, "bridge CEX bit count")
	data = append(data, 1)
	data = bridgeAppendUnsigned(data, uint32(p.NRegs))
	for i := int32(0); i < p.NRegs; i++ {
		data = append(data, BRIDGE_VALUE_0)
	}
	return WriteBridgePackage(w, BRIDGE_RESULTS, data)
}
func Gia_ManToBridgeResult(w io.Writer, result int32, cex *Abc_Cex_t, po int32) int32 {
	var err error
	switch result {
	case 0:
		err = Gia_ManFromBridgeCex(w, cex)
	case 1:
		err = Gia_ManFromBridgeHolds(w, po)
	case -1:
		err = Gia_ManFromBridgeUnknown(w, po)
	default:
		panic("invalid bridge result")
	}
	if err != nil {
		return 0
	}
	return 1
}
func DecodeBridgeNetlist(data []byte) (p *BridgeNetlist, inits []int32, err error) {
	reader := bytes.NewReader(data)
	read := func() (uint32, error) {
		v, e := binary.ReadUvarint(reader)
		if e != nil {
			return 0, e
		}
		if v > 0xffffffff {
			return 0, fmt.Errorf("bridge integer exceeds 32 bits")
		}
		return uint32(v), nil
	}
	inputs, e := read()
	if e != nil {
		return nil, nil, e
	}
	regs, e := read()
	if e != nil {
		return nil, nil, e
	}
	gates, e := read()
	if e != nil {
		return nil, nil, e
	}
	// Every record needs at least one byte; bound counts before allocation.
	if uint64(inputs)+uint64(regs)+uint64(gates) > 1<<28 || uint64(gates)*2+uint64(regs) > uint64(reader.Len()) {
		return nil, nil, fmt.Errorf("invalid bridge graph dimensions")
	}
	p = &BridgeNetlist{Inputs: int32(inputs), Registers: int32(regs)}
	literals := make([]int32, 2, 2+int(inputs+regs+gates))
	literals[0] = -999
	literals[1] = 1
	for i := uint32(0); i < inputs+regs; i++ {
		literals = append(literals, int32(2*(i+1)))
	}
	literal := func(wire uint32) (int32, error) {
		id := wire >> 1
		if id == 0 || int(id) >= len(literals) {
			return 0, fmt.Errorf("bridge literal references unavailable node %d", id)
		}
		return literals[id] ^ int32(wire&1), nil
	}
	for i := uint32(0); i < gates; i++ {
		a, e := read()
		if e != nil {
			return nil, nil, e
		}
		b, e := read()
		if e != nil {
			return nil, nil, e
		}
		if a&1 != 0 {
			return nil, nil, fmt.Errorf("unsupported bridge gate kind")
		}
		la, e := literal(a >> 1)
		if e != nil {
			return nil, nil, e
		}
		lb, e := literal(b)
		if e != nil {
			return nil, nil, e
		}
		p.Ands = append(p.Ands, [2]int32{la, lb})
		literals = append(literals, int32(2*(inputs+regs+i+1)))
	}
	encodedRegs := make([]uint32, regs)
	for i := range encodedRegs {
		encodedRegs[i], e = read()
		if e != nil {
			return nil, nil, e
		}
	}
	outputs, e := read()
	if e != nil {
		return nil, nil, e
	}
	if uint64(outputs) > uint64(reader.Len()) {
		return nil, nil, fmt.Errorf("invalid bridge output count")
	}
	for i := uint32(0); i < outputs; i++ {
		v, e := read()
		if e != nil {
			return nil, nil, e
		}
		lit, e := literal(v)
		if e != nil {
			return nil, nil, e
		}
		p.Outputs = append(p.Outputs, lit^1)
	}
	// Incoming external packages include fairness metadata; the historic ABC
	// sender omits it. Accept either and reject unsupported fairness/constraints.
	if reader.Len() > 0 {
		version, e := read()
		if e != nil {
			return nil, nil, e
		}
		fair, e := read()
		if e != nil {
			return nil, nil, e
		}
		constraints, e := read()
		if e != nil {
			return nil, nil, e
		}
		if version != 1 || fair != 0 || constraints != 0 || reader.Len() != 0 {
			return nil, nil, fmt.Errorf("unsupported bridge fairness/constraints")
		}
	}
	for _, v := range encodedRegs {
		if v&3 != BRIDGE_VALUE_0 {
			return nil, nil, fmt.Errorf("unsupported nonzero register initialization")
		}
		inits = append(inits, int32(v&3))
		lit, e := literal(v >> 2)
		if e != nil {
			return nil, nil, e
		}
		p.RegisterInputs = append(p.RegisterInputs, lit)
	}
	return p, inits, nil
}
func Gia_ManFromBridgeReadBody(size int32, data []byte, inits *[]int32) (*BridgeNetlist, error) {
	if size < 0 || int(size) > len(data) {
		return nil, fmt.Errorf("bridge size exceeds buffer")
	}
	p, values, err := DecodeBridgeNetlist(data[:size])
	if inits != nil {
		*inits = values
	}
	return p, err
}
func Gia_ManFromBridgeReadPackage(r io.Reader, typ, size *int32, buffer *[]byte) int32 {
	kind, data, err := ReadBridgePackage(r)
	if err != nil {
		return 0
	}
	*typ = kind
	*size = int32(len(data))
	*buffer = data
	return 1
}
func Gia_ManFromBridge(r io.Reader, inits *[]int32) (*BridgeNetlist, error) {
	if _, _, err := ReadBridgePackage(r); err != nil {
		return nil, err
	}
	_, data, err := ReadBridgePackage(r)
	if err != nil {
		return nil, err
	}
	p, values, err := DecodeBridgeNetlist(data)
	if err != nil {
		return nil, err
	}
	if _, _, err = ReadBridgePackage(r); err != nil {
		return nil, err
	}
	if inits != nil {
		*inits = values
	}
	return p, nil
}
func Gia_ManToBridgeAbsNetlistTest(name string, p *BridgeNetlist, typ int32) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	if Gia_ManToBridgeAbsNetlist(f, p, typ) == 0 {
		return io.ErrShortWrite
	}
	return nil
}
func Gia_ManFromBridgeTest(name string) (*BridgeNetlist, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Gia_ManFromBridge(f, nil)
}
