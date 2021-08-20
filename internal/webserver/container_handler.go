package webserver

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mdouchement/logger"
	"github.com/mdouchement/openstackswift/internal/database"
	"github.com/mdouchement/openstackswift/internal/model"
	"github.com/mdouchement/openstackswift/internal/webserver/serializer"
	"github.com/mdouchement/openstackswift/internal/webserver/weberror"
	"github.com/ncw/swift/v2"
)

type container struct {
	logger logger.Logger
	db     database.Client
}

func (h *container) List(c echo.Context) error {
	c.Set("handler_method", "container.List")

	containers, err := h.db.ListContainers()
	if err != nil {
		return weberror.New(http.StatusInternalServerError, err.Error())
	}

	//

	if c.Request().Header.Get("Accept") == "text/plain" {
		return c.String(http.StatusOK, serializer.TextContainers(containers))
	}
	// "application/json"
	return c.JSON(http.StatusOK, serializer.Containers(containers))
}

func (h *container) Show(c echo.Context) error {
	c.Set("handler_method", "container.Show")

	container, err := h.db.FindContainerByName(c.Param("container"))
	if err != nil {
		if h.db.IsNotFound(err) {
			return weberror.New(http.StatusNotFound, swift.ContainerNotFound.Text)
		}

		return weberror.New(http.StatusInternalServerError, err.Error())
	}

	objects, err := h.db.FindObjectsByContainerID(container.ID)
	if err != nil && !h.db.IsNotFound(err) {
		return weberror.New(http.StatusInternalServerError, err.Error())
	}

	//

	c.Response().Header().Set("Date", time.Now().UTC().Format(http.TimeFormat))
	c.Response().Header().Set("X-Timestamp", strconv.FormatInt(container.CreatedAt.Unix(), 10))
	c.Response().Header().Set("X-Container-Object-Count", strconv.Itoa(len(objects)))

	switch c.Request().Method {
	case http.MethodHead:
		return c.NoContent(http.StatusOK)
	case http.MethodGet:
		prefix := c.Request().Header.Get("prefix")

		if c.Request().Header.Get("Accept") == "text/plain" {
			return c.String(http.StatusOK, serializer.TextObjects(objects, prefix))
		}
		// "application/json"
		return c.JSON(http.StatusOK, serializer.Objects(objects, prefix))
	}
	return weberror.New(http.StatusNotFound, swift.BadRequest.Text)
}

func (h *container) Create(c echo.Context) error {
	c.Set("handler_method", "container.Create")

	container, err := h.db.FindContainerByName(c.Param("container"))
	if err != nil && !h.db.IsNotFound(err) {
		return weberror.New(http.StatusInternalServerError, err.Error())
	}

	if h.db.IsNotFound(err) {
		container = &model.Container{Name: c.Param("container")}
		err = h.db.Save(container)
		if err != nil {
			return weberror.New(http.StatusInternalServerError, err.Error())
		}
	}

	//

	c.Response().Header().Set("Date", time.Now().UTC().Format(http.TimeFormat))
	c.Response().Header().Set("X-Timestamp", strconv.FormatInt(container.CreatedAt.Unix(), 10))
	return c.NoContent(http.StatusCreated)
}

func (h *container) Delete(c echo.Context) error {
	c.Set("handler_method", "container.Delete")

	container, err := h.db.FindContainerByName(c.Param("container"))
	if err != nil {
		if h.db.IsNotFound(err) {
			return weberror.New(http.StatusNotFound, swift.ContainerNotFound.Text)
		}

		return weberror.New(http.StatusInternalServerError, err.Error())
	}

	//

	objects, err := h.db.FindObjectsByContainerID(container.ID)
	if err != nil && !h.db.IsNotFound(err) {
		return weberror.New(http.StatusInternalServerError, err.Error())
	}

	if len(objects) > 0 {
		return weberror.New(http.StatusConflict, swift.ContainerNotEmpty.Text)
	}

	//

	err = h.db.DeleteContainer(container.ID)
	if err != nil {
		return weberror.New(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
