package valobj

import (
	"drawwwingame/domain/internal"
	"math"
	"strconv"
	"time"
)

type GameTimer struct {
	dur           time.Duration
	end_time      time.Time
	interval_func func(time.Duration)
	end_func      func()
	turn_channel  chan (struct{})
}

func NewGameTimer(d time.Duration, interval_func func(time.Duration), end_func func()) (*GameTimer, error) {
	if int64(d) <= 0 {
		return nil, internal.ErrorArg
	}
	s := new(GameTimer)
	s.dur = d
	s.end_time = time.Now().Add(-time.Minute)
	s.interval_func = interval_func
	s.end_func = end_func
	s.turn_channel = make(chan struct{}, 1)
	return s, nil
}

func (tim *GameTimer) Minutes() int {
	return int(math.Floor(tim.dur.Minutes()))
}

func (tim *GameTimer) Start() {
	tim.end_time = time.Now().Add(tim.dur)
	tim.turn_channel = make(chan struct{}, 1)
	if tim.interval_func != nil {
		go func() {
			for tim.InProgress() {
				select {
				case <-tim.turn_channel:
					return
				default:
					tim.interval_func(tim.LestTime())
					time.Sleep(500 * time.Millisecond)
				}
			}
			select {
			case <-tim.turn_channel:
				return
			default:
				tim.end_func()
			}
		}()
	}
}

func (tim *GameTimer) End() {
	if !tim.InProgress() {
		return
	}
	close(tim.turn_channel)
	tim.end_time = time.Now().Add(-time.Minute)
}

func (tim *GameTimer) InProgress() bool {
	return time.Now().Before(tim.end_time)
}

func (tim *GameTimer) LestTime() time.Duration {
	return time.Until(tim.end_time)
}
func (tim *GameTimer) LestTimeSeconds() int {
	t := tim.LestTime()
	return int(math.Floor(t.Seconds()))
}
func (tim *GameTimer) LestTimeString() string {
	seconds := tim.LestTimeSeconds()
	return strconv.Itoa(seconds/60) + "m" +
		strconv.Itoa(seconds%60) + "s"
}
