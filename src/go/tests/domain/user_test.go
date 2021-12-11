package domain_test

import (
	"drawwwingame/domain"
	"drawwwingame/domain/valobj"
	"strconv"
	"testing"
)

func getTestUser(t *testing.T, user_ver int) (*domain.User, *valobj.PasswordString) {
	name, err := valobj.NewNameString("test33" + strconv.Itoa(user_ver))
	check(t, "NewNameString", err, nil)
	pass, err := valobj.NewPasswordString("126436483574ad")
	check(t, "NewPasswordString", err, nil)
	email, err := valobj.NewEmailString(domain.GetTestGmailAddress(user_ver))
	check(t, "NewEmailString", err, nil)
	user, err := domain.NewUsernCreate(name, email, pass)
	check(t, "NewUsernCreate", err, nil)
	return user, pass
}

func TestUser(t *testing.T) {
	user, pass := getTestUser(t, 0)
	m := user.GetAllMap()
	m2 := user.ToMapString()

	check(t, "map uuid", strconv.Itoa(m["uuid"].(int)), m2["uuid"])
	check(t, "map tempid", m["tempid"].(string), m2["tempid"])
	check(t, "map name", m["name"].(string), m2["username"])
	check(t, "map email", m["email"].(string), m2["email"])

	check(t, "User uuid int", m["uuid"].(int), user.GetUuidInt())
	check(t, "User name string", m["name"].(string), user.GetNameString())

	uuid, err := valobj.NewUuidInt(m["uuid"].(int))
	check(t, "User uuid error", err, nil)
	tempid, err := valobj.NewTempIdString(m["tempid"].(string))
	check(t, "User tempid error", err, nil)
	name, err := valobj.NewNameString(m["name"].(string))
	check(t, "User name error", err, nil)

	user2, err := domain.NewUserById(uuid, tempid)
	check(t, "NewUserById error", err, nil)
	check(t, "NewUserById", user.Equal(user2), true)
	err = user.UpdateTempid()
	check(t, "UpdateTempid error", err, nil)
	check(t, "UpdateTempid", user.Equal(user2), false)
	user2, err = domain.NewUserByNamePassword(name, pass)
	check(t, "NewUserByNamePassword error", err, nil)
	check(t, "NewUserByNamePassword", user.Equal(user2), true)

	name, err = valobj.NewNameString("test876")
	check(t, "NewNameString error", err, nil)
	email_name, err := valobj.NewEmailString(domain.GetTestGmailAddress(1))
	check(t, "NewEmailString error", err, nil)
	email := valobj.NewEmailObject(email_name)
	err = user.UpdateExceptId(name, email)
	check(t, "UpdateExceptId error", err, nil)
	check(t, "UpdateExceptId equal", user.Equal(user2), false)
	m = user.GetAllMap()
	check(t, "UpdateExceptId name", m["name"].(string), name.ToString())
	check(t, "UpdateExceptId email", m["email"].(string), email.EmailString())

	err = user.SendEmail("test", "test")
	check(t, "SendEmail error", err, nil)
	err = user.SendAuthorizeEmail("aa")
	check(t, "SendAuthorizeEmail error", err, nil)

	check(t, "EmailAuthorized", user.EmailAuthorized(), false)
	auth, err := domain.Encrypt(strconv.Itoa(m["uuid"].(int)) + "-" + m["tempid"].(string))
	check(t, "encrypt EmailAuthorized", err, nil)
	user, err = domain.Authorize(auth)
	check(t, "Authorize error", err, nil)
	check(t, "EmailAuthorized", user.EmailAuthorized(), true)
}

func equalGroupObject(t *testing.T, arg string, g1 *domain.GroupObject, g2 *domain.GroupObject) {
	m1 := g1.GetMap()
	m2 := g1.GetMap()
	for key, val := range m1 {
		check(t, arg+":"+key, val, m2[key])
	}
}

func TestGroupObject(t *testing.T) {
	user, _ := getTestUser(t, 0)
	defer domain.SqlHandle.DeleteForTest()
	group_id, err := valobj.NewGroupIdInt(0)
	check(t, "NewGroupIdInt", err, nil)
	group, err := user.UpdateGroup(group_id)
	check(t, "UpdateGroup error", err, nil)
	check(t, "UpdateGroup group id", group.GetGroupId().ToInt(), group_id.ToInt())
	check(t, "UpdateGroup uuid", group.GetUuid().ToInt(), user.GetUuidInt())
	group2, err := user.GetGroup()
	check(t, "GetGroup error", err, nil)
	equalGroupObject(t, "GetGroup", group, group2)
	group2, err = domain.NewExistGroupObjectById(group.GetUuid())
	check(t, "NewExistGroupObjectById error", err, nil)
	equalGroupObject(t, "NewExistGroupObjectById", group, group2)

	patterns := []int{1, 2, 4, 8, 7, 12, 15}

	for _, v := range patterns {
		err = group.UpdateRole(
			valobj.NewBoolean(v&1 > 0),
			valobj.NewBoolean(v&(1<<1) > 0),
			valobj.NewBoolean(v&(1<<2) > 0),
		)
		check(t, "UpdateRole error:"+strconv.Itoa(v), err, nil)
		err = group.UpdateReady(
			valobj.NewBoolean(v&(1<<3) > 0),
		)
		check(t, "UpdateReady error:"+strconv.Itoa(v), err, nil)
		check(t, "GroupRole admin:"+strconv.Itoa(v),
			group.IsAdmin(), v&1 > 0)
		check(t, "GroupRole can answer:"+strconv.Itoa(v),
			group.CanAnswer(), v&(1<<1) > 0)
		check(t, "GroupRole can writer:"+strconv.Itoa(v),
			group.CanWriter(), v&(1<<2) > 0)
		ready, err := group.GroupIsReady()
		check(t, "GroupRole ready error:"+strconv.Itoa(v), err, nil)
		check(t, "GroupRole ready:"+strconv.Itoa(v),
			ready, v&(1<<3) > 0)
	}
	err = group.StartGame(10, 7, 1)
	check(t, "StartGame error", err, nil)
	game := domain.GameGroup[0]
	check(t, "Game not nill", game != nil, true)
}
