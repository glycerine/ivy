package abcgo

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [starter.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Wrapper for calling ABC.]

  Synopsis    [A demo program illustrating parallel execution of ABC.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 22, 2009.]

  Revision    [$Id: starter.c,v 1.00 2009/10/22 00:00:00 alanmi Exp $]

***********************************************************************/

// To compile on Linux run:  gcc -pthread -o starter starter.c

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>
#ifdef WIN32
//#include "pthread.h"
#else
//#include <pthread.h>
#include <unistd.h>
#endif

// the max number of commands to execute from the input file
const MAX_COMM_NUM = 1000

// time printing
// Original macro:
// #define ABC_PRT(a,t)    (printf("%s = ", (a)), printf("%7.2f sec\n", (float)(t)/(float)(CLOCKS_PER_SEC)))


// the number of currently running threads
var nThreadsRunning int32;

// mutext to control access to the number of threads
var mutex sync.Mutex  ;

// procedure for duplicating strings
func Abc_UtilStrsav(s string) string { if s {
    return strcpy(malloc(len(s)+1), s);
} else {
    return nil;
}  }

/**Function*************************************************************

  Synopsis    [This procedures executes one call to system().]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_RunThread(Command any) any {
    // perform the call
    if system( (char *)Command ) {
        mutex.Lock();
        fmt.Fprintf(os.Stderr, "The following command has returned non-zero exit status:\n");
        fmt.Fprintf(os.Stderr, "\"%s\"\n", (char *)Command);
        fmt.Fprintf(os.Stderr, "Sorry for the inconvenience.\n");
        fflush( os.Stdout );
        mutex.Unlock();
    }

    // decrement the number of threads runining 
    mutex.Lock();
    nThreadsRunning--;
    mutex.Unlock();

    // quit this thread
    //printf("...Finishing %s\n", (char *)Command);
    free( Command );
    pthread_exit( nil );
    assert(false, "0");
    return nil;
}

/**Function*************************************************************

  Synopsis    [Takes file with commands to be executed and the number of CPUs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func main() {
    var pFile *os.File;
    var pOutput *os.File = os.Stdout;
    var ThreadIds [MAX_COMM_NUM]pthread_t;
    var pBufferCopy string;
    var Buffer [MAX_COMM_NUM]byte;
    var i, nCPUs, nLines, Counter int32;
    var clk int64  = time.Now().UnixNano();
    
    // check command line arguments
    if len(os.Args) != 3 { fmt.Fprintf(os.Stderr, "Wrong number of command line arguments.\n"); goto usage; }

    // get the number of CPUs
    nCPUs = strconv.Atoi(os.Args[1]);
    if nCPUs <= 0 { fmt.Fprintf(pOutput, "Cannot read an integer represting the number of CPUs.\n"); goto usage; }

    // open the file and make sure it is available
    pFile = os.Open(os.Args[2]);
    if pFile == nil { fmt.Fprintf(pOutput, "Input file \"%s\" cannot be opened.\n", os.Args[2]); goto usage; }

    // read commands and execute at most <num> of them at a time
//    assert(mutex == PTHREAD_MUTEX_INITIALIZER);
    for  fgets( Buffer, MAX_COMM_NUM, pFile ) != nil  {
        // get the command from the file
        if Buffer[0] == '\n' || Buffer[0] == '\r' || Buffer[0] == '\t' || 
             Buffer[0] == ' ' || Buffer[0] == '#' {
            continue;
        }

        if Buffer[strlen(Buffer)-1] == '\n' {
            Buffer[strlen(Buffer)-1] = 0; }
        if Buffer[strlen(Buffer)-1] == '\r' {
            Buffer[strlen(Buffer)-1] = 0; }

        // wait till there is an empty thread
        for {
            mutex.Lock();
            Counter = nThreadsRunning;
            mutex.Unlock();
            if Counter < nCPUs - 1 {
                break; }
//            Sleep( 100 );
        }

        // increament the number of threads running
        mutex.Lock();
        nThreadsRunning++;
        fmt.Printf("Calling:  %s\n", (char *)Buffer);  
        fflush( os.Stdout );
        mutex.Unlock();

        // create thread to execute this command
        pBufferCopy = Abc_UtilStrsav( Buffer );
        assert(pthread_create( &ThreadIds[nLines], nil, Abc_RunThread, (void *)pBufferCopy ) == 0, "pthread_create( &ThreadIds[nLines], nil, Abc_RunThread, (void *)pBufferCopy ) == 0");
        if nLines++; nLines == MAX_COMM_NUM { fmt.Fprintf(pOutput, "Cannot execute more than %d commands from file \"%s\".\n", nLines, os.Args[2]); break; }
    }

    // wait for all the threads to finish
    for {
        mutex.Lock();
        Counter = nThreadsRunning;
        mutex.Unlock();
        if Counter == 0 {
            break; }
    }

    // cleanup
    assert(pthread_mutex_destroy(&mutex) == 0, "pthread_mutex_destroy(&mutex) == 0");
//    assert(mutex == NULL);
    _ = (pFile).Close();
    fmt.Printf("Finished processing commands in file \"%s\".  ", os.Args[2]);
    // inlining of macro ABC_PRT
    fmt.Printf("%s = ", ("Total time"));
fmt.Printf("%7.2f sec\n", float32((time.Now().UnixNano() - clk))/float32((int64(time.Second))));;
    return ;

usage: 
    // skip the path name till the binary name
    for  i = len(os.Args[0]) - 1; i > 0; i--  {
        if os.Args[0][i-1] == '\\' || os.Args[0][i-1] == '/' {
            break; }
    }
    // print usage message
    fmt.Fprintf(pOutput, "usage: %s <num> <file>\n", os.Args[0]+i);
    fmt.Fprintf(pOutput, "       executes command listed in <file> in parallel on <num> CPUs\n");
    fmt.Fprintf(pOutput, "\n");
    os.Exit(int( 1));

}

