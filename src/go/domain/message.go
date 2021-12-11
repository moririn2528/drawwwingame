package domain

import (
	"drawwwingame/domain/internal"
	"drawwwingame/domain/valobj"
	"strconv"
	"strings"
)

func ToGroupAllUser(group_id *valobj.GroupIdInt) (*valobj.MessageTo, error) {
	uuids, err := SqlHandle.GetUuidInGroup(group_id)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	res := make([]*valobj.UuidInt, 0)
	for _, u := range uuids {
		_, ok := Clients.GetConnByUuid(u)
		if ok {
			res = append(res, u)
		}
	}
	return valobj.NewMessageTo(res), nil
}

type MessageObject struct {
	id    *valobj.MessageId
	uuid  *valobj.UuidInt
	name  *valobj.NameString
	group *valobj.GroupIdInt
	t     *valobj.MessageTypeString
	info  *valobj.MessageString
	str   *valobj.MessageString
}

var MessageObjectKey = []string{
	"id", "uuid", "name", "group_id", "type", "info", "message",
}

func NewMessageObjectByUser(user *User, id *valobj.MessageId,
	group_id *valobj.GroupIdInt, message_type *valobj.MessageTypeString,
	message_info *valobj.MessageString, message *valobj.MessageString) *MessageObject {
	s := new(MessageObject)
	if id == nil {
		s.id = valobj.NewMessageId()
	} else {
		s.id = id
	}
	s.t = message_type
	s.info = message_info
	s.str = message
	s.uuid = user.uuid
	s.name = user.name
	s.group = group_id
	return s
}

func NewMessageObject(id *valobj.MessageId, uuid *valobj.UuidInt, name *valobj.NameString,
	group *valobj.GroupIdInt, t *valobj.MessageTypeString, info *valobj.MessageString, str *valobj.MessageString) *MessageObject {
	s := new(MessageObject)
	if id == nil {
		s.id = valobj.NewMessageId()
	} else {
		s.id = id
	}
	s.t = t
	s.info = info
	s.str = str
	s.uuid = uuid
	s.name = name
	s.group = group
	return s
}

func NewGetMessageObject(id *valobj.MessageId) (*MessageObject, error) {
	return SqlHandle.GetMessage(id)
}

func (mes *MessageObject) GetTypeString() string {
	return mes.t.ToString()
}
func (mes *MessageObject) GetMessageString() string {
	return mes.str.ToString()
}
func (mes *MessageObject) GetUuidInt() int {
	return mes.uuid.ToInt()
}
func (mes *MessageObject) GetGroupId() int {
	return mes.group.ToInt()
}
func (mes *MessageObject) GetNameString() string {
	return mes.name.ToString()
}

// func (mes *MessageObject) GetGroupRoleInt() int {
// 	return mes.group.GetRoleInt()
// }

func (mes *MessageObject) GetMap() map[string]interface{} {
	return map[string]interface{}{
		"id":       mes.id.ToInt(),
		"uuid":     mes.uuid.ToInt(),
		"name":     mes.name.ToString(),
		"group_id": mes.group.ToInt(),
		"type":     mes.t.ToString(),
		"info":     mes.info.ToString(),
		"message":  mes.str.ToString(),
	}
}

func (mes *MessageObject) SaveSql() error {
	var err error = SqlHandle.CreateMessage(mes)
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	return nil
}

func (mes *MessageObject) ToOutputWebsocketWithTo(to *valobj.MessageTo) (*OutputWebSocketMessage, error) {
	var err error
	var info *valobj.MessageString
	if mes.t.IsText() {
		info_lines := []string{
			strconv.Itoa(mes.GetUuidInt()),
			mes.info.ToString(),
		}
		info, err = valobj.NewMessageString(strings.Join(info_lines, ":"))
		if err != nil {
			Log(err)
			return nil, ErrorInternal
		}
	} else {
		info = mes.info
	}
	return NewOutputWebSocketMessage(
		to, mes.id, mes.name, mes.group, mes.t, info, mes.str,
	), nil
}

func (mes *MessageObject) ToOutputWebsocket() (*OutputWebSocketMessage, error) {
	var to *valobj.MessageTo
	var err error
	game := GameGroup[mes.group.ToInt()]
	if game != nil {
		if mes.t.IsWriter() {
			to = valobj.NewMessageTo(game.writer)
		}
		if mes.t.IsAnswer() {
			to = valobj.NewMessageTo(game.member)
		}
	}
	if to == nil {
		to, err = ToGroupAllUser(mes.group)
		if err != nil {
			return nil, err
		}
	}
	return mes.ToOutputWebsocketWithTo(to)
}

func (mes *MessageObject) Valid() bool {
	var g int
	if mes.t.IsAnswer() || mes.t.IsWriter() || mes.t.IsLines() {
		if mes.group == nil {
			return false
		}
		g = mes.group.ToInt()
		if GameGroup[g] == nil {
			return false
		}
	}
	if mes.t.IsAnswer() {
		return GameGroup[g].roles.IsAnswer(mes.uuid)
	}
	if mes.t.IsWriter() || mes.t.IsLines() {
		return GameGroup[g].roles.IsWriter(mes.uuid)
	}
	return true
}

func (mes *MessageObject) ToString() string {
	return internal.MapToString(mes.GetMap())
}

