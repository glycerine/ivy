package csat

// GateType identifies a Boolean gate accepted by ABC_AddGate.
type GateType int32

const (
	CSAT_CONST GateType = iota
	CSAT_BPI
	CSAT_BPPI
	CSAT_BAND
	CSAT_BNAND
	CSAT_BOR
	CSAT_BNOR
	CSAT_BXOR
	CSAT_BXNOR
	CSAT_BINV
	CSAT_BBUF
	CSAT_BMUX
	CSAT_BDFF
	CSAT_BSDFF
	CSAT_BTRIH
	CSAT_BTRIL
	CSAT_BBUS
	CSAT_BPPO
	CSAT_BPO
	CSAT_BCNF
	CSAT_BDC
)

type CSAT_StatusT int32

const (
	UNDETERMINED CSAT_StatusT = iota
	UNSATISFIABLE
	SATISFIABLE
	TIME_OUT
	FRAME_OUT
	NO_TARGET
	ABORTED
	SEQ_SATISFIABLE
)

type CSAT_CallerT int32

const (
	BLS CSAT_CallerT = iota
	SATORI
	NONE
)

type CSAT_OptionT int32

const (
	BASE_LINE CSAT_OptionT = iota
	IMPLICT_LEARNING
	EXPLICT_LEARNING
)

type CSAT_Target_ResultT struct {
	Status  CSAT_StatusT
	NumDec  int32
	NumImp  int32
	NumCftg int32
	NumCfts int32
	Time    float64
	NoSig   int32
	Names   []string
	Values  []int32
}

type ABC_UINT64_T = uint64

type ABC_ManagerStruct_t struct {
	// Fields are private so the manager maintains graph and solver invariants.
	state *managerState
}

type ABC_Manager_t = ABC_ManagerStruct_t
type ABC_Manager = *ABC_ManagerStruct_t
