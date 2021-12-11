package valobj

import (
	"crypto/sha256"
	"drawwwingame/domain/internal"
	"encoding/hex"
	"math/rand"
	"net/smtp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	alphanum_char            = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	tempid_size              = 20
	tempid_valid_hour        = 24
	email_limit_send_par_day = 10
	max_int                  = int(^uint(0) >> 1)
)

type UuidInt struct {
	num int
}

func NewUuidIntRandom() *UuidInt {
	s := new(UuidInt)
	s.num = rand.Int()
	return s
}

func NewUuidInt(num int) (*UuidInt, error) {
	if num < 0 {
		internal.LogStringf("minus arg")
		return nil, internal.ErrorArg
	}
	s := new(UuidInt)
	s.num = num
	return s, nil
}

func NewUuidIntByString(str string) (*UuidInt, error) {
	num, err := strconv.Atoi(str)
	if err != nil {
		return nil, internal.ErrorArg
	}
	return NewUuidInt(num)
}

func (id *UuidInt) ToInt() int {
	return id.num
}
func (id *UuidInt) Equal(id2 *UuidInt) bool {
	return id.num == id2.num
}

type Datetime struct {
	time.Time
}

func NewDatetimeNow() *Datetime {
	return &Datetime{time.Now()}
}

func NewDatetime(tim time.Time) *Datetime {
	return &Datetime{tim}
}

func (date1 *Datetime) EqualDay(date2 *Datetime) bool {
	return date1.Year() == date2.Year() && date1.Month() == date2.Month() &&
		date1.Day() == date2.Day()
}
func (date *Datetime) IsToday() bool {
	return date.EqualDay(NewDatetimeNow())
}

func (date *Datetime) AddHour(hour int) *Datetime {
	return &Datetime{date.Add(time.Hour * time.Duration(hour))}
}
func (date *Datetime) Before(date2 *Datetime) bool {
	return date.Time.Before(date2.Time)
}
func (date *Datetime) After(date2 *Datetime) bool {
	return date.Time.After(date2.Time)
}
func (date *Datetime) Equal(date2 *Datetime) bool {
	return date.Time.Equal(date2.Time)
}

type NameString struct {
	str string
}

func NewNameString(str string) (*NameString, error) {
	if len(str) > 30 {
		internal.LogStringf("length too large")
		return nil, internal.ErrorArg
	}
	if !utf8.Valid([]byte(str)) {
		internal.LogStringf("utf8 not validate")
		return nil, internal.ErrorArg
	}
	for _, r := range str {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			continue
		}
		if unicode.In(r, unicode.Hiragana) || unicode.In(r, unicode.Katakana) || unicode.In(r, unicode.Han) {
			continue
		}
		internal.LogStringf("rune not match")
		return nil, internal.ErrorArg
	}
	s := new(NameString)
	s.str = str
	return s, nil
}

func NewNameStringNil() *NameString {
	return &NameString{
		str: "None",
	}
}

func (pas *NameString) ToString() string {
	return pas.str
}

type TempIdString struct {
	*AlphanumString
}
type TempIdValidString struct {
	*TempIdString
	expiredAt *Datetime
}

func NewTempIdStringRandom() *TempIdString {
	return &TempIdString{
		NewAlphanumStringRandom(tempid_size),
	}
}

func NewTempIdString(str string) (*TempIdString, error) {
	if len(str) != tempid_size {
		internal.LogStringf("error length")
		return nil, internal.ErrorArg
	}
	s, err := NewAlphanumString(str)
	if err != nil {
		return nil, err
	}
	return &TempIdString{s}, nil
}

func NewTempIdValidStringRandom() *TempIdValidString {
	return &TempIdValidString{
		NewTempIdStringRandom(),
		NewDatetimeNow().AddHour(tempid_valid_hour),
	}
}

func NewTempIdValidString(str string, tim time.Time) (*TempIdValidString, error) {
	if tim.Before(time.Now()) {
		internal.LogStringf("error time is expired")
		return nil, internal.ErrorExpired
	}
	tempid, err := NewTempIdString(str)
	if err != nil {
		return nil, err
	}
	return &TempIdValidString{
		tempid,
		NewDatetime(tim),
	}, nil
}

func (tempid *TempIdValidString) GetExpiredAt() time.Time {
	return tempid.expiredAt.Time
}

type PasswordString struct {
	str string
}

func createHashPassword(password string) string {
	str := "wcRysg" + password + "wj6cXyZhiCViRqw3UAnQ"
	hash_hex := sha256.Sum256([]byte(str))
	return hex.EncodeToString(hash_hex[:])
}

