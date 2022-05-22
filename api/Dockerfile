FROM golang:1.17.0 as builder
WORKDIR /go/src/github.com/yosupo06/library-checker-judge/api
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build .

FROM alpine
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /go/src/github.com/yosupo06/library-checker-judge/api/api .
COPY ./langs.toml .
ENV PORT 50051
CMD ["./api"]
