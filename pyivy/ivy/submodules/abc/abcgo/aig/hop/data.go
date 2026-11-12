package hop

// Hop_ObjData accesses caller-owned conversion scratch on the regular node.
// Complemented edges share that scratch with their regular node.
func Hop_ObjData(obj *Hop_Obj_t) any              { return Hop_Regular(obj).pData }
func Hop_ObjSetData(obj *Hop_Obj_t, value any)    { Hop_Regular(obj).pData = value }
func Hop_ObjIData(obj *Hop_Obj_t) int32           { return Hop_Regular(obj).iData }
func Hop_ObjSetIData(obj *Hop_Obj_t, value int32) { Hop_Regular(obj).iData = value }
