import logging
import os

from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.grpc import GrpcInstrumentorClient
from opentelemetry.instrumentation.logging import LoggingInstrumentor
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor


def setup_observability(app):
    service_name = os.environ.get("OTEL_SERVICE_NAME", "api-gateway")
    endpoint = os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4317")

    resource = Resource.create({SERVICE_NAME: service_name})
    provider = TracerProvider(resource=resource)
    provider.add_span_processor(BatchSpanProcessor(OTLPSpanExporter(endpoint=endpoint, insecure=True)))
    trace.set_tracer_provider(provider)

    FastAPIInstrumentor.instrument_app(app)
    GrpcInstrumentorClient().instrument()
    LoggingInstrumentor().instrument(set_logging_format=True)

    logging.getLogger(__name__).info("OpenTelemetry tracing enabled for %s", service_name)
