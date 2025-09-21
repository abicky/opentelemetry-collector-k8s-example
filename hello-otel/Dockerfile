FROM golang:1.25 AS builder

WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o hello

FROM gcr.io/distroless/base-debian12

COPY --from=builder /work/hello /usr/bin/hello

ENTRYPOINT ["hello"]
