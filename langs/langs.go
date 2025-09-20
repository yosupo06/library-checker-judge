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
	ID:        "checker",
	Source:    "checker.cpp",
	ImageName: "library-checker-images-gcc",
	Compile:   []string{"g++", "-O2", "-std=c++17", "-march=native", "-o", "checker", "checker.cpp"},
	Exec:      []string{"./checker", "input.in", "actual.out", "expect.out"},
}
var LANG_VERIFIER = Lang{
	ID:        "verifier",
	Source:    "verifier.cpp",
	ImageName: "library-checker-images-gcc",
	Compile:   []string{"g++", "-O2", "-std=c++17", "-march=native", "-o", "verifier", "verifier.cpp"},
	Exec:      []string{"./verifier"},
}
var LANG_GENERATOR = Lang{
	ID:        "generator",
	Source:    "generator.cpp",
	ImageName: "library-checker-images-gcc",
	Compile:   []string{"g++", "-O2", "-std=c++17", "-march=native", "-o", "generator", "generator.cpp"},
	Exec:      []string{"./generator", "0"},
}
var LANG_MODEL_SOLUTION Lang

//go:embed langs.toml
var langToml string

func init() {
	var data struct {
		Langs []Lang `toml:"langs"`
	}
	if _, err := toml.Decode(langToml, &data); err != nil {
		slog.Error("toml decode failed", slog.Any("err", err))
		os.Exit(1)
	}
	LANGS = data.Langs

	if lang, ok := GetLang("cpp"); !ok {
		slog.Error("cpp is not found in langs")
		os.Exit(1)
	} else {
		LANG_MODEL_SOLUTION = lang
		LANG_MODEL_SOLUTION.Compile = addProblemIncludeFlags(LANG_MODEL_SOLUTION.Compile, LANG_MODEL_SOLUTION.Source)
	}

	LANG_CHECKER.Compile = addProblemIncludeFlags(LANG_CHECKER.Compile, LANG_CHECKER.Source)
	LANG_VERIFIER.Compile = addProblemIncludeFlags(LANG_VERIFIER.Compile, LANG_VERIFIER.Source)
	LANG_GENERATOR.Compile = addProblemIncludeFlags(LANG_GENERATOR.Compile, LANG_GENERATOR.Source)
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

func addProblemIncludeFlags(args []string, source string) []string {
	flags := []string{"-iquote", "/problem", "-iquote", "/problem/common", "-iquote", "/problem/sol"}
	insertIdx := -1
	for i, arg := range args {
		if arg == source {
			insertIdx = i
			break
		}
	}
	if insertIdx == -1 {
		return append(args, flags...)
	}

	res := make([]string, 0, len(args)+len(flags))
	res = append(res, args[:insertIdx]...)
	res = append(res, flags...)
	res = append(res, args[insertIdx:]...)
	return res
}
