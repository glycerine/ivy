package acb

import "github.com/glycerine/abcgo/misc/vec"

// ACB's source indices are C int values, represented as int32 here. The
// shared Go vector package indexes slices with int.
func Vec_IntAlloc(n int32) *Vec_Int_t                { return vec.Vec_IntAlloc(int(n)) }
func Vec_IntDup(v *Vec_Int_t) *Vec_Int_t             { return vec.Vec_IntDup(v) }
func Vec_IntFree(v *Vec_Int_t)                       { vec.Vec_IntFree(v) }
func Vec_IntSize(v *Vec_Int_t) int32                 { return int32(vec.Vec_IntSize(v)) }
func Vec_IntEntry(v *Vec_Int_t, i int32) int32       { return vec.Vec_IntEntry(v, int(i)) }
func Vec_IntWriteEntry(v *Vec_Int_t, i, value int32) { vec.Vec_IntWriteEntry(v, int(i), value) }
func Vec_IntPush(v *Vec_Int_t, value int32)          { vec.Vec_IntPush(v, value) }
func Vec_IntClear(v *Vec_Int_t)                      { vec.Vec_IntClear(v) }
func Vec_IntArray(v *Vec_Int_t) []int32              { return vec.Vec_IntArray(v) }
func Vec_StrSize(v *Vec_Str_t) int32                 { return int32(vec.Vec_StrSize(v)) }
func Vec_PtrSize(v *Vec_Ptr_t) int32                 { return int32(vec.Vec_PtrSize(v)) }
func Vec_PtrEntry(v *Vec_Ptr_t, i int32) any         { return vec.Vec_PtrEntry(v, int(i)) }
func Vec_IntCap(v *Vec_Int_t) int32                  { return int32(vec.Vec_IntCap(v)) }
func Vec_StrCap(v *Vec_Str_t) int32                  { return int32(vec.Vec_StrCap(v)) }
func Vec_StrCountEntry(v *Vec_Str_t, x byte) int32   { return int32(vec.Vec_StrCountEntry(v, x)) }
func Vec_StrCountLarger(v *Vec_Str_t, x byte) int32  { return int32(vec.Vec_StrCountLarger(v, x)) }
func Vec_IntFill(v *Vec_Int_t, n int32, x int32)     { vec.Vec_IntFill(v, int(n), x) }
func Hash_Int2ManInsert(v *Hash_IntMan_t, a, b, data int32) int32 {
	return vec.Hash_Int2ManInsert(v, a, b, data)
}
func Hash_IntObjData0(v *Hash_IntMan_t, id int32) int32  { return vec.Hash_IntObjData0(v, id) }
func Hash_IntObjData1(v *Hash_IntMan_t, id int32) int32  { return vec.Hash_IntObjData1(v, id) }
func Vec_WrdSize(v *Vec_Wrd_t) int32                     { return int32(vec.Vec_WrdSize(v)) }
func Vec_WrdFill(v *Vec_Wrd_t, n int32, x word)          { vec.Vec_WrdFill(v, int(n), x) }
func Vec_FltFill(v *Vec_Flt_t, n int32, x float32)       { vec.Vec_FltFill(v, int(n), x) }
func Vec_WecInit(v *Vec_Wec_t, n int32)                  { vec.Vec_WecInit(v, int(n)) }
func Vec_FltSize(v *Vec_Flt_t) int32                     { return int32(vec.Vec_FltSize(v)) }
func Vec_WecSize(v *Vec_Wec_t) int32                     { return int32(vec.Vec_WecSize(v)) }
func Vec_IntErase(v *Vec_Int_t)                          { vec.Vec_IntErase(v) }
func Vec_WrdErase(v *Vec_Wrd_t)                          { vec.Vec_WrdErase(v) }
func Vec_WrdFreeP(v **Vec_Wrd_t)                         { vec.Vec_WrdFreeP(v) }
func Vec_FltErase(v *Vec_Flt_t)                          { vec.Vec_FltErase(v) }
func Vec_WecErase(v *Vec_Wec_t)                          { vec.Vec_WecErase(v) }
func Vec_StrEntry(v *Vec_Str_t, i int32) byte            { return vec.Vec_StrEntry(v, int(i)) }
func Vec_StrWriteEntry(v *Vec_Str_t, i int32, x byte)    { vec.Vec_StrWriteEntry(v, int(i), x) }
func Vec_WecEntry(v *Vec_Wec_t, i int32) *Vec_Int_t      { return vec.Vec_WecEntry(v, int(i)) }
func Vec_WrdEntry(v *Vec_Wrd_t, i int32) word            { return vec.Vec_WrdEntry(v, int(i)) }
func Vec_WrdEntryP(v *Vec_Wrd_t, i int32) *word          { return vec.Vec_WrdEntryP(v, int(i)) }
func Vec_WrdWriteEntry(v *Vec_Wrd_t, i int32, x word)    { vec.Vec_WrdWriteEntry(v, int(i), x) }
func Vec_FltEntry(v *Vec_Flt_t, i int32) float32         { return vec.Vec_FltEntry(v, int(i)) }
func Vec_FltWriteEntry(v *Vec_Flt_t, i int32, x float32) { vec.Vec_FltWriteEntry(v, int(i), x) }
func Vec_IntUpdateEntry(v *Vec_Int_t, i int32, x int32)  { vec.Vec_IntUpdateEntry(v, int(i), x) }
func Vec_IntPushArray(v *Vec_Int_t, a []int32, n int32)  { vec.Vec_IntPushArray(v, a, int(n)) }
func Vec_IntAddToEntry(v *Vec_Int_t, i, x int32) int32   { return vec.Vec_IntAddToEntry(v, int(i), x) }
func Vec_StrPush(v *Vec_Str_t, x byte)                   { vec.Vec_StrPush(v, x) }
func Vec_WrdPush(v *Vec_Wrd_t, x word)                   { vec.Vec_WrdPush(v, x) }
func Vec_FltPush(v *Vec_Flt_t, x float32)                { vec.Vec_FltPush(v, x) }
func Vec_WecPushLevel(v *Vec_Wec_t) *Vec_Int_t           { return vec.Vec_WecPushLevel(v) }
func Vec_QueSetPriority(v *Vec_Que_t, costs *[]float32)  { vec.Vec_QueSetPriority(v, costs) }
func Vec_FltArrayP(v *Vec_Flt_t) *[]float32              { return vec.Vec_FltArrayP(v) }
func Vec_IntRemove(v *Vec_Int_t, x int32) int32          { return int32(vec.Vec_IntRemove(v, x)) }
func Vec_IntGrow(v *Vec_Int_t, n int32)                  { vec.Vec_IntGrow(v, int(n)) }
func Vec_StrGrow(v *Vec_Str_t, n int32)                  { vec.Vec_StrGrow(v, int(n)) }
func Vec_StrErase(v *Vec_Str_t)                          { vec.Vec_StrErase(v) }
func Vec_PtrPush(v *Vec_Ptr_t, x any)                    { vec.Vec_PtrPush(v, x) }
func Vec_QueFreeP(v **Vec_Que_t)                         { vec.Vec_QueFreeP(v) }
func Vec_StrArray(v *Vec_Str_t) []byte                   { return vec.Vec_StrArray(v) }
func Vec_StrClear(v *Vec_Str_t)                          { vec.Vec_StrClear(v) }
func Vec_StrReleaseArray(v *Vec_Str_t) []byte            { return vec.Vec_StrReleaseArray(v) }
func Vec_IntAppend(v, q *Vec_Int_t)                      { vec.Vec_IntAppend(v, q) }
func Vec_IntTwoMerge(v, q *Vec_Int_t) *Vec_Int_t         { return vec.Vec_IntTwoMerge(v, q) }
func Vec_WecAlloc(n int32) *Vec_Wec_t                    { return vec.Vec_WecAlloc(int(n)) }
func Vec_IntStartNatural(n int32) *Vec_Int_t             { return vec.Vec_IntStartNatural(int(n)) }
func Vec_PtrReverseOrder(v *Vec_Ptr_t)                   { vec.Vec_PtrReverseOrder(v) }
func Vec_StrFreeP(v **Vec_Str_t)                         { vec.Vec_StrFreeP(v) }
func Vec_BitFreeP(v **Vec_Bit_t)                         { vec.Vec_BitFreeP(v) }
func Vec_BitFree(v *Vec_Bit_t)                           { vec.Vec_BitFree(v) }
func Vec_IntPrint(v *Vec_Int_t)                          { vec.Vec_IntPrint(v) }
func Vec_IntPushThree(v *Vec_Int_t, a, b, c int32)       { vec.Vec_IntPushThree(v, a, b, c) }
func Vec_IntPushTwo(v *Vec_Int_t, a, b int32)            { vec.Vec_IntPushTwo(v, a, b) }
func Vec_WrdFree(v *Vec_Wrd_t)                           { vec.Vec_WrdFree(v) }
func Vec_IntEntryLast(v *Vec_Int_t) int32                { return vec.Vec_IntEntryLast(v) }
func Vec_IntPop(v *Vec_Int_t) int32                      { return vec.Vec_IntPop(v) }
func Vec_IntReverseOrder(v *Vec_Int_t)                   { vec.Vec_IntReverseOrder(v) }
func Vec_IntPushUnique(v *Vec_Int_t, x int32) int32      { return int32(vec.Vec_IntPushUnique(v, x)) }
func Vec_IntCheckUniqueSmall(v *Vec_Int_t) int32         { return int32(vec.Vec_IntCheckUniqueSmall(v)) }
func Vec_IntUniqify(v *Vec_Int_t) int32                  { return int32(vec.Vec_IntUniqify(v)) }
func Vec_IntInvert(v *Vec_Int_t, fill int32) *Vec_Int_t  { return vec.Vec_IntInvert(v, fill) }
func Vec_QueTopPriority(v *Vec_Que_t) float32            { return vec.Vec_QueTopPriority(v) }
func Vec_QuePop(v *Vec_Que_t) int32                      { return int32(vec.Vec_QuePop(v)) }
func Vec_IntRemapArray(mapping, old, new *Vec_Int_t, n int32) {
	vec.Vec_IntRemapArray(mapping, old, new, int(n))
}
func Vec_IntMemory(v *Vec_Int_t) float64             { return vec.Vec_IntMemory(v) }
func Vec_StrMemory(v *Vec_Str_t) float64             { return vec.Vec_StrMemory(v) }
func Vec_PtrWriteEntry(v *Vec_Ptr_t, i int32, x any) { vec.Vec_PtrWriteEntry(v, int(i), x) }

