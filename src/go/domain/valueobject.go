package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"net/smtp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	alphanum_char            = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	tempid_size              = 20
	tempid_valid_hour        = 24
	email_limit_send_par_day = 10
)

type AlphanumString struct {
	str string
}

func NewAlphanumString(str string) (*AlphanumString, error) {
	for _, r := range str {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			Log(ErrorString)
			return nil, ErrorString
		}
	}
	s := new(AlphanumString)
	s.str = str
	return s, nil
}

func NewAlphanumStringRandom(length int) *AlphanumString {
	s := new(AlphanumString)
	runes := make([]byte, length)
	for i := 0; i < length; i++ {
		j := rand.Intn(len(alphanum_char))
		runes[i] = alphanum_char[j]
	}
	s.str = string(runes)
	return s
}

func (alp *AlphanumString) ToString() string {
	return alp.str
}

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
		Log(ErrorInt)
		return nil, ErrorInt
	}
	s := new(UuidInt)
	s.num = num
	return s, nil
}

func NewUuidIntByString(str string) (*UuidInt, error) {
	num, err := strconv.Atoi(str)
	if err != nil {
		return nil, ErrorArg
	}
	return NewUuidInt(num)
}

func (id *UuidInt) ToInt() int {
	return id.num
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
		Log(ErrorString)
		return nil, ErrorString
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
		Log(ErrorExpired)
		return nil, ErrorExpired
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
		Log(ErrorString)
		return nil, ErrorString
	}
	for _, r := range str {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			Log(ErrorString)
			return nil, ErrorString
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
		Log(ErrorString)
		return nil, ErrorString
	}
	index := strings.LastIndex(str, "@")
	if len(str) <= index+1 {
		Log(ErrorString)
		return nil, ErrorString
	}
	local := str[:index]
	domain := str[index+1:]

	if !isEmailLocalString(local) || !isDomainString(domain) {
		Log(ErrorString)
		return nil, ErrorString
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
		LogStringf("email count error")
		return nil, ErrorArg
	}
	if lastday.After(time.Now()) {
		LogStringf("sending email last day error")
		return nil, ErrorArg
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
		return nil, NewError("sending count limit")
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

type Boolean struct {
	b bool
}

func NewBoolean(b bool) *Boolean {
	s := new(Boolean)
	s.b = b
	return s
}

func NewBooleanByByte(b []byte) *Boolean {
	if len(b) == 0 {
		return NewBoolean(false)
	}
	for _, bit := range b {
		if bit != 0 {
			return NewBoolean(true)
		}
	}
	return NewBoolean(false)
}

func (b *Boolean) ToBool() bool {
	return b.b
}
func (b *Boolean) ToInt() int {
	if b.b {
		return 1
	}
	return 0
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
		Log(err)
		return nil, ErrorArg
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

func (e *EmailObject) SendEmail(subject, message string) (*EmailObject, error) {
	auth := smtp.PlainAuth(
		"",
		GMAIL_ADDRESS,
		GMAIL_PASSWORD,
		"smtp.gmail.com",
	)
	to := e.EmailString()
	if !e.count.canSendNow() {
		return nil, ErrorSendingEmailLimit
	}
	count, err := e.count.IncrementCountNow()
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	err = smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		GMAIL_ADDRESS,
		[]string{to},
		[]byte("To: "+to+"\r\n"+
			"Subject:"+subject+"\r\n"+
			"\r\n"+
			message),
	)
	if err != nil {
		Log(err)
		return nil, ErrorInternal
	}
	return NewEmailObjectAll(e.email, count, e.authorized), nil
}

func (e *EmailObject) Authorize() *EmailObject {
	return NewEmailObjectAll(e.email, e.count, NewBoolean(true))
}
func (e *EmailObject) IsAuthorized() bool {
	return e.authorized.ToBool()
}

type NameString struct {
	str string
}

func NewNameString(str string) (*NameString, error) {
	if len(str) > 30 {
		Log(ErrorString)
		return nil, ErrorString
	}
	for _, r := range str {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			continue
		}
		if unicode.In(r, unicode.Hiragana) || unicode.In(r, unicode.Katakana) || unicode.In(r, unicode.Han) {
			continue
		}
		Log(ErrorString)
		return nil, ErrorString
	}
	s := new(NameString)
	s.str = str
	return s, nil
}

func (pas *NameString) ToString() string {
	return pas.str
}

type GroupIdInt struct {
	id int // -1: not set
}

func NewGroupIdInt(id int) (*GroupIdInt, error) {
	if id < -1 || GROUP_NUMBER <= id {
		return nil, ErrorArg
	}
	s := new(GroupIdInt)
	s.id = id
	return s, nil
}
func NewGroupIdIntNoSet() *GroupIdInt {
	s := new(GroupIdInt)
	s.id = -1
	return s
}

func NewGroupIdIntByString(id_str string) (*GroupIdInt, error) {
	id, err := strconv.Atoi(id_str)
	if err != nil {
		return nil, ErrorArg
	}
	return NewGroupIdInt(id)
}

func (g *GroupIdInt) ToInt() int {
	return g.id
}
