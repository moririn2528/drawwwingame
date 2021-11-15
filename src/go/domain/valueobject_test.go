package domain

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	DEBUG_MODE = false
	InitValTestLocal()
	defer Close()
	status := m.Run()
	os.Exit(status)
}

func check(t *testing.T, arg, actual, expect interface{}) {
	if expect != actual {
		t.Errorf("\ninput: %v\noutput: %v\nexpect %v", arg, actual, expect)
	}
}

func TestAlphanumString(t *testing.T) {
	patterns := map[string]error{
		"aaaaa":                                nil,
		"1234567890poiuytrewqasdfghjklmnbvcxz": nil,
		"2345QWERTYUIOPLKJHGFDSAZXCVBNM":       nil,
		"!sdfgh":                               ErrorString,
		"1,2.":                                 ErrorString,
	}
	for key, val := range patterns {
		str, err := NewAlphanumString(key)
		check(t, key, err, val)
		if err == nil {
			check(t, key, key, str.ToString())
		}
	}
	lens := []int{1, 10, 123}
	for _, l := range lens {
		a := NewAlphanumStringRandom(l)
		_, err := NewAlphanumString(a.ToString())
		check(t, l, err, nil)
	}
}

func TestUuidInt(t *testing.T) {
	patterns := map[int]error{
		12345:   nil,
		0:       nil,
		860503:  nil,
		2528:    nil,
		-1:      ErrorInt,
		-123456: ErrorInt,
	}
	for key, val := range patterns {
		uuid, err := NewUuidInt(key)
		check(t, key, err, val)
		if err == nil {
			check(t, key, key, uuid.ToInt())
		}
	}
	lens := []int{1, 10, 123}
	for _, l := range lens {
		a := NewUuidIntRandom()
		_, err := NewUuidInt(a.ToInt())
		check(t, l, err, nil)
	}
}

func TestDatetime(t *testing.T) {
	loc, _ := time.LoadLocation("Local")
	times := []time.Time{
		time.Date(2003, 12, 11, 12, 12, 12, 111, loc),
		time.Date(2004, 1, 11, 12, 12, 12, 111, loc),
		time.Date(2004, 1, 13, 12, 12, 12, 111, loc),
		time.Date(2004, 12, 11, 12, 12, 12, 111, loc),
		time.Date(2007, 3, 12, 9, 12, 12, 111, loc),
	}
	same_day := []time.Time{
		time.Date(2004, 1, 13, 12, 12, 12, 111, loc),
		time.Date(2004, 1, 13, 12, 13, 12, 111, loc),
		time.Date(2004, 1, 13, 12, 12, 12, 2345, loc),
		time.Date(2004, 1, 13, 12, 12, 23, 111, loc),
	}
	tim := time.Date(2007, 3, 12, 9, 12, 12, 234, loc)
	for i, a := range times {
		for j, b := range times {
			check(t, a.String()+" "+b.String(),
				NewDatetime(a).EqualDay(NewDatetime(b)),
				i == j)
		}
	}
	for _, a := range same_day {
		for _, b := range same_day {
			check(t, a.String()+" "+b.String(),
				NewDatetime(a).EqualDay(NewDatetime(b)),
				true)
		}
	}
	for _, a := range times {
		for _, b := range times {
			check(t, a.String()+" "+b.String(),
				NewDatetime(a).Before(NewDatetime(b)),
				a.Before(b))
			check(t, a.String()+" "+b.String(),
				NewDatetime(a).After(NewDatetime(b)),
				a.After(b))
		}
		check(t, a.String(), NewDatetime(a).AddHour(5).Time, a.Add(time.Hour*5))
		check(t, a.String(), NewDatetime(a).AddHour(1000).Time, a.Add(time.Hour*1000))
		check(t, a.String(), NewDatetime(a).AddHour(-5).Time, a.Add(-time.Hour*5))
		check(t, a.String(), NewDatetime(a).AddHour(-1000).Time, a.Add(-time.Hour*1000))
	}
	check(t, tim.String(),
		NewDatetime(tim).Before(NewDatetime(tim)),
		tim.Before(tim))
	check(t, tim.String(),
		NewDatetime(tim).After(NewDatetime(tim)),
		tim.After(tim))
}

