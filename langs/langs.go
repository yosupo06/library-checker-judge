package langs

import (
	_ "embed"
	"log/slog"
	"os"
	"slices"

	"github.com/BurntSushi/toml"
)

type Lang struct {
	ID              string   `toml:"id"`
	Name            string   `toml:"name"`
	Version         string   `toml:"version"`
	Source          string   `toml:"source"`
	Compile         []string `toml:"compile"`
	Exec            []string `toml:"exec"`
	ImageName       string   `toml:"image_name"`
	AdditionalFiles []string `toml:"additional_files"`
}

var LANGS []Lang
var LANG_CHECKER = Lang{
	Source:    "checker.cpp",
	ImageName: "library-checker-images-gcc",
	Compile:   []string{"g++", "-O2", "-std=c++14", "-DEVAL", "-march=native", "-o", "checker", "checker.cpp"},
	Exec:      []string{"./checker", "input.in", "actual.out", "expect.out"},
}

//go:embed langs.toml
var langToml string

func init() {
	var data struct {
		Langs []Lang `toml:"langs"`
	}
	if _, err := toml.Decode(langToml, &data); err != nil {
		slog.Error("toml decode failed", slog.Any("err", err))
		os.Exit(0)
	}
	LANGS = data.Langs
}

func GetLang(id string) (Lang, bool) {
	if idx := slices.IndexFunc(LANGS, func(lang Lang) bool {
		return lang.ID == id
	}); idx == -1 {
		return Lang{}, false
	} else {
		return LANGS[idx], true
	}
}
