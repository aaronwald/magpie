FROM mcr.microsoft.com/devcontainers/go:1-1.23-bookworm

# Install unzip
RUN apt-get update && apt-get install -y unzip && rm -rf /var/lib/apt/lists/*

# Download and install protobuf
RUN PB_REL="https://github.com/protocolbuffers/protobuf/releases" && \
  curl -LO $PB_REL/download/v25.1/protoc-25.1-linux-x86_64.zip && \
  unzip protoc-25.1-linux-x86_64.zip -d /usr/local/ && \
  rm protoc-25.1-linux-x86_64.zip
