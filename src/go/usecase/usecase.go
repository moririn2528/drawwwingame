package usecase

import (
	domain "drawwwingame/domain"
	"drawwwingame/domain/valobj"
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
	ErrorInputTypeString    = errors.New("input type string error")
	ErrorInputMessageInfo   = errors.New("input message info error")
	ErrorInputMessage       = errors.New("input message error")
	ErrorSendingEmailLimit  = errors.New("sending email limit")
	ErrorInternal           = errors.New("internal error")
	ErrorNotEmailAuthorized = errors.New("email not authorized")
	ErrorWebsocketInput     = errors.New("error input in websocket")
	ErrorNoMatter           = domain.ErrorNoMatter
)

func userWrap(uuid_str, tempid_str, name_str, email_str, password_str string) (*valobj.UuidInt, *valobj.TempIdString, *valobj.NameString, *valobj.EmailString, *valobj.PasswordString, error) {
	var err error
	var out_err error
	var uuid *valobj.UuidInt
	var tempid *valobj.TempIdString
	var name *valobj.NameString
	var email *valobj.EmailString
	var password *valobj.PasswordString
	if uuid_str != "" {
		uuid, err = valobj.NewUuidIntByString(uuid_str)
		if err != nil {
			domain.Log1(err)
			out_err = err
		}
	}
	if tempid_str != "" {
		tempid, err = valobj.NewTempIdString(tempid_str)
		if err != nil {
			domain.Log1(err)
			out_err = err
		}
	}
	if name_str != "" {
		name, err = valobj.NewNameString(name_str)
		if err != nil {
			domain.Log1(err)
			out_err = err
		}
	}
	if email_str != "" {
		email, err = valobj.NewEmailString(email_str)
		if err != nil {
			domain.Log1(err)
			out_err = err
		}
	}
	if password_str != "" {
		password, err = valobj.NewPasswordString(password_str)
		if err != nil {
			domain.Log1(err)
			out_err = err
		}
	}
	return uuid, tempid, name, email, password, out_err
}

func getUser(uuid *valobj.UuidInt, tempid *valobj.TempIdString, name *valobj.NameString, email *valobj.EmailString, password *valobj.PasswordString, errs ...error) (*domain.User, error) {
	var err error
	var user *domain.User
	for _, e := range errs {
		if e != nil {
			return nil, e
		}
	}
	if uuid != nil && tempid != nil {
		user, err = domain.NewUserById(uuid, tempid)
		if err != nil {
			domain.Log1(err)
			return nil, ErrorInternal
		}
		return user, nil
	}
	if name != nil && email != nil && password != nil {
		user, err = domain.NewUsernCreate(name, email, password)
		if err != nil {
			domain.Log1(err)
			return nil, ErrorInternal
		}
		return user, nil
	}
	if name != nil && password != nil {
		user, err = domain.NewUserByNamePassword(name, password)
		if err != nil {
			domain.Log1(err)
			return nil, ErrorInternal
		}
		return user, nil
	}
	domain.LogStringf("no match case")
	return nil, ErrorInternal
}

func getUserByRow(uuid_str, tempid_str, name_str, email_str, password_str string) (*domain.User, error) {
	return getUser(
		userWrap(uuid_str, tempid_str, name_str, email_str, password_str),
	)
}