func Vec_WecMemory(v *Vec_Wec_t) float64             { return vec.Vec_WecMemory(v) }
func Vec_WrdMemory(v *Vec_Wrd_t) float64             { return vec.Vec_WrdMemory(v) }
func Vec_PtrGrow(v *Vec_Ptr_t, n int32)              { vec.Vec_PtrGrow(v, int(n)) }
func Hash_IntManRef(v *Hash_IntMan_t) *Hash_IntMan_t { return vec.Hash_IntManRef(v) }
func Hash_IntManDeref(v *Hash_IntMan_t)              { vec.Hash_IntManDeref(v) }

func Vec_PtrErase(v *Vec_Ptr_t)                   { vec.Vec_PtrErase(v) }
func Vec_IntGetEntry(v *Vec_Int_t, i int32) int32 { return vec.Vec_IntGetEntry(v, int(i)) }
func Vec_IntSetEntry(v *Vec_Int_t, i, x int32)    { vec.Vec_IntSetEntry(v, int(i), x) }
func Vec_PtrMemory(v *Vec_Ptr_t) float64          { return float64(len(vec.Vec_PtrArray(v)) * 8) }

func Vec_StrAlloc(n int32) *Vec_Str_t            { return vec.Vec_StrAlloc(int(n)) }
func Vec_StrAppend(v *Vec_Str_t, s string)       { vec.Vec_StrAppend(v, s) }
func Vec_StrFree(v *Vec_Str_t)                   { vec.Vec_StrFree(v) }
func Vec_IntStartFull(n int32) *Vec_Int_t        { return vec.Vec_IntStartFull(int(n)) }
func Vec_IntEntryP(v *Vec_Int_t, i int32) *int32 { return vec.Vec_IntEntryP(v, int(i)) }
func Vec_IntFreeP(v **Vec_Int_t)                 { vec.Vec_IntFreeP(v) }
func Vec_BitWriteEntry(v *Vec_Bit_t, i, x int32) { vec.Vec_BitWriteEntry(v, int(i), int(x)) }
func Vec_BitStart(n int32) *Vec_Bit_t            { return vec.Vec_BitStart(int(n)) }
func Vec_BitEntry(v *Vec_Bit_t, i int32) int32   { return int32(vec.Vec_BitEntry(v, int(i))) }
func Vec_IntSort(v *Vec_Int_t, reverse int32)    { vec.Vec_IntSort(v, int(reverse)) }
func Vec_IntShrink(v *Vec_Int_t, n int32)        { vec.Vec_IntShrink(v, int(n)) }
func Vec_IntSelectSortCost(a []int32, n int32, cost *Vec_Int_t) {
	vec.Vec_IntSelectSortCost(a, int(n), cost)
}
func Vec_WrdStart(n int32) *Vec_Wrd_t            { return vec.Vec_WrdStart(int(n)) }
func Vec_IntLimit(v *Vec_Int_t) []int32          { return vec.Vec_IntLimit(v) }
func Vec_IntSelectSort(a []int32, n int32)       { vec.Vec_IntSelectSort(a, int(n)) }
func Vec_IntDrop(v *Vec_Int_t, i int32)          { vec.Vec_IntDrop(v, int(i)) }
func Vec_IntInsert(v *Vec_Int_t, i, value int32) { vec.Vec_IntInsert(v, int(i), value) }

func Vec_QueAlloc(n int32) *Vec_Que_t               { return vec.Vec_QueAlloc(int(n)) }
func Vec_QueIsMember(p *Vec_Que_t, n int32) int32   { return int32(vec.Vec_QueIsMember(p, int(n))) }
func Vec_QueUpdate(p *Vec_Que_t, n int32)           { vec.Vec_QueUpdate(p, int(n)) }
func Vec_QuePush(p *Vec_Que_t, n int32)             { vec.Vec_QuePush(p, int(n)) }
func Vec_QueClear(p *Vec_Que_t)                     { vec.Vec_QueClear(p) }
func Vec_WrdSetEntry(p *Vec_Wrd_t, i int32, w word) { vec.Vec_WrdSetEntry(p, int(i), w) }
