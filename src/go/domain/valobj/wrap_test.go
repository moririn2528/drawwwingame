package valobj

import (
	"drawwwingame/domain/internal"
	"strconv"
	"testing"
)

func TestAlphanumString(t *testing.T) {
	patterns := map[string]error{
		"aaaaa":                                nil,
		"1234567890poiuytrewqasdfghjklmnbvcxz": nil,
		"2345QWERTYUIOPLKJHGFDSAZXCVBNM":       nil,
		"!sdfgh":                               internal.ErrorArg,
		"1,2.":                                 internal.ErrorArg,
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

func TestCountingInt(t *testing.T) {
	patterns := []int{-1, 0, 1, 5, 10, 10000}
	for _, count_max := range patterns {
		count := NewCountingInt(count_max)
		str := strconv.Itoa(count_max)
		if count_max <= 0 {
			check(t, str+",getmax", count.GetMax(), max_int)
		} else {
			check(t, str+",getmax", count.GetMax(), count_max)
		}
		for i := 0; i < 15; i++ {
			check(t, str+",toint", count.ToInt(), i)
			ok := count.Increment()
			check(t, str+",increment", ok, count_max <= 0 || i < count_max)
			if !ok {
				break
			}
		}
	}
}