func TestTempidString(t *testing.T) {
	patterns := map[string]error{
		"1234567890POIUYTREWQ": nil,
		"ASDFGHJKLMNBVCXZqwer": nil,
		"tyuiopasdfghjklzxcvb": nil,
		"bnmsdfgoh3434yinSDF2": nil,
		"#Dbik2grvfbndthgfdsa": ErrorString,
		"dffben3":              ErrorString,
	}
	for key, val := range patterns {
		tempid, err := NewTempIdString(key)
		check(t, key, err, val)
		if err == nil {
			check(t, key, key, tempid.ToString())
		}
	}
	for i := 0; i < 10; i++ {
		tempid := NewTempIdStringRandom()
		_, err := NewTempIdString(tempid.ToString())
		check(t, "NewTempIdStringRandom", err, nil)
	}

	hours := map[int]error{
		1: nil, 123: nil, 1000000: nil,
		-1: ErrorExpired, -1234: ErrorExpired, -1000000: ErrorExpired}
	for key, val := range patterns {
		if val != nil {
			continue
		}
		for h, expect_error := range hours {
			tm := time.Now().Add(time.Hour * time.Duration(h))
			tempid, err := NewTempIdValidString(key, tm)
			check(t, "NewTempIdValidString error", err, expect_error)
			if err != nil {
				break
			}
			check(t, "NewTempIdValidString string", tempid.ToString(), key)
			check(t, "NewTempIdValidString expired time", tempid.GetExpiredAt(), tm)
		}
	}
	for i := 0; i < 10; i++ {
		tempid := NewTempIdValidStringRandom()
		_, err := NewTempIdValidString(tempid.ToString(), tempid.GetExpiredAt())
		check(t, "NewTempIdValidStringRandom", err, nil)
	}
}

func TestPasswordString(t *testing.T) {
	patterns := map[string]error{
		"aaaaa":                                ErrorString,
		"aaaaaa":                               nil,
		strings.Repeat("a", 100):               nil,
		strings.Repeat("a", 101):               ErrorString,
		"asdfghjklqwertyuiopzxcvbnm":           nil,
		"QWERTYUIOPASDFGHJKLMNBVCXZ0987654321": nil,
		";--":                                  ErrorString,
		";":                                    ErrorString,
		"1234sasdfdbf244t":                     nil,
		"1234567654323":                        nil,
		"ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb": nil,
	}
	for key, val := range patterns {
		pass, err := NewPasswordString(key)
		check(t, key, err, val)
		if err != nil {
			continue
		}
		for key2, _ := range patterns {
			pass2, err := NewPasswordString(key2)
			if err != nil {
				continue
			}
			check(t, key+", "+key2, pass.ToString() == pass2.ToString(), key == key2)
		}
		check(t, key, pass.ToString() == key, false)
	}
}

func TestIsDomainString(t *testing.T) {
	patterns := map[string]bool{
		"":                       false,
		strings.Repeat("a", 253): true,
		strings.Repeat("a", 254): false,
		"[12.3.4.4]":             true,
		"[255.0.0.1]":            true,
		"[123.256.0.0]":          false,
		"[12.23.24.35":           false,
		"1.1.1.1]":               false,
		"test":                   true,
		"qwertyuiopasdfghjklzxcvbnm-QWERTYUIOPLKJHGFDSAZXCVBNM.1234567890-a": true,
		"-a.a": false,
		"a--a": true,
		"a..a": false,
	}
	for key, val := range patterns {
		check(t, key, isDomainString(key), val)
	}
}

