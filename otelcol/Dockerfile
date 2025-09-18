FROM golang:1.25 AS builder

COPY build ./
RUN CGO_ENABLED=0 GOOS=linux go build -gcflags="all=-N -l" -o /otelcol

FROM golang:1.25 AS dlv
RUN go install github.com/go-delve/delve/cmd/dlv@latest

FROM gcr.io/distroless/base-debian12

COPY --from=builder /otelcol /usr/bin/otelcol
COPY --from=dlv /go/bin/dlv /usr/bin/dlv

ENTRYPOINT ["dlv", "--listen=:2345", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/usr/bin/otelcol", "--"]
