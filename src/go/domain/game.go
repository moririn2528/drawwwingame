package domain

import (
	"drawwwingame/domain/collection"
	"drawwwingame/domain/internal"
	"drawwwingame/domain/valobj"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var GameGroup [internal.GROUP_NUMBER]*Game
var WaitingRoom [internal.GROUP_NUMBER]*WaitingRoomStruct

type Game struct {
	group_id       *valobj.GroupIdInt
	member         []*valobj.UuidInt
	answer, writer []*valobj.UuidInt
	roles          *collection.RoleMap
	step           *valobj.CountingInt
	tim            *valobj.GameTimer
	theme          *valobj.MessageString //init: nil
	writer_number  int
	init_time      time.Time
}

func NewGame(group *valobj.GroupIdInt, steps, minutes, writer_number int) (*Game, error) {
	var err error
	if steps <= 0 || minutes <= 0 || writer_number <= 0 {
		return nil, ErrorArg
	}
	game := new(Game)
	game.group_id = group
	game.member = WaitingRoom[group.ToInt()].GetAllUuid()
	WaitingRoom[group.ToInt()].DeleteAll()
	game.step = valobj.NewCountingInt(0, steps)

	interval_func := func(t time.Duration) {
		err := game.sendInfoMessage("time", strconv.Itoa(int(math.Floor(t.Seconds()))))
		if err != nil {
			Log(err)
		}
	}
	end_func := func() {
		err := game.End(nil)
		if err != nil {
			Log(err)
		}
	}
	game.tim, err = valobj.NewGameTimer(time.Duration(minutes)*time.Minute,
		interval_func, end_func)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	game.writer_number = writer_number
	err = game.SplitMember()
	if err != nil {
		return nil, err
	}
	err = game.sendInitInfoMessage(valobj.NewMessageTo(game.member))
	if err != nil {
		Log(err)
		return nil, err
	}
	game.init_time = time.Now()
	GameGroup[game.group_id.ToInt()] = game
	return game, nil
}

func (game *Game) sendInitInfoMessage(to *valobj.MessageTo) error {

	message_line := []string{
		strconv.Itoa(game.step.GetMax()),
		strconv.Itoa(game.tim.Minutes()),
		strconv.Itoa(game.writer_number),
	}
	message := strings.Join(message_line, ":")
	return game.sendInfoMessageTo(to, "init", message)
}

func (game *Game) SplitMember() error {
	var err error
	var member []*GroupObject
	member, err = SqlHandle.GetGroupObjectsByUuid(game.member)
	viewer := []*valobj.UuidInt{}
	game.answer = []*valobj.UuidInt{}
	game.writer = []*valobj.UuidInt{}
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	var temp []*GroupObject
	for _, user := range member {
		if user.CanAnswer() && user.CanWriter() {
			temp = append(temp, user)
			continue
		}
		if user.CanAnswer() {
			game.answer = append(game.answer, user.GetUuid())
		} else if user.CanWriter() {
			game.writer = append(game.writer, user.GetUuid())
		} else {
			viewer = append(viewer, user.GetUuid())
		}
	}
	if game.writer_number < len(game.writer) {
		return ErrorArg
	}
	add_writer_number := game.writer_number - len(game.writer)
	if len(temp) < add_writer_number {
		return ErrorArg
	}
	for i := range temp {
		j := i + rand.Intn(len(temp)-i)
		if i == j {
			continue
		}
		temp[i], temp[j] = temp[j], temp[i]
	}
	for i, user := range temp {
		if i < add_writer_number {
			game.writer = append(game.writer, user.GetUuid())
		} else {
			game.answer = append(game.answer, user.GetUuid())
		}
	}

	game.roles = collection.NewRoleMap(game.answer, game.writer)
	return internal.LogAll(
		game.sendInfoMessageTo(valobj.NewMessageTo(game.answer),
			"role", "answer"),
		game.sendInfoMessageTo(valobj.NewMessageTo(game.writer),
			"role", "writer"),
		game.sendInfoMessageTo(valobj.NewMessageTo(viewer),
			"role", "viewer"),
		game.sendInfoMessageTo(valobj.NewMessageTo(game.member),
			"turn", strconv.Itoa(game.step.ToInt()+1)),
	)
}

func (game *Game) GetAnswerUuid() []*valobj.UuidInt {
	return game.answer
}
func (game *Game) GetWriterUuid() []*valobj.UuidInt {
	return game.writer
}

func (game *Game) SetTheme(name *valobj.MessageString) {
	if game.theme != nil {
		return
	}
	game.theme = name
}

func (game *Game) WriterUserSize() int {
	return len(game.writer)
}

func (game *Game) sendInfoMessageTo(to *valobj.MessageTo, info, message string) error {
	mes_info, err := valobj.NewMessageString("game:" + info)
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	mes, err := valobj.NewMessageString(message)
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	msg := NewOutputWebSocketMessage(
		to, valobj.NewMessageIdNil(),
		valobj.NewNameStringNil(), game.group_id,
		valobj.NewInfoMessageTypeString(),
		mes_info, mes,
	)
	Broadcast <- msg
	return nil
}

func (game *Game) sendInfoMessage(info, message string) error {
	return game.sendInfoMessageTo(valobj.NewMessageTo(game.member),
		info, message)
}

func (game *Game) Start() error {
	to := valobj.NewMessageTo(game.writer)
	err := game.sendInfoMessageTo(to, "start", game.theme.ToString())
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	to = valobj.NewMessageToExcept(game.member, game.writer)
	err = game.sendInfoMessageTo(to, "start", "")
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	game.tim.Start()
	return nil
}

func (game *Game) End(answer_id *valobj.MessageId) error {
	mes := make([]string, 3)
	var err error
	if answer_id != nil {
		game.tim.End()
		user, err := SqlHandle.GetUserFromMessageId(answer_id)
		if err != nil {
			Log(err)
			return ErrorInternal
		}
		mes[0] = "win"
		mes[2] = user.GetNameString()
	} else {
		mes[0] = "lose"
	}
	mes[1] = game.theme.ToString()

	err = game.sendInfoMessage("end", strings.Join(mes, ":"))
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	ok := game.step.Increment()
	if !ok {
		return game.Finish()
	}
	err = game.SplitMember()
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	game.theme = nil
	return nil
}

func (game *Game) Finish() error {
	err := game.sendInfoMessage("finish", "")
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	GameGroup[game.group_id.ToInt()] = nil
	return nil
}

func (game *Game) uuidIn(uuid *valobj.UuidInt, uuids []*valobj.UuidInt) bool {
	for _, u := range uuids {
		if u.ToInt() == uuid.ToInt() {
			return true
		}
	}
	return false
}

func (game *Game) InMember(uuid *valobj.UuidInt) bool {
	return game.uuidIn(uuid, game.member)
}
func (game *Game) IsWriter(uuid *valobj.UuidInt) bool {
	return game.uuidIn(uuid, game.writer)
}

func (game *Game) sendNowInfoMessage(uuid *valobj.UuidInt) error {
	to := valobj.NewMessageTo([]*valobj.UuidInt{uuid})
	theme := ""
	if game.IsWriter(uuid) && game.theme != nil {
		theme = game.theme.ToString()
	}
	tim_str := ""
	if game.tim.InProgress() {
		tim_str = strconv.Itoa(game.tim.LestTimeSeconds())
	}
	message_lines := []string{
		strconv.Itoa(game.step.ToInt() + 1),
		strconv.FormatBool(game.tim.InProgress()),
		tim_str, theme,
	}
	err := game.sendInfoMessageTo(to, "now", strings.Join(message_lines, ":"))
	if err != nil {
		return err
	}
	msg, mark, err := SqlHandle.GetGameMessage(game.group_id, game.init_time)
	if err != nil {
		Log(err)
		return err
	}
	info, err := valobj.NewMessageString("#before")
	if err != nil {
		LogStringf("error internal")
		return ErrorInternal
	}
	for _, m := range msg {
		out_msg, err := m.ToOutputWebsocketWithTo(to)
		if err != nil {
			Log(err)
			return err
		}
		out_msg.message_info = info
		Broadcast <- out_msg
	}
	for _, m := range mark {
		out_mark, err := m.ToOutputWebsocketWithTo(to)
		if err != nil {
			Log(err)
			return err
		}
		out_mark.message_info = info
		Broadcast <- out_mark
	}
	return nil
}

func (game *Game) SendInfoMessageBefore(uuid *valobj.UuidInt) error {
	if !game.InMember(uuid) {
		LogStringf("user is not member")
		return ErrorArg
	}
	to := valobj.NewMessageTo([]*valobj.UuidInt{uuid})
	err := game.sendInitInfoMessage(to)
	if err != nil {
		return err
	}
	var mes string
	if game.uuidIn(uuid, game.answer) {
		mes = "answer"
	} else if game.uuidIn(uuid, game.writer) {
		mes = "writer"
	} else {
		mes = "viewer"
	}
	err = game.sendInfoMessageTo(to, "role", mes)
	if err != nil {
		return err
	}
	err = game.sendNowInfoMessage(uuid)
	if err != nil {
		return err
	}
	return nil
}

type WaitingRoomStruct struct {
	group_id *valobj.GroupIdInt
	ready    map[int]bool
}

func NewWaitingRoomStruct(group_id *valobj.GroupIdInt) *WaitingRoomStruct {
	s := new(WaitingRoomStruct)
	s.ready = make(map[int]bool)
	s.group_id = group_id
	return s
}

func (room *WaitingRoomStruct) Append(uuid *valobj.UuidInt) {
	room.ready[uuid.ToInt()] = false
}

func (room *WaitingRoomStruct) Pop(uuid *valobj.UuidInt) {
	delete(room.ready, uuid.ToInt())
}

func (room *WaitingRoomStruct) AppendAllInGroup(group_id *valobj.GroupIdInt) error {
	uuids, err := SqlHandle.GetUuidInGroup(group_id)
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	for _, u := range uuids {
		_, ok := room.ready[u.ToInt()]
		if !ok {
			room.ready[u.ToInt()] = false
		}
	}
	return nil
}

func (room *WaitingRoomStruct) GetAllUuid() []*valobj.UuidInt {
	uuids := make([]*valobj.UuidInt, 0)
	for u, _ := range room.ready {
		uuid, err := valobj.NewUuidInt(u)
		if err != nil {
			delete(room.ready, u)
		}
		uuids = append(uuids, uuid)
	}
	return uuids
}

func (room *WaitingRoomStruct) DeleteAll() {
	room.ready = make(map[int]bool)
}

func (room *WaitingRoomStruct) SendInfoBefore(to_uuid *valobj.UuidInt) error {
	to := valobj.NewMessageTo([]*valobj.UuidInt{to_uuid})
	uuids := room.GetAllUuid()
	users, err := SqlHandle.GetUsersByUuid(uuids)
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	user_map := make(map[int]*User)
	for _, u := range users {
		user_map[u.GetUuidInt()] = u
	}
	send := func(uuid *valobj.UuidInt, str string) error {
		user, ok := user_map[uuid.ToInt()]
		if !ok {
			LogStringf("user no database")
			return ErrorInternal
		}
		mes, err := valobj.NewMessageString(str)
		if err != nil {
			Log(err)
			return ErrorArg
		}
		msg := NewMessageObject(
			valobj.NewMessageIdNil(), uuid, user.name,
			room.group_id, valobj.NewInfoMessageTypeString(),
			valobj.NewMessageStringNil(), mes,
		)
		out_msg, err := msg.ToOutputWebsocketWithTo(to)
		if err != nil {
			Log(err)
			return ErrorInternal
		}
		Broadcast <- out_msg
		return nil
	}
	for u, r := range room.ready {
		if u == to_uuid.ToInt() {
			continue
		}
		uuid, err := valobj.NewUuidInt(u)
		if err != nil {
			LogStringf("error NewUuidInt")
			return ErrorInternal
		}
		err = send(uuid, "join")
		if err != nil {
			return err
		}
		if r {
			err = send(uuid, "ready")
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (room *WaitingRoomStruct) Exist(uuid *valobj.UuidInt) bool {
	_, ok := room.ready[uuid.ToInt()]
	return ok
}

func (room *WaitingRoomStruct) UpdateReady(uuid *valobj.UuidInt, ready *valobj.Boolean) error {
	if !room.Exist(uuid) {
		LogStringf("no user in WaitingRoomStruct ready, %v", uuid)
		return ErrorArg
	}
	room.ready[uuid.ToInt()] = ready.ToBool()
	return nil
}

func (room *WaitingRoomStruct) IsReady(uuid *valobj.UuidInt) (bool, error) {
	if !room.Exist(uuid) {
		LogStringf("no user in WaitingRoomStruct ready, %v", uuid)
		return false, ErrorArg
	}
	return room.ready[uuid.ToInt()], nil
}

func (room *WaitingRoomStruct) GroupIsReady() bool {
	for _, ok := range room.ready {
		if !ok {
			return false
		}
	}
	return len(room.ready) > 0
}

func (room *WaitingRoomStruct) Leave(uuid *valobj.UuidInt) error {
	if !room.Exist(uuid) {
		return nil
	}
	room.Pop(uuid)

	users, err := SqlHandle.GetUsersByUuid([]*valobj.UuidInt{uuid})
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	if len(users) != 1 {
		LogStringf("users length: %v", len(users))
		return ErrorInternal
	}
	user := users[0]
	mes, err := valobj.NewMessageString("leave")
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	msg := NewMessageObject(
		valobj.NewMessageIdNil(), uuid, user.name,
		room.group_id, valobj.NewInfoMessageTypeString(),
		valobj.NewMessageStringNil(), mes,
	)
	all_user := room.GetAllUuid()
	to := valobj.NewMessageTo(all_user)
	out_msg, err := msg.ToOutputWebsocketWithTo(to)
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	Broadcast <- out_msg
	return nil
}

func initGame() error {
	for i := 0; i < internal.GROUP_NUMBER; i++ {
		group_id, err := valobj.NewGroupIdInt(i)
		if err != nil {
			Log(err)
			return err
		}
		WaitingRoom[i] = NewWaitingRoomStruct(group_id)
	}
	return nil
}
