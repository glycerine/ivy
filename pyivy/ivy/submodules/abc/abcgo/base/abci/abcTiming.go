package abci

import (
	"fmt"
	"math"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
)

// Abc_Time_t is ABC's rise/fall timing pair.
type Abc_Time_t = baseabc.Abc_Time_t

const TOLERANCE = 0.001
const abcTimingInfinity float32 = 1.0e30

// Abc_ManTime_t owns the timing annotations attached to one network. The
// manager is stored on the network itself, matching C's pManTime lifetime.
type Abc_ManTime_t struct {
	network                  *Abc_Ntk_t
	tArrDef, tReqDef         Abc_Time_t
	tInDriveDef, tOutLoadDef Abc_Time_t
	arrivals, required       []Abc_Time_t
	inDrive, outLoad         []Abc_Time_t
	stopped                  bool
}

func timingCIs(n *Abc_Ntk_t) []*Abc_Obj_t {
	v := make([]*Abc_Obj_t, 0)
	for i := int32(0); i < baseabc.Abc_NtkCiNum(n); i++ {
		v = append(v, baseabc.Abc_NtkCi(n, i))
	}
	return v
}
func timingCOs(n *Abc_Ntk_t) []*Abc_Obj_t {
	v := make([]*Abc_Obj_t, 0)
	for i := int32(0); i < baseabc.Abc_NtkCoNum(n); i++ {
		v = append(v, baseabc.Abc_NtkCo(n, i))
	}
	return v
}
func timingManager(n *Abc_Ntk_t) *Abc_ManTime_t {
	if n == nil {
		return nil
	}
	m, _ := baseabc.NetworkTimingManager(n).(*Abc_ManTime_t)
	return m
}
func ensureTiming(n *Abc_Ntk_t) *Abc_ManTime_t {
	if m := timingManager(n); m != nil {
		return m
	}
	m := &Abc_ManTime_t{network: n, tReqDef: Abc_Time_t{}}
	baseabc.NetworkSetTimingManager(n, m)
	m.expand(baseabc.Abc_NtkObjNumMax(n)+1, false)
	for _, o := range timingCIs(n) {
		m.arrivals[o.ID] = m.tArrDef
	}
	for _, o := range timingCOs(n) {
		m.required[o.ID] = m.tReqDef
	}
	return m
}
func (m *Abc_ManTime_t) expand(size int32, progressive bool) {
	if m == nil || size <= int32(len(m.arrivals)) {
		return
	}
	n := int(size)
	if progressive {
		n *= 2
	}
	if n < 100 {
		n = 100
	}
	a := make([]Abc_Time_t, n)
	copy(a, m.arrivals)
	for i := len(m.arrivals); i < n; i++ {
		a[i] = Abc_Time_t{-abcTimingInfinity, -abcTimingInfinity}
	}
	m.arrivals = a
	r := make([]Abc_Time_t, n)
	copy(r, m.required)
	for i := len(m.required); i < n; i++ {
		r[i] = Abc_Time_t{abcTimingInfinity, abcTimingInfinity}
	}
	m.required = r
}
func (m *Abc_ManTime_t) Stop() {
	if m == nil || m.stopped {
		return
	}
	m.stopped = true
	m.arrivals = nil
	m.required = nil
	m.inDrive = nil
	m.outLoad = nil
	if m.network != nil && baseabc.NetworkTimingManager(m.network) == m {
		baseabc.NetworkSetTimingManager(m.network, nil)
	}
	m.network = nil
}
func Abc_FloatEqual(x, y float32) int32 {
	if math.Abs(float64(x-y)) < TOLERANCE {
		return 1
	}
	return 0
}
func Abc_ManTimeStart(n *Abc_Ntk_t) *Abc_ManTime_t { return ensureTiming(n) }
func Abc_ManTimeStop(m *Abc_ManTime_t) {
	if m != nil {
		m.Stop()
	}
}
func Abc_ManTimeExpand(m *Abc_ManTime_t, size, progressive int32) {
	if m != nil {
		m.expand(size, progressive != 0)
	}
}
func Abc_ManTimeDup(old, n *Abc_Ntk_t) {
	s := timingManager(old)
	if s == nil || n == nil {
		return
	}
	d := ensureTiming(n)
	d.tArrDef = s.tArrDef
	d.tReqDef = s.tReqDef
	d.tInDriveDef = s.tInDriveDef
	d.tOutLoadDef = s.tOutLoadDef
	for i, o := range timingCIs(n) {
		if i < len(timingCIs(old)) {
			d.arrivals[o.ID] = s.arrivals[timingCIs(old)[i].ID]
		}
	}
	for i, o := range timingCOs(n) {
		if i < len(timingCOs(old)) {
			d.required[o.ID] = s.required[timingCOs(old)[i].ID]
		}
	}
	d.inDrive = append([]Abc_Time_t(nil), s.inDrive...)
	d.outLoad = append([]Abc_Time_t(nil), s.outLoad...)
}
func Abc_NodeArrival(o *Abc_Obj_t) *Abc_Time_t {
	if o == nil {
		return nil
	}
	m := timingManager(o.Network())
	if m == nil || o.ID >= len(m.arrivals) {
		return nil
	}
	return &m.arrivals[o.ID]
}
func Abc_NodeRequired(o *Abc_Obj_t) *Abc_Time_t {
	if o == nil {
		return nil
	}
	m := timingManager(o.Network())
	if m == nil || o.ID >= len(m.required) {
		return nil
	}
	return &m.required[o.ID]
}
func Abc_NtkReadDefaultArrival(n *Abc_Ntk_t) *Abc_Time_t {
	m := timingManager(n)
	if m == nil {
		return nil
	}
	return &m.tArrDef
}
func Abc_NtkReadDefaultRequired(n *Abc_Ntk_t) *Abc_Time_t {
	m := timingManager(n)
	if m == nil {
		return nil
	}
	return &m.tReqDef
}
func Abc_NodeReadArrival(o *Abc_Obj_t) *Abc_Time_t {
	return Abc_NodeArrival(o)
}
func Abc_NodeReadRequired(o *Abc_Obj_t) *Abc_Time_t {
	return Abc_NodeRequired(o)
}
func Abc_NtkReadDefaultArrivalWorst(n *Abc_Ntk_t) float32 {
	t := Abc_NtkReadDefaultArrival(n)
	if t == nil {
		return 0
	}
	return maxF(t.Rise, t.Fall)
}
func Abc_NtkReadDefaultRequiredWorst(n *Abc_Ntk_t) float32 {
	t := Abc_NtkReadDefaultRequired(n)
	if t == nil {
		return 0
	}
	return minF(t.Rise, t.Fall)
}
func Abc_NodeReadArrivalAve(o *Abc_Obj_t) float32 {
	t := Abc_NodeReadArrival(o)
	return (t.Rise + t.Fall) * .5
}
func Abc_NodeReadRequiredAve(o *Abc_Obj_t) float32 {
	t := Abc_NodeReadRequired(o)
	return (t.Rise + t.Fall) * .5
}
func Abc_NodeReadArrivalWorst(o *Abc_Obj_t) float32 {
	t := Abc_NodeReadArrival(o)
	return maxF(t.Rise, t.Fall)
}
func Abc_NodeReadRequiredWorst(o *Abc_Obj_t) float32 {
	t := Abc_NodeReadRequired(o)
	return minF(t.Rise, t.Fall)
}
func maxF(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
func minF(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
func Abc_NtkTimeSetDefaultArrival(n *Abc_Ntk_t, r, f float32) {
	m := ensureTiming(n)
	m.tArrDef = Abc_Time_t{r, f}
	for _, o := range timingCIs(n) {
		m.expand(int32(o.ID+1), true)
		m.arrivals[o.ID] = m.tArrDef
	}
}
func Abc_NtkTimeSetDefaultRequired(n *Abc_Ntk_t, r, f float32) {
	m := ensureTiming(n)
	m.tReqDef = Abc_Time_t{r, f}
	for _, o := range timingCOs(n) {
		m.expand(int32(o.ID+1), true)
		m.required[o.ID] = m.tReqDef
	}
}
func Abc_NtkTimeSetArrival(n *Abc_Ntk_t, id int32, r, f float32) {
	if n == nil || id < 0 {
		return
	}
	m := ensureTiming(n)
	m.expand(id+1, true)
	m.arrivals[id] = Abc_Time_t{r, f}
}
func Abc_NtkTimeSetRequired(n *Abc_Ntk_t, id int32, r, f float32) {
	if n == nil || id < 0 {
		return
	}
	m := ensureTiming(n)
	m.expand(id+1, true)
	m.required[id] = Abc_Time_t{r, f}
}
func Abc_NtkReadDefaultInputDrive(n *Abc_Ntk_t) *Abc_Time_t {
	m := timingManager(n)
	if m == nil {
		return nil
	}
	return &m.tInDriveDef
}
func Abc_NtkReadDefaultOutputLoad(n *Abc_Ntk_t) *Abc_Time_t {
	m := timingManager(n)
	if m == nil {
		return nil
	}
	return &m.tOutLoadDef
}
func Abc_NodeReadInputDrive(n *Abc_Ntk_t, i int32) *Abc_Time_t {
	m := timingManager(n)
	if m == nil {
		return nil
	}
	if i < 0 || i >= int32(len(timingCIs(n))) {
		return nil
	}
	if len(m.inDrive) == 0 {
		return nil
	}
	return &m.inDrive[i]
}
func Abc_NodeReadOutputLoad(n *Abc_Ntk_t, i int32) *Abc_Time_t {
	m := timingManager(n)
	if m == nil {
		return nil
	}
	if i < 0 || i >= int32(len(timingCOs(n))) {
		return nil
	}
	if len(m.outLoad) == 0 {
		return nil
	}
	return &m.outLoad[i]
}
func Abc_NodeReadInputDriveWorst(n *Abc_Ntk_t, i int32) float32 {
	t := Abc_NodeReadInputDrive(n, i)
	if t == nil {
		return 0
	}
	return maxF(t.Rise, t.Fall)
}
func Abc_NodeReadOutputLoadWorst(n *Abc_Ntk_t, i int32) float32 {
	t := Abc_NodeReadOutputLoad(n, i)
	if t == nil {
		return 0
	}
	return maxF(t.Rise, t.Fall)
}
func Abc_NtkTimeSetDefaultInputDrive(n *Abc_Ntk_t, r, f float32) {
	m := ensureTiming(n)
	m.tInDriveDef = Abc_Time_t{r, f}
	for i := range m.inDrive {
		if m.inDrive[i] == (Abc_Time_t{}) {
			m.inDrive[i] = m.tInDriveDef
		}
	}
}
func Abc_NtkTimeSetDefaultOutputLoad(n *Abc_Ntk_t, r, f float32) {
	m := ensureTiming(n)
	m.tOutLoadDef = Abc_Time_t{r, f}
	for i := range m.outLoad {
		if m.outLoad[i] == (Abc_Time_t{}) {
			m.outLoad[i] = m.tOutLoadDef
		}
	}
}
func Abc_NtkTimeSetInputDrive(n *Abc_Ntk_t, i int32, r, f float32) {
	if n == nil || i < 0 || i >= int32(len(timingCIs(n))) {
		return
	}
	m := ensureTiming(n)
	if len(m.inDrive) == 0 {
		m.inDrive = make([]Abc_Time_t, len(timingCIs(n)))
		for j := range m.inDrive {
			m.inDrive[j] = m.tInDriveDef
		}
	}
	m.inDrive[i] = Abc_Time_t{r, f}
}
func Abc_NtkTimeSetOutputLoad(n *Abc_Ntk_t, i int32, r, f float32) {
	if n == nil || i < 0 || i >= int32(len(timingCOs(n))) {
		return
	}
	m := ensureTiming(n)
	if len(m.outLoad) == 0 {
		m.outLoad = make([]Abc_Time_t, len(timingCOs(n)))
		for j := range m.outLoad {
			m.outLoad[j] = m.tOutLoadDef
		}
	}
	m.outLoad[i] = Abc_Time_t{r, f}
}
func Abc_NtkTimeInitialize(n, old *Abc_Ntk_t) {
	if n == nil {
		return
	}
	m := timingManager(n)
	if m == nil {
		return
	}
	m.expand(baseabc.Abc_NtkObjNumMax(n), false)
	if old != nil {
		if s := timingManager(old); s != nil {
			m.tArrDef = s.tArrDef
			m.tReqDef = s.tReqDef
		}
	}
	for i, o := range timingCIs(n) {
		if old != nil && i < len(timingCIs(old)) {
			if t := Abc_NodeReadArrival(timingCIs(old)[i]); t != nil {
				m.arrivals[o.ID] = *t
			}
		} else {
			m.arrivals[o.ID] = m.tArrDef
		}
	}
	for i, o := range timingCOs(n) {
		if old != nil && i < len(timingCOs(old)) {
			if t := Abc_NodeReadRequired(timingCOs(old)[i]); t != nil {
				m.required[o.ID] = *t
			}
		} else {
			m.required[o.ID] = m.tReqDef
		}
	}
}
func Abc_NtkTimeScale(n *Abc_Ntk_t, s float32) {
	m := timingManager(n)
	if m == nil {
		return
	}
	scale := func(t *Abc_Time_t) { t.Rise *= s; t.Fall *= s }
	scale(&m.tArrDef)
	scale(&m.tReqDef)
	for _, o := range timingCIs(n) {
		scale(&m.arrivals[o.ID])
	}
	for _, o := range timingCOs(n) {
		scale(&m.required[o.ID])
	}
}
func Abc_NtkTimePrepare(n *Abc_Ntk_t) {
	if n == nil {
		return
	}
	m := ensureTiming(n)
	m.expand(baseabc.Abc_NtkObjNumMax(n), false)
	for _, o := range n.Objects() {
		if o.Kind == baseabc.ObjectNode {
			m.arrivals[o.ID] = Abc_Time_t{-abcTimingInfinity, -abcTimingInfinity}
			m.required[o.ID] = Abc_Time_t{abcTimingInfinity, abcTimingInfinity}
		}
		if o.IsCO() {
			m.arrivals[o.ID] = Abc_Time_t{-abcTimingInfinity, -abcTimingInfinity}
		}
		if o.IsCI() {
			m.required[o.ID] = Abc_Time_t{abcTimingInfinity, abcTimingInfinity}
		}
	}
}
func Abc_NtkTimePrint(n *Abc_Ntk_t) {
	if timingManager(n) == nil {
		fmt.Println("There is no timing manager")
		return
	}
	fmt.Printf("Default arrival = %8f\nDefault required = %8f\n", Abc_NtkReadDefaultArrival(n).Fall, Abc_NtkReadDefaultRequired(n).Fall)
	fmt.Printf("Inputs (%d):\n", len(timingCIs(n)))
	for _, o := range timingCIs(n) {
		fmt.Printf("%20s   arrival = %8f   required = %8f\n", baseabc.Abc_ObjName(o), Abc_NodeReadArrivalWorst(o), Abc_NodeReadRequiredWorst(o))
	}
	fmt.Printf("Outputs (%d):\n", len(timingCOs(n)))
	for _, o := range timingCOs(n) {
		fmt.Printf("%20s   arrival = %8f   required = %8f\n", baseabc.Abc_ObjName(o), Abc_NodeReadArrivalWorst(o), Abc_NodeReadRequiredWorst(o))
	}
}
func Abc_NtkGetCiArrivalTimes(n *Abc_Ntk_t) *Abc_Time_t {
	v := make([]Abc_Time_t, len(timingCIs(n)))
	for i, o := range timingCIs(n) {
		if t := Abc_NodeArrival(o); t != nil {
			v[i] = *t
		}
	}
	if len(v) == 0 {
		return nil
	}
	return &v[0]
}
func Abc_NtkGetCoRequiredTimes(n *Abc_Ntk_t) *Abc_Time_t {
	v := make([]Abc_Time_t, len(timingCOs(n)))
	for i, o := range timingCOs(n) {
		if t := Abc_NodeRequired(o); t != nil {
			v[i] = *t
		}
	}
	if len(v) == 0 {
		return nil
	}
	return &v[0]
}
func Abc_NtkGetCiArrivalFloats(n *Abc_Ntk_t) []float32 {
	v := make([]float32, len(timingCIs(n)))
	if timingManager(n) == nil {
		return v
	}
	for i, o := range timingCIs(n) {
		v[i] = Abc_NodeReadArrivalWorst(o)
	}
	return v
}
func Abc_NtkGetCoRequiredFloats(n *Abc_Ntk_t) []float32 {
	if timingManager(n) == nil {
		return nil
	}
	v := make([]float32, len(timingCOs(n)))
	for i, o := range timingCOs(n) {
		v[i] = Abc_NodeReadRequiredWorst(o)
	}
	return v
}

// Edge slacks are stored as IEEE float bits in Vec_Int_t, as in Abc_Int2Float.
func Abc_NtkDelayTraceSlackStart(n *Abc_Ntk_t) *Vec_Int_t {
	if n == nil {
		return nil
	}
	v := vec.Vec_IntAlloc(int(baseabc.Abc_NtkObjNumMax(n)))
	vec.Vec_IntFill(v, int(baseabc.Abc_NtkObjNumMax(n)), -1)
	for _, o := range n.ObjectsOfKind(baseabc.ObjectNode) {
		vec.Vec_IntWriteEntry(v, o.ID, int32(vec.Vec_IntSize(v)))
		for range o.Fanins() {
			vec.Vec_IntPush(v, -1)
		}
	}
	return v
}
func Abc_NtkDelayTraceSlack(v *Vec_Int_t, o *Abc_Obj_t, i int32) float32 {
	return math.Float32frombits(uint32(vec.Vec_IntEntry(v, int(vec.Vec_IntEntry(v, o.ID)+i))))
}
func Abc_NtkDelayTraceSetSlack(v *Vec_Int_t, o *Abc_Obj_t, i int32, x float32) {
	vec.Vec_IntWriteEntry(v, int(vec.Vec_IntEntry(v, o.ID)+i), int32(math.Float32bits(x)))
}
func Abc_NtkDelayTraceCritPath_rec(slacks *Vec_Int_t, node, leaf *Abc_Obj_t, best *Vec_Int_t) int32 {
	if node == nil {
		return 0
	}
	if node.IsCI() {
		return bool32(leaf == nil || leaf == node)
	}
	if node.Kind != baseabc.ObjectNode {
		return 0
	}
	if baseabc.Abc_NodeIsTravIdCurrent(node) != 0 {
		return bool32(vec.Vec_IntEntry(best, node.ID) >= 0)
	}
	baseabc.Abc_NodeSetTravIdCurrent(node)
	bestIdx := -1
	minSlack := float32(math.MaxFloat32)
	for i, e := range node.Fanins() {
		if e.Object == nil || Abc_NtkDelayTraceCritPath_rec(slacks, e.Object, leaf, best) == 0 {
			continue
		}
		s := Abc_NtkDelayTraceSlack(slacks, node, int32(i))
		if bestIdx < 0 || s < minSlack {
			bestIdx = i
			minSlack = s
		}
	}
	if bestIdx >= 0 {
		vec.Vec_IntWriteEntry(best, node.ID, int32(bestIdx))
	}
	return bool32(bestIdx >= 0)
}
func bool32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func Abc_NtkDelayTraceCritPathCollect_rec(slacks *Vec_Int_t, node *Abc_Obj_t, best *Vec_Int_t, path *Vec_Ptr_t) {
	if node == nil {
		return
	}
	if node.Kind == baseabc.ObjectNode {
		i := vec.Vec_IntEntry(best, node.ID)
		if i >= 0 && int(i) < len(node.Fanins()) {
			Abc_NtkDelayTraceCritPathCollect_rec(slacks, node.Fanins()[i].Object, best, path)
		}
	}
	vec.Vec_PtrPush(path, node)
}

// Abc_NodeDelayTraceArrival propagates rise/fall through the node's actual MIO
// pin phase and block-delay table, preserving the source timing calculation.
func Abc_NodeDelayTraceArrival(node *Abc_Obj_t, slacks *Vec_Int_t) {
	if node == nil || node.Kind != baseabc.ObjectNode {
		return
	}
	out := Abc_NodeArrival(node)
	if out == nil {
		return
	}
	*out = Abc_Time_t{Rise: -abcTimingInfinity, Fall: -abcTimingInfinity}
	if baseabc.Abc_ObjIsBarBuf(node) != 0 && len(node.Fanins()) > 0 {
		if in := Abc_NodeArrival(node.Fanins()[0].Object); in != nil {
			*out = *in
		}
		return
	}
	gate, ok := node.Data.(*mio.Mio_Gate_t)
	if !ok || gate == nil {
		return
	}
	pin := mio.Mio_GateReadPins(gate)
	for i, e := range node.Fanins() {
		if e.Object == nil || pin == nil {
			break
		}
		in := Abc_NodeArrival(e.Object)
		if in == nil {
			pin = mio.Mio_PinReadNext(pin)
			continue
		}
		phase := mio.Mio_PinReadPhase(pin)
		dr, df := float32(mio.Mio_PinReadDelayBlockRise(pin)), float32(mio.Mio_PinReadDelayBlockFall(pin))
		if phase != mio.MIO_PHASE_INV {
			out.Rise = maxF(out.Rise, in.Rise+dr)
			out.Fall = maxF(out.Fall, in.Fall+df)
		}
		if phase != mio.MIO_PHASE_NONINV {
			out.Rise = maxF(out.Rise, in.Fall+dr)
			out.Fall = maxF(out.Fall, in.Rise+df)
		}
		if slacks != nil {
			slack := float32(math.MaxFloat32)
			if phase != mio.MIO_PHASE_INV {
				slack = minF(slack, float32(math.Abs(float64(in.Rise+dr-out.Rise))))
				slack = minF(slack, float32(math.Abs(float64(in.Fall+df-out.Fall))))
			}
			if phase != mio.MIO_PHASE_NONINV {
				slack = minF(slack, float32(math.Abs(float64(in.Fall+dr-out.Rise))))
				slack = minF(slack, float32(math.Abs(float64(in.Rise+df-out.Fall))))
			}
			Abc_NtkDelayTraceSetSlack(slacks, node, int32(i), slack)
		}
		pin = mio.Mio_PinReadNext(pin)
	}
}
func Abc_NtkSetNodeLevelsArrival(old *Abc_Ntk_t) {
	if old == nil {
		return
	}
	for _, o := range timingCIs(old) {
		if o.Copy != nil {
			o.Copy.Level = int(Abc_NodeReadArrivalWorst(o))
		}
	}
}
func Abc_NtkDelayTrace(n *Abc_Ntk_t, out, in *Abc_Obj_t, print int32) float32 {
	if n == nil {
		return 0
	}
	Abc_NtkTimePrepare(n)
	var slacks *Vec_Int_t
	if out != nil || in != nil || print != 0 {
		slacks = Abc_NtkDelayTraceSlackStart(n)
	}
	for _, o := range baseabc.Abc_NtkDfs(n, 1) {
		Abc_NodeDelayTraceArrival(o, slacks)
	}
	worst := float32(0)
	for _, o := range timingCOs(n) {
		if f := o.Fanins(); len(f) > 0 && f[0].Object != nil {
			worst = maxF(worst, Abc_NodeReadArrivalWorst(f[0].Object))
		}
	}
	if slacks != nil {
		vec.Vec_IntFree(slacks)
	}
	return worst
}

// Incremental level maintenance is owned by base/abc; these adapters preserve
// the abci C-facing entry points without duplicating the graph bookkeeping.
func Abc_ObjLevelNew(o *Abc_Obj_t) int32               { return baseabc.Abc_ObjLevelNew(o) }
func Abc_ObjReverseLevelNew(o *Abc_Obj_t) int32        { return baseabc.Abc_ObjReverseLevelNew(o) }
func Abc_ObjRequiredLevel(o *Abc_Obj_t) int32          { return baseabc.Abc_ObjRequiredLevel(o) }
func Abc_ObjReverseLevel(o *Abc_Obj_t) int32           { return baseabc.Abc_ObjReverseLevel(o) }
func Abc_ObjSetReverseLevel(o *Abc_Obj_t, level int32) { baseabc.Abc_ObjSetReverseLevel(o, level) }
func Abc_NtkStartReverseLevels(n *Abc_Ntk_t, increase int32) {
	baseabc.Abc_NtkStartReverseLevels(n, increase)
}
func Abc_NtkStopReverseLevels(n *Abc_Ntk_t)              { baseabc.Abc_NtkStopReverseLevels(n) }
func Abc_NtkUpdateLevel(o *Abc_Obj_t, levels *Vec_Vec_t) { baseabc.Abc_NtkUpdateLevel(o, levels) }
func Abc_NtkUpdateReverseLevel(o *Abc_Obj_t, levels *Vec_Vec_t) {
	baseabc.Abc_NtkUpdateReverseLevel(o, levels)
}
func Abc_NtkUpdate(old, replacement *Abc_Obj_t, levels *Vec_Vec_t) {
	baseabc.Abc_NtkUpdate(old, replacement, levels)
}
