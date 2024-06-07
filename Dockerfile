# Only for local usage
# GitHub Actions builds images in a different way
ARG GOLANG_VERSION=1.22
ARG UBUNTU_VERSION=oracular

FROM golang:${GOLANG_VERSION} as build

WORKDIR /go/src/dns-lookuper
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=local
RUN go install \
    -installsuffix "static" \
    -ldflags "                                            \
        -X main.Version=${VERSION}                        \
        -X main.GoVersion=$(go version | cut -d " " -f 3) \
        -X main.Compiler=$(go env CC)                     \
        -X main.Platform=$(go env GOOS)/$(go env GOARCH)  \
    " ./...

FROM ubuntu:${UBUNTU_VERSION} as runtime

COPY --from=build /go/bin/dns-lookuper /dns-lookuper
ENTRYPOINT [ "/dns-lookuper" ]
