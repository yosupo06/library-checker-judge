package main

import (
	"log"

	"github.com/BurntSushi/toml"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
)

func ReadLangs(tomlPath string) []*pb.Lang {
	var tomlData struct {
		Langs []struct {
			ID      string `toml:"id"`
			Name    string `toml:"name"`
			Version string `toml:"version"`
		}
	}
	if _, err := toml.DecodeFile(tomlPath, &tomlData); err != nil {
		log.Fatal(err)
	}
	var langs []*pb.Lang
	for _, lang := range tomlData.Langs {
		if lang.ID == "checker" {
			continue
		}
		langs = append(langs, &pb.Lang{
			Id:      lang.ID,
			Name:    lang.Name,
			Version: lang.Version,
		})
	}
	return langs
}
