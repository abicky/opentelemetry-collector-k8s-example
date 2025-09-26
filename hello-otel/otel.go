package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	// Some of the environment variables that the Go SDK doesn't support
	propagatorsKey     = "OTEL_PROPAGATORS"
	defaultPropagators = "tracecontext,baggage"

	// Environment variables for exporter selection
	// cf. https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/#exporter-selection
	tracesExporterKey      = "OTEL_TRACES_EXPORTER"
	metricsExporterKey     = "OTEL_METRICS_EXPORTER"
	logsExporterKey        = "OTEL_LOGS_EXPORTER"
	defaultTracesExporter  = "otlp"
	defaultMetricsExporter = "otlp"
	defaultLogsExporter    = "otlp"

	// Custom environment variables similar to the OTLP exporter (https://opentelemetry.io/docs/specs/otel/protocol/exporter/)
	consoleTracesPrettyPrintKey      = "OTEL_EXPORTER_CONSOLE_TRACES_PRETTY_PRINT"
	consoleLogsPrettyPrintKey        = "OTEL_EXPORTER_CONSOLE_LOGS_PRETTY_PRINT"
	consoleMetricsPrettyPrintKey     = "OTEL_EXPORTER_CONSOLE_METRICS_PRETTY_PRINT"
	defaultConsoleTracesPrettyPrint  = "false"
	defaultConsoleLogsPrettyPrint    = "false"
	defaultConsoleMetricsPrettyPrint = "false"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline using the environment variables
// described on https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/.
// If it does not return an error, make sure to call shutdown for proper cleanup.
//
// Although the SDK specification states that an unknown enum value must be ignored with a warning,
// this function returns an error instead to make such issues more noticeable to users.
func setupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	prop, err := newPropagator()
	if err != nil {
		handleErr(err)
		return
	}
	otel.SetTextMapPropagator(prop)

	tracerProvider, err := newTracerProvider(ctx)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	meterProvider, err := newMeterProvider(ctx)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	loggerProvider, err := newLoggerProvider(ctx)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return
}

func getEnv(envName, defaultValue string) string {
	if v := os.Getenv(envName); v != "" {
		return v
	}
	return defaultValue
}

func newPropagator() (propagation.TextMapPropagator, error) {
	var propagators []propagation.TextMapPropagator
	for _, propagator := range strings.Split(getEnv(propagatorsKey, defaultPropagators), ",") {
		switch propagator {
		case "tracecontext":
			propagators = append(propagators, propagation.TraceContext{})
		case "baggage":
			propagators = append(propagators, propagation.Baggage{})
		default:
			return nil, fmt.Errorf("unsupported propagator: %q from %s=%q", propagator, propagatorsKey, os.Getenv(propagatorsKey))
		}
	}

	return propagation.NewCompositeTextMapPropagator(propagators...), nil
}

func newTracerProvider(ctx context.Context) (*sdktrace.TracerProvider, error) {
	var opts []sdktrace.TracerProviderOption
	for _, exporter := range strings.Split(getEnv(tracesExporterKey, defaultTracesExporter), ",") {
		switch exporter {
		case "otlp":
			exp, err := otlptracegrpc.New(ctx)
			if err != nil {
				return nil, err
			}
			opts = append(opts, sdktrace.WithBatcher(exp))
		case "console":
			var stdoutOpts []stdouttrace.Option
			if getEnv(consoleTracesPrettyPrintKey, defaultConsoleTracesPrettyPrint) == "true" {
				stdoutOpts = append(stdoutOpts, stdouttrace.WithPrettyPrint())
			}
			exp, err := stdouttrace.New(stdoutOpts...)
			if err != nil {
				return nil, err
			}
			opts = append(opts, sdktrace.WithBatcher(exp))
		case "none":
			// Do nothing
		default:
			return nil, fmt.Errorf("unsupported exporter: %q from %s=%q", exporter, tracesExporterKey, os.Getenv(tracesExporterKey))
		}
	}

	return sdktrace.NewTracerProvider(opts...), nil
}

func newMeterProvider(ctx context.Context) (*sdkmetric.MeterProvider, error) {
	var opts []sdkmetric.Option
	for _, exporter := range strings.Split(getEnv(metricsExporterKey, defaultMetricsExporter), ",") {
		switch exporter {
		case "otlp":
			exp, err := otlpmetricgrpc.New(ctx)
			if err != nil {
				return nil, err
			}
			opts = append(opts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)))
		case "console":
			var stdoutOpts []stdoutmetric.Option
			if getEnv(consoleMetricsPrettyPrintKey, defaultConsoleMetricsPrettyPrint) == "true" {
				stdoutOpts = append(stdoutOpts, stdoutmetric.WithPrettyPrint())
			}
			exp, err := stdoutmetric.New(stdoutOpts...)
			if err != nil {
				return nil, err
			}
			opts = append(opts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)))
		case "none":
			// Do nothing
		default:
			return nil, fmt.Errorf("unsupported exporter: %q from %s=%q", exporter, metricsExporterKey, os.Getenv(metricsExporterKey))
		}
	}

	return sdkmetric.NewMeterProvider(opts...), nil
}

func newLoggerProvider(ctx context.Context) (*sdklog.LoggerProvider, error) {
	var opts []sdklog.LoggerProviderOption
	for _, exporter := range strings.Split(getEnv(logsExporterKey, defaultLogsExporter), ",") {
		switch exporter {
		case "otlp":
			exp, err := otlploggrpc.New(ctx)
			if err != nil {
				return nil, err
			}
			opts = append(opts, sdklog.WithProcessor(sdklog.NewBatchProcessor(exp)))
		case "console":
			var stdoutOpts []stdoutlog.Option
			if getEnv(consoleLogsPrettyPrintKey, defaultConsoleLogsPrettyPrint) == "true" {
				stdoutOpts = append(stdoutOpts, stdoutlog.WithPrettyPrint())
			}
			exp, err := stdoutlog.New(stdoutOpts...)
			if err != nil {
				return nil, err
			}
			opts = append(opts, sdklog.WithProcessor(sdklog.NewBatchProcessor(exp)))
		case "none":
			// Do nothing
		default:
			return nil, fmt.Errorf("unsupported exporter: %q from %s=%q", exporter, logsExporterKey, os.Getenv(logsExporterKey))
		}
	}

	return sdklog.NewLoggerProvider(opts...), nil
}
