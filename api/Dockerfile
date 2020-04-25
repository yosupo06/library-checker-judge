FROM golang:1.14.2-stretch as builder
WORKDIR /go/src/github.com/yosupo06/library-checker-judge/api
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build .

ENV PORT 50051
CMD ["/go/src/github.com/yosupo06/library-checker-judge/api/api"]
