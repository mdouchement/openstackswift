package webserver

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mdouchement/logger"
	"github.com/mdouchement/openstackswift/internal/webserver/weberror"
	"github.com/ncw/swift/v2"
)

type keystone3 struct {
	logger logger.Logger

	tenant   string
	domain   string
	username string
	password string
}

func (h *keystone3) Authenticate(c echo.Context) error {
	c.Set("handler_method", "keystone3.Authenticate")

	// Filter params
	var params keystone3params
	if err := c.Bind(&params); err != nil {
		return weberror.New(http.StatusBadRequest, swift.BadRequest.Text)
	}

	// Authorization
	authorized, err := h.authorize(params)
	if err != nil {
		return weberror.New(http.StatusBadRequest, swift.BadRequest.Text)
	}
	if !authorized {
		return weberror.New(http.StatusUnauthorized, swift.AuthorizationFailed.Text)
	}

	// Render response
	c.Response().Header().Set("X-Subject-Token", CraftToken(h.username))
	return c.JSON(http.StatusCreated, keystone3reponse{
		Token: Token{
			IssuedAt:  time.Now().Format(time.RFC3339),
			ExpiresAt: time.Now().AddDate(0, 1, 0).Format(time.RFC3339),
			Catalog: []Catalog{
				{
					Type: "object-store",
					ID:   "050726f278654128aba89757ae25950c",
					Name: "swift",
					Endpoints: []Endpoint{
						{
							ID:        "068d1b359ee84b438266cb736d81de97",
							Interface: swift.EndpointTypePublic,
							Region:    "RegionOne",
							RegionID:  "RegionOne",
							URL:       c.Scheme() + "://" + c.Request().Host + "/v1/AUTH_" + h.username,
						},
					},
				},
			},
		},
	})
}

func (h *keystone3) authorize(params keystone3params) (bool, error) {
	for _, method := range params.Auth.Identity.Methods {
		switch method {
		case "password":
			return params.Auth.Scope.Project.Domain.Name == h.domain &&
				params.Auth.Scope.Project.Name == h.tenant &&
				params.Auth.Identity.Password.User.Name == h.username &&
				params.Auth.Identity.Password.User.Password == h.password, nil
		}
	}

	return false, errors.New("unsupported Auth.Identity.Methods")
}

//
//
//
//
// Params
//
//
//
//

// V3 Authentication request
// http://docs.openstack.org/developer/keystone/api_curl_examples.html
// http://developer.openstack.org/api-ref-identity-v3.html
// Code imported from: https://github.com/ncw/swift
type keystone3params struct {
	Auth struct {
		Identity struct {
			Methods  []string        `json:"methods"`
			Password *v3AuthPassword `json:"password,omitempty"`
		} `json:"identity"`
		Scope *v3Scope `json:"scope,omitempty"`
	} `json:"auth"`
}

//
// Response
//

type keystone3reponse struct {
	Token `json:"token"`
}

type Token struct {
	ExpiresAt string    `json:"expires_at"`
	IssuedAt  string    `json:"issued_at"`
	Catalog   []Catalog `json:"catalog"`
}

type Catalog struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Endpoints []Endpoint `json:"endpoints"`
}

type Endpoint struct {
	ID        string             `json:"id"`
	RegionID  string             `json:"region_id"`
	URL       string             `json:"url"`
	Region    string             `json:"region"`
	Interface swift.EndpointType `json:"interface"`
}

//
// Types
//

type v3Scope struct {
	Project *v3Project `json:"project,omitempty"`
	Domain  *v3Domain  `json:"domain,omitempty"`
	Trust   *v3Trust   `json:"OS-TRUST:trust,omitempty"`
}

type v3Domain struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type v3Project struct {
	ID     string    `json:"id,omitempty"`
	Name   string    `json:"name,omitempty"`
	Domain *v3Domain `json:"domain,omitempty"`
}

type v3Trust struct {
	ID string `json:"id"`
}

type v3User struct {
	Domain   *v3Domain `json:"domain,omitempty"`
	ID       string    `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	Password string    `json:"password,omitempty"`
}

type v3AuthToken struct {
	ID string `json:"id"`
}

type v3AuthPassword struct {
	User v3User `json:"user"`
}
