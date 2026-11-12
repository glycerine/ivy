// Shared partition structure and parameters from place_gordian.h.
package place

const (
	CLIQUE_PENALTY              = 1.0
	IGNORE_NETSIZE              = 20
	LARGEST_FINAL_SIZE          = 20
	PARTITION_AREA_ONLY         = true
	REALLOCATE_PARTITIONS       = false
	FINAL_REALLOCATE_PARTITIONS = false
	IGNORE_COG                  = false
	MAX_PARTITION_NONSYMMETRY   = 0.30
	REPARTITION_LEVEL_DEPTH     = 4
	REPARTITION_TARGET_FRACTION = 0.15
	REPARTITION_FM              = false
	REPARTITION_HMETIS          = true
	FM_MAX_BIN                  = 10
	FM_MAX_PASSES               = 10
)

type Partition struct {
	m_numMembers               int32
	m_members                  []*ConcreteCell
	m_bounds                   Rect
	m_done, m_leaf, m_vertical bool
	m_area                     float32
	m_level                    int32
	m_sub1, m_sub2             *Partition
}