func NewPasswordString(str string) (*PasswordString, error) {
	if len(str) < 6 || len(str) > 100 {
		internal.LogStringf("length error")
		return nil, internal.ErrorArg
	}
	for _, r := range str {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			internal.LogStringf("error rune contain")
			return nil, internal.ErrorArg
		}
	}
	s := new(PasswordString)
	s.str = createHashPassword(str)
	return s, nil
}

func (pas *PasswordString) ToString() string {
	return pas.str
}

type EmailString struct {
	str string
}

func isDomainString(str string) bool {
	if len(str) == 0 || len(str) > 253 {
		return false
	}
	if str[0] == '[' {
		if len(str) <= 2 || str[len(str)-1] != ']' {
			return false
		}
		str_arr := strings.Split(str[1:len(str)-1], ".")
		if len(str_arr) != 4 {
			return false
		}
		for _, s := range str_arr {
			for _, r := range s {
				if !unicode.IsNumber(r) {
					return false
				}
			}
			num, err := strconv.Atoi(s)
			if err != nil || num > 255 {
				return false
			}
		}
		return true
	}
	str_arr := strings.Split(str, ".")
	for _, s := range str_arr {
		if len(s) == 0 || s[0] == '-' {
			return false
		}
		for _, r := range s {
			if r != '-' && !unicode.IsLetter(r) && !unicode.IsNumber(r) {
				return false
			}
		}
	}
	return true
}

func isEmailLocalString(str string) bool {
	quote_flag := false
	if len(str) == 0 || len(str) > 64 || str[0] == '.' || str[len(str)-1] == '.' {
		return false
	}
	for i, r := range str {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || strings.ContainsRune("!#$%&'*+-/=?^_`{|}~", r) {
			continue
		}
		if r == '"' && (i == 0 || str[i-1] != '\\') {
			quote_flag = !quote_flag
			continue
		}
		if str[i] == '.' {
			if i > 0 && str[i-1] == '.' && !quote_flag {
				return false
			}
			continue
		}
		if strings.ContainsRune("(),:;<>@[]", r) {
			if quote_flag {
				continue
			} else {
				return false
			}
		}
		if r == '\\' {
			if !quote_flag {
				return false
			}
			continue
		}
		if strings.ContainsRune("\\\" \t", r) {
			if !quote_flag {
				return false
			}
			if i > 0 && str[i-1] == '\\' {
				continue
			} else {
				return false
			}
		}
		return false
	}
	return !quote_flag
}

func NewEmailString(str string) (*EmailString, error) {
	if len(str) > 254 || strings.Count(str, "@") == 0 {
		internal.LogStringf("error length")
		return nil, internal.ErrorArg
	}
	index := strings.LastIndex(str, "@")
	if len(str) <= index+1 {
		internal.LogStringf("not contain @")
		return nil, internal.ErrorArg
	}
	local := str[:index]
	domain := str[index+1:]

	if !isEmailLocalString(local) || !isDomainString(domain) {
		internal.LogStringf("error email")
		return nil, internal.ErrorArg
	}
	s := new(EmailString)
	s.str = str
	return s, nil
}

func (ema *EmailString) ToString() string {
	return ema.str
}

type SendingEmailCount struct {
	count   int
	lastDay *Datetime
}

func NewSendingEmailCount() *SendingEmailCount {
	s := new(SendingEmailCount)
	s.count = 0
	s.lastDay = NewDatetimeNow()
	return s
}

func NewSendingEmailCountnSet(count int, lastday time.Time) (*SendingEmailCount, error) {
	s := new(SendingEmailCount)
	if count < 0 || count > email_limit_send_par_day {
		internal.LogStringf("email count error")
		return nil, internal.ErrorArg
	}
	if lastday.After(time.Now()) {
		internal.LogStringf("sending email last day error")
		return nil, internal.ErrorArg
	}
	s.count = count
	s.lastDay = NewDatetime(lastday)
	return s, nil
}

func (send *SendingEmailCount) canSendNow() bool {
	if send.lastDay.IsToday() {
		return send.count < email_limit_send_par_day
	}
	return true
}

func (send *SendingEmailCount) IncrementCountNow() (*SendingEmailCount, error) {
	if !send.canSendNow() {
		internal.LogStringf("sending count limit")
		return nil, internal.ErrorSendingEmailLimit
	}
	if send.lastDay.IsToday() {
		return &SendingEmailCount{
			count: send.count + 1, lastDay: NewDatetimeNow(),
		}, nil
	} else {
		return &SendingEmailCount{
			count: 1, lastDay: NewDatetimeNow(),
		}, nil
	}
}

func (send *SendingEmailCount) GetCount() int {
	return send.count
}
func (send *SendingEmailCount) GetLastDay() time.Time {
	return send.lastDay.Time
}

type EmailObject struct {
	email      *EmailString
	count      *SendingEmailCount
	authorized *Boolean
}

