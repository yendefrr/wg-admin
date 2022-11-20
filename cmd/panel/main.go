package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"go/wg-admin/internal/app/panel"
	"log"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "configs-path", "configs/panel.toml", "path to configs file")
}

func main() {
	flag.Parse()

	config := panel.NewConfig()
	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal(err)
	}

	if err := panel.Start(config); err != nil {
		log.Fatal(err)
	}
}
