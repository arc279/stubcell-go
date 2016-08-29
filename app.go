package main

import (
	"fmt"
	"net/http"
	"os"
	"path"

	flags "github.com/jessevdk/go-flags"

	"github.com/flosch/pongo2"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"

	"github.com/BurntSushi/toml"
)

type Options struct {
	ConfigFile string `short:"c" long:"config" description:"config file location"`
}

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

func (v Route) CreateHandler(cfg *Config) func(c echo.Context) error {
	tpl, err := func(v Route) (*pongo2.Template, error) {
		fname := path.Join(cfg.BaseDir, v.Response.File)
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
		_, err = w.Write([]byte(out))
		if err != nil {
			return err
		}

		return nil
	}
}

func (cfg Config) Run(debug bool) {
	e := echo.New()
	e.SetDebug(debug)
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.Request.Cors.AllowOrigins,
		AllowMethods: cfg.Request.Cors.AllowMethods,
	}))

	// add routing
	r := e.Router()
	r.Add("GET", "/debug", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	}, e)

	for _, v := range cfg.Request.Routes {
		r.Add(v.Method, v.Path, v.CreateHandler(&cfg), e)
	}

	// create server
	server := func() *standard.Server {
		ec := engine.Config{
			Address: fmt.Sprintf(":%d", cfg.Port),
		}
		if cfg.Tls.Enable {
			ec.TLSCertFile = cfg.Tls.Crt
			ec.TLSKeyFile = cfg.Tls.Key
		}
		return standard.WithConfig(ec)
	}()

	e.Run(server)
}

func main() {
	var opts Options
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	ConfigFile := fmt.Sprintf(".%s.toml", path.Base(os.Args[0]))
	if opts.ConfigFile != "" {
		ConfigFile = opts.ConfigFile
	}
	var BaseDir = path.Dir(ConfigFile)

	var cfg = func(baseDir, configFile string) Config {
		var cfg Config
		_, err := toml.DecodeFile(configFile, &cfg)
		if err != nil {
			panic(err)
		}
		cfg.BaseDir = baseDir
		return cfg
	}(BaseDir, ConfigFile)
	fmt.Printf("%+v\n", cfg)

	cfg.Run(true)
}
