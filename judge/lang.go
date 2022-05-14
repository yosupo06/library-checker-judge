package main

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Lang struct {
	Source  string `toml:"source"`
	Compile string `toml:"compile"`
	Exec    string `toml:"exec"`
}

var langs map[string]Lang

func init() {
	var tomlData struct {
		Langs []struct {
			Lang
			ID string `toml:"id"`
		} `toml:"langs"`
	}
	if _, err := toml.DecodeFile("../api/langs.toml", &tomlData); err != nil {
		log.Fatal(err)
	}
	langs = make(map[string]Lang)
	for _, lang := range tomlData.Langs {
		langs[lang.ID] = lang.Lang
	}
	if _, ok := langs["checker"]; !ok {
		log.Fatal("lang file don't have checker")
	}
}
