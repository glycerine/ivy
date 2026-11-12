# Original-C extrab fixtures

These generators execute genuine original ABC algorithms. Go tests contain the
expected checksums and do not require a native compiler.

## Unateness

```sh
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddUnate.c -o /tmp/extrab-unate.o
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddSymm.c -o /tmp/extrab-symm.o
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddMisc.c -o /tmp/extrab-misc.o
cc -DABC_USE_STDINT_H -I src -Wl,--gc-sections abcgo/bdd/extrab/test_vectors/unate_oracle.c /tmp/extrab-unate.o /tmp/extrab-symm.o /tmp/extrab-misc.o src/bdd/cudd/*.o src/bdd/mtr/*.o src/bdd/epd/*.o src/misc/st/st.o src/misc/extra/extraUtilUtil.o -lm -o /tmp/extrab-unate-oracle
/tmp/extrab-unate-oracle
```

Expected: `94e5759806dcde0c` (all 256 three-variable functions; cofactor checks,
fast info fields, and slow info count).

## Bounded threshold search

```sh
cc -DABC_USE_STDINT_H -DABC_USE_CUDD -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddThresh.c -o /tmp/extrab-threshold.o
cc -DABC_USE_STDINT_H -I src -Wl,--gc-sections abcgo/bdd/extrab/test_vectors/threshold_oracle.c /tmp/extrab-threshold.o -lm -o /tmp/extrab-threshold-oracle
/tmp/extrab-threshold-oracle
```

Expected: `23d92fb53fb7aab7` (all three-/four-variable functions; Chow values,
returned threshold, mutated truth table, and weights).

## Inequality heuristic

```sh
cc -DABC_USE_STDINT_H -I src -Wl,--gc-sections abcgo/bdd/extrab/test_vectors/threshold_heuristic_oracle.c /tmp/extrab-threshold.o -lm -o /tmp/extrab-threshold-heuristic-oracle
/tmp/extrab-threshold-heuristic-oracle
```

Expected: `14b0eda9522e02d5` (all nonconstant three-/four-variable functions).
The adapter executes the original Chow sort/grouping and inequality weight search.
For its monotone functions, complete prime covers equal minimal true/maximal false
assignments; the adapter enumerates those in original SOP traversal order. It
avoids ABC's global frame/allocator interface and the C heuristic's buffer-overrun
bug, without replacing the original weight-search algorithm or truncating covers.

Function sections discard unrelated original CLI dependencies. There are no
native stubs. Floating infinity configuration does not affect these integer and
Boolean fixtures.

## Auto translation spaces and equations

With `/tmp/extrab-misc.o` and `/tmp/extrab-symm.o` built as above:

```sh
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddAuto.c -o /tmp/extrab-auto.o
cc -DABC_USE_STDINT_H -I src -Wl,--gc-sections abcgo/bdd/extrab/test_vectors/auto_oracle.c /tmp/extrab-auto.o /tmp/extrab-misc.o /tmp/extrab-symm.o src/bdd/cudd/*.o src/bdd/mtr/*.o src/bdd/epd/*.o src/misc/st/st.o src/misc/extra/extraUtilUtil.o -lm -o /tmp/extrab-auto-oracle
/tmp/extrab-auto-oracle
```

Expected `4b8b81243ecf473a`. Includes every pair of three-variable functions,
fast/self spaces, both signed recursions, canonical-variable cubes, positive and
negative equation families, and conversions for every three-variable matrix.

## Scheduled/monolithic image

```sh
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddImage.c -o /tmp/extrab-image.o
cc -DABC_USE_STDINT_H -I src -Wl,--gc-sections abcgo/bdd/extrab/test_vectors/image_oracle.c /tmp/extrab-image.o /tmp/extrab-misc.o src/bdd/cudd/*.o src/bdd/mtr/*.o src/bdd/epd/*.o src/misc/st/st.o src/misc/extra/extraUtilUtil.o -lm -o /tmp/extrab-image-oracle
/tmp/extrab-image-oracle
```

Expected `2c5b1f2faff964eb`. Exercises 128 seeded source-valid schedules with zero,
one or two partitions and every two-variable care set. Both stored initial images
and every tree/monolithic recomputation are hashed. Go tests additionally exercise
more partitions than manager variables, a case unsafe in the original scheduler.

## K-map, miscellaneous, symmetry and deadline families

All these are genuine original-C function-section builds using existing ABC CUDD
objects. No native test is needed by the Go suite. Build the source family then
its matching harness, using the following template (`NAME` and `FILE` below):

```sh
cc -DABC_USE_STDINT_H -I src -ffunction-sections -fdata-sections -c src/bdd/extrab/extraBddFILE.c -o /tmp/extrab-NAME.o
cc -DABC_USE_STDINT_H -I src -Wl,--gc-sections abcgo/bdd/extrab/test_vectors/NAME_oracle.c /tmp/extrab-NAME.o /tmp/extrab-misc.o src/bdd/cudd/*.o src/bdd/mtr/*.o src/bdd/epd/*.o src/misc/st/st.o src/misc/extra/extraUtilUtil.o -lm -o /tmp/extrab-NAME-oracle
/tmp/extrab-NAME-oracle
```

Use `NAME=kmap FILE=Kmap`, `NAME=symm FILE=Symm`, or `NAME=time FILE=Time`.
The miscellaneous object supplies original support helpers to symmetry and cube
helpers to K-map. It can be omitted for time. For `NAME=misc FILE=Misc`, include `/tmp/extrab-misc.o`
only once, and add `/tmp/extrab-symm.o` if the linker requires its support helper.

Expected output:

| Harness | Hash | Sequence |
| --- | --- | --- |
| kmap | `aa69523ca10d5418` | Exact relation-map ASCII for all 256 three-variable functions |
| misc | `869b30057848fe11` | Boolean transfer/permutation/polarity/cube transformations |
| misc | `8e62023c841daec4` | All bounded ISOP count/phase-guide cases |
| misc | `3889814f756ac6b6` | Exact paired-literal PLA text |
| symm | `72c5fd61011d557d` | All symmetry metadata/prefilters and 65,536 singleton relations |
| time | `f6ee7ba969cdfbe5` | Unlimited timed AND/abstraction pairs and remapped transfers |

The time fixture passes the source's `TimeOut=0` to disable its process-CPU-clock
deadline. Go tests separately cover standard-library absolute deadlines, expired
and unlimited values, and deterministic cancellation within recursion. They do not
pretend process CPU ticks and wall/monotonic time use identical units.
