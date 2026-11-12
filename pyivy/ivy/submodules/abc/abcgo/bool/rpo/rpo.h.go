package rpo

// Rpo_LCI_Edge_t is a grouping candidate. Go fields replace the C bitfields;
// node indices refer to positions in the current literal-cluster slice.
type Rpo_LCI_Edge_t struct {
	Visited        bool
	ConnectionType Operator_t
	Idx1, Idx2     int32
}
