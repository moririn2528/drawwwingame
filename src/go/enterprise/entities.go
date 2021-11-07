package domain

import "errors"

const (
	WEBSOCKET_READ_BUFFER_SIZE  = 1024
	WEBSOCKET_WRITE_BUFFER_SIZE = 1024
)

type Message struct {
	Uuid    string `json:"uuid"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

//error
var (
	ErrorParse = errors.New("parse error")
)
