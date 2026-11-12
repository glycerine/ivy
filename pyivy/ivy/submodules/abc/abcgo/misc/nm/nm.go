// Package nm maintains object-ID/name mappings. Names may belong to several
// object types; an ID has exactly one name. Port of ABC's misc/nm package.
package nm

// Manager is a name manager. It is not safe for concurrent mutation.
type Manager struct {
	byID   map[int32]*Entry
	byName map[string]*Entry
}
type Nm_Man_t = Manager

// Entry records one object and participates in a ring of namesakes.
type Entry struct {
	Type      uint32
	ObjId     uint32
	Name      string
	pNameSake *Entry
}
type Nm_Entry_t = Entry
