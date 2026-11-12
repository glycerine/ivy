package mem

// Byte arena entry APIs return mutable []byte rather than C char pointers.
// Entries remain stable across arena growth. Restart/recycle invalidate ownership
// of previously fetched entries, as in the C allocator.
