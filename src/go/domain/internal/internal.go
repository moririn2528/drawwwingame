package internal

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
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
	DEBUG_MODE          = true
	GMAIL_PASSWORD      string
	GMAIL_ADDRESS       string
	TEST_GMAIL_ADDRESS0 string
	TEST_GMAIL_ADDRESS1 string
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
	ErrorInternal           = errors.New("internal error")
	ErrorArg                = errors.New("argument error")
	ErrorConnection         = errors.New("connection error")
	ErrorNoMatter           = errors.New("error no matter")
)

var (
	Goroutine_cancel = make(chan struct{}, 1)
)

func PrintLog(str string, dep, max_repeat int) {
	if max_repeat <= 0 {
		max_repeat = 10
	}
	if !DEBUG_MODE {
		return
	}
	s := []string{}
	s = append(s, fmt.Sprintf("\nerror:%v\nnests", str))

	for i := 0; i < max_repeat; i++ {
		pc, file, line, ok := runtime.Caller(dep + i)
		if !ok {
			break
		}
		f := runtime.FuncForPC(pc)
		s = append(s, fmt.Sprintf("file:%s:%d\nfunc:%s", file, line, f.Name()))
	}
	log.Println(strings.Join(s, "\n"))
}

func Log(err error) {
	PrintLog(err.Error(), 2, -1)
}
func LogStringf(str string, arg ...interface{}) {
	PrintLog(fmt.Sprintf(str, arg...), 2, -1)
}

func setVar(arg *string, name string, m *map[string]string) {
	var ok bool
	name = strings.ToLower(name)
	*arg, ok = (*m)[name]
	if ok {
		return
	}
	name = strings.ToUpper(name)
	*arg = (*m)[name]
}

func SetInternalVar(m map[string]string) {
	setVar(&TEST_GMAIL_ADDRESS0, "TEST_GMAIL_ADDRESS0", &m)
	setVar(&TEST_GMAIL_ADDRESS1, "TEST_GMAIL_ADDRESS1", &m)
	setVar(&GMAIL_ADDRESS, "GMAIL_ADDRESS", &m)
	setVar(&GMAIL_PASSWORD, "GMAIL_PASSWORD", &m)
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
	SetInternalVar(m)
}

func LogAll(err ...error) error {
	var out_err error
	for i, e := range err {
		if e != nil {
			log.Println("error: arg", i)
			Log(e)
			out_err = ErrorInternal
		}
	}
	return out_err
}

func MapToString(m map[string]interface{}) string {
	lines := []string{}
	for key, val := range m {
		val_str := ""
		switch val := val.(type) {
		case string:
			val_str = val
		case int:
			val_str = strconv.Itoa(val)
		case int64:
			val_str = strconv.FormatInt(val, 10)
		case bool:
			val_str = strconv.FormatBool(val)
		}
		lines = append(lines, key+": "+val_str)
	}
	return "[" + strings.Join(lines, "\n") + "]"
}
