// Package abcmain owns the ABC framework lifecycle, state and command registry.
// Component packages register their commands and adapters without an import cycle.
package abcmain

import (
	"bufio"
	"io"
	"sync"
	"time"

	"github.com/glycerine/abcgo/bool/dec"
)

const ABC_VERSION = "UC Berkeley, ABC 1.01"
const ABC_MAX_STR = 1 << 15

type Command struct {
	Name, Group    string
	Handler        func(*Frame, []string) int32
	ChangesNetwork bool
}
type Dispatcher interface{ Execute(*Frame, string) int32 }
type NetworkOperations interface {
	Clone(any) (any, error)
	Close(any) error
	Unmap(any) error
	Outputs(any) int
}
type DataBridge interface {
	Import(*Frame, string, any) error
	Export(*Frame, string) (any, error)
}
type TimingBridge interface {
	SetTiming(*Frame, string, int, float32, float32) error
}
type ResourceLimits interface {
	SetMemoryLimit(int64) error
	SetCPULimit(time.Duration) error
}
type Counterexample struct {
	PIs, Registers, Output, Frame int32
	Bits                          []uint64
}
type networkEntry struct {
	value any
	step  int32
}

// Frame keeps undo history outside concrete network objects. All component
// resources stay owned by the frame until a read-and-take API transfers them.
type Frame struct {
	mu                         sync.RWMutex
	lifecycle                  sync.Mutex
	initialized, ended, closed bool
	closing                    bool
	initializers               []*Abc_FrameInitializer_t
	commands                   map[string]Command
	aliases                    map[string][]string
	flags                      map[string]string
	history                    []string
	resources                  map[string]any
	strings                    map[string]string
	ints                       map[string][]int32
	lists                      map[string][]any
	rows                       map[string][][]int32
	current                    networkEntry
	backups                    []networkEntry
	steps                      int32
	manDec                     *dec.Dec_Man_t
	inputMu                    sync.Mutex
	inputReader                *bufio.Reader
	inputSource                io.Reader
	sourceDepth                int
	autoexec                   bool
	Version, Binary            string
	In                         io.Reader
	Out, Err                   io.Writer
	Dispatcher                 Dispatcher
	Networks                   NetworkOperations
	Bridge                     DataBridge
	Timing                     TimingBridge
	Limits                     ResourceLimits
	ModelChecker               ModelChecker
	LibraryLoader              LibraryLoader
	libraries                  []*Library
	Status, NFrames            int32
	Cex                        *Counterexample
	abcState                   *ABCFrameState
	BatchMode, BridgeMode      bool
	MaxLoad                    float32
	TimeCommand, TimeTotal     time.Duration
}
type Abc_Frame_t = Frame
type Abc_Frame_t_ = Frame
type Abc_FrameInitializer_t struct{ Init, Destroy func(*Frame) }
type Abc_Frame_Callback_BmcFrameDone_Func func(frame, output, status int32)
