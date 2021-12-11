package valobj

type MessageTo struct {
	uuids []*UuidInt
}

func NewMessageTo(uuids []*UuidInt) *MessageTo {
	mes := new(MessageTo)
	mes.uuids = make([]*UuidInt, len(uuids))
	copy(mes.uuids, uuids)
	return mes
}

func NewMessageToAll(uuids ...[]*UuidInt) *MessageTo {
	mes := new(MessageTo)
	mes.uuids = make([]*UuidInt, 0)
	for _, u := range uuids {
		mes.uuids = append(mes.uuids, u...)
	}
	return mes
}

func NewMessageToExcept(uuids []*UuidInt, except []*UuidInt) *MessageTo {
	mes := new(MessageTo)
	m := make(map[int]bool)
	m_temp := make(map[int]bool)
	for _, u := range uuids {
		m_temp[u.ToInt()] = true
	}
	for _, u := range except {
		_, ok := m_temp[u.ToInt()]
		if !ok {
			continue
		}
		m[u.ToInt()] = true
	}
	mes.uuids = make([]*UuidInt, len(uuids)-len(m))
	i := 0
	for _, u := range uuids {
		_, ok := m[u.ToInt()]
		if !ok {
			mes.uuids[i] = u
			i++
		}
	}
	return mes
}

func (to *MessageTo) GetUUids() []*UuidInt {
	return to.uuids
}
