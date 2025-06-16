module github.com/yosupo06/library-checker-judge/langs

go 1.21

require (
	github.com/BurntSushi/toml v1.4.0
	github.com/yosupo06/library-checker-judge/executor v0.0.0-00010101000000-000000000000
)

require github.com/google/uuid v1.6.0 // indirect

replace github.com/yosupo06/library-checker-judge/executor => ../executor
