package usecase

import (
	domain "drawwwingame/domain"
	"errors"
	"log"
	"strconv"

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

func getUser(uuid_str, tempid_str, name_str, email_str, password_str string) (*domain.User, error) {
	var err error
	var uuid *domain.UuidInt
	var tempid *domain.TempIdString
	var name *domain.NameString
	var email *domain.EmailString
	var password *domain.PasswordString
	var user *domain.User
	if uuid_str != "" {
		uuid, err = domain.NewUuidIntByString(uuid_str)
		if err != nil {
			domain.Log(err)
			return nil, err
		}
	}
	if tempid_str != "" {
		tempid, err = domain.NewTempIdString(tempid_str)
		if err != nil {
			domain.Log(err)
			return nil, err
		}
	}
	if name_str != "" {
		name, err = domain.NewNameString(name_str)
		if err != nil {
			domain.Log(err)
			return nil, err
		}
	}
	if email_str != "" {
		email, err = domain.NewEmailString(email_str)
		if err != nil {
			domain.Log(err)
			return nil, err
		}
	}
	if password_str != "" {
		password, err = domain.NewPasswordString(password_str)
		if err != nil {
			domain.Log(err)
			return nil, err
		}
	}

	if uuid != nil && tempid != nil {
		user, err = domain.NewUserById(uuid, tempid)
		if err != nil {
			domain.Log(err)
			return nil, err
		}
		return user, nil
	}
	if name != nil && email != nil && password != nil {
		user, err = domain.NewUsernCreate(name, email, password)
		if err != nil {
			domain.Log(err)
			return nil, err
		}
		return user, nil
	}
	if name != nil && password != nil {
		user, err = domain.NewUserByNamePassword(name, password)
		if err != nil {
			domain.Log(err)
			return nil, err
		}
		return user, nil
	}
	domain.LogStringf("no match case")
	return nil, ErrorInternal
}

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
			user, err := getUser(msg.Uuid, msg.Tempid, msg.Name, "", "")
			if err != nil {
				domain.Log(err)
				return err
			}
			if msg.Type == "info" {
				domain.Clients[ws] = user
				continue
			}
			domain.Broadcast <- msg
		}
	}
}

func Connect2WebSocket(c echo.Context, uuid_str, tempid_str string) error {
	ws, err := domain.ConnectWebSocket(c)
	if err != nil {
		domain.Log(err)
		return ErrorInternal
	}
	domain.Clients[ws] = nil
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
			for client, user := range domain.Clients {
				if user == nil {
					continue
				}
				if strconv.Itoa(user.GetUuidInt()) == msg.Uuid {
					continue
				}
				if msg.GroupId == "-1" || strconv.Itoa(user.GetGroupId()) != msg.GroupId {
					continue
				}
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

func Register(input_username, input_email, input_password string) error {
	user, err := getUser("", "", input_username, input_email, input_password)
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

func Authorize(crypto_str string) (string, error) {
	user, err := domain.Authorize(crypto_str)
	if err != nil {
		return "", err
	}
	return user.GetNameString(), nil
}

func Login(input_username, input_password string) (map[string]string, error) {
	s := make(map[string]string)
	user, err := getUser("", "", input_username, "", input_password)
	if err != nil {
		return s, err
	}
	if !user.EmailAuthorized() {
		return s, ErrorNotEmailAuthorized
	}
	return user.ToMapString(), nil
}

func GetGroup(e echo.Context) map[string]string {

	return nil
}

func SetGroup(uuid_str, tempid_str, group_id_str string) error {
	group_id, err := domain.NewGroupIdIntByString(group_id_str)
	if err != nil {
		return ErrorInputGroupid
	}
	user, err := getUser(uuid_str, tempid_str, "", "", "")
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
