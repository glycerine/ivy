package espresso

// util_old.h only declared an external SIS runtime, allocator and process API;
// it supplied no Espresso algorithms. Go slices/GC, io/os, strconv and the
// monotonic timing helpers replace those dependencies in the port. No foreign
// allocator, image-dumping or fork interface is emulated here.
