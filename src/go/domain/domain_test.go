package domain

import (
	"drawwwingame/domain/internal"
	"fmt"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	internal.InitValTestLocal()
	log.Println("test start")
	status := m.Run()
	log.Println("test end")
	Close()
	os.Exit(status)
}

func check(t *testing.T, arg, actual, expect interface{}) {
	if expect != actual {
		PrintLog(fmt.Sprintf("\ninput: %v\noutput: %v\nexpect: %v", arg, actual, expect), 2)
		t.Errorf("check fail")
		Close()
		os.Exit(1)
	}
}
