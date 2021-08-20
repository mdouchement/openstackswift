package middleware

import (
	"fmt"
	"net/http/httputil"

	"github.com/labstack/echo/v4"
)

func Dumpper() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			payload, err := httputil.DumpRequest(c.Request(), false)
			if err != nil {
				fmt.Println("DumpRequest:", err.Error())
			}
			fmt.Println(string(payload))

			err = next(c)
			if err != nil {
				return err
			}

			// payload, err = httputil.DumpResponse(c.Response(), false)
			// if err != nil {
			// 	fmt.Println("DumpResponse:", err.Error())
			// }
			// fmt.Println(string(payload))
			return nil
		}
	}
}
