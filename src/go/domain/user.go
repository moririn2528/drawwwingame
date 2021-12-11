package domain

import (
	"drawwwingame/domain/internal"
	"drawwwingame/domain/valobj"
	"log"
	"strconv"
	"strings"
)

type User struct {
	uuid   *valobj.UuidInt
	tempid *valobj.TempIdValidString
	name   *valobj.NameString
	email  *valobj.EmailObject
}

func NewUsernCreate(name *valobj.NameString, email *valobj.EmailString, password *valobj.PasswordString) (*User, error) {
	user := new(User)
	user.uuid = valobj.NewUuidIntRandom()
	user.tempid = valobj.NewTempIdValidStringRandom()
	user.name = name
	user.email = valobj.NewEmailObject(email)
	err := SqlHandle.CreateUser(user, password)
	return user, err
}

func NewUserById(uuid *valobj.UuidInt, tempid *valobj.TempIdString) (*User, error) {
	sql_user, err := SqlHandle.GetUserById(uuid, tempid)
	if err != nil {
		return nil, err
	}
	return sql_user.ToUser()
}

func NewUserByNamePassword(name *valobj.NameString, password *valobj.PasswordString) (*User, error) {
	sql_user, err := SqlHandle.GetUser(name, password)
	if err != nil {
		return nil, err
	}
	return sql_user.ToUser()
}

func NewUser(uuid *valobj.UuidInt, tempid *valobj.TempIdValidString, name *valobj.NameString, email *valobj.EmailObject) *User {
	user := new(User)
	user.uuid = uuid
	user.tempid = tempid
	user.name = name
	user.email = email
	return user
}

func (user *User) GetUuidInt() int {
	return user.uuid.ToInt()
}

// func (user *User) GetTempidString() string {
// 	return user.tempid.ToString()
// }
// func (user *User) GetTempidExpiredAt() time.Time {
// 	return user.tempid.GetExpiredAt()
// }
func (user *User) GetNameString() string {
	return user.name.ToString()
}

// func (user *User) GetEmailString() string {
// 	return user.email.EmailString()
// }
// func (user *User) GetSendEmailCount() int {
// 	return user.email.GetCount()
// }
// func (user *User) GetSendEmailLastDay() time.Time {
// 	return user.email.GetLastDay()
// }

func (user *User) ToMapString() map[string]string {
	return map[string]string{
		"uuid":     strconv.Itoa(user.uuid.ToInt()),
		"tempid":   user.tempid.ToString(),
		"username": user.name.ToString(),
		"email":    user.email.EmailString(),
	}
}

func (user *User) UpdateTempid() error {
	user.tempid = valobj.NewTempIdValidStringRandom()
	err := SqlHandle.UpdateUser(user)
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	return nil
}

func (user *User) UpdateExceptId(name *valobj.NameString, email *valobj.EmailObject) error {
	if name != nil {
		user.name = name
	}
	if email != nil {
		user.email = email
	}
	err := SqlHandle.UpdateUser(user)
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	return nil
}

func (user *User) SendEmail(subject, message string) error {
	if !user.email.CanSendNow() {
		return ErrorSendingEmailLimit
	}
	var err error
	user.email, err = user.email.SendEmail(subject, message)
	if err != nil {
		return err
	}

	err = SqlHandle.UpdateUser(user)
	if err != nil {
		return err
	}
	return nil
}

func (user *User) EmailAuthorized() bool {
	return user.email.IsAuthorized()
}

func (user *User) SendAuthorizeEmail(path string) error {
	if user.EmailAuthorized() {
		return ErrorUnnecessary
	}
	str, err := Encrypt(strconv.Itoa(user.uuid.ToInt()) + "-" + user.tempid.ToString())
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	if !user.email.CanSendNow() {
		return ErrorSendingEmailLimit
	}
	path_all := internal.WORK_SPACE + path + "/" + str
	email, err := user.email.SendEmail("メール認証",
		"以下のリンクを踏んでメールアドレスを認証してください。\n"+
			path_all+"\n"+
			"このメールに心当たりがない方はお問い合わせください。\n",
	)
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	user.email = email
	err = SqlHandle.UpdateUser(user)
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	return nil
}

func (user *User) Equal(u *User) bool {
	return user.uuid.ToInt() == u.uuid.ToInt() &&
		user.tempid.ToString() == u.tempid.ToString() &&
		user.name.ToString() == u.name.ToString() &&
		user.email.EmailString() == u.email.EmailString()
}

func (user *User) GetAllMap() map[string]interface{} {
	return map[string]interface{}{
		"uuid":               user.uuid.ToInt(),
		"tempid":             user.tempid.ToString(),
		"name":               user.name.ToString(),
		"email":              user.email.EmailString(),
		"expire_tempid_at":   user.tempid.GetExpiredAt(),
		"send_email_count":   user.email.GetCount(),
		"send_last_email_at": user.email.GetLastDay(),
		"email_authorized":   user.email.IsAuthorized(),
	}
}

func Authorize(encrypt_str string) (*User, error) {
	str, err := Decrypt(encrypt_str)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	log.Println(str)
	if !strings.Contains(str, "-") {
		LogStringf("string hyphen error")
		return nil, ErrorInternal
	}
	str_arr := strings.Split(str, "-")
	uuid_int, err := strconv.Atoi(str_arr[0])
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	uuid, err := valobj.NewUuidInt(uuid_int)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	tempid, err := valobj.NewTempIdString(str_arr[1])
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	user, err := NewUserById(uuid, tempid)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	if user.email.IsAuthorized() {
		LogStringf("user")
		return nil, ErrorUnnecessary
	}
	user.email = user.email.Authorize()
	err = SqlHandle.UpdateUser(user)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	return user, nil
}

