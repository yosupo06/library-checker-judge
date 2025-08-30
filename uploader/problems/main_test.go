package main

import (
    "testing"
)

func TestToSourceURL(t *testing.T) {
    url := toSourceURL("/dummy/paty/to/sample/aplusb/info.toml")
    if url != "https://github.com/yosupo06/library-checker-problems/tree/master/sample/aplusb" {
        t.Fatal("URL is differ", url)
    }
}

