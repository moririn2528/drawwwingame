package domain

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	WEBSOCKET_READ_BUFFER_SIZE  = 1024
	WEBSOCKET_WRITE_BUFFER_SIZE = 1024
	DATABASE_WORKPLACE          = "drawwwingame"
	DATABASE_INIT_INFO          = "user483:Te9SLqyciALe@tcp(127.0.0.1:3306)/drawwwingame?parseTime=true&loc=Asia%2FTokyo"
	DATABASE                    = "mysql"
	ALPHANUM_CHAR               = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	ALPHANUM_SIZE               = 20
	WORK_SPACE                  = "localhost:1213/"
	GROUP_NUMBER                = 20
)

var (
	DEBUG_MODE         = true
	GMAIL_PASSWORD     string
	GMAIL_ADDRESS      string
	TEST_GMAIL_ADDRESS string
)

var ( ///error
	ErrorParse              = errors.New("parse error")
	ErrorNoUser             = errors.New("no user")
	ErrorDuplicateUserName  = errors.New("duplicate user name")
	ErrorDuplicateUserEmail = errors.New("duplicate user email")
	ErrorDuplicateUserId    = errors.New("duplicate user id")
	ErrorString             = errors.New("invalid string")
	ErrorInt                = errors.New("invalid int")
	ErrorExpired            = errors.New("expired")
	ErrorSendingEmailLimit  = errors.New("send email limit")
	ErrorUnnecessary        = errors.New("unnecessary")
	ErrorInternal           = errors.New("iternal error")
	ErrorArg                = errors.New("argument error")
)

func createLog(str string, dep int) string {
	pc, file, line, _ := runtime.Caller(dep)
	f := runtime.FuncForPC(pc)
	s := fmt.Sprintf("\ncall:%s\nfile:%s:%d\nerror:%v\n", f.Name(), file, line, str)
	return s
}

func Log(err error) {
	if !DEBUG_MODE {
		return
	}
	log.Println(createLog(err.Error(), 2))
}
func NewError(err string) error {
	return errors.New(strings.Replace(createLog(err, 2), "\n", " ", -1))
}
func LogStringf(str string, arg ...interface{}) {
	if !DEBUG_MODE {
		return
	}
	log.Println(createLog(fmt.Sprintf(str, arg...), 2))
}
func ErrorIsEach(err error, errs ...error) {
	if err == nil {
		return
	}
	for _, e := range errs {
		if err == e {
			return
		}
	}
	createLog(fmt.Sprintf("error not contain, %v", err), 2)
}

