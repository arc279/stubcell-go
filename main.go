package main

import (
	"fmt"
	"os"
	"path"

	"./handler"

	flags "github.com/jessevdk/go-flags"

	"github.com/BurntSushi/toml"
)

type Options struct {
	ConfigFile string `short:"c" long:"config" description:"config file location"`
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

	var cfg = func(baseDir, configFile string) handler.Config {
		var cfg handler.Config
		_, err := toml.DecodeFile(configFile, &cfg)
		if err != nil {
			panic(err)
		}
		cfg.BaseDir = baseDir
		return cfg
	}(BaseDir, ConfigFile)
	fmt.Printf("%+v\n", cfg)

	// run
	e, s := cfg.NewEcho(true), cfg.NewServer()
	e.Run(s)
}
