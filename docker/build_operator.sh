
#!/bin/bash
set -ex

case "$(uname -m)" in
    x86_64)
        arch="amd64"
        ;;
    aarch64)
        arch="arm64"
        ;;
    armv7l)
        arch="arm"
        ;;
    *)
        echo "Unsupported architecture: $(uname -m)" >&2
        exit 1
        ;;
esac

# Install dependencies

(
    cd /tmp
    wget -q "https://go.dev/dl/go1.27.2.linux-${arch}.tar.gz" -O go.tar.gz
    tar -xf go.tar.gz
    mv go /usr/local
    ln -s /usr/local/go/bin/go /usr/bin/go
    go version
)

# Build

export PATH="$PATH:$HOME/go/bin"
export GOPATH="$HOME/go"
mkdir -p build

GO111MODULE=on go mod vendor
CGO_ENABLED=0 go build -o build/postgres-operator -v -ldflags "$OPERATOR_LDFLAGS" cmd/main.go
