package amap

// Amap_OutName returns the mapper-selected gate name from an output record.
func Amap_OutName(output *Amap_Out_t) string { return output.pName }

// Amap_OutType returns the C-compatible input, internal-node, or output tag.
func Amap_OutType(output *Amap_Out_t) int16 { return output.Type }

// Amap_OutFanCount returns the number of fanin indices stored in the record.
func Amap_OutFanCount(output *Amap_Out_t) int16 { return output.nFans }

// Amap_OutFan returns a borrowed fanin index by position.
func Amap_OutFan(output *Amap_Out_t, index int16) int32 { return output.pFans[index] }
