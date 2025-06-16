FROM rust:alpine as init-builder
WORKDIR /library-checker-init
COPY init /library-checker-init
RUN cargo build --release --target=x86_64-unknown-linux-musl

FROM mcr.microsoft.com/dotnet/sdk:7.0

ENV DOTNET_EnableWriteXorExecute=0
RUN dotnet new console -o /opt/C-Sharp
COPY resources/C-Sharp.csproj /opt/C-Sharp
RUN dotnet restore /opt/C-Sharp -r linux-x64

COPY --from=init-builder /library-checker-init/target/x86_64-unknown-linux-musl/release/library-checker-init /usr/bin

LABEL library-checker-image=true
