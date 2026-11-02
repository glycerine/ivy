package abc

// Hand-written skeleton of the bz_stream type in src/misc/bzlib/bzlib.h. The
// prototypes in this header are implemented in bzlib.c (not part of this
// skeleton).

// BzStream is bz_stream.
type BzStream struct {
	next_in        []byte                  // char * : advanced as input is consumed
	avail_in       uint32                  // unsigned int
	total_in_lo32  uint32                  // unsigned int
	total_in_hi32  uint32                  // unsigned int
	next_out       []byte                  // char * : advanced as output is produced
	avail_out      uint32                  // unsigned int
	total_out_lo32 uint32                  // unsigned int
	total_out_hi32 uint32                  // unsigned int
	state          any                     // void *
	bzalloc        func(any, int, int) any // void *(*bzalloc)(void *,int,int)
	bzfree         func(any, any)          // void (*bzfree)(void *,void *)
	opaque         any                     // void *
}