func TestIsEmailLocalString(t *testing.T) {
	patterns := map[string]bool{
		"":                      false,
		strings.Repeat("a", 64): true,
		strings.Repeat("a", 65): false,
		".sdfgg":                false,
		"sdfgg.":                false,
		"sdf..sdfghn":           false,
		"!#$%&'*+-/=?^_`{|}~":   true,
		"asdfghjklqwertyuiopmnbvcxz1234567890QWERTYUIOPLKJHGFDSAZXCVBNM": true,
		"\"(),:;<>@[]..\"":        true,
		"\"\\\\a\\ a\\\ta\\\"a\"": true,
		"asdf\\\\asdf":            false,
		"sdf(sfdg":                false,
		"\"\"\"":                  false,
	}
	for key, val := range patterns {
		check(t, key, isEmailLocalString(key), val)
	}
}

func TestNewEmailString(t *testing.T) {
	patterns := map[string]error{
		strings.Repeat("a", 50) + "@" + strings.Repeat("a", 203): nil,
		strings.Repeat("a", 50) + "@" + strings.Repeat("a", 204): ErrorString,
		"test1@gmail.com":    nil,
		"test@test@test.com": ErrorString,
		"testest.com":        ErrorString,
	}
	for key, val := range patterns {
		a, err := NewEmailString(key)
		check(t, key, err, val)
		if err == nil {
			check(t, key, key, a.ToString())
		}
	}
}

func TestSendingEmailCount(t *testing.T) {
	loc, _ := time.LoadLocation("Local")
	patterns := map[[2]interface{}]error{
		{email_limit_send_par_day, time.Date(2003, 1, 1, 1, 1, 1, 1, loc)}:     nil,
		{email_limit_send_par_day + 2, time.Date(2003, 1, 1, 1, 1, 1, 1, loc)}: ErrorArg,
		{-1, time.Date(2003, 1, 1, 1, 1, 1, 1, loc)}:                           ErrorArg,
		{0, time.Now().Add(time.Hour * 100)}:                                   ErrorArg,
		{0, time.Now()}:                                                        nil,
	}
	for arg, val := range patterns {
		em, err := NewSendingEmailCountnSet(arg[0].(int), arg[1].(time.Time))
		arg_str := strconv.Itoa(arg[0].(int)) + "," + arg[1].(time.Time).String()
		check(t, arg_str,
			err, val)
		if err != nil {
			continue
		}
		check(t, arg_str+", getcount", em.GetCount(), arg[0].(int))
		check(t, arg_str+", getlastday", em.GetLastDay(), arg[1].(time.Time))
		for i := 0; i <= email_limit_send_par_day; i++ {
			check(t, arg_str+", email send: "+strconv.Itoa(i), em.canSendNow(), i < email_limit_send_par_day)
			if i == email_limit_send_par_day {
				break
			}
			em, err = em.IncrementCountNow()
			check(t, arg_str+", increment email send: "+strconv.Itoa(i), err, nil)
		}
	}
}

func TestBoolean(t *testing.T) {
	patterns := map[string]bool{
		string([]byte{0}): false,
		"":                false,
		"0":               true,
		"000":             true,
	}
	for key, val := range patterns {
		check(t, key, NewBooleanByByte([]byte(key)).ToBool(), val)
	}
}

func TestEmailObject(t *testing.T) {
	var err error
	DEBUG_MODE = true
	e, _ := NewEmailString(TEST_GMAIL_ADDRESS)
	email := NewEmailObject(e)
	email, err = email.SendEmail("test", "test")
	check(t, "test email send", err, nil)
	email = email.Authorize()
	check(t, "is authorized", email.IsAuthorized(), true)
}

func TestNameString(t *testing.T) {
	patterns := map[string]error{
		"test":        nil,
		"moririn2528": nil,
		"te24わっふぁ":    nil,
		"te24わっふぁｗ":   nil,
		"te24わっふぁ　ｗ":  ErrorString,
		"te24わっふぁ ｗ":  ErrorString,
		"   ":         ErrorString,
		"お手紙クレジット":    nil,
		";--":         ErrorString,
	}
	for key, val := range patterns {
		name, err := NewNameString(key)
		check(t, key, err, val)
		if err != nil {
			continue
		}
		check(t, key, name.ToString(), key)
	}
}
