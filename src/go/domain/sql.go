package domain

import (
	"drawwwingame/domain/internal"
	"drawwwingame/domain/valobj"
	"time"
)

type SqlHandler interface {
	Init() error
	GetUserById(*valobj.UuidInt, *valobj.TempIdString) (*SqlUser, error)
	GetUsersByUuid([]*valobj.UuidInt) ([]*User, error)
	GetUser(*valobj.NameString, *valobj.PasswordString) (*SqlUser, error)
	CreateUser(*User, *valobj.PasswordString) error
	UpdateUser(*User) error
	GetGroupUserCount(*valobj.GroupIdInt) (int, error)
	CreateMessage(*MessageObject) error
	GetMessage(*valobj.MessageId) (*MessageObject, error)
	GetUserInGroup(*valobj.GroupIdInt) ([]*User, error)
	GetUuidInGroup(*valobj.GroupIdInt) ([]*valobj.UuidInt, error)
	SetMessageMark(*MessageMark) error
	GetMarksOnMessage(*valobj.MessageId) ([]int, error)
	GetAllGroupObjectByGroupId(*valobj.GroupIdInt) ([]*GroupObject, error)
	GetGroupObjectsByUuid([]*valobj.UuidInt) ([]*GroupObject, error)
	GetGroupObjectByUuid(*valobj.UuidInt) (*GroupObject, error)
	SetAllGroupObject([]*GroupObject) error
	GetnSetGroupObjectById(*valobj.UuidInt, *valobj.GroupIdInt) (*GroupObject, error)
	UpdateGroupObject(*GroupObject) error
	DeleteForTest()
	GetUserFromMessageId(*valobj.MessageId) (*User, error)
	GetGameMessage(*valobj.GroupIdInt, time.Time) ([]*MessageObject, []*MessageMark, error)
}

var (
	SqlHandle     SqlHandler
	SqlUserMapKey = []string{
		"uuid", "tempid", "name", "email", "expire_tempid_at",
		"send_email_count", "send_last_email_at", "email_authorized",
	}
)

type SqlUser struct {
	Uuid            int       `db:"uuid"`
	Tempid          string    `db:"tempid"`
	Name            string    `db:"name"`
	Email           string    `db:"email"`
	ExpireTempidAt  time.Time `db:"expire_tempid_at"`
	SendEmailCount  int       `db:"send_email_count"`
	SendLastEmailAt time.Time `db:"send_last_email_at"`
	EmailAuthorized []byte    `db:"email_authorized"`
}

func (sql_user *SqlUser) ToUser() (*User, error) {
	update := false
	var err error
	user := new(User)
	user.uuid, err = valobj.NewUuidInt(sql_user.Uuid)
	if err != nil {
		return nil, err
	}
	user.tempid, err = valobj.NewTempIdValidString(sql_user.Tempid, sql_user.ExpireTempidAt)
	if err != nil {
		if err != internal.ErrorExpired {
			return nil, err
		}
		// expired
		update = true
		user.tempid = valobj.NewTempIdValidStringRandom()
	}
	user.name, err = valobj.NewNameString(sql_user.Name)
	if err != nil {
		return nil, err
	}
	email, err := valobj.NewEmailString(sql_user.Email)
	if err != nil {
		return nil, err
	}
	user.email, err = valobj.NewEmailObjectnSet(email, sql_user.SendEmailCount, sql_user.SendLastEmailAt, valobj.NewBooleanByByte(sql_user.EmailAuthorized))
	if err != nil {
		return nil, err
	}
	if update {
		err = SqlHandle.UpdateUser(user)
		if err != nil {
			return nil, err
		}
	}
	return user, nil
}

type SqlGroupUser struct {
	Uuid      int    `db:"uuid"`
	GroupId   int    `db:"group_id"`
	Admin     []byte `db:"admin"`
	CanAnswer []byte `db:"can_answer"`
	CanWriter []byte `db:"can_writer"`
}

func (sql_group *SqlGroupUser) ToGroupRole() *valobj.GroupRole {
	admin := valobj.NewBooleanByByte(sql_group.Admin)
	can_answer := valobj.NewBooleanByByte(sql_group.CanAnswer)
	can_writer := valobj.NewBooleanByByte(sql_group.CanWriter)
	return valobj.NewGroupRole(admin, can_answer, can_writer)
}

func (sql_group *SqlGroupUser) ToGroupObject() (*GroupObject, error) {
	uuid, err := valobj.NewUuidInt(sql_group.Uuid)
	if err != nil {
		return nil, err
	}
	group_id, err := valobj.NewGroupIdInt(sql_group.GroupId)
	if err != nil {
		return nil, err
	}
	group := NewGroupObject(uuid, group_id, sql_group.ToGroupRole())
	return group, nil
}

type SqlMessage struct {
	Id      int    `db:"id"`
	Uuid    int    `db:"uuid"`
	GroupId int    `db:"group_id"`
	Name    string `db:"name"`
	T       string `db:"type"`
	Info    string `db:"info"`
	Str     string `db:"message"`
}

func (mes *SqlMessage) ToMessageObject() (*MessageObject, error) {
	id, err := valobj.NewMessageIdByInt(mes.Id)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	uuid, err := valobj.NewUuidInt(mes.Uuid)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	group_id, err := valobj.NewGroupIdInt(mes.GroupId)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	name, err := valobj.NewNameString(mes.Name)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	t, err := valobj.NewMessageTypeString(mes.T)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	info, err := valobj.NewMessageString(mes.Info)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	str, err := valobj.NewMessageString(mes.Str)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	return NewMessageObject(id, uuid, name, group_id, t, info, str), nil
}

type SqlMessageMark struct {
	Uuid        int    `db:"uuid"`
	Groupid     int    `db:"group_id"`
	MessageId   int    `db:"message_id"`
	Mark        string `db:"mark"`
	MessageType string `db:"type"`
}

func (m *SqlMessageMark) ToMessageMark() (*MessageMark, error) {
	uuid, err := valobj.NewUuidInt(m.Uuid)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	groupid, err := valobj.NewGroupIdInt(m.Groupid)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	id, err := valobj.NewMessageIdByInt(m.MessageId)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	mark, err := valobj.NewMessageMarkString(m.Mark)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	message_type, err := valobj.NewMessageTypeString(m.MessageType)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	return NewMessageMark(uuid, groupid, id, mark, message_type)
}
