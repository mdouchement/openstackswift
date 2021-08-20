package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ncw/swift/v2"
)

func Authenticate(token string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if c.Request().Header.Get("X-Auth-Token") != token {
				return c.JSON(http.StatusUnauthorized, swift.AuthorizationFailed)
			}

			return next(c)
		}
	}
}
