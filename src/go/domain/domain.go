package domain

import (
	"drawwwingame/domain/internal"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

var ( ///error
	ErrorParse              = internal.ErrorParse
	ErrorNoUser             = internal.ErrorNoUser
	ErrorDuplicateUserName  = internal.ErrorDuplicateUserName
	ErrorDuplicateUserEmail = internal.ErrorDuplicateUserEmail
	ErrorDuplicateUserId    = internal.ErrorDuplicateUserId
	ErrorString             = internal.ErrorString
	ErrorInt                = internal.ErrorInt
	ErrorExpired            = internal.ErrorExpired
	ErrorSendingEmailLimit  = internal.ErrorSendingEmailLimit
	ErrorUnnecessary        = internal.ErrorUnnecessary
	ErrorInternal           = internal.ErrorInternal
	ErrorArg                = internal.ErrorArg
	ErrorDuplicate          = errors.New("duplicate error")
	ErrorConnection         = internal.ErrorConnection
	ErrorNoMatter           = internal.ErrorNoMatter
)

var (
	Goroutine_cancel = make(chan struct{}, 1)
)

func PrintLog(str string, dep int) {
	internal.PrintLog(str, dep+1, -1)
}
func Log(err error) {
	internal.PrintLog(err.Error(), 2, -1)
}
func Log1(err error) {
	internal.PrintLog(err.Error(), 2, 1)
}
func LogStringf(str string, arg ...interface{}) {
	internal.PrintLog(fmt.Sprintf(str, arg...), 2, -1)
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
	internal.PrintLog(fmt.Sprintf("error not contain, %v", err), 2, -1)
}

func GetDatabaseInfo() (string, string) {
	return internal.DATABASE, internal.DATABASE_INIT_INFO
}
func GetTestGmailAddress(ver int) string {
	if ver == 0 {
		return internal.TEST_GMAIL_ADDRESS0
	}
	if ver == 1 {
		return internal.TEST_GMAIL_ADDRESS1
	}
	return ""
}
func SetInternalVar(m map[string]string) {
	internal.SetInternalVar(m)
}

func Init() error {
	rand.Seed(time.Now().UnixNano())
	// if SqlHandle == nil {
	// 	// test version
	// 	InitValTestLocal()
	// 	return nil
	// }
	err := SqlHandle.Init()
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	err = initGame()
	if err != nil {
		Log(err)
		return ErrorInternal
	}
	return nil
}

func Close() {
	close(internal.Goroutine_cancel)
	close(Goroutine_cancel)
}
