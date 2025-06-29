// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"errors"
	"sync"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/collector/component"
)

func Meter(settings component.TelemetrySettings) metric.Meter {
	return settings.MeterProvider.Meter("github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver")
}

func Tracer(settings component.TelemetrySettings) trace.Tracer {
	return settings.TracerProvider.Tracer("github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver")
}

// TelemetryBuilder provides an interface for components to report telemetry
// as defined in metadata and user config.
type TelemetryBuilder struct {
	meter                                    metric.Meter
	mu                                       sync.Mutex
	registrations                            []metric.Registration
	KafkaBrokerClosed                        metric.Int64Counter
	KafkaBrokerConnects                      metric.Int64Counter
	KafkaBrokerThrottlingDuration            metric.Int64Histogram
	KafkaReceiverBytes                       metric.Int64Counter
	KafkaReceiverBytesUncompressed           metric.Int64Counter
	KafkaReceiverCurrentOffset               metric.Int64Gauge
	KafkaReceiverLatency                     metric.Int64Histogram
	KafkaReceiverMessages                    metric.Int64Counter
	KafkaReceiverOffsetLag                   metric.Int64Gauge
	KafkaReceiverPartitionClose              metric.Int64Counter
	KafkaReceiverPartitionStart              metric.Int64Counter
	KafkaReceiverUnmarshalFailedLogRecords   metric.Int64Counter
	KafkaReceiverUnmarshalFailedMetricPoints metric.Int64Counter
	KafkaReceiverUnmarshalFailedSpans        metric.Int64Counter
}

// TelemetryBuilderOption applies changes to default builder.
type TelemetryBuilderOption interface {
	apply(*TelemetryBuilder)
}

type telemetryBuilderOptionFunc func(mb *TelemetryBuilder)

func (tbof telemetryBuilderOptionFunc) apply(mb *TelemetryBuilder) {
	tbof(mb)
}

// Shutdown unregister all registered callbacks for async instruments.
func (builder *TelemetryBuilder) Shutdown() {
	builder.mu.Lock()
	defer builder.mu.Unlock()
	for _, reg := range builder.registrations {
		reg.Unregister()
	}
}

// NewTelemetryBuilder provides a struct with methods to update all internal telemetry
// for a component
func NewTelemetryBuilder(settings component.TelemetrySettings, options ...TelemetryBuilderOption) (*TelemetryBuilder, error) {
	builder := TelemetryBuilder{}
	for _, op := range options {
		op.apply(&builder)
	}
	builder.meter = Meter(settings)
	var err, errs error
	builder.KafkaBrokerClosed, err = builder.meter.Int64Counter(
		"otelcol_kafka_broker_closed",
		metric.WithDescription("The total number of connections closed."),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.KafkaBrokerConnects, err = builder.meter.Int64Counter(
		"otelcol_kafka_broker_connects",
		metric.WithDescription("The total number of connections opened."),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.KafkaBrokerThrottlingDuration, err = builder.meter.Int64Histogram(
		"otelcol_kafka_broker_throttling_duration",
		metric.WithDescription("The throttling duration in ms imposed by the broker when receiving messages."),
		metric.WithUnit("ms"),
	)
	errs = errors.Join(errs, err)
	builder.KafkaReceiverBytes, err = builder.meter.Int64Counter(
		"otelcol_kafka_receiver_bytes",
		metric.WithDescription("The size in bytes of received messages seen by the broker."),
		metric.WithUnit("By"),
	)
	errs = errors.Join(errs, err)
	builder.KafkaReceiverBytesUncompressed, err = builder.meter.Int64Counter(
		"otelcol_kafka_receiver_bytes_uncompressed",
		metric.WithDescription("The uncompressed size in bytes of received messages seen by the client."),
		metric.WithUnit("By"),
	)
	errs = errors.Join(errs, err)
	builder.KafkaReceiverCurrentOffset, err = builder.meter.Int64Gauge(
		"otelcol_kafka_receiver_current_offset",
		metric.WithDescription("Current message offset"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.KafkaReceiverLatency, err = builder.meter.Int64Histogram(
		"otelcol_kafka_receiver_latency",
		metric.WithDescription("The time it took in ms to receive a batch of messages."),
		metric.WithUnit("ms"),
	)
	errs = errors.Join(errs, err)
	builder.KafkaReceiverMessages, err = builder.meter.Int64Counter(
		"otelcol_kafka_receiver_messages",
		metric.WithDescription("The number of received messages."),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.KafkaReceiverOffsetLag, err = builder.meter.Int64Gauge(
		"otelcol_kafka_receiver_offset_lag",
		metric.WithDescription("Current offset lag"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.KafkaReceiverPartitionClose, err = builder.meter.Int64Counter(
		"otelcol_kafka_receiver_partition_close",
		metric.WithDescription("Number of finished partitions"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.KafkaReceiverPartitionStart, err = builder.meter.Int64Counter(
		"otelcol_kafka_receiver_partition_start",
		metric.WithDescription("Number of started partitions"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.KafkaReceiverUnmarshalFailedLogRecords, err = builder.meter.Int64Counter(
		"otelcol_kafka_receiver_unmarshal_failed_log_records",
		metric.WithDescription("Number of log records failed to be unmarshaled"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.KafkaReceiverUnmarshalFailedMetricPoints, err = builder.meter.Int64Counter(
		"otelcol_kafka_receiver_unmarshal_failed_metric_points",
		metric.WithDescription("Number of metric points failed to be unmarshaled"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.KafkaReceiverUnmarshalFailedSpans, err = builder.meter.Int64Counter(
		"otelcol_kafka_receiver_unmarshal_failed_spans",
		metric.WithDescription("Number of spans failed to be unmarshaled"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	return &builder, errs
}
