package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"

	"github.com/stretchr/testify/assert"
)

func TestDebug(t *testing.T) {
	var cfg Config
	toml.Decode("", &cfg)

	e := cfg.NewEcho(true)

	req, _ := http.NewRequest("GET", "/debug", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(rec, e.Logger()))

	r := e.Router()
	r.Find("GET", "/debug", c)

	h := c.Handler()
	h(c)

	assert.Equal(t, rec.Code, 200)
	assert.Equal(t, rec.Body.String(), "Hello, World!")
}

func TestIndex(t *testing.T) {
	var ConfigToml = `
[[request.routes]]
method = "GET"
path = "/"
  [request.routes.response]
  status = 200
  text = """
<h1>It works!!</h1>
"""
    [request.routes.response.headers]
    Content-Type = "text/html; charset=utf-8"
    Host = "example.com"
`

	var cfg Config
	toml.Decode(ConfigToml, &cfg)

	e := cfg.NewEcho(true)

	req, _ := http.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(rec, e.Logger()))

	r := e.Router()
	r.Find("GET", "/", c)

	h := c.Handler()
	h(c)

	assert.Equal(t, rec.Code, 200)
	assert.Equal(t, rec.Header()["Host"][0], "example.com")
	assert.Equal(t, rec.Header()["Content-Type"][0], "text/html; charset=utf-8")
	assert.Equal(t, rec.Body.String(), "<h1>It works!!</h1>\n")
}

func TestWithPathParams(t *testing.T) {
	var ConfigToml = `
[[request.routes]]
method = "GET"
path = "/hello/:name"
  [request.routes.response]
  status = 200
  text = """
{
    "name": "{{ name }}"
}
"""
`

	var cfg Config
	toml.Decode(ConfigToml, &cfg)

	e := cfg.NewEcho(true)

	req, _ := http.NewRequest("GET", "/hello/foo", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(rec, e.Logger()))

	r := e.Router()
	r.Find("GET", "/hello/foo", c)

	h := c.Handler()
	h(c)

	assert.Equal(t, rec.Code, 200)
	assert.Equal(t, rec.Body.String(), `{
    "name": "foo"
}
`)
}

func TestWithQueryParams(t *testing.T) {
	var ConfigToml = `
[[request.routes]]
method = "GET"
path = "/hello"
  [request.routes.response]
  status = 200
  text = """
{
    "name": "{{ name|first }}"
}
"""
`

	var cfg Config
	toml.Decode(ConfigToml, &cfg)

	e := cfg.NewEcho(true)

	req, _ := http.NewRequest("GET", "/hello?name=foo", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(rec, e.Logger()))

	r := e.Router()
	r.Find("GET", "/hello", c)

	h := c.Handler()
	h(c)

	assert.Equal(t, rec.Code, 200)
	assert.Equal(t, rec.Body.String(), `{
    "name": "foo"
}
`)
}

func TestWithFormParams(t *testing.T) {
	var ConfigToml = `
[[request.routes]]
method = "POST"
path = "/hello"
  [request.routes.response]
  status = 201
  text = """
{
    "name": "{{ name|first }}"
}
"""
`

	var cfg Config
	toml.Decode(ConfigToml, &cfg)

	e := cfg.NewEcho(true)

	values := url.Values{}
	values.Set("name", "foo")
	req, _ := http.NewRequest("POST", "/hello", strings.NewReader(values.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(rec, e.Logger()))

	r := e.Router()
	r.Find("POST", "/hello", c)

	h := c.Handler()
	h(c)

	assert.Equal(t, req.Method, "POST")
	assert.Equal(t, rec.Code, 201)
	assert.Equal(t, rec.Body.String(), `{
    "name": "foo"
}
`)
}
