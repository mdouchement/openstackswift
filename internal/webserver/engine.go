package webserver

import (
	"fmt"
	"net/http"
	"path"
	"sort"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mdouchement/logger"
	"github.com/mdouchement/openstackswift/internal/database"
	"github.com/mdouchement/openstackswift/internal/storage"
	middlewarepkg "github.com/mdouchement/openstackswift/internal/webserver/middleware"
)

// A Controller is an Iversion Of Control pattern used to init the server package.
type Controller struct {
	Version  string
	Logger   logger.Logger
	Database database.Client
	Storage  storage.Backend
	//
	Tenant   string
	Domain   string
	Username string
	Password string
}

// EchoEngine instantiates the wep server.
func EchoEngine(ctrl Controller) *echo.Echo {
	engine := echo.New()
	// engine.Use(middleware.Recover())
	engine.Use(middleware.Gzip())
	engine.Use(middlewarepkg.Logger(ctrl.Logger))
	// engine.Use(middlewarepkg.Dumpper())

	engine.HTTPErrorHandler = middlewarepkg.NewHTTPErrorHandler(ctrl.Logger)

	engine.Pre(middleware.Rewrite(map[string]string{
		"/": "/version",
	}))

	//
	//
	//

	router := engine.Group("")

	// Generic handlers
	//
	router.GET("/version", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{
			"version": ctrl.Version,
		})
	})

	// Keystone
	//
	k3 := keystone3{
		logger:   ctrl.Logger,
		tenant:   ctrl.Tenant,
		domain:   ctrl.Domain,
		username: ctrl.Username,
		password: ctrl.Password,
	}
	router.POST("/v3/auth/tokens", k3.Authenticate)

	// Swift
	//
	// https://docs.openstack.org/api-ref/object-store/index.html
	//

	swift := router.Group("/v1/AUTH_" + ctrl.Username)
	auth := middlewarepkg.Authenticate(CraftToken(ctrl.Username))

	// Container
	//
	container := container{
		logger: ctrl.Logger,
		db:     ctrl.Database,
	}
	swift.GET("", container.List, auth)
	swift.HEAD("/:container", container.Show, auth) // check existence
	swift.GET("/:container", container.Show, auth)
	swift.PUT("/:container", container.Create, auth)
	swift.POST("/:container", container.Update, auth)
	swift.DELETE("/:container", container.Delete, auth)

	// Object
	//
	object := object{
		logger:  ctrl.Logger,
		db:      ctrl.Database,
		storage: ctrl.Storage,
	}
	swift.HEAD("/:container/:object", object.Show, auth)
	swift.GET("/:container/:object", object.Download, auth)
	swift.PUT("/:container/:object", func(c echo.Context) error {
		switch {
		case c.Request().Header.Get("X-Copy-From") != "":
			c.Set("object_source", c.Request().Header.Get("X-Copy-From"))
			c.Set("object_destination", path.Join(c.Param("container"), c.Param("object")))
			return object.Copy(c)
		case c.Request().Header.Get("X-Object-Manifest") != "":
			return object.Manifest(c)
		default:
			return object.Upload(c)
		}
	}, auth)
	swift.Add("COPY", "/:container/:object", func(c echo.Context) error {
		c.Set("object_source", path.Join(c.Param("container"), c.Param("object")))
		c.Set("object_destination", c.Request().Header.Get("Destination"))
		return object.Copy(c)
	}, auth)

	swift.POST("/:container/:object", object.Update, auth)
	swift.DELETE("/:container/:object", object.Delete, auth)

	return engine
}

// PrintRoutes prints the Echo engin exposed routes.
func PrintRoutes(e *echo.Echo) {
	ignored := map[string]bool{
		"":   true,
		".":  true,
		"/*": true,
	}

	routes := e.Routes()
	sort.Slice(routes, func(i int, j int) bool {
		return routes[i].Path < routes[j].Path
	})

	fmt.Println("Routes:")
	for _, route := range routes {
		if ignored[route.Path] {
			continue
		}
		fmt.Printf("%6s %s\n", route.Method, route.Path)
	}
}

// CraftToken returns the auth token for a user.
func CraftToken(username string) string {
	return "tk_" + username
}
