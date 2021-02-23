ARG BASE_IMAGE=gcr.io/distroless/static:nonroot

# Build the manager binary
FROM gcr.io/gcp-runtimes/go1-builder:1.13 as builder

ENV PATH "/usr/local/go/bin:$PATH"

# Inject custom root certificate authorities if needed.
# Docker does not have a good conditional copy statement and requires that a
# source file exists to complete the copy function without error. Therefore, the
# README.md file will be copied to the image every time even if there are no
# .crt files.
COPY ./certs/* /usr/local/share/ca-certificates/
RUN update-ca-certificates

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY pkg/ pkg/


# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM ${BASE_IMAGE}
WORKDIR /
COPY --from=builder /workspace/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
