package valobj

import (
	"drawwwingame/domain/internal"
	"testing"
	"time"
)

func TestGameTimer(t *testing.T) {
	patterns := map[string]error{
		"-1s":   internal.ErrorArg,
		"0s":    internal.ErrorArg,
		"100ms": nil,
		"0.5s":  nil,
	}
	for key, val := range patterns {
		dur, err := time.ParseDuration(key)
		check(t, "ParseDuration error: "+key, err, nil)
		if err != nil {
			continue
		}
		tm, err := NewGameTimer(dur)
		check(t, "NewGameTimer error: "+key, err, val)
		if err != nil {
			continue
		}
		check(t, "InProgress 1: "+key, tm.InProgress(), false)
		time.Sleep(dur)
		check(t, "InProgress 2: "+key, tm.InProgress(), false)
		tm.Start()
		check(t, "InProgress 3: "+key, tm.InProgress(), true)
		time.Sleep(dur + time.Millisecond)
		check(t, "InProgress 4: "+key, tm.InProgress(), false)
	}
}