func readWebSocket(ws *websocket.Conn) error {
	defer domain.DeleteConnectionByConn(ws)

	for {
		select {
		case <-domain.Goroutine_cancel:
			return nil
		default:
			msg, err := domain.NewInputWebSocketMessage(ws)
			if err == domain.ErrorNoMatter {
				return ErrorNoMatter
			}
			if err != nil {
				domain.Log1(err)
				return ErrorInternal
			}
			user, err := msg.GetUser()
			if err != nil {
				domain.Log1(err)
				return err
			}
			t, err := msg.GetType()
			if err != nil {
				domain.Log1(err)
				return ErrorInputTypeString
			}
			group, err := msg.GetnSetGroupObject()
			if err != nil {
				domain.Log1(err)
				return ErrorInputTypeString
			}
			if t.IsInfo() {
				// not broadcast
				if msg.Message == "uuid" {
					domain.Clients.Append(ws, user, group)
					if group.InGame() {
						err = group.SendGameInfoBefore()
						if err != nil {
							domain.Log1(err)
							return ErrorInternal
						}
					}
					continue
				}

				// broadcast
				msg_obj, err := msg.ToMessageObject()
				if err != nil {
					domain.Log1(err)
					return ErrorWebsocketInput
				}
				out_msg, err := msg_obj.ToOutputWebsocket()
				if err != nil {
					domain.Log1(err)
					return ErrorWebsocketInput
				}

				group, ok := domain.Clients.GetGroupByConn(ws)
				if !ok {
					domain.LogStringf("GetGroupByConn error")
					return ErrorInternal
				}
				switch msg.Message {
				case "join":
					domain.Broadcast <- out_msg
					group.AppendWaitingRoom()
					err = group.SendWaitingRoomInfoBefore()
					if err != nil {
						domain.Log1(err)
						return ErrorInternal
					}

				case "ready":
					domain.Broadcast <- out_msg
					err = group.UpdateReady(valobj.NewBoolean(true))
					if err != nil {
						domain.Log1(err)
						return ErrorInternal
					}
					ok = group.GroupIsReady()
					if !ok {
						break
					}
					err = group.StartGame(5, 1, 1)
					if err != nil {
						domain.Log(err)
						return ErrorInternal
					}

				default:
					domain.LogStringf("assert false")
					return ErrorWebsocketInput
				}
				continue
			}
			if t.IsMark() {
				mark, err := msg.ToMessageMark()
				if err != nil {
					domain.Log1(err)
					return ErrorWebsocketInput
				}
				err = mark.SaveSql()
				if err != nil {
					domain.Log1(err)
					return ErrorInternal
				}
				err = mark.Setting()
				if err != nil {
					domain.Log1(err)
					return ErrorInternal
				}
				out_msg, err := mark.ToOutputWebsocket()
				if err != nil {
					log.Println(err)
					return ErrorInternal
				}
				domain.Broadcast <- out_msg
				continue
			}
			message, err := msg.ToMessageObject()
			if err != nil {
				log.Println(err)
				return ErrorWebsocketInput
			}
			message.SaveSql()
			out_msg, err := message.ToOutputWebsocket()
			if err != nil {
				domain.Log1(err)
				return ErrorInternal
			}
			domain.Broadcast <- out_msg
		}
	}
}

func Connect2WebSocket(c echo.Context) error {
	ws, err := domain.ConnectWebSocket(c)
	if err != nil {
		domain.Log(err)
		return ErrorInternal
	}
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
			err := msg.Send()
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func Register(input_username, input_email, input_password string) error {
	user, err := getUserByRow("", "", input_username, input_email, input_password)
	if err != nil {
		return err
	}
	err = user.SendAuthorizeEmail("auth")
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
	user, err := getUserByRow("", "", input_username, "", input_password)
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

func SetGroup(uuid_str, tempid_str, group_str string) error {
	uuid, tempid, _, _, _, err := userWrap(uuid_str, tempid_str, "", "", "")
	if err != nil {
		log.Println("userWrap", err)
		return err
	}
	_, err = getUser(uuid, tempid, nil, nil, nil)
	if err != nil {
		log.Println("getUser", err)
		return ErrorInternal
	}
	group_id, err := valobj.NewGroupIdIntByString(group_str)
	if err != nil {
		log.Println("valobj.NewGroupIdIntByString", err)
		return ErrorInputGroupid
	}
	_, err = domain.NewGetnSetGroupObjectById(uuid, group_id)
	if err != nil {
		log.Println("domain.NewGetnSetGroupObjectById", err)
		return ErrorInputGroupid
	}
	return nil
}

func SetGroupRole(uuid_str, tempid_str string, can_answer_bool, can_writer_bool bool) error {
	uuid, tempid, _, _, _, err := userWrap(uuid_str, tempid_str, "", "", "")
	if err != nil {
		return err
	}
	_, err = getUser(uuid, tempid, nil, nil, nil)
	if err != nil {
		log.Println(err)
		return ErrorInternal
	}
	group, err := domain.NewExistGroupObjectById(uuid)
	if err != nil {
		log.Println(err)
		return ErrorInternal
	}
	can_answer := valobj.NewBoolean(can_answer_bool)
	can_writer := valobj.NewBoolean(can_writer_bool)
	err = group.UpdateRole(nil, can_answer, can_writer)
	if err != nil {
		log.Println(err)
		return ErrorInternal
	}
	return nil
}

func Init() error {
	err := domain.Init()
	if err != nil {
		log.Println(err)
		return ErrorInternal
	}
	go sendWebSocketMesage()
	return nil
}

func Close() {
	domain.Close()
}
