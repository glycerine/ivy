// Go port of xsatBQueue.h by Bruno Schmitt (2016).
package xsat

type xSAT_BQueue_t struct{data []uint32;first int;sum uint64}
func xSAT_BQueueNew(capacity int)*xSAT_BQueue_t{if capacity<=0{panic("queue capacity must be positive")};return &xSAT_BQueue_t{data:make([]uint32,0,capacity)}}
func xSAT_BQueueFree(p *xSAT_BQueue_t){*p=xSAT_BQueue_t{}}
func xSAT_BQueuePush(p *xSAT_BQueue_t,value uint32){capacity:=cap(p.data);if capacity==0{panic("push to freed queue")};if len(p.data)==capacity{p.sum-=uint64(p.data[p.first]);p.first=(p.first+1)%capacity}else{p.data=p.data[:len(p.data)+1]};index:=(p.first+len(p.data)-1)%capacity;p.data[:capacity][index]=value;p.sum+=uint64(value)}
func xSAT_BQueuePop(p *xSAT_BQueue_t)int{if len(p.data)==0{panic("pop empty queue")};value:=p.data[:cap(p.data)][p.first];p.first=(p.first+1)%cap(p.data);p.data=p.data[:len(p.data)-1];p.sum-=uint64(value);return int(value)}
func xSAT_BQueueAvg(p *xSAT_BQueue_t)uint32{if len(p.data)==0{panic("average empty queue")};return uint32(p.sum/uint64(len(p.data)))}
func xSAT_BQueueIsValid(p *xSAT_BQueue_t)int{return boolInt(len(p.data)==cap(p.data))}
func xSAT_BQueueClean(p *xSAT_BQueue_t){p.data=p.data[:0];p.first=0;p.sum=0}
