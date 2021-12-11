package valobj

import (
	"strconv"
	"testing"
)

func TestMessageTo(t *testing.T) {
	patterns := map[int]int{
		0: 1, 1: 0, 2: 3, 10: 10,
	}
	checkUuids := func(arg string, u1, u2 []*UuidInt) {
		check(t, arg+",length", len(u1), len(u2))
		m := make(map[int]bool)
		for _, u := range u1 {
			m[u.ToInt()] = true
		}
		for _, u := range u2 {
			_, ok := m[u.ToInt()]
			check(t, arg+",not exist,"+strconv.Itoa(u.ToInt()), ok, true)
		}
	}
	for a, b := range patterns {
		var ua, ub []*UuidInt
		for i := 0; i < a; i++ {
			ua = append(ua, NewUuidIntRandom())
		}
		for i := 0; i < b; i++ {
			ub = append(ub, NewUuidIntRandom())
		}
		uc := make([]*UuidInt, len(ua))
		copy(uc, ua)
		uc = append(uc, ub...)
		to := NewMessageTo(ua)
		checkUuids("pattern: A"+strconv.Itoa(a)+","+strconv.Itoa(a),
			to.GetUUids(), ua)
		to = NewMessageToExcept(uc, ub)
		checkUuids("pattern: B"+strconv.Itoa(a)+","+strconv.Itoa(a),
			to.GetUUids(), ua)
	}
}
