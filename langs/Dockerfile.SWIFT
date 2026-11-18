FROM rust:alpine AS init-builder
WORKDIR /library-checker-init
COPY init /library-checker-init
RUN cargo build --release --target=x86_64-unknown-linux-musl

FROM swift:6.4.2-jammy AS libraries-builder
WORKDIR /build
COPY swift/ ./
RUN sh install.sh

FROM swift:6.4.2-jammy

COPY --from=init-builder /library-checker-init/target/x86_64-unknown-linux-musl/release/library-checker-init /usr/bin
COPY --from=libraries-builder /opt/swift-libraries /opt/swift-libraries

LABEL library-checker-image=true
