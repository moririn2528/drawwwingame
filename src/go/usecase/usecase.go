package usecase

import (
	domain "drawwwingame/domain"
	"errors"
	"log"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

//Error
var (
	ErrorInputName          = errors.New("input name error")
	ErrorInputPassword      = errors.New("input password error")
	ErrorInputEmail         = errors.New("input email error")
	ErrorInputUuid          = errors.New("input uuid error")
	ErrorInputTempid        = errors.New("input temp id error")
	ErrorInputGroupid       = errors.New("input group id error")
	ErrorSendingEmailLimit  = errors.New("sending email limit")
	ErrorInternal           = errors.New("internal error")
	ErrorNotEmailAuthorized = errors.New("email not authorized")
)

func readWebSocket(ws *websocket.Conn) error {
	defer ws.Close()

	for {
		select {
		case <-domain.Goroutine_cancel:
			return nil
		default:
			var msg domain.Message
			err := ws.ReadJSON(&msg)

			if err != nil {
				delete(domain.Clients, ws)
				log.Printf("Error: domain, readWebSocket, ReadJSON, %v", err)
				return err
			}
			domain.Broadcast <- msg
		}
	}
}

func Connect2WebSocket(c echo.Context, uuid_str, tempid_str string) error {
	uuid, err := domain.NewUuidIntByString(uuid_str)
	if err != nil {
		domain.Log(err)
		return ErrorInputUuid
	}
	tempid, err := domain.NewTempIdString(tempid_str)
	if err != nil {
		domain.Log(err)
		return ErrorInputTempid
	}
	user, err := domain.NewUserById(uuid, tempid)
	if err != nil {
		domain.Log(err)
		return ErrorInternal
	}
	ws, err := domain.ConnectWebSocket(c)
	if err != nil {
		domain.Log(err)
		return ErrorInternal
	}
	domain.Clients[ws] = user
	go readWebSocket(ws)
	return err
}

func sendWebSocketMesage() error {
	for {
		select {
		case <-domain.Goroutine_cancel:
			return nil
		default:
			msg := <-domain.Broadcast
			for client := range domain.Clients {
				err := client.WriteJSON(msg)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(domain.Clients, client)
				}
			}
		}
	}
}

func Register(input_username, input_password, input_email string) error {
	username, err := domain.NewNameString(input_username)
	if err != nil {
		domain.Log(err)
		return ErrorInputName
	}
	password, err := domain.NewPasswordString(input_password)
	if err != nil {
		domain.Log(err)
		return ErrorInputPassword
	}
	email, err := domain.NewEmailString(input_email)
	if err != nil {
		domain.Log(err)
		return ErrorInputEmail
	}
	user, err := domain.NewUsernCreate(username, email, password)
	if err != nil {
		return err
	}
	_, err = user.SendAuthorizeEmail("auth")
	domain.ErrorIsEach(err,
		domain.ErrorUnnecessary,
		domain.ErrorSendingEmailLimit,
		domain.ErrorInternal,
	)
	if err == domain.ErrorSendingEmailLimit {
		return ErrorSendingEmailLimit
	}
	if err == domain.ErrorInternal {
		return ErrorInternal
	}
	return nil
}

func Authorize(crypto_str string) (map[string]string, error) {
	s := make(map[string]string)
	user, err := domain.Authorize(crypto_str)
	if err != nil {
		return s, err
	}
	return user.GetMapString(), nil
}

func Login(input_username, input_password string) (map[string]string, error) {
	s := make(map[string]string)
	username, err := domain.NewNameString(input_username)
	if err != nil {
		domain.Log(err)
		return s, ErrorInputName
	}
	password, err := domain.NewPasswordString(input_password)
	if err != nil {
		domain.Log(err)
		return s, ErrorInputPassword
	}
	user, err := domain.NewUserByNamePassword(username, password)
	if err != nil {
		return s, err
	}
	if !user.EmailAuthorized() {
		return s, ErrorNotEmailAuthorized
	}
	return user.GetMapString(), nil
}

func GetGroup(e echo.Context) map[string]string {

	return nil
}

func SetGroup(uuid_str, tempid_str, group_id_str string) error {
	uuid, err := domain.NewUuidIntByString(uuid_str)
	if err != nil {
		return ErrorInputUuid
	}
	tempid, err := domain.NewTempIdString(tempid_str)
	if err != nil {
		return ErrorInputTempid
	}
	group_id, err := domain.NewGroupIdIntByString(group_id_str)
	if err != nil {
		return ErrorInputGroupid
	}
	user, err := domain.NewUserById(uuid, tempid)
	if err != nil {
		return ErrorInternal
	}
	_, err = user.UpdateExceptId(nil, nil, group_id)
	if err != nil {
		domain.Log(err)
		return ErrorInternal
	}
	return nil
}

func Init() error {
	err := domain.Init()
	if err != nil {
		return ErrorInternal
	}
	go sendWebSocketMesage()
	return nil
}

func Close() {
	domain.Close()
}
