package main

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Category struct {
	Name     string
	Problems []string
}

type List struct {
	Category []Category
}

var list *List

func loadList() {
	if _, err := toml.DecodeFile("list.toml", &list); err != nil {
		log.Fatal(err)
	}
}
