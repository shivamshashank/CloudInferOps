from prometheus_client import Counter, Histogram, Gauge

# Counters for total requests and errors
INFERENCE_REQUESTS_TOTAL = Counter(
    "cloudinferops_inference_requests_total",
    "Total number of inference requests processed.",
    ["model", "provider"],
)

INFERENCE_ERRORS_TOTAL = Counter(
    "cloudinferops_inference_errors_total",
    "Total number of failed inference requests.",
    ["model", "provider", "error_type"],
)

MODEL_REQUESTS_TOTAL = Counter(
    "cloudinferops_inference_model_requests_total",
    "Total number of inference requests grouped by model.",
    ["model"],
)

# Latency histogram
INFERENCE_LATENCY_SECONDS = Histogram(
    "cloudinferops_inference_latency_seconds",
    "Inference latency in seconds.",
    ["model", "provider"],
    buckets=(0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0),
)

# Token metrics
INFERENCE_TOKENS_TOTAL = Counter(
    "cloudinferops_inference_tokens_total",
    "Total number of tokens processed (prompt or completion).",
    ["model", "provider", "token_type"],
)

INFERENCE_TOKENS_PER_SECOND = Gauge(
    "cloudinferops_inference_tokens_per_second",
    "Inference speed in tokens per second.",
    ["model", "provider"],
)

# Time to first token (TTFT)
INFERENCE_TTFT_SECONDS = Histogram(
    "cloudinferops_inference_ttft_seconds",
    "Time to first token (TTFT) in seconds.",
    ["model", "provider"],
    buckets=(0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0),
)
