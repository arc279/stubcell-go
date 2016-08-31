package main

import (
	"fmt"
	"os"
	"path"

	"github.com/arc279/stubcell-go/handler"

	flags "github.com/jessevdk/go-flags"

	"github.com/BurntSushi/toml"
)

type Options struct {
	ConfigFile string `short:"c" long:"config" description:"config file location"`
	Debug      bool   `long:"debug" description:"set debug flag"`
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

	var cfg = func(configFile string) handler.Config {
		var cfg handler.Config
		_, err := toml.DecodeFile(configFile, &cfg)
		if err != nil {
			panic(err)
		}
		cfg.BaseDir = path.Dir(configFile)
		return cfg
	}(ConfigFile)
	fmt.Printf("%+v\n", cfg)

	e, s := cfg.NewEcho(opts.Debug), cfg.NewServer()
	e.Run(s)
}
