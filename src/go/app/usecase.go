package usecase

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	domain "drawwwingame/entity"
)

var (
	upgrader  WebSocketCreate
	clients   = make(map[WebSocket]bool)
	broadcast = make(chan domain.Message)
)

type WebSocket interface {
	ReadJSON(interface{}) error
	WriteJSON(interface{}) error
	Close() error
}

type EchoContext interface {
	FormValue(string) string
	Param(string) string
	QueryParam(string) string
	JSON(int, interface{}) error
	String(int, string) error
}

type WebSocketCreate interface {
	Set(int, int)
	Create(EchoContext) (WebSocket, error)
}

//upgrader.Upgrade(c.Response(), c.Request(), nil)

func HandleConnection(c EchoContext) error {
	ws, err := upgrader.Upgrade(c)

	if err != nil {
		return err
	}
	defer ws.Close()

	clients[ws] = true
	log.Println("handleConnection ok")

	for {
		var msg domain.Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error: ReadJSON %v", err)
			delete(clients, ws)
			return errors.New("ReadJSON error")
		}
		fmt.Printf("message: %s\n", msg.Message)
		broadcast <- msg
	}
}

func HandleMessage() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}

}

func TestPage(c EchoContext) error {
	return c.String(http.StatusOK, "Hello")
}
