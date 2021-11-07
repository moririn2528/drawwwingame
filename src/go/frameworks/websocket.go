package infra

import (
	domain "drawwwingame/entity"
	interf "drawwwingame/interface"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

//database

// type SqlHandler struct {
// 	db *sqlx.DB
// }

// type SqlResult struct {
// 	LastInsertId() (int64, error)
// 	RowsAffected() (int64, error)
// }

// func (handler *SqlHandler) NamedExec(query string, arg ...interface{}) {

// }

//echo

type HttpResponseWriter struct {
	Writer http.ResponseWriter
}

type HttpRequest struct {
	Request *http.Request
}

type EchoResponse struct {
	Res *echo.Response
}

type EchoContextHandler struct {
	C echo.Context
}

func (e *EchoContextHandler) Request() *HttpRequest {
	req := new(HttpRequest)
	req.Request = e.C.Request()
	return req
}

func (e *EchoContextHandler) Response() *EchoResponse {
	res := new(EchoResponse)
	res.Res = e.C.Response()
	return res
}

func (e *EchoResponse) Writer() *HttpResponseWriter {
	res := new(HttpResponseWriter)
	res.Writer = e.Res.Writer
	return res
}

type HttpHeader struct {
	Header *http.Header
}

type EchoHandlerFunc struct {
	Func echo.HandlerFunc
}

type EchoMiddlewareFunc struct {
	Func echo.MiddlewareFunc
}

type EchoHandler struct {
	E *echo.Echo
}

type EchoRoute struct {
	Route *echo.Route
}

func NewEchoHandlerFunc(f echo.HandlerFunc) *EchoHandlerFunc {
	handler := new(EchoHandlerFunc)
	handler.Func = f
	return handler
}

func NewEchoMiddlewareFunc(f echo.MiddlewareFunc) *EchoMiddlewareFunc {
	middle := new(EchoMiddlewareFunc)
	middle.Func = f
	return middle
}

func NewEchoHandler() *EchoHandler {
	e := new(EchoHandler)
	e.E = echo.New()
	return e
}

func NewEchoRoute(route *echo.Route) *EchoRoute {
	r := new(EchoRoute)
	r.Route = route
	return r
}

func parseEchoMiddlewareFunc(m []interf.EchoMiddlewareFunc) ([]echo.MiddlewareFunc, error) {
	s := make([]echo.MiddlewareFunc, len(m))
	for i, v := range m {
		middle, ok := v.(EchoMiddlewareFunc)
		if !ok {
			log.Println("Error: infra, parseEchoMiddlewareFunc, EchoMiddlewareFunc, parse error")
			return s, domain.ErrorParse
		}
		s[i] = middle.Func
	}
	return s, nil
}

func echoHttpWrapper(f func(string, echo.HandlerFunc, ...echo.MiddlewareFunc) *echo.Route, path string, h interf.EchoHandlerFunc, m []interf.EchoMiddlewareFunc) interf.EchoRoute {
	middle_func, err := parseEchoMiddlewareFunc(m)
	if err != nil {
		return new(EchoMiddlewareFunc)
	}
	handle, ok := h.(EchoHandlerFunc)
	if !ok {
		log.Println("Error: infra, EchoHandler, GET, EchoHandlerFunc, Parse Error")
		return new(EchoMiddlewareFunc)
	}
	return *NewEchoRoute(f(path, handle.Func, middle_func...))
}

func (e *EchoHandler) GET(path string, h interf.EchoHandlerFunc, m ...interf.EchoMiddlewareFunc) interf.EchoRoute {
	return echoHttpWrapper(e.E.GET, path, h, m)
}
func (e *EchoHandler) POST(path string, h interf.EchoHandlerFunc, m ...interf.EchoMiddlewareFunc) interf.EchoRoute {
	return echoHttpWrapper(e.E.POST, path, h, m)
}
func (e *EchoHandler) PUT(path string, h interf.EchoHandlerFunc, m ...interf.EchoMiddlewareFunc) interf.EchoRoute {
	return echoHttpWrapper(e.E.PUT, path, h, m)
}
func (e *EchoHandler) DELETE(path string, h interf.EchoHandlerFunc, m ...interf.EchoMiddlewareFunc) interf.EchoRoute {
	return echoHttpWrapper(e.E.DELETE, path, h, m)
}
func (e *EchoHandler) Use(m interf.EchoMiddlewareFunc) {
	middle, ok := m.(EchoMiddlewareFunc)
	if !ok {
		log.Println("Error: infra, EchoHandler, Use, Parse Error")
		return
	}
	e.E.Use(middle.Func)
}
func (e *EchoHandler) Start(address string) error {
	return e.E.Start(address)
}
func (e *EchoHandler) LoggerFatal(i ...interface{}) {
	e.E.Logger.Fatal(i)
}

//websocket

type WebSocketHandler struct {
	Conn *websocket.Conn
}

type WebSocketUpgrader struct {
	Upgrader *websocket.Upgrader
}

func (sock_upg *WebSocketUpgrader) Upgrade(w HttpResponseWriter, r HttpRequest, head HttpHeader) (*WebSocketHandler, error) {
	var err error
	sock := new(WebSocketHandler)
	sock.Conn, err = sock_upg.Upgrader.Upgrade(w.Writer, r.Request, *(head.Header))
	return sock, err
}

func (sock *WebSocketHandler) ReadJSON(v interface{}) error {
	return sock.Conn.ReadJSON(v)
}

func (sock *WebSocketHandler) WriteJSON(v interface{}) error {
	return sock.Conn.WriteJSON(v)
}

func (sock *WebSocketHandler) Close() error {
	return sock.Conn.Close()
}