func NewEmailObject(email *EmailString) *EmailObject {
	s := new(EmailObject)
	s.email = email
	s.count = NewSendingEmailCount()
	s.authorized = NewBoolean(false)
	return s
}
func NewEmailObjectAll(email *EmailString, count *SendingEmailCount, auth *Boolean) *EmailObject {
	s := new(EmailObject)
	s.email = email
	s.count = count
	s.authorized = auth
	return s
}

func NewEmailObjectnSet(email *EmailString, count int, day time.Time, auth *Boolean) (*EmailObject, error) {
	s := new(EmailObject)
	var err error
	s.email = email
	s.count, err = NewSendingEmailCountnSet(count, day)
	if err != nil {
		internal.Log(err)
		return nil, internal.ErrorArg
	}
	s.authorized = auth
	return s, nil
}

func (e *EmailObject) EmailString() string {
	return e.email.ToString()
}
func (e *EmailObject) GetCount() int {
	return e.count.GetCount()
}
func (e *EmailObject) GetLastDay() time.Time {
	return e.count.GetLastDay()
}

func (e *EmailObject) CanSendNow() bool {
	return e.count.canSendNow()
}

func (e *EmailObject) SendEmail(subject, message string) (*EmailObject, error) {
	auth := smtp.PlainAuth(
		"",
		internal.GMAIL_ADDRESS,
		internal.GMAIL_PASSWORD,
		"smtp.gmail.com",
	)
	to := e.EmailString()
	if !e.count.canSendNow() {
		return nil, internal.ErrorSendingEmailLimit
	}
	count, err := e.count.IncrementCountNow()
	if err != nil {
		internal.Log(err)
		return nil, internal.ErrorInternal
	}
	err = smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		internal.GMAIL_ADDRESS,
		[]string{to},
		[]byte("To: "+to+"\r\n"+
			"Subject:"+subject+"\r\n"+
			"\r\n"+
			message),
	)
	if err != nil {
		internal.Log(err)
		return nil, internal.ErrorInternal
	}
	return NewEmailObjectAll(e.email, count, e.authorized), nil
}

func (e *EmailObject) Authorize() *EmailObject {
	return NewEmailObjectAll(e.email, e.count, NewBoolean(true))
}
func (e *EmailObject) IsAuthorized() bool {
	return e.authorized.ToBool()
}

type GroupIdInt struct {
	id int
}

func NewGroupIdInt(id int) (*GroupIdInt, error) {
	if id < 0 || internal.GROUP_NUMBER <= id {
		return nil, internal.ErrorArg
	}
	s := new(GroupIdInt)
	s.id = id
	return s, nil
}

func NewGroupIdIntByString(id_str string) (*GroupIdInt, error) {
	id, err := strconv.Atoi(id_str)
	if err != nil {
		return nil, internal.ErrorArg
	}
	return NewGroupIdInt(id)
}

func (g *GroupIdInt) ToInt() int {
	return g.id
}
func (g *GroupIdInt) Equal(g2 *GroupIdInt) bool {
	return g.id == g2.id
}

type GroupRole struct {
	admin     *Boolean
	canAnswer *Boolean
	canWriter *Boolean
}

func NewGroupRole(admin, canAnswer, canWriter *Boolean) *GroupRole {
	s := new(GroupRole)
	s.admin = admin
	s.canAnswer = canAnswer
	s.canWriter = canWriter
	return s
}

func NewGroupRoleByRow(admin, canAnswer, canWriter bool) *GroupRole {
	return NewGroupRole(
		NewBoolean(admin),
		NewBoolean(canAnswer),
		NewBoolean(canWriter),
	)
}

func (role *GroupRole) IsAdmin() bool {
	return role.admin.ToBool()
}
func (role *GroupRole) CanWriter() bool {
	return role.canWriter.ToBool()
}
func (role *GroupRole) CanAnswer() bool {
	return role.canAnswer.ToBool()
}

func NewGroupRoleNoSet() *GroupRole {
	s := new(GroupRole)
	s.admin = NewBoolean(false)
	s.canAnswer = NewBoolean(true)
	s.canWriter = NewBoolean(true)
	return s
}

func (role *GroupRole) Update(admin, canAnswer, canWriter *Boolean) bool {
	updated := false
	if admin != nil && !role.admin.Equal(admin) {
		role.admin = admin
		updated = true
	}
	if canAnswer != nil && !role.canAnswer.Equal(canAnswer) {
		role.canAnswer = canAnswer
		updated = true
	}
	if canWriter != nil && !role.canWriter.Equal(canWriter) {
		role.canWriter = canWriter
		updated = true
	}
	return updated
}

func (role *GroupRole) Equal(r *GroupRole) bool {
	return role.canAnswer.Equal(r.canAnswer) &&
		role.canWriter.Equal(r.canWriter) &&
		role.admin.Equal(r.admin)
}
