package main

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Lang struct {
	ID              string   `toml:"id"`
	Source          string   `toml:"source"`
	Compile         []string `toml:"compile"`
	Exec            []string `toml:"exec"`
	ImageName       string   `toml:"image_name"`
	AdditionalFiles []string `toml:"additional_files"`
}

var langs map[string]Lang

func ReadLangs(tomlPath string) map[string]Lang {
	var tomlData struct {
		Langs []Lang `toml:"langs"`
	}
	if _, err := toml.DecodeFile(tomlPath, &tomlData); err != nil {
		log.Fatalln(err)
	}
	langs = make(map[string]Lang)
	for _, lang := range tomlData.Langs {
		langs[lang.ID] = lang
	}
	if _, ok := langs["checker"]; !ok {
		log.Fatal("lang file don't have checker")
	}
	return langs
}
