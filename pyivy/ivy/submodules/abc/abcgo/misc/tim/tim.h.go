// The original tim.h describes a manager coupled to the combinational inputs
// and outputs of an AIG. The Go API uses slices instead of Vec containers;
// box delay tables retain ABC's [id, inputs, outputs, delays...] representation.
package tim

const TIM_ETERNITY = 1000000000
