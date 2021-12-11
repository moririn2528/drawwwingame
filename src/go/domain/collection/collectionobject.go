package collection

import (
	"drawwwingame/domain/valobj"
)

type RoleMap struct {
	m map[int]int
}

func NewRoleMap(answer, writer []*valobj.UuidInt) *RoleMap {
	s := new(RoleMap)
	s.m = make(map[int]int)
	for _, u := range answer {
		s.m[u.ToInt()] |= 1
	}
	for _, u := range writer {
		s.m[u.ToInt()] |= 1 << 1
	}
	return s
}

func (r RoleMap) SetAnswer(uuid *valobj.UuidInt) {
	r.m[uuid.ToInt()] = 1
}
func (r RoleMap) SetWriter(uuid *valobj.UuidInt) {
	r.m[uuid.ToInt()] = 1 << 1
}

func (r RoleMap) IsAnswer(uuid *valobj.UuidInt) bool {
	return r.m[uuid.ToInt()]&1 > 0
}
func (r RoleMap) IsWriter(uuid *valobj.UuidInt) bool {
	return r.m[uuid.ToInt()]&(1<<1) > 0
}
