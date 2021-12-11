package domain

import (
	"drawwwingame/domain/valobj"
	"testing"
	"time"
)

func TestSqlUser(t *testing.T) {
	sql_user := SqlUser{
		Uuid:            1234,
		Tempid:          valobj.NewTempIdStringRandom().ToString(),
		Name:            "testuser",
		Email:           GetTestGmailAddress(0),
		ExpireTempidAt:  time.Now().Add(time.Hour),
		SendEmailCount:  1,
		SendLastEmailAt: time.Now().Add(-time.Hour * 100),
		EmailAuthorized: []byte{1},
	}
	user, err := sql_user.ToUser()
	check(t, "ToUser error", err, nil)
	check(t, "ToUser uuid", sql_user.Uuid, user.GetUuidInt())
	check(t, "ToUser tempid", sql_user.Tempid, user.tempid.ToString())
	check(t, "ToUser tempid expired", sql_user.ExpireTempidAt.Equal(user.tempid.GetExpiredAt()), true)
	check(t, "ToUser name", sql_user.Name, user.GetNameString())
	check(t, "ToUser email", sql_user.Email, user.email.EmailString())
	check(t, "ToUser email count", sql_user.SendEmailCount, user.email.GetCount())
	check(t, "ToUser email last", sql_user.SendLastEmailAt.Equal(user.email.GetLastDay()), true)
	check(t, "ToUser email auth", user.email.IsAuthorized(), true)
}

func TestSqlGroupUser(t *testing.T) {
	sql_group := SqlGroupUser{
		Uuid:      1234,
		GroupId:   0,
		Admin:     []byte{0x00},
		CanAnswer: []byte{0x01},
		CanWriter: []byte{0x01},
		Ready:     []byte{0x00},
	}
	role := sql_group.ToGroupRole()
	check(t, "ToGroupRole admin", role.IsAdmin(), false)
	check(t, "ToGroupRole can answer", role.CanAnswer(), true)
	check(t, "ToGroupRole admin", role.CanWriter(), true)
	check(t, "ToGroupReady", sql_group.ToGroupReady().ToBool(), false)
	group, err := sql_group.ToGroupObject()
	check(t, "ToGroupObject error", err, nil)
	check(t, "ToGroupObject uuid", group.GetUuid().ToInt(), sql_group.Uuid)
	check(t, "ToGroupObject group id", group.GetGroupId().ToInt(), sql_group.GroupId)
}

func TestSqlMessage(t *testing.T) {
	sql_mes := SqlMessage{
		Id:      valobj.NewMessageId().ToInt(),
		Uuid:    112345,
		GroupId: 0,
		Name:    "test123",
		T:       valobj.NewInfoMessageTypeString().ToString(),
		Info:    "test info",
		Str:     "test message",
	}
	mes, err := sql_mes.ToMessageObject()
	check(t, "ToMessageObject error", err, nil)
	check(t, "ToMessageObject uuid", mes.GetUuidInt(), sql_mes.Uuid)
	check(t, "ToMessageObject message id", mes.id.ToInt(), sql_mes.Id)
	check(t, "ToMessageObject group id", mes.group.ToInt(), sql_mes.GroupId)
	check(t, "ToMessageObject user name", mes.name.ToString(), sql_mes.Name)
	check(t, "ToMessageObject message type", mes.t.ToString(), sql_mes.T)
	check(t, "ToMessageObject message info", mes.info.ToString(), sql_mes.Info)
	check(t, "ToMessageObject message", mes.str.ToString(), sql_mes.Str)
}
