package abcmain

import "sync"

// ABCFrameState stores opaque ABC-owned frame values without importing the
// packages that define their concrete types. Replacement methods model the
// pointer-to-pointer ownership transfer used by ABC: the input is cleared.
type ABCFrameState struct {
	mu       sync.RWMutex
	gia      any
	cex      any
	cexVec   any
	poEquivs any
	statuses any
	status   int32
	frames   int32
}

// ReadABCFrameState returns the frame's shared opaque state container.
func ReadABCFrameState(f *Frame) *ABCFrameState {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.abcState == nil {
		f.abcState = &ABCFrameState{}
	}
	return f.abcState
}

func (s *ABCFrameState) Gia() any     { s.mu.RLock(); defer s.mu.RUnlock(); return s.gia }
func (s *ABCFrameState) SetGia(v any) { s.mu.Lock(); s.gia = v; s.mu.Unlock() }
func (s *ABCFrameState) TakeGia() any {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := s.gia
	s.gia = nil
	return v
}
func (s *ABCFrameState) Cex() any      { s.mu.RLock(); defer s.mu.RUnlock(); return s.cex }
func (s *ABCFrameState) CexVec() any   { s.mu.RLock(); defer s.mu.RUnlock(); return s.cexVec }
func (s *ABCFrameState) PoEquivs() any { s.mu.RLock(); defer s.mu.RUnlock(); return s.poEquivs }
func (s *ABCFrameState) Statuses() any { s.mu.RLock(); defer s.mu.RUnlock(); return s.statuses }

func consume(v *any) any {
	if v == nil {
		return nil
	}
	x := *v
	*v = nil
	return x
}

func (s *ABCFrameState) ReplaceCex(v *any) {
	x := consume(v)
	s.mu.Lock()
	s.cex = x
	s.cexVec = nil
	s.mu.Unlock()
}
func (s *ABCFrameState) ReplaceCexVec(v *any) {
	x := consume(v)
	s.mu.Lock()
	s.cexVec = x
	s.cex = nil
	s.mu.Unlock()
}
func (s *ABCFrameState) ReplacePoEquivs(v *any) {
	x := consume(v)
	s.mu.Lock()
	s.poEquivs = x
	s.mu.Unlock()
}
func (s *ABCFrameState) ReplaceStatuses(v *any) {
	x := consume(v)
	s.mu.Lock()
	s.statuses = x
	s.mu.Unlock()
}
func (s *ABCFrameState) VerificationStatus() (int32, int32) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status, s.frames
}
func (s *ABCFrameState) SetVerificationStatus(status, frames int32) {
	s.mu.Lock()
	s.status, s.frames = status, frames
	s.mu.Unlock()
}
