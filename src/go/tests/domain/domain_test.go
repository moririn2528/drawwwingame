package domain_test

import (
	"drawwwingame/domain"
	infra "drawwwingame/infrastructure"
	"fmt"
	"log"
	"os"
	"testing"
)

func Init() {
	log.Println("test initialize start")
	var err error
	domain.SqlHandle, err = infra.NewSqlHandler()
	if err != nil {
		panic(err)
	}
	err = domain.Init()
	if err != nil {
		panic(err)
	}
}

func Close() {
	domain.SqlHandle.DeleteForTest()
	domain.Close()
}

func TestMain(m *testing.M) {
	Init()
	log.Println("test start")
	status := m.Run()
	log.Println("test end")
	Close()
	os.Exit(status)
}

func check(t *testing.T, arg, actual, expect interface{}) {
	if expect != actual {
		domain.PrintLog(fmt.Sprintf("\ninput: %v\noutput: %v\nexpect: %v", arg, actual, expect), 2)
		t.Errorf("check fail")
		Close()
		os.Exit(1)
	}
}
