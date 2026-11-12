// Go port of the comparator-based sorting helpers in xsatUtils.h.
package xsat
import "sort"
func boolInt(b bool)int{if b{return 1};return 0}
func xSAT_UtilSelectSort(data []any,n int,less func(any,any)bool){data=data[:n];for i:=0;i<n;i++{best:=i;for j:=i+1;j<n;j++{if less(data[j],data[best]){best=j}};data[i],data[best]=data[best],data[i]}}
func xSAT_UtilSort(data []any,n int,less func(any,any)bool){data=data[:n];sort.Slice(data,func(i,j int)bool{return less(data[i],data[j])})}
