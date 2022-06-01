package main

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Lang struct {
	ID        string   `toml:"id"`
	Source    string   `toml:"source"`
	Compile   []string `toml:"compile"`
	Exec      []string `toml:"exec"`
	ImageName string   `toml:"image_name"`
}

var langs map[string]Lang

func init() {
	var tomlData struct {
		Langs []Lang `toml:"langs"`
	}
	if _, err := toml.DecodeFile("../langs/langs.toml", &tomlData); err != nil {
		log.Fatal(err)
	}
	langs = make(map[string]Lang)
	for _, lang := range tomlData.Langs {
		langs[lang.ID] = lang
	}
	if _, ok := langs["checker"]; !ok {
		log.Fatal("lang file don't have checker")
	}
}
