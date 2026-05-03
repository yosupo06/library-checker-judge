FROM rust:alpine AS init-builder
WORKDIR /library-checker-init
COPY init /library-checker-init
RUN cargo build --release --target=x86_64-unknown-linux-musl

FROM ruby:2.7.1-alpine

COPY --from=init-builder /library-checker-init/target/x86_64-unknown-linux-musl/release/library-checker-init /usr/bin

LABEL library-checker-image=true
