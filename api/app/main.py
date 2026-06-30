import logging
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from prometheus_client import make_asgi_app
from app.router import router

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("cloudinferops-gateway")

app = FastAPI(
    title="CloudInferOps Inference Gateway",
    version="1.0.0",
    description="Python FastAPI gateway for managing model serving backends.",
)

# CORS configuration
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include router
app.include_router(router)

# Mount Prometheus metrics endpoint at /metrics
metrics_app = make_asgi_app()
app.mount("/metrics", metrics_app)

# Initialize OpenTelemetry instrumentation
try:
    from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor

    FastAPIInstrumentor.instrument_app(app)
    logger.info("OpenTelemetry tracing successfully instrumented.")
except Exception as e:
    logger.warning(f"Could not instrument OpenTelemetry: {str(e)}")
