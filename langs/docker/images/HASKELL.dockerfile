FROM rust:alpine as init-builder
WORKDIR /library-checker-init
COPY init /library-checker-init
RUN cargo build --release --target=x86_64-unknown-linux-musl

FROM haskell:9.4.4-slim

RUN cabal update && cabal install --lib \
    QuickCheck-2.14.2 \
    array-0.5.4.0 \
    attoparsec-0.14.4 \
    bitvec-1.1.3.0 \
    bytestring-0.11.4.0 \
    containers-0.6.7 \
    deepseq-1.4.8.0 \
    exceptions-0.10.7 \
    extra-1.7.12 \
    fgl-5.8.1.1 \
    hashable-1.4.2.0 \
    heaps-0.4 \
    integer-logarithms-1.0.3.1 \
    lens-5.2 \
    linear-base-0.3.0 \
    massiv-1.0.3.0 \
    mono-traversable-1.0.15.3 \
    mtl-2.3.1 \
    mutable-containers-0.3.4.1 \
    mwc-random-0.15.0.2 \
    parallel-3.2.2.0 \
    parsec-3.1.16.1 \
    primitive-0.7.4.0 \
    psqueues-0.2.7.3 \
    random-1.2.1.1 \
    reflection-2.1.6 \
    regex-tdfa-1.3.2 \
    safe-exceptions-0.1.7.3 \
    text-2.0.2 \
    tf-random-0.5 \
    transformers-0.6.1.0 \
    unboxing-vector-0.2.0.0 \
    unordered-containers-0.2.19.1 \
    utility-ht-0.0.16 \
    vector-0.13.0.0 \
    vector-algorithms-0.9.0.1 \
    vector-stream-0.1.0.0 \
    vector-th-unbox-0.2.2

COPY --from=init-builder /library-checker-init/target/x86_64-unknown-linux-musl/release/library-checker-init /usr/bin

LABEL library-checker-image=true
