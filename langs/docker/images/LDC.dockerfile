FROM rust:alpine as init-builder
WORKDIR /library-checker-init
COPY init /library-checker-init
RUN cargo build --release --target=x86_64-unknown-linux-musl

FROM ubuntu:22.04 as builder

RUN apt-get update
RUN apt-get install -y wget xz-utils
RUN wget https://github.com/ldc-developers/ldc/releases/download/v1.29.0/ldc2-1.29.0-linux-x86_64.tar.xz
RUN tar -xf ldc2-1.29.0-linux-x86_64.tar.xz -C /opt

FROM ubuntu:22.04

RUN apt-get update
RUN apt-get install -y libxml2 gcc

COPY --from=builder /opt/ldc2-1.29.0-linux-x86_64/ /opt/ldc2-1.29.0-linux-x86_64/

ENV PATH $PATH:/opt/ldc2-1.29.0-linux-x86_64/bin

COPY --from=init-builder /library-checker-init/target/x86_64-unknown-linux-musl/release/library-checker-init /usr/bin

LABEL library-checker-image=true
