package domain

import (
	"drawwwingame/domain/valobj"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	Broadcast = make(chan *OutputWebSocketMessage, 100)
	Clients   = NewClientsMap()
)

type ClientsMap struct {
	uuid  map[int]*websocket.Conn
	user  map[*websocket.Conn]*User
	group map[*websocket.Conn]*GroupObject
}

func DeleteConnectionByConn(ws *websocket.Conn) error {
	defer ws.Close()
	user, ok := Clients.user[ws]
	if !ok {
		LogStringf("delete error")
		return ErrorInternal
	}
	group, ok := Clients.group[ws]
	if !ok {
		LogStringf("delete error")
		return ErrorInternal
	}
	if group.InWaitingRoom() {
		err := group.LeaveFromWaitingRoom()
		if err != nil {
			Log(err)
			return ErrorInternal
		}
	}
	delete(Clients.uuid, user.GetUuidInt())
	delete(Clients.user, ws)
	delete(Clients.group, ws)
	return nil
}

func DeleteConnectionByUuid(uuid *valobj.UuidInt) error {
	ws, ok := Clients.uuid[uuid.ToInt()]
	if !ok {
		LogStringf("delete error")
		return ErrorInternal
	}
	return DeleteConnectionByConn(ws)
}

func NewClientsMap() *ClientsMap {
	s := new(ClientsMap)
	s.uuid = make(map[int]*websocket.Conn)
	s.user = make(map[*websocket.Conn]*User)
	s.group = make(map[*websocket.Conn]*GroupObject)
	return s
}

func (c ClientsMap) Append(ws *websocket.Conn, user *User, group *GroupObject) {
	c.uuid[user.GetUuidInt()] = ws
	c.user[ws] = user
	c.group[ws] = group
}

func (c ClientsMap) GetConnByUuid(uuid *valobj.UuidInt) (*websocket.Conn, bool) {
	u := uuid.ToInt()
	s, ok := c.uuid[u]
	if !ok {
		return nil, false
	}
	if s == nil {
		DeleteConnectionByUuid(uuid)
		return nil, false
	}
	return s, true
}

func (c ClientsMap) GetUserByConn(conn *websocket.Conn) (*User, bool) {
	u, ok := c.user[conn]
	if u == nil {
		DeleteConnectionByConn(conn)
		return nil, false
	}
	if !ok {
		return nil, false
	}
	return u, true
}

func (c ClientsMap) GetGroupByConn(conn *websocket.Conn) (*GroupObject, bool) {
	g, ok := c.group[conn]
	if g == nil {
		DeleteConnectionByConn(conn)
		return nil, false
	}
	if !ok {
		return nil, false
	}
	return g, true
}

func ConnectWebSocket(c echo.Context) (*websocket.Conn, error) {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return nil, err
	}
	return ws, nil
}

type InputWebSocketMessage struct {
	Uuid        string `json:"uuid"`
	Tempid      string `json:"tempid"`
	Name        string `json:"username"`
	GroupId     string `json:"group_id"`
	Type        string `json:"type"`
	MessageInfo string `json:"message_info"` // able null
	Message     string `json:"message"`
}

func NewInputWebSocketMessage(ws *websocket.Conn) (*InputWebSocketMessage, error) {
	s := new(InputWebSocketMessage)
	err := ws.ReadJSON(s)
	if err != nil {
		if websocket.IsCloseError(err,
			websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			log.Println(err)
			return nil, ErrorNoMatter
		} else {
			Log(err)
			return nil, ErrorConnection
		}
	}
	return s, nil
}

func (msg *InputWebSocketMessage) IsMark() bool {
	t, err := valobj.NewMessageTypeString(msg.Type)
	return err == nil && t.IsMark()
}
func (msg *InputWebSocketMessage) GetType() (*valobj.MessageTypeString, error) {
	return valobj.NewMessageTypeString(msg.Type)
}

func (msg *InputWebSocketMessage) GetUser() (*User, error) {
	uuid, err := valobj.NewUuidIntByString(msg.Uuid)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	tempid, err := valobj.NewTempIdString(msg.Tempid)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	user, err := NewUserById(uuid, tempid)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	if user.GetNameString() != msg.Name {
		LogStringf("input name error")
		return nil, ErrorArg
	}
	return user, nil
}

