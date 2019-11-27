# Run from parent directory(/path/to/library-checker-judge)
FROM golang:1.13.4-stretch as builder
WORKDIR /go/src/github.com/yosupo06/library-checker-judge/
COPY ./api ./api
COPY ./compiler/langs.toml ./compiler/langs.toml
WORKDIR /go/src/github.com/yosupo06/library-checker-judge/api
RUN CGO_ENABLED=0 GOOS=linux go build .

ENV PORT 50051
CMD ["/go/src/github.com/yosupo06/library-checker-judge/api/api"]
