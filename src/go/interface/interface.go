package interf

import "drawwwingame/usecase"

//database

// type SqlHandler interface {
// 	NamedExec(string, interface{}) (SqlRows, error)
// 	NamedQuery(string, interface{}) (SqlResult, error)
// 	Select(interface{}, string, ...interface{}) error
// }

// type SqlResult interface {
// 	LastInsertId() (int64, error)
// 	RowsAffected() (int64, error)
// }

// type SqlRows interface {
// }

// echo
var (
	MiddlewareLogger  EchoMiddlewareFunc
	MiddlewareRecover EchoMiddlewareFunc
)

type HttpResponseWriter interface {
}

type HttpRequest interface {
}

type HttpHeader interface {
}

type EchoResponse interface {
	Writer() HttpResponseWriter
}

type EchoContextHandler interface {
	Request() *HttpRequest
	Response() *EchoResponse
}

type EchoHandlerFunc interface {
}

type EchoMiddlewareFunc interface {
}

type EchoRoute interface {
}

type EchoHandler interface {
	GET(string, EchoHandlerFunc, ...EchoMiddlewareFunc) EchoRoute
	POST(string, EchoHandlerFunc, ...EchoMiddlewareFunc) EchoRoute
	PUT(string, EchoHandlerFunc, ...EchoMiddlewareFunc) EchoRoute
	DELETE(string, EchoHandlerFunc, ...EchoMiddlewareFunc) EchoRoute
	Use(EchoMiddlewareFunc)
	Start(string) error
	LoggerFatal(...interface{})
}

//websocket
type WebSocketHandler interface {
	ReadJSON(interface{}) error
	WriteJSON(interface{}) error
	Close() error
}

type WebSocketUpgrader interface {
	Upgrade(HttpResponseWriter, HttpRequest, HttpHeader) (WebSocketHandler, error)
}

type WebSocket struct {
	Socket WebSocketHandler
}

type WebSocketCreate struct {
	Upgrader WebSocketUpgrader
}

func NewWebSocket(upg WebSocketUpgrader, w HttpResponseWriter, r HttpRequest, head HttpHeader) (*WebSocket, error) {
	var err error
	sock := new(WebSocket)
	sock.Socket, err = upg.Upgrade(w, r, head)
	return sock, err
}

func (sock *WebSocket) ReadJSON(v interface{}) error {
	return sock.Socket.ReadJSON(v)
}

func (sock *WebSocket) WriteJSON(v interface{}) error {
	return sock.Socket.WriteJSON(v)
}

func (sock *WebSocket) Close() error {
	return sock.Socket.Close()
}

func (cre *WebSocketCreate) Set(readBufferSize int, writeBufferSize int){
	cre.Upgrader = 
	
}

func (cre *WebSocketCreate) Create(e EchoContext, readBufferSize int, writeBufferSize int) (WebSocket, error){
	cre.Upgrader.Upgrade(e)
	func(r *HttpRequest) bool {
		return true
	})
	
}
//main run
func Run(e EchoHandler) {
	e.Use(MiddlewareLogger)
	e.Use(MiddlewareRecover)

	e.GET("/ws", usecase.HandleConnection)
	go usecase.HandleMessage()
	e.GET("/testpage", usecase.TestPage)

	e.LoggerFatal(e.Start(":1213"))
}
