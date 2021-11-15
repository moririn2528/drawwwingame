package infra

import (
	"drawwwingame/domain"
	"log"
	"os"
	"reflect"
	"testing"
)

var (
	Handle *SqlHandler
)

func TestMain(m *testing.M) {
	domain.InitValTestLocal()
	status := m.Run()
	os.Exit(status)
}

func check(t *testing.T, arg, actual, expect interface{}) {
	if expect != actual {
		t.Errorf("\ninput: %v\noutput: %v\nexpect: %v", arg, actual, expect)
	}
}

func TestSqlHandler(t *testing.T) {
	Handle, err := NewSqlHandler()
	defer Handle.DeleteForTest()
	check(t, "NewSqlHandler", err, nil)
	err = Handle.Init()
	check(t, "Init", err, nil)
	name, err := domain.NewNameString("test77")
	check(t, "NewPasswordString", err, nil)
	e, err := domain.NewEmailString(domain.TEST_GMAIL_ADDRESS)
	check(t, "NewEmailString", err, nil)
	email := domain.NewEmailObject(e)
	user := domain.NewUser(domain.NewUuidIntRandom(), domain.NewTempIdValidStringRandom(), name, email)
	log.Printf("user.EmailAuthorized(): %v", user.EmailAuthorized())
	pass, err := domain.NewPasswordString("testpass")
	check(t, "NewPasswordString", err, nil)
	err = Handle.CreateUser(user, pass)
	check(t, "CreateUser", err, nil)
	u1, err := Handle.GetUser(name, pass)
	log.Printf("u1.EmailAuthorized(): %v", u1.EmailAuthorized)
	check(t, "GetUser", err, nil)
	uuid, err := domain.NewUuidInt(user.GetUuidInt())
	check(t, "NewUuidInt", err, nil)
	tempid, err := domain.NewTempIdString(user.GetTempidString())
	check(t, "NewUuidInt", err, nil)
	u2, err := Handle.GetUserById(uuid, tempid)
	check(t, "GetUserById", err, nil)
	check(t, "GetUser, GetUserById", reflect.DeepEqual(*u1, *u2), true)
	check(t, "user uuid", u1.Uuid, user.GetUuidInt())
	check(t, "user tempid string", u1.Tempid, user.GetTempidString())
	check(t, "user tempid expired at", u1.ExpireTempidAt.Format("2021/11/11 11:11:11"), user.GetTempidExpiredAt().Format("2021/11/11 11:11:11"))
	check(t, "user tempid name", u1.Name, user.GetNameString())
	check(t, "user tempid email", u1.Email, user.GetEmailString())
	check(t, "user tempid send email count", u1.SendEmailCount, user.GetSendEmailCount())
	check(t, "user tempid send email last day", u1.SendLastEmailAt.Format("2021/11/11 11:11:11"), user.GetSendEmailLastDay().Format("2021/11/11 11:11:11"))
	check(t, "user tempid email authorized", domain.NewBooleanByByte(u1.EmailAuthorized).ToBool(), user.EmailAuthorized())
	user2 := domain.NewUser(uuid, domain.NewTempIdValidStringRandom(), name, email)
	err = Handle.UpdateUser(user2)
	check(t, "UpdateUser", err, nil)
	check(t, "UpdateUser, tempid distinct", user2.GetTempidString() != user.GetTempidString(), true)
}
