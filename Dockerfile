FROM golang:1.25-alpine AS builder

WORKDIR /workspace

COPY go.mod go.sum ./

RUN go mod download

COPY ./cmd ./cmd
COPY ./pkg ./pkg

RUN CGO_ENABLED=0 go build \
    -ldflags "-X github.com/forkspacer/forkspacer/pkg/version.Version=${VERSION} \
    -X github.com/forkspacer/forkspacer/pkg/version.GitCommit=${GIT_COMMIT} \
    -X github.com/forkspacer/forkspacer/pkg/version.BuildDate=${BUILD_DATE}" \
    -o api ./cmd/main.go

FROM gcr.io/distroless/static-debian12:latest

WORKDIR /workspace

COPY --from=builder /workspace/api .

ENTRYPOINT ["/workspace/api"]
