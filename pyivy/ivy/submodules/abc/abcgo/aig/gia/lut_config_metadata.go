package gia

// Gia_ManLutConfigs returns the manager's LUT configuration metadata.
func Gia_ManLutConfigs(p *Gia_Man_t) *Vec_Int_t { return p.vLutConfigs }

// Gia_ManSetLutConfigs installs LUT configuration metadata on the manager.
func Gia_ManSetLutConfigs(p *Gia_Man_t, values *Vec_Int_t) { p.vLutConfigs = values }
