package valobj

import (
	"drawwwingame/domain/internal"
	"strconv"
	"strings"
	"testing"
)

func TestMessageTypeString(t *testing.T) {
	patterns := map[string]error{
		"info":        nil,
		"lines":       nil,
		"mark:writer": nil,
		"mark:answer": nil,
		"text:writer": nil,
		"text:answer": nil,
		"test":        internal.ErrorArg,
		"info:writer": internal.ErrorArg,
	}
	patterns_int := map[string]int{
		"info":        1 << 1,
		"lines":       0,
		"mark:writer": 1 + 1<<3,
		"mark:answer": 1 + 1<<2,
		"text:writer": 1 << 3,
		"text:answer": 1 << 2,
	}
	for key, val := range patterns {
		mes, err := NewMessageTypeString(key)
		check(t, "NewMessageTypeString:"+key, err, val)
		if err != nil {
			continue
		}
		check(t, "NewMessageTypeString, string:"+key,
			mes.ToString(), key)
	}
	for key, val := range patterns_int {
		mes, err := NewMessageTypeString(key)
		if err != nil {
			panic("assert false")
		}
		check(t, "IsMark:"+key, mes.IsMark(), val&1 > 0)
		check(t, "IsInfo:"+key, mes.IsInfo(), val&(1<<1) > 0)
		check(t, "IsAnswer:"+key, mes.IsAnswer(), val&(1<<2) > 0)
		check(t, "IsWriter:"+key, mes.IsWriter(), val&(1<<3) > 0)
	}
	check(t, "NewInfoMessageTypeString", NewInfoMessageTypeString().IsInfo(), true)
}

func TestMessageString(t *testing.T) {
	patterns := map[string]error{
		"testA":                   nil,
		strings.Repeat("a", 1000): nil,
		strings.Repeat("a", 1001): internal.ErrorArg,
		"test;--":                 internal.ErrorArg,
		"のんのんノンノン佃煮食べるasafou12r53":             nil,
		"qwertyuioplkjhgfdsazxcvbnm1234567890": nil,
		"QWERTYUIOPLKJHGFDSAZXCVBNM":           nil,
	}
	check(t, "NewMessageStringNil", NewMessageStringNil().ToString(), "")
	for key, val := range patterns {
		mes, err := NewMessageString(key)
		check(t, "NewMessageString:"+key, err, val)
		if err != nil {
			continue
		}
		check(t, "NewMessageString.ToString:"+key, mes.ToString(), key)
	}
}

func TestMessageMarkString(t *testing.T) {
	patterns := map[string]error{
		"AA": internal.ErrorArg,
		"A":  nil,
		"B":  nil,
		"C":  nil,
		"D":  internal.ErrorArg,
		"a":  internal.ErrorArg,
	}
	patterns_int := map[string]int{
		"A": 0,
		"B": 1,
		"C": 2,
	}
	for key, val := range patterns {
		mes, err := NewMessageMarkString(key)
		check(t, "NewMessageString:"+key, err, val)
		if err != nil {
			continue
		}
		check(t, "NewMessageString.ToString:"+key, mes.ToString(), key)
	}
	for key, val := range patterns_int {
		mes, err := NewMessageMarkString(key)
		if err != nil {
			panic("assert false")
		}
		check(t, "NewMessageString.ToString:"+key, mes.ToInt(), val)
	}
}

func TestMessageId(t *testing.T) {
	const N = 10
	patterns := map[int]error{
		-1:    internal.ErrorArg,
		0:     nil,
		N - 1: nil,
		N:     internal.ErrorArg,
	}
	for i := 0; i < N; i++ {
		id := NewMessageId()
		check(t, "NewMessageId:"+strconv.Itoa(i), id.ToInt(), i)
	}
	for key, val := range patterns {
		mes1, err := NewMessageIdByInt(key)
		check(t, "NewMessageIdByInt:"+strconv.Itoa(key), err, val)
		mes2, err := NewMessageIdByString(strconv.Itoa(key))
		check(t, "NewMessageIdByInt:"+strconv.Itoa(key), err, val)
		if err != nil {
			continue
		}
		check(t, "NewMessageIdByInt.ToInt1:"+strconv.Itoa(key), mes1.ToInt(), key)
		check(t, "NewMessageIdByInt.ToInt2:"+strconv.Itoa(key), mes2.ToInt(), key)
	}
}
