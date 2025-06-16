FROM rust:alpine as init-builder
WORKDIR /library-checker-init
COPY init /library-checker-init
RUN cargo build --release --target=x86_64-unknown-linux-musl

FROM rust:1.71.1-slim

COPY --from=init-builder /library-checker-init/target/x86_64-unknown-linux-musl/release/library-checker-init /usr/bin

LABEL library-checker-image=true