func (msg *InputWebSocketMessage) GetGroupObject() (*GroupObject, error) {
	uuid, err := valobj.NewUuidIntByString(msg.Uuid)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	group, err := NewExistGroupObjectById(uuid)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	return group, nil
}

func (msg *InputWebSocketMessage) GetnSetGroupObject() (*GroupObject, error) {
	uuid, err := valobj.NewUuidIntByString(msg.Uuid)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	group, err := valobj.NewGroupIdIntByString(msg.GroupId)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	return NewGetnSetGroupObjectById(uuid, group)
}

func (msg *InputWebSocketMessage) ToMessageMark() (*MessageMark, error) {
	uuid, err := valobj.NewUuidIntByString(msg.Uuid)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	t, err := valobj.NewMessageTypeString(msg.Type)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	if !t.IsMark() {
		LogStringf("type isnot mark")
		return nil, ErrorArg
	}
	id, err := valobj.NewMessageIdByString(msg.MessageInfo)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	group, err := valobj.NewGroupIdIntByString(msg.GroupId)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	mark, err := valobj.NewMessageMarkString(msg.Message)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	message_type, err := valobj.NewMessageTypeString(msg.Type)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	message_mark, err := NewMessageMark(uuid, group, id, mark, message_type)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	return message_mark, nil
}

func (msg *InputWebSocketMessage) ToMessageObject() (*MessageObject, error) {
	var id *valobj.MessageId
	var err error
	id = valobj.NewMessageId()
	uuid, err := valobj.NewUuidIntByString(msg.Uuid)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	name, err := valobj.NewNameString(msg.Name)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	group, err := valobj.NewGroupIdIntByString(msg.GroupId)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	t, err := valobj.NewMessageTypeString(msg.Type)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	if t.IsMark() {
		LogStringf("type is mark")
		return nil, ErrorArg
	}
	info, err := valobj.NewMessageString(msg.MessageInfo)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	message, err := valobj.NewMessageString(msg.Message)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	mes := NewMessageObject(id, uuid, name, group, t, info, message)
	if !mes.Valid() {
		LogStringf("message object isnot validate")
		return nil, ErrorInternal
	}
	return mes, nil
}

type OutputWebSocketJsonMessage struct {
	Id          int    `json:"id"`
	Name        string `json:"username"`
	GroupId     int    `json:"group_id"`
	Type        string `json:"type"`
	MessageInfo string `json:"message_info"` // able null
	Message     string `json:"message"`
}

type OutputWebSocketMessage struct {
	to           *valobj.MessageTo
	id           *valobj.MessageId
	name         *valobj.NameString
	group_id     *valobj.GroupIdInt
	t            *valobj.MessageTypeString
	message_info *valobj.MessageString
	message      *valobj.MessageString
}

func NewOutputWebSocketMessage(
	to *valobj.MessageTo,
	id *valobj.MessageId,
	name *valobj.NameString,
	group_id *valobj.GroupIdInt,
	t *valobj.MessageTypeString,
	message_info *valobj.MessageString,
	message *valobj.MessageString) *OutputWebSocketMessage {
	return &OutputWebSocketMessage{
		to:           to,
		id:           id,
		name:         name,
		group_id:     group_id,
		t:            t,
		message_info: message_info,
		message:      message,
	}
}

func (mes *OutputWebSocketMessage) ToJsonMessage() *OutputWebSocketJsonMessage {
	s := new(OutputWebSocketJsonMessage)
	s.Id = mes.id.ToInt()
	s.Name = mes.name.ToString()
	s.GroupId = mes.group_id.ToInt()
	s.Type = mes.t.ToString()
	s.MessageInfo = mes.message_info.ToString()
	s.Message = mes.message.ToString()
	return s
}

func (mes *OutputWebSocketMessage) Send() error {
	for _, u := range mes.to.GetUUids() {
		ws, ok := Clients.GetConnByUuid(u)
		if !ok {
			LogStringf("ClientsByUuid error, %v", u)
			return ErrorInternal
		}
		err := ws.WriteJSON(mes.ToJsonMessage())
		if err != nil {
			LogStringf("WriteJSON error")
			DeleteConnectionByConn(ws)
			return ErrorInternal
		}
	}
	return nil
}
