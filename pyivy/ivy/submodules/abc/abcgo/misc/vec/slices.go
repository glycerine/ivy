package vec

// storage exposes capacity for reservation/fill operations; data itself always
// contains the logical live prefix. No size or capacity is cached.

func (p *Vec_Flt_t) storage() []float32 { return p.data[:cap(p.data)] }
func (p *Vec_Flt_t) resize(n int)       { p.data = p.data[:n] }

func (p *Vec_Ptr_t) storage() []any { return p.data[:cap(p.data)] }
func (p *Vec_Ptr_t) resize(n int)   { p.data = p.data[:n] }

func (p *Vec_Wrd_t) storage() []uint64 { return p.data[:cap(p.data)] }
func (p *Vec_Wrd_t) resize(n int)      { p.data = p.data[:n] }

func (p *Vec_Str_t) storage() []byte { return p.data[:cap(p.data)] }
func (p *Vec_Str_t) resize(n int)    { p.data = p.data[:n] }

func (p *Vec_Wec_t) storage() []Vec_Int_t { return p.data[:cap(p.data)] }
func (p *Vec_Wec_t) resize(n int)         { p.data = p.data[:n] }

func (p *Vec_Int_t) storage() []int32 { return p.data[:cap(p.data)] }
func (p *Vec_Int_t) resize(n int)     { p.data = p.data[:n] }
