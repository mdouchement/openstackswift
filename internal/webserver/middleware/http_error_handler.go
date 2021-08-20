package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mdouchement/logger"
	"github.com/mdouchement/openstackswift/internal/webserver/weberror"
)

// NewHTTPErrorHandler is a middleware that formats rendered errors.
func NewHTTPErrorHandler(log logger.Logger) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		if !c.Response().Committed {
			var err2 error

			switch err := err.(type) {
			case *echo.HTTPError:
				err2 = weberror.New(err.Code, err.Error())
				err2 = c.JSON(weberror.StatusCode(err2), err2)
			case *weberror.Error:
				err2 = c.JSON(weberror.StatusCode(err), err)
			default:
				err = weberror.New(http.StatusInternalServerError, err.Error())
				err2 = c.JSON(weberror.StatusCode(err), err)
			}

			log.Error(err)
			if err2 != nil {
				log.Errorf("HTTPErrorHandler: %s", err2)
			}
		}
	}
}
