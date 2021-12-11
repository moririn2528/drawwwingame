package domain_test

import (
	"drawwwingame/domain"
	"drawwwingame/domain/valobj"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
)

func TestMessageObject(t *testing.T) {
	user0, group0 := getTestUsernGroup(t, 0)
	defer domain.SqlHandle.DeleteForTest()

	messages := [][3]string{
		{"info", "", "test"},
		{"info", "start:5:5:1", "start"},
		{"lines", "", "#111111,1,1.0,10:10-20:20-5:15"},
		{"mark:answer", "", "1:A"},
		{"mark:writer", "", "1:B"},
		{"text:answer", "", "213425bfnsdzoinあたら"},
		{"text:writer", "", "テスト"},
	}
	for _, arg := range messages {
		ty, err := valobj.NewMessageTypeString(arg[0])
		check(t, "NewMessageTypeString: "+arg[0], err, nil)
		str_info, err := valobj.NewMessageString(arg[1])
		check(t, "NewMessageString: "+arg[1], err, nil)
		str, err := valobj.NewMessageString(arg[2])
		check(t, "NewMessageString: "+arg[2], err, nil)
		var id *valobj.MessageId
		if rand.Intn(2) == 0 {
			id = valobj.NewMessageId()
		}
		message := domain.NewMessageObjectByUser(user0, id, group0.GetGroupId(),
			ty, str_info, str)
		check(t, fmt.Sprintf("GetTypeString: %v", arg),
			message.GetTypeString(), arg[0])
		check(t, fmt.Sprintf("GetMessageString: %v", arg),
			message.GetMessageString(), arg[2])
		check(t, fmt.Sprintf("GetUuidInt: %v", arg),
			message.GetUuidInt(), user0.GetUuidInt())
		check(t, fmt.Sprintf("GetGroupId: %v", arg),
			message.GetGroupId(), group0.GetGroupId().ToInt())
		check(t, fmt.Sprintf("GetNameString: %v", arg),
			message.GetNameString(), user0.GetNameString())

		err = message.SaveSql()
		check(t, fmt.Sprintf("SaveSql error: %v", arg), err, nil)
		_, err = message.ToOutputWebsocket()
		check(t, fmt.Sprintf("ToOutputWebsocket error: %v", arg), err, nil)
	}
	for _, arg := range messages {
		ty, err := valobj.NewMessageTypeString(arg[0])
		check(t, "NewMessageTypeString: "+arg[0], err, nil)
		str_info, err := valobj.NewMessageString(arg[1])
		check(t, "NewMessageString: "+arg[1], err, nil)
		str, err := valobj.NewMessageString(arg[2])
		check(t, "NewMessageString: "+arg[2], err, nil)
		var id *valobj.MessageId
		if rand.Intn(2) == 0 {
			id = valobj.NewMessageId()
		}
		uuid, err := valobj.NewUuidInt(user0.GetUuidInt())
		check(t, "NewUuidInt", err, nil)
		name, err := valobj.NewNameString(user0.GetNameString())
		check(t, "NewNameString", err, nil)
		message := domain.NewMessageObject(id, uuid, name, group0.GetGroupId(), ty, str_info, str)
		check(t, fmt.Sprintf("GetTypeString: %v", arg),
			message.GetTypeString(), arg[0])
		check(t, fmt.Sprintf("GetMessageString: %v", arg),
			message.GetMessageString(), arg[2])
		check(t, fmt.Sprintf("GetUuidInt: %v", arg),
			message.GetUuidInt(), user0.GetUuidInt())
		check(t, fmt.Sprintf("GetGroupId: %v", arg),
			message.GetGroupId(), group0.GetGroupId().ToInt())
		check(t, fmt.Sprintf("GetNameString: %v", arg),
			message.GetNameString(), user0.GetNameString())
	}
	for i, arg := range messages {
		id, err := valobj.NewMessageIdByInt(i)
		check(t, "NewMessageIdByInt error: "+strconv.Itoa(i), err, nil)
		mes, err := domain.NewGetMessageObject(id)
		check(t, "NewGetMessageObject error: "+strconv.Itoa(i), err, nil)
		check(t, "NewGetMessageObject GetTypeString: "+strconv.Itoa(i),
			mes.GetTypeString(), arg[0])
		check(t, "NewGetMessageObject GetMessageString: "+strconv.Itoa(i),
			mes.GetMessageString(), arg[2])
	}
}

func TestMessageMark(t *testing.T) {
	user0, group0 := getTestUsernGroup(t, 0)
	user1, _ := getTestUsernGroup(t, 1)
	patterns := [][4]interface{}{
		{user0, "A", 0, "1:0:0:2"},
		{user1, "C", 0, "1:0:1:2"},
		{user1, "A", 0, "2:0:0:2"},
		{user0, "B", 1, "0:1:0:2"},
		{user1, "A", 1, "1:1:0:2"},
	}
	name, err := valobj.NewNameString(user0.GetNameString())
	check(t, "NewNameString error", err, nil)
	for i := 0; i < 2; i++ {
		ty, err := valobj.NewMessageTypeString("text:answer")
		check(t, "NewMessageTypeString error", err, nil)
		msg, err := valobj.NewMessageString("test")
		check(t, "NewMessageTypeString error", err, nil)
		id := valobj.NewMessageId()
		mes := domain.NewMessageObject(id, user0.GetUuid(), name, group0.GetGroupId(), ty, msg, msg)
		mes.SaveSql()
	}
	mark_type, err := valobj.NewMessageTypeString("mark:answer")
	check(t, "NewMessageTypeString error", err, nil)
	group_id, err := valobj.NewGroupIdInt(0)
	check(t, "NewGroupIdInt error", err, nil)
	_, err = domain.NewGame(group_id, 10, 10, 1)
	check(t, "NewGame error", err, nil)
	for _, arg := range patterns {
		mark, err := valobj.NewMessageMarkString(arg[1].(string))
		check(t, "NewMessageMarkString error", err, nil)
		id, err := valobj.NewMessageIdByInt(arg[2].(int))
		check(t, "NewMessageIdByInt error", err, nil)
		m := domain.NewMessageMarkByUser(arg[0].(*domain.User), group_id, id, mark, mark_type)

		err = m.SaveSql()
		check(t, "SaveSql error", err, nil)
		marks, err := m.GetMessageMarks()
		check(t, "GetMessageMarks error", err, nil)
		_, err = m.ToOutputWebsocket()
		check(t, "ToOutputWebsocket error", err, nil)
		mes, err := marks.ToMessageString()
		check(t, "ToMessageString error", err, nil)
		check(t, "ToMessageString", mes.ToString(), arg[3].(string))
	}
}
