package handler

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/flosch/pongo2"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
)

type Route struct {
	Path     string `toml:"path"`
	Method   string `toml:"method"`
	Response struct {
		Status  int               `toml:"status"`
		File    string            `toml:"file"`
		Text    string            `toml:"text"`
		Headers map[string]string `toml:"headers"`
	} `toml:"response"`
}

type Config struct {
	BaseDir string
	Port    int `toml:"port"`
	Tls     struct {
		Enable bool   `toml:"enable"`
		Crt    string `toml:"crt"`
		Key    string `toml:"key"`
	} `toml:"tls"`
	Request struct {
		Cors struct {
			AllowOrigins []string `toml:"allow_origins"`
			AllowMethods []string `toml:"allow_methods"`
		} `toml:"cors"`
		Routes []Route `toml:"routes"`
	} `toml:"request"`
}

func (self Config) FilePath(s string) string {
	if path.IsAbs(s) {
		return s
	} else {
		return path.Join(self.BaseDir, s)
	}
}

func (self Config) NewServer() engine.Server {
	ec := engine.Config{
		Address: fmt.Sprintf(":%d", self.Port),
	}
	if self.Tls.Enable {
		ec.TLSCertFile = self.FilePath(self.Tls.Crt)
		ec.TLSKeyFile = self.FilePath(self.Tls.Key)
	}
	return standard.WithConfig(ec)
}

func (self Config) NewEcho(debug bool) *echo.Echo {
	e := echo.New()
	e.SetDebug(debug)
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: self.Request.Cors.AllowOrigins,
		AllowMethods: self.Request.Cors.AllowMethods,
	}))

	// add routing
	r := e.Router()
	if debug {
		r.Add("GET", "/debug", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		}, e)
	}

	for _, v := range self.Request.Routes {
		r.Add(v.Method, v.Path, v.CreateHandler(&self), e)
	}

	return e
}

func (v Route) CreateHandler(cfg *Config) func(c echo.Context) error {
	tpl, err := func(v Route) (*pongo2.Template, error) {
		fname := cfg.FilePath(v.Response.File)
		fstat, err := os.Stat(fname)
		if err == nil && !fstat.IsDir() {
			return pongo2.FromFile(fname)
		} else {
			return pongo2.FromString(v.Response.Text)
		}
	}(v)
	if err != nil {
		panic(err)
	}

	return func(c echo.Context) error {
		p2c := pongo2.Context{}
		// path params
		for _, k := range c.ParamNames() {
			p2c[k] = c.Param(k)
		}
		// form params
		for k, v := range c.FormParams() {
			p2c[k] = v // 配列になってる
		}
		// query params
		for k, v := range c.QueryParams() {
			p2c[k] = v // 配列になってる
		}

		out, err := tpl.Execute(p2c)
		if err != nil {
			return err
		}

		h := c.Response().Header().(*standard.Header).Header
		for k, v := range v.Response.Headers {
			h.Set(k, v)
		}

		w := c.Response().(*standard.Response).ResponseWriter
		w.WriteHeader(v.Response.Status)
		_, err = w.Write([]byte(out))
		if err != nil {
			return err
		}

		return nil
	}
}