func (user *User) GetUuid() *valobj.UuidInt {
	return user.uuid
}
func (user *User) GetGroup() (*GroupObject, error) {
	return NewExistGroupObjectById(user.uuid)
}
func (user *User) UpdateGroup(group_id *valobj.GroupIdInt) (*GroupObject, error) {
	return SqlHandle.GetnSetGroupObjectById(user.uuid, group_id)
}

func (user *User) ToString() string {
	return internal.MapToString(user.GetAllMap())
}

type GroupObject struct {
	uuid *valobj.UuidInt
	id   *valobj.GroupIdInt
	role *valobj.GroupRole
}

func NewGroupObject(uuid *valobj.UuidInt, id *valobj.GroupIdInt, role *valobj.GroupRole) *GroupObject {
	s := new(GroupObject)
	s.uuid = uuid
	s.id = id
	s.role = role
	return s
}

func NewGroupObjectByRow(uuid, id int, admin, canAnswer, canWriter bool) (*GroupObject, error) {
	s := new(GroupObject)
	var err error
	s.uuid, err = valobj.NewUuidInt(uuid)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	s.id, err = valobj.NewGroupIdInt(id)
	if err != nil {
		Log(err)
		return nil, ErrorArg
	}
	s.role = valobj.NewGroupRoleByRow(admin, canWriter, canAnswer)
	return s, nil
}

func NewGetnSetGroupObjectById(uuid *valobj.UuidInt, group_id *valobj.GroupIdInt) (*GroupObject, error) {
	return SqlHandle.GetnSetGroupObjectById(uuid, group_id)
}

func NewExistGroupObjectById(uuid *valobj.UuidInt) (*GroupObject, error) {
	return SqlHandle.GetGroupObjectByUuid(uuid)
}

func (group *GroupObject) GetGroupId() *valobj.GroupIdInt {
	return group.id
}
func (group *GroupObject) GetUuid() *valobj.UuidInt {
	return group.uuid
}
func (group *GroupObject) IsAdmin() bool {
	return group.role.IsAdmin()
}
func (group *GroupObject) CanWriter() bool {
	return group.role.CanWriter()
}
func (group *GroupObject) CanAnswer() bool {
	return group.role.CanAnswer()
}

var GroupObjectMapKeys = []string{
	"uuid", "group_id", "admin", "can_answer", "can_writer",
}

func (group *GroupObject) GetMap() map[string]interface{} {
	return map[string]interface{}{
		"uuid":       group.uuid.ToInt(),
		"group_id":   group.id.ToInt(),
		"admin":      group.role.IsAdmin(),
		"can_answer": group.role.CanAnswer(),
		"can_writer": group.role.CanWriter(),
	}
}
func (group *GroupObject) SaveSql() error {
	return SqlHandle.UpdateGroupObject(group)
}
func (group *GroupObject) UpdateRole(admin, can_answer, can_writer *valobj.Boolean) error {
	updated := group.role.Update(admin, can_answer, can_writer)
	if !updated {
		return nil
	}
	err := SqlHandle.UpdateGroupObject(group)
	if err != nil {
		return ErrorInternal
	}
	return nil
}

func (group *GroupObject) StartGame(steps, minutes, write_number int) error {
	if GameGroup[group.id.ToInt()] != nil {
		LogStringf("duplicate error")
		return ErrorDuplicate
	}
	game, err := NewGame(group.id, steps, minutes, write_number)
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	GameGroup[group.id.ToInt()] = game
	return nil
}

func (group *GroupObject) Equal(g *GroupObject) bool {
	return group.uuid.Equal(g.uuid) &&
		group.id.Equal(g.id) &&
		group.role.Equal(g.role)
}

func (group *GroupObject) InWaitingRoom() bool {
	id := group.id.ToInt()
	return WaitingRoom[id].Exist(group.uuid)
}
func (group *GroupObject) AppendWaitingRoom() {
	id := group.id.ToInt()
	WaitingRoom[id].Append(group.uuid)
}
func (group *GroupObject) PopWaitingRoom() {
	id := group.id.ToInt()
	WaitingRoom[id].Pop(group.uuid)
}
func (group *GroupObject) SendWaitingRoomInfoBefore() error {
	id := group.id.ToInt()
	return WaitingRoom[id].SendInfoBefore(group.uuid)
}
func (group *GroupObject) UpdateReady(ready *valobj.Boolean) error {
	id := group.id.ToInt()
	return WaitingRoom[id].UpdateReady(group.uuid, ready)
}
func (group *GroupObject) GroupIsReady() bool {
	id := group.id.ToInt()
	ok := WaitingRoom[id].GroupIsReady()
	if !ok {
		return false
	}
	return GameGroup[id] == nil
}
func (group *GroupObject) LeaveFromWaitingRoom() error {
	id := group.id.ToInt()
	return WaitingRoom[id].Leave(group.uuid)
}

func (group *GroupObject) InGame() bool {
	id := group.id.ToInt()
	if GameGroup[id] == nil {
		return false
	}
	return GameGroup[id].InMember(group.uuid)
}
func (group *GroupObject) SendGameInfoBefore() error {
	id := group.id.ToInt()
	return GameGroup[id].SendInfoMessageBefore(group.uuid)
}

func (group *GroupObject) ToString() string {
	return internal.MapToString(group.GetMap())
}
