package bsat

import "github.com/glycerine/abcgo/misc/mem"

/**CFile****************************************************************

  FileName    [satMem.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Memory management.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: satMem.c,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

// SAT's allocators share the ported Go byte pools. Clients receive slices,
// never string-backed writable memory or byte-to-struct pointer overlays.
type Sat_MmFixed_t_ = mem.Mem_Fixed_t
type Sat_MmFlex_t_ = mem.Mem_Flex_t
type Sat_MmStep_t_ = mem.Mem_Step_t

func Sat_MmFixedStart(size int32) *Sat_MmFixed_t             { return mem.Mem_FixedStart(size) }
func Sat_MmFixedStop(p *Sat_MmFixed_t, verbose int32)        { mem.Mem_FixedStop(p, verbose) }
func Sat_MmFixedEntryFetch(p *Sat_MmFixed_t) []byte          { return mem.Mem_FixedEntryFetch(p) }
func Sat_MmFixedEntryRecycle(p *Sat_MmFixed_t, entry []byte) { mem.Mem_FixedEntryRecycle(p, entry) }
func Sat_MmFixedRestart(p *Sat_MmFixed_t)                    { mem.Mem_FixedRestart(p) }
func Sat_MmFixedReadMemUsage(p *Sat_MmFixed_t) int32         { return mem.Mem_FixedReadMemUsage(p) }
func Sat_MmFlexStart() *Sat_MmFlex_t                         { return mem.Mem_FlexStart() }
func Sat_MmFlexStop(p *Sat_MmFlex_t, verbose int32)          { mem.Mem_FlexStop(p, verbose) }
func Sat_MmFlexEntryFetch(p *Sat_MmFlex_t, n int32) []byte   { return mem.Mem_FlexEntryFetch(p, n) }
func Sat_MmFlexReadMemUsage(p *Sat_MmFlex_t) int32           { return mem.Mem_FlexReadMemUsage(p) }
func Sat_MmStepStart(n int32) *Sat_MmStep_t                  { return mem.Mem_StepStart(n) }
func Sat_MmStepStop(p *Sat_MmStep_t, verbose int32)          { mem.Mem_StepStop(p, verbose) }
func Sat_MmStepRestart(p *Sat_MmStep_t)                      { mem.Mem_StepRestart(p) }
func Sat_MmStepEntryFetch(p *Sat_MmStep_t, n int32) []byte   { return mem.Mem_StepEntryFetch(p, n) }
func Sat_MmStepEntryRecycle(p *Sat_MmStep_t, entry []byte, n int32) {
	mem.Mem_StepEntryRecycle(p, entry, n)
}
func Sat_MmStepReadMemUsage(p *Sat_MmStep_t) int32 { return mem.Mem_StepReadMemUsage(p) }
