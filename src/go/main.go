package main

import (
	infra "drawwwingame/infrastructure"
	interf "drawwwingame/interface"

	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := infra.NewEchoHandler()
	interf.MiddlewareLogger = middleware.Logger()
	interf.MiddlewareRecover = middleware.Recover()
	interf.Run(e)
}
