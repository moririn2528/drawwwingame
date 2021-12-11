package valobj

import (
	"drawwwingame/domain/internal"
	"math/rand"
	"unicode"
)

type AlphanumString struct {
	str string
}

func NewAlphanumString(str string) (*AlphanumString, error) {
	for _, r := range str {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			internal.LogStringf("String error")
			return nil, internal.ErrorArg
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
func (alp *AlphanumString) Equal(alp2 *AlphanumString) bool {
	return alp.str == alp2.str
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
func (b *Boolean) Equal(b2 *Boolean) bool {
	return b.b == b2.b
}

type CountingInt struct {
	count     int
	count_max int
}

func NewCountingInt(start, count_max int) *CountingInt {
	s := new(CountingInt)
	s.count = start
	if count_max <= 0 {
		s.count_max = max_int
	} else {
		s.count_max = count_max
	}
	return s
}

func (c *CountingInt) ToInt() int {
	return c.count
}
func (c *CountingInt) GetMax() int {
	return c.count_max
}

func (c *CountingInt) Increment() bool {
	c.count++
	return c.count < c.count_max
}
