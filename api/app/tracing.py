import os
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor, ConsoleSpanExporter
from opentelemetry.sdk.resources import Resource

# Create provider and resource info
resource = Resource.create(
    attributes={"service.name": "cloudinferops-gateway", "service.version": "1.0.0"}
)
provider = TracerProvider(resource=resource)
trace.set_tracer_provider(provider)

# Determine OTLP/Tempo endpoint from environment
otlp_endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

if otlp_endpoint:
    try:
        from opentelemetry.exporter.otlp.proto.http.trace_exporter import (
            OTLPSpanExporter,
        )

        otlp_processor = BatchSpanProcessor(OTLPSpanExporter(endpoint=otlp_endpoint))
        provider.add_span_processor(otlp_processor)
    except ImportError:
        # Fallback if dependency is not installed
        provider.add_span_processor(BatchSpanProcessor(ConsoleSpanExporter()))
else:
    # Use console exporter by default for local development/testing
    provider.add_span_processor(BatchSpanProcessor(ConsoleSpanExporter()))

tracer = trace.get_tracer("cloudinferops-gateway-tracer")


def get_tracer():
    return tracer
