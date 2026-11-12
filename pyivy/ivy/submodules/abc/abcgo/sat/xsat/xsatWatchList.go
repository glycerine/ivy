// Go port of xsatWatchList.h by Bruno Schmitt (2016).
package xsat

type xSAT_Watcher_t struct{Blocker int;CRef uint32}
type xSAT_WatchList_t struct{data []xSAT_Watcher_t}
type xSAT_VecWatchList_t struct{data []*xSAT_WatchList_t}
func xSAT_WatchListFree(v *xSAT_WatchList_t){v.data=nil}
func xSAT_WatchListSize(v *xSAT_WatchList_t)int{return len(v.data)}
func xSAT_WatchListShrink(v *xSAT_WatchList_t,n int){if n<0||n>len(v.data){panic("invalid watch shrink")};v.data=v.data[:n]}
func xSAT_WatchListPush(v *xSAT_WatchList_t,w xSAT_Watcher_t){v.data=append(v.data,w)}
func xSAT_WatchListArray(v *xSAT_WatchList_t)[]xSAT_Watcher_t{return v.data}
func xSAT_WatchListRemove(v *xSAT_WatchList_t,ref uint32){for i,w:=range v.data{if w.CRef==ref{copy(v.data[i:],v.data[i+1:]);v.data=v.data[:len(v.data)-1];return}}}
func xSAT_VecWatchListAlloc(capacity int)*xSAT_VecWatchList_t{if capacity<0{panic("negative capacity")};return &xSAT_VecWatchList_t{data:make([]*xSAT_WatchList_t,0,capacity)}}
func xSAT_VecWatchListFree(v *xSAT_VecWatchList_t){v.data=nil}
func xSAT_VecWatchListPush(v *xSAT_VecWatchList_t){v.data=append(v.data,&xSAT_WatchList_t{})}
func xSAT_VecWatchListEntry(v *xSAT_VecWatchList_t,i int)*xSAT_WatchList_t{return v.data[i]}