type MarkCountOfMessage struct {
	group_id     *valobj.GroupIdInt
	message_id   *valobj.MessageId
	message_type *valobj.MessageTypeString
	count        []int
	user_count   int
}

func (mark *MarkCountOfMessage) Setting() error {
	if mark.count[0] != mark.user_count {
		return nil
	}
	if mark.message_type.IsAnswer() {
		GameGroup[mark.group_id.ToInt()].End(mark.message_id)
		return nil
	}
	if mark.message_type.IsWriter() {
		group_id := mark.group_id.ToInt()
		theme := GameGroup[group_id].theme
		if theme != nil && theme.ToString() != "" {
			return nil
		}
		message, err := NewGetMessageObject(mark.message_id)
		if err != nil {
			Log(err)
			return ErrorInternal
		}
		GameGroup[group_id].SetTheme(message.str)
		return GameGroup[group_id].Start()
	}
	LogStringf("assert false")
	return ErrorInternal
}

func NewMarkCountOfMessage(group_id *valobj.GroupIdInt, message_id *valobj.MessageId, message_type *valobj.MessageTypeString) (*MarkCountOfMessage, error) {
	s := new(MarkCountOfMessage)
	var err error
	groupIdInt := group_id.ToInt()
	if GameGroup[groupIdInt] == nil {
		LogStringf("GameGroup is nil, %v", groupIdInt)
		return nil, ErrorInternal
	}
	s.group_id = group_id
	s.message_id = message_id
	s.message_type = message_type
	s.user_count = GameGroup[group_id.ToInt()].WriterUserSize()
	s.count, err = SqlHandle.GetMarksOnMessage(message_id)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	return s, nil
}

func (mark *MarkCountOfMessage) ToMessageString() (*valobj.MessageString, error) {
	var strs []string
	for _, num := range mark.count {
		strs = append(strs, strconv.Itoa(num))
	}
	strs = append(strs, strconv.Itoa(mark.user_count))
	return valobj.NewMessageString(strings.Join(strs, ":"))
}

type MessageMark struct {
	uuid       *valobj.UuidInt
	group_id   *valobj.GroupIdInt
	message_id *valobj.MessageId
	mark       *valobj.MessageMarkString
	t          *valobj.MessageTypeString
}

func NewMessageMark(uuid *valobj.UuidInt, group *valobj.GroupIdInt, id *valobj.MessageId, mark *valobj.MessageMarkString, message_type *valobj.MessageTypeString) (*MessageMark, error) {
	m := new(MessageMark)
	m.uuid = uuid
	m.message_id = id
	m.mark = mark
	m.group_id = group
	m.t = message_type
	if !m.Valid() {
		LogStringf("not validate")
		return nil, ErrorArg
	}
	return m, nil
}

func NewMessageMarkByUser(user *User, group_id *valobj.GroupIdInt, id *valobj.MessageId, mark *valobj.MessageMarkString, message_type *valobj.MessageTypeString) (*MessageMark, error) {
	return NewMessageMark(user.uuid, group_id, id, mark, message_type)
}

var MessageMarkKey = []string{"uuid", "group_id", "message_id", "mark", "type"}

func (m *MessageMark) GetMap() map[string]interface{} {
	return map[string]interface{}{
		"uuid":       m.uuid.ToInt(),
		"group_id":   m.group_id.ToInt(),
		"message_id": m.message_id.ToInt(),
		"mark":       m.mark.ToString(),
		"type":       m.t.ToString(),
	}
}

func (m *MessageMark) SaveSql() error {
	err := SqlHandle.SetMessageMark(m)
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	return nil
}

func (m *MessageMark) GetMessageMarks() (*MarkCountOfMessage, error) {
	return NewMarkCountOfMessage(m.group_id, m.message_id, m.t)
}

func (m *MessageMark) Setting() error {
	mark, err := m.GetMessageMarks()
	if err != nil {
		return err
	}
	return mark.Setting()
}

func (m *MessageMark) ToOutputWebsocketWithTo(to *valobj.MessageTo) (*OutputWebSocketMessage, error) {
	marks, err := m.GetMessageMarks()
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	message, err := marks.ToMessageString()
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	return NewOutputWebSocketMessage(
		to, m.message_id,
		valobj.NewNameStringNil(),
		m.group_id, m.t,
		valobj.NewMessageStringNil(), message,
	), nil
}

func (m *MessageMark) ToOutputWebsocket() (*OutputWebSocketMessage, error) {
	var err error
	g := m.group_id.ToInt()
	game := GameGroup[g]
	if game == nil {
		LogStringf("GameGroup not found")
		return nil, ErrorInternal
	}
	var to *valobj.MessageTo
	if m.t.IsWriter() {
		to = valobj.NewMessageTo(game.writer)
	} else {
		to, err = ToGroupAllUser(m.group_id)
		if err != nil {
			Log(err)
			return nil, ErrorInternal
		}
	}
	return m.ToOutputWebsocketWithTo(to)
}

func (m *MessageMark) Valid() bool {
	var g int
	if !m.t.IsAnswer() && !m.t.IsWriter() {
		return false
	}
	if m.group_id == nil {
		return false
	}
	g = m.group_id.ToInt()
	if GameGroup[g] == nil {
		return false
	}
	return GameGroup[g].roles.IsWriter(m.uuid)
}

func (mark *MessageMark) ToString() string {
	return internal.MapToString(mark.GetMap())
}
