package main

import "os"

//#include <misc/util/abc_global.h>

//ABC_NAMESPACE_IMPL_START

// int Abc_RealMain(int argc, char *argv[]);


//ABC_NAMESPACE_IMPL_END

func main() {
   os.Exit(int( ABC_NAMESPACE_PREFIX Abc_RealMain(len(os.Args), os.Args)));
}
