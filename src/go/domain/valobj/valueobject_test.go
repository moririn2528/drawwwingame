package valobj

import (
	"drawwwingame/domain/internal"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

func Close() {

}

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	internal.InitValTestLocal()
	status := m.Run()
	os.Exit(status)
}

func check(t *testing.T, arg, actual, expect interface{}) {
	if expect != actual {
		internal.PrintLog(fmt.Sprintf("\ninput: %v\noutput: %v\nexpect: %v", arg, actual, expect), 2)
		t.Errorf("check fail")
	}
}
