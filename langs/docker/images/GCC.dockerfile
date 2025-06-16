FROM rust:alpine as init-builder
WORKDIR /library-checker-init
COPY init /library-checker-init
RUN cargo build --release --target=x86_64-unknown-linux-musl

FROM ubuntu as builder
RUN apt-get update
RUN apt-get install -y git
WORKDIR /workdir
RUN git clone https://github.com/atcoder/ac-library/ -b v1.5.1

FROM gcc:14.2
COPY --from=builder /workdir/ac-library/atcoder /opt/ac-library/atcoder
COPY init /usr/bin

COPY --from=init-builder /library-checker-init/target/x86_64-unknown-linux-musl/release/library-checker-init /usr/bin

LABEL library-checker-image=true
