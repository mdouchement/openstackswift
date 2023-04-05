package webserver

import (
	"strings"
	"net/http"
	"strconv"
	"time"
	"fmt"

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
	db database.Client
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

func GetPathInt(c echo.Context, name string) (int, error) {
	val := c.QueryParam(name)
	if val == "" {
		return 0, fmt.Errorf("%v path parameter value is empty or not specified", name)
	}
	return strconv.Atoi(val)
}

func setHeadersFromMeta(c echo.Context, metas []*model.Meta) error {
	for _, meta := range metas {
		c.Response().Header().Set(meta.Key, meta.Value)
	}
	return nil
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

	limit, err := GetPathInt(c, "limit")
	if (err != nil) {
		limit = -1
	}

	h.logger.Debugf("container Show: container %v limit=%v prefix=%v", c.Param("container"), limit, c.QueryParam("prefix"))

	objects, err := h.db.FindObjectsByContainerID(container.ID, limit, c.QueryParam("prefix"))

	/*
	// Delimiter spliter maybe can be removed
	delimiter := c.QueryParam("delimiter")
	if (delimiter != "") {
		for i := 0; i < len(objects); i++ {
			item := objects[i]
			item.Key = item.Key[strings.LastIndex(item.Key, string([]rune(delimiter)[0]))+1:]
		}
	}*/

	if err != nil && !h.db.IsNotFound(err) {
		return weberror.New(http.StatusInternalServerError, err.Error())
	}

	// get metas and set headers

	metas, err := h.db.FindMeta(container.ID, "")

	if err != nil && !h.db.IsNotFound(err) {
		return weberror.New(http.StatusInternalServerError, err.Error())
	}

	if metas != nil {
		setHeadersFromMeta(c, metas)
	}

	//

	container.Count = len(objects)
	c.Response().Header().Set("Date", time.Now().UTC().Format(http.TimeFormat))
	c.Response().Header().Set("X-Timestamp", strconv.FormatInt(container.CreatedAt.Unix(), 10))
	c.Response().Header().Set("X-Container-Object-Count", strconv.Itoa(container.Count))
	c.Response().Header().Set("X-Container-Bytes-Used", "0")

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
		container = &model.Container{Name: c.Param("container"), Count: 0, Bytes: 0}
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

// sets meta data for container
func (h *container) Update(c echo.Context) error {
	c.Set("handler_method", "container.Update")

	container, err := h.db.FindContainerByName(c.Param("container"))
	if err != nil && !h.db.IsNotFound(err) {
		return weberror.New(http.StatusInternalServerError, err.Error())
	}

	if h.db.IsNotFound(err) {
		return weberror.New(http.StatusNotFound, swift.ContainerNotFound.Text)
	}

	// Create and update metadata
	for key, values := range c.Request().Header {
		if (!strings.HasPrefix(key, "X-Container-Meta-") && len(values) > 0 ) {
			continue
		}
		// no string to mark only container
		_, err := h.db.AddMeta(container.ID, "", key, values[0])
		if err != nil {
			return weberror.New(http.StatusInternalServerError, err.Error())
		}
	}

	c.Response().Header().Set("Date", time.Now().UTC().Format(http.TimeFormat))
	c.Response().Header().Set("X-Timestamp", strconv.FormatInt(container.CreatedAt.Unix(), 10))
	return c.NoContent(http.StatusAccepted)
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

	objects, err := h.db.FindObjectsByContainerID(container.ID, 1, "")
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
