package valobj

import (
	"drawwwingame/domain/internal"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	messageId   *CountingInt
	messageType = []string{
		"info", "lines",
		"mark:writer", "mark:answer",
		"text:writer", "text:answer",
	}
)

const (
	message_char = ",.-:# \n"
)

type MessageTypeString struct {
	str string
}

func matchTypeString(str string) bool {
	for _, s := range messageType {
		if s == str {
			return true
		}
	}
	return false
}

func NewMessageTypeString(str string) (*MessageTypeString, error) {
	if !matchTypeString(str) {
		internal.LogStringf("no match")
		return nil, internal.ErrorArg
	}
	s := new(MessageTypeString)
	s.str = str
	return s, nil
}

func NewInfoMessageTypeString() *MessageTypeString {
	return &MessageTypeString{str: "info"}
}

func (t *MessageTypeString) ToString() string {
	return t.str
}

func (t *MessageTypeString) checkarg(pos int, check string) bool {
	strs := strings.Split(t.str, ":")
	if len(strs) <= pos {
		return false
	}
	return strs[pos] == check
}

func (t *MessageTypeString) IsMark() bool {
	return t.checkarg(0, "mark")
}
func (t *MessageTypeString) IsInfo() bool {
	return t.str == "info"
}
func (t *MessageTypeString) IsAnswer() bool {
	return t.checkarg(1, "answer")
}
func (t *MessageTypeString) IsWriter() bool {
	return t.checkarg(1, "writer")
}
func (t *MessageTypeString) IsLines() bool {
	return t.str == "lines"
}
func (t *MessageTypeString) IsText() bool {
	return t.checkarg(0, "text")
}

type MessageString struct {
	str string
}

func NewMessageString(str string) (*MessageString, error) {
	if len(str) > 1000 {
		internal.LogStringf("length too large")
		return nil, internal.ErrorArg
	}
	if !utf8.Valid([]byte(str)) {
		internal.LogStringf("utf8 not validated")
		return nil, internal.ErrorArg
	}
	if strings.Contains(str, "--") {
		internal.LogStringf("contain --")
		return nil, internal.ErrorArg
	}
	for _, r := range str {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			continue
		}
		if unicode.In(r, unicode.Hiragana) || unicode.In(r, unicode.Katakana) || unicode.In(r, unicode.Han) {
			continue
		}
		if strings.ContainsRune(message_char, r) {
			continue
		}
		internal.LogStringf("error rune contain")
		return nil, internal.ErrorArg
	}
	s := new(MessageString)
	s.str = str
	return s, nil
}

func NewMessageStringNil() *MessageString {
	return &MessageString{
		str: "",
	}
}

func (mes *MessageString) ToString() string {
	return mes.str
}

var MessageMarksAllString = "ABC"

type MessageMarkString struct {
	str string
}

func NewMessageMarkString(str string) (*MessageMarkString, error) {
	if len(str) > 1 {
		internal.LogStringf("too large length")
		return nil, internal.ErrorArg
	}
	if !strings.Contains(MessageMarksAllString, str) {
		internal.LogStringf("mark not match")
		return nil, internal.ErrorArg
	}
	s := new(MessageMarkString)
	s.str = str
	return s, nil
}

func (mark *MessageMarkString) ToString() string {
	return mark.str
}

func (mark *MessageMarkString) ToInt() int {
	return int(mark.str[0]) - int('A')
}

type MessageId struct {
	id int
}

func SetMessageId(start int) {
	messageId = NewCountingInt(start, -1)
}

func NewMessageId() *MessageId {
	s := new(MessageId)
	s.id = messageId.ToInt()
	ok := messageId.Increment()
	if !ok {
		internal.LogStringf("message id increment error")
		panic("message id increment error")
	}
	return s
}

func NewMessageIdByInt(id int) (*MessageId, error) {
	s := new(MessageId)
	if id < -1 || messageId.ToInt() <= id {
		internal.LogStringf("input message id dont validate ")
		return nil, internal.ErrorArg
	}
	s.id = id
	return s, nil
}

func NewMessageIdByString(str string) (*MessageId, error) {
	var err error
	id, err := strconv.Atoi(str)
	if err != nil {
		internal.Log(err)
		return nil, internal.ErrorArg
	}
	return NewMessageIdByInt(id)
}

func NewMessageIdNil() *MessageId {
	s := new(MessageId)
	s.id = -1
	return s
}

func (m *MessageId) ToInt() int {
	return m.id
}
