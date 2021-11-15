package main

import (
	"drawwwingame/domain"
	infra "drawwwingame/infrastructure"
	interf "drawwwingame/interface"
	"log"
)

func main() {
	var err error
	domain.SqlHandle, err = infra.NewSqlHandler()
	if err != nil {
		log.Printf("Error: main, NewUserHandler, %v", err)
		return
	}
	interf.Run()
}
