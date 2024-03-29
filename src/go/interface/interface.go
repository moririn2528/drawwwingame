package interf

import (
	"drawwwingame/usecase"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ParamStruct struct {
	Username  string `param:"username" query:"username" form:"username" json:"username"`
	Password  string `param:"password" query:"password" form:"password" json:"password"`
	Email     string `param:"email" query:"email" form:"email" json:"email"`
	Uuid      string `param:"uuid" query:"uuid" form:"uuid" json:"uuid"`
	Tempid    string `param:"tempid" query:"tempid" form:"tempid" json:"tempid"`
	GroupId   string `param:"group_id" query:"group_id" form:"group_id" json:"group_id"`
	Admin     bool   `param:"admin" query:"admin" form:"admin" json:"admin"`
	CanAnswer bool   `param:"can_answer" query:"can_answer" form:"can_answer" json:"can_answer"`
	CanWriter bool   `param:"can_writer" query:"can_writer" form:"can_writer" json:"can_writer"`
	IsReady   bool   `param:"is_ready" query:"is_ready" form:"is_ready" json:"is_ready"`
}

func connect2WebSocket(c echo.Context) error {
	err := usecase.Connect2WebSocket(c)
	if err != nil {
		log.Printf("Error: ConnectWebSocket, %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func testPage(c echo.Context) error {
	return c.String(http.StatusOK, "Hello")
}

func register(c echo.Context) error {
	u := ParamStruct{}
	err := c.Bind(&u)
	if err != nil {
		log.Printf("Error: Bind, %v", err)
		return c.String(http.StatusInternalServerError, "")
	}
	log.Println("username", u.Username)
	log.Println("password", u.Password)
	err = usecase.Register(u.Username, u.Email, u.Password)
	if err != nil {
		log.Printf("Error: register, %v", err)
		return c.String(http.StatusInternalServerError, "")
	}
	c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, "*")
	return c.String(http.StatusOK, "")
}

func login(c echo.Context) error {
	u := ParamStruct{}
	err := c.Bind(&u)
	if err != nil {
		log.Printf("Error: Bind, %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	user, err := usecase.Login(u.Username, u.Password)
	if err == usecase.ErrorNotEmailAuthorized {
		return c.String(http.StatusUnauthorized, "not email authorized")
	}
	if err != nil {
		log.Printf("Error: register, %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, user)
}

// GET /auth/:str
func authorize(c echo.Context) error {
	str := c.Param("str")
	name, err := usecase.Authorize(str)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.String(http.StatusOK, "OK! Authorized. Thank you, "+name+".")
}

func testOptions(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderAccessControlAllowMethods, "GET,POST,HEAD,OPTIONS")
	c.Response().Header().Set(echo.HeaderAccessControlAllowHeaders, "Content-Type,Origin")
	return c.NoContent(http.StatusOK)
}

func serverHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, "*")
		return next(c)
	}
}

func getGroup(c echo.Context) error {
	u := ParamStruct{}
	err := c.Bind(&u)
	if err != nil {
		log.Printf("Error: Bind, %v", err)
		return c.String(http.StatusInternalServerError, "input format error")
	}

	return nil
}

func setGroup(c echo.Context) error {
	u := ParamStruct{}
	err := c.Bind(&u)
	if err != nil {
		log.Printf("Error: Bind, %v", err)
		return c.String(http.StatusInternalServerError, "input format error")
	}
	err = usecase.SetGroup(u.Uuid, u.Tempid, u.GroupId)
	if err != nil {
		log.Printf("Error: setGroup, usecase.SetGroup, %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.String(http.StatusOK, "OK!")
}

func setGroupRole(c echo.Context) error {
	u := ParamStruct{}
	err := c.Bind(&u)
	if err != nil {
		log.Printf("Error: Bind, %v", err)
		return c.String(http.StatusInternalServerError, "input format error")
	}
	err = usecase.SetGroupRole(u.Uuid, u.Tempid, u.CanAnswer, u.CanWriter)
	if err != nil {
		log.Printf("Error: setGroupRole, %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.String(http.StatusOK, "OK")
}

func Init() bool {
	err := usecase.Init()
	return err == nil
}

func Close() {
	usecase.Close()
}

func Run() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(serverHeader)

	flag := Init()
	defer Close()
	if !flag {
		return
	}

	e.OPTIONS("/login", testOptions)
	e.OPTIONS("/register", testOptions)
	e.OPTIONS("/group", testOptions)
	e.OPTIONS("/group/role", testOptions)

	e.GET("/ws", connect2WebSocket)
	e.GET("/testpage", testPage)
	e.POST("/register", register)
	e.POST("/login", login)
	e.POST("/group", setGroup)
	e.POST("/group/role", setGroupRole)
	e.GET("/auth/:str", authorize)

	e.Logger.Fatal(e.Start(":1213"))
}
