package abc

// Disabled ISCAS EDIF reader boundary from src/base/io/ioReadEdif.c.

// ReadABCEDIF is the owned-byte counterpart of Io_ReadEdif. The source returns
// unconditionally before opening its file or consulting fCheck, so the Go port
// likewise reports unsupported without inspecting data, check, limits or Stop.
// The file-local Io_ReadEdifNetwork body is unreachable in this source snapshot
// and has no other caller; implementing it would expose behavior ABC itself
// deliberately disables.
func ReadABCEDIF(data []byte, check bool, limits IOLimits) (*ABCNetwork, error) {
	return nil, ErrUnsupported
}
