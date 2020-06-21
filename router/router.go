package route

import (
	chat "chatting-example/chat"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func healthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "Health Check success")
}

func Init() *echo.Echo {
	e := echo.New()
	e.Debug = true
	e.Use(middleware.Recover())

	//CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))
	e.GET("/chat/:room_id/", chat.Start)
	e.GET("/health/", healthCheck)
	return e
}
