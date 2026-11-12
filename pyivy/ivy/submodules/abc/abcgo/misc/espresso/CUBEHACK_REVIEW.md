# Cube geometry helper review

Ported all four cubehack.c functions. Binary geometries up to 100 variables
reuse the source's preallocated metadata/mask/scratch buffers; larger geometries
use normal cube_setup. Resizing updates packed headers, full/binary masks,
counts, output state, and fast binary distance masks. Undefine releases the
shared geometry. Espresso default flags preserve the source values.

## Compactness and readability

A single backing-buffer identity replaces C's persistent called_before flag;
it carries exactly the information needed to know whether the current cube
storage belongs to the 100-variable cache. No second geometry representation is
maintained. Counts retain the shared int32 convention. Setup, resizing and
teardown remain short functions with explicit ownership.

## Bug review

The C cautious helper returns early whenever num_binary_vars matches, even
when a multiple-valued tail changes num_vars and masks. The regression builds
two binary plus one four-valued variable, requests two binary variables, and
requires size four/output -1. The port compares both variable counts.

The source called_before flag can also survive another caller replacing global
cube storage; resizing then indexes arrays assuming 100-variable capacity.
A regression replaces the geometry with three multiple-valued variables before
requesting two binary variables. Backing identity invalidates and reconstructs
the cache. Negative or packed-width-overflow counts are rejected explicitly.

Tests exercise 0/1/16/17/100/101/150 variables, shrink/grow transitions, actual
buffer reuse, teardown/restart, every variable mask, full-set cardinality,
external setup invalidation, mixed-valued transitions, and default flags.
Combined complement/cofactor/irred/geometry plus supporting set/cover/mincov
subset passes tests, race tests, and vet. Full-package compilation was rerun;
the next blocker is the unrelated cvrin.go draft, being ported by its owner.
