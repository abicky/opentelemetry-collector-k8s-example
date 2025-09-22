package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const name = "github.com/abicky/opentelemetry-collector-k8s-example"

var (
	tracer             = otel.Tracer(name)
	meter              = otel.Meter(name)
	logger             = otelslog.NewLogger(name)
	invocationCnt      metric.Int64Counter
	initialWaitSeconds = 1 * time.Second
)

func init() {
	var err error
	invocationCnt, err = meter.Int64Counter("hello.invocations",
		metric.WithDescription("The number of invocations"),
		metric.WithUnit("{invocation}"))
	if err != nil {
		panic(err)
	}

	if v := os.Getenv("INITIAL_WAIT_SECONDS"); v != "" {
		initialWaitSeconds, err = time.ParseDuration(v)
		if err != nil {
			panic(err)
		}
	}
}

func run() (err error) {
	shutdown, err := setupOTelSDK(context.Background())
	if err != nil {
		return
	}
	defer func() {
		err = errors.Join(err, shutdown(context.Background()))
	}()

	ctx, span := tracer.Start(context.Background(), "run")
	defer span.End()

	attr := attribute.String("key1", "value1")
	invocationCnt.Add(ctx, 1, metric.WithAttributes(attr))
	span.SetAttributes(attr)
	logger.InfoContext(ctx, "Hello World!", slog.String("key1", "value1"))
	fmt.Println("[INFO] Hello World!]")

	return
}

func main() {
	// Sleep a little so that the OpenTelemetry Collector can detect a newly created pod
	time.Sleep(initialWaitSeconds)
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