type Message struct {
	Uuid    string `json:"uuid"`
	Tempid  string `json:"tempid"`
	Name    string `json:"username"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

//websocket
var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	Clients          = make(map[*websocket.Conn]*User)
	Broadcast        = make(chan Message)
	Goroutine_cancel = make(chan struct{})
)

func ConnectWebSocket(c echo.Context) (*websocket.Conn, error) {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return nil, err
	}
	return ws, nil
}

//user
type SqlUser struct {
	Uuid            int       `db:"uuid"`
	Tempid          string    `db:"tempid"`
	Name            string    `db:"name"`
	Email           string    `db:"email"`
	ExpireTempidAt  time.Time `db:"expire_tempid_at"`
	SendEmailCount  int       `db:"send_email_count"`
	SendLastEmailAt time.Time `db:"send_last_email_at"`
	EmailAuthorized []byte    `db:"email_authorized"`
	GroupId         int       `db:"group_id"`
}

type SqlHandler interface {
	Init() error
	GetUserById(*UuidInt, *TempIdString) (*SqlUser, error)
	GetUser(*NameString, *PasswordString) (*SqlUser, error)
	CreateUser(*User, *PasswordString) error
	UpdateUser(*User) error
}

var (
	SqlHandle SqlHandler
)

type User struct {
	uuid   *UuidInt
	tempid *TempIdValidString
	name   *NameString
	email  *EmailObject
	group  *GroupIdInt
}

func (sql_user *SqlUser) ToUser() (*User, error) {
	update := false
	var err error
	user := new(User)
	user.uuid, err = NewUuidInt(sql_user.Uuid)
	if err != nil {
		return nil, err
	}
	user.tempid, err = NewTempIdValidString(sql_user.Tempid, sql_user.ExpireTempidAt)
	if err != nil {
		if err != ErrorExpired {
			return nil, err
		}
		// expired
		update = true
		user.tempid = NewTempIdValidStringRandom()
	}
	user.name, err = NewNameString(sql_user.Name)
	if err != nil {
		return nil, err
	}
	email, err := NewEmailString(sql_user.Email)
	if err != nil {
		return nil, err
	}
	user.email, err = NewEmailObjectnSet(email, sql_user.SendEmailCount, sql_user.SendLastEmailAt, NewBooleanByByte(sql_user.EmailAuthorized))
	if err != nil {
		return nil, err
	}
	if update {
		err = SqlHandle.UpdateUser(user)
		if err != nil {
			return nil, err
		}
	}
	user.group, err = NewGroupIdInt(sql_user.GroupId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func NewUsernCreate(name *NameString, email *EmailString, password *PasswordString) (*User, error) {
	user := new(User)
	user.uuid = NewUuidIntRandom()
	user.tempid = NewTempIdValidStringRandom()
	user.name = name
	user.email = NewEmailObject(email)
	user.group = NewGroupIdIntNoSet()
	err := SqlHandle.CreateUser(user, password)
	return user, err
}

func NewUserById(uuid *UuidInt, tempid *TempIdString) (*User, error) {
	sql_user, err := SqlHandle.GetUserById(uuid, tempid)
	if err != nil {
		return nil, err
	}
	return sql_user.ToUser()
}

func NewUserByNamePassword(name *NameString, password *PasswordString) (*User, error) {
	sql_user, err := SqlHandle.GetUser(name, password)
	if err != nil {
		return nil, err
	}
	return sql_user.ToUser()
}

func NewUser(uuid *UuidInt, tempid *TempIdValidString, name *NameString, email *EmailObject, group *GroupIdInt) *User {
	user := new(User)
	user.uuid = uuid
	user.tempid = tempid
	user.name = name
	user.email = email
	user.group = group
	return user
}

func (user *User) GetMapString() map[string]string {
	return map[string]string{
		"uuid":     strconv.Itoa(user.uuid.ToInt()),
		"tempid":   user.tempid.ToString(),
		"username": user.name.ToString(),
		"email":    user.email.EmailString(),
		"group_id": strconv.Itoa(user.group.ToInt()),
	}
}

func (user *User) GetUuidInt() int {
	return user.uuid.ToInt()
}
func (user *User) GetTempidString() string {
	return user.tempid.ToString()
}
func (user *User) GetTempidExpiredAt() time.Time {
	return user.tempid.GetExpiredAt()
}
func (user *User) GetNameString() string {
	return user.name.ToString()
}
func (user *User) GetEmailString() string {
	return user.email.EmailString()
}
func (user *User) GetSendEmailCount() int {
	return user.email.GetCount()
}
func (user *User) GetSendEmailLastDay() time.Time {
	return user.email.GetLastDay()
}
func (user *User) GetGoupId() int {
	return user.group.ToInt()
}

func (user *User) ToMapString() map[string]string {
	return map[string]string{
		"uuid":   strconv.Itoa(user.uuid.ToInt()),
		"tempid": user.tempid.ToString(),
		"name":   user.name.ToString(),
		"email":  user.email.EmailString(),
	}
}

func (user *User) UpdateTempid() (*User, error) {
	new_user := NewUser(user.uuid, NewTempIdValidStringRandom(), user.name, user.email, user.group)
	err := SqlHandle.UpdateUser(new_user)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	return new_user, nil
}

func (user *User) UpdateExceptId(name *NameString, email *EmailObject, group_id *GroupIdInt) (*User, error) {
	u := new(User)
	u.uuid = user.uuid
	u.tempid = user.tempid
	if name == nil {
		u.name = user.name
	} else {
		u.name = name
	}
	if email == nil {
		u.email = user.email
	} else {
		u.email = email
	}
	if group_id == nil {
		u.group = user.group
	} else {
		u.group = group_id
	}
	err := SqlHandle.UpdateUser(u)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	return u, nil
}

func (user *User) SendEmail(subject, message string) (*User, error) {
	if !user.email.count.canSendNow() {
		return nil, ErrorSendingEmailLimit
	}
	email, err := user.email.SendEmail(subject, message)
	if err != nil {
		return nil, err
	}
	new_user := NewUser(user.uuid, user.tempid, user.name, email, user.group)
	err = SqlHandle.UpdateUser(new_user)
	if err != nil {
		return nil, err
	}
	return new_user, nil
}

func (user *User) EmailAuthorized() bool {
	return user.email.IsAuthorized()
}

func (user *User) SendAuthorizeEmail(path string) (*User, error) {
	if user.EmailAuthorized() {
		return nil, ErrorUnnecessary
	}
	str, err := Encrypt(strconv.Itoa(user.uuid.ToInt()) + "-" + user.tempid.ToString())
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	if !user.email.count.canSendNow() {
		return nil, ErrorSendingEmailLimit
	}
	path_all := WORK_SPACE + path + "/" + str
	email, err := user.email.SendEmail("メール認証",
		"以下のリンクを踏んでメールアドレスを認証してください。\n"+
			path_all+"\n"+
			"このメールに心当たりがない方はお問い合わせください。\n",
	)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	new_user := NewUser(user.uuid, user.tempid, user.name, email, user.group)
	err = SqlHandle.UpdateUser(new_user)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	return new_user, nil
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
	uuid, err := NewUuidInt(uuid_int)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	tempid, err := NewTempIdString(str_arr[1])
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
		return nil, ErrorUnnecessary
	}
	new_user := NewUser(user.uuid, user.tempid, user.name, user.email.Authorize(), user.group)
	err = SqlHandle.UpdateUser(new_user)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	return new_user, nil
}

func InitValTestLocal() {
	file, err := os.Open("C:/Users/stran/Documents/GitHub/drawwwingame/secret/val.txt")
	if err != nil {
		panic(err)
	}
	temp := make([]byte, 1000)
	n, err := file.Read(temp)
	if err != nil {
		panic(err)
	}
	data := string(temp[:n])
	data_arr := strings.Split(data, "\n")
	m := make(map[string]string)
	for _, v := range data_arr {
		if len(v) <= 2 {
			continue
		}
		arr := strings.Split(v, ":")
		a := strings.TrimSpace(arr[0])
		b := strings.TrimSpace(arr[1])
		m[a] = b
	}
	GMAIL_ADDRESS = m["GMAIL_ADDRESS"]
	GMAIL_PASSWORD = m["GMAIL_PASSWORD"]
	TEST_GMAIL_ADDRESS = m["TEST_GMAIL_ADDRESS"]
}

func Init() error {
	rand.Seed(time.Now().UnixNano())
	if SqlHandle == nil {
		// test version
		InitValTestLocal()
		return nil
	}
	err := SqlHandle.Init()
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	return nil
}

func Close() {
	close(Goroutine_cancel)
}
