package cnf

import "sync"

var cnfMsopOnce sync.Once
var cnfMsopSizes []byte
var cnfMsops []string

// cnfReadMsops adapts the shared ABC cover table to the manager's source-style
// parallel SOP and size arrays without decoding the compressed source twice.
func cnfReadMsops() ([]byte, []string) {
	cnfMsopOnce.Do(func() {
		covers := cnfMsopCovers()
		cnfMsopSizes = make([]byte, len(covers))
		cnfMsops = make([]string, len(covers))
		for i, cover := range covers {
			cnfMsopSizes[i] = byte(len(cover))
			cnfMsops[i] = string(cover)
		}
	})
	return cnfMsopSizes, cnfMsops
}
