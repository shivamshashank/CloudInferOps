import os
import time
import uuid
import httpx
from fastapi import HTTPException
from typing import List, Dict, Any
from app.providers.base import BaseProvider
from app.schemas import (
    ChatCompletionRequest,
    ChatCompletionResponse,
    ChatCompletionChoice,
    ChatCompletionUsage,
    ChatMessage,
)
from app.metrics import (
    INFERENCE_REQUESTS_TOTAL,
    INFERENCE_ERRORS_TOTAL,
    INFERENCE_LATENCY_SECONDS,
    INFERENCE_TOKENS_TOTAL,
    INFERENCE_TOKENS_PER_SECOND,
    INFERENCE_TTFT_SECONDS,
    MODEL_REQUESTS_TOTAL,
)


class VLLMProvider(BaseProvider):
    def __init__(self):
        # Default local URL or Kubernetes service URL
        self.host = os.getenv("VLLM_HOST", "http://localhost:8000").rstrip("/")

    async def get_available_models(self) -> List[Dict[str, Any]]:
        # vLLM has an OpenAI-compatible /v1/models endpoint
        url = f"{self.host}/v1/models"
        try:
            async with httpx.AsyncClient() as client:
                resp = await client.get(url, timeout=5.0)
                if resp.status_code != 200:
                    raise HTTPException(
                        status_code=502,
                        detail=f"vLLM returned status {resp.status_code}",
                    )

                data = resp.json()
                models = []
                # vLLM OpenAI API returns: {"object": "list", "data": [{"id": "model-name", ...}]}
                for m in data.get("data", []):
                    models.append({"id": m.get("id"), "object": "model"})
                return models
        except Exception as e:
            if isinstance(e, HTTPException):
                raise e
            raise HTTPException(
                status_code=503,
                detail=f"Could not connect to vLLM at {self.host}: {str(e)}",
            )

    async def chat_completions(
        self, request: ChatCompletionRequest
    ) -> ChatCompletionResponse:
        url = f"{self.host}/v1/chat/completions"
        model = request.model
        provider_name = "vllm"

        MODEL_REQUESTS_TOTAL.labels(model=model).inc()
        INFERENCE_REQUESTS_TOTAL.labels(model=model, provider=provider_name).inc()

        start_time = time.time()

        # Forward direct OpenAI payload to vLLM
        payload = request.dict()

        try:
            async with httpx.AsyncClient() as client:
                resp = await client.post(url, json=payload, timeout=60.0)

                if resp.status_code != 200:
                    INFERENCE_ERRORS_TOTAL.labels(
                        model=model,
                        provider=provider_name,
                        error_type=f"http_{resp.status_code}",
                    ).inc()
                    raise HTTPException(
                        status_code=resp.status_code,
                        detail=f"vLLM returned error: {resp.text}",
                    )

                data = resp.json()
                latency = time.time() - start_time

                # Update Latency
                INFERENCE_LATENCY_SECONDS.labels(
                    model=model, provider=provider_name
                ).observe(latency)

                # Mock TTFT
                ttft = latency * 0.3
                INFERENCE_TTFT_SECONDS.labels(
                    model=model, provider=provider_name
                ).observe(ttft)

                # Parse choices
                choices = []
                for idx, c in enumerate(data.get("choices", [])):
                    msg = c.get("message", {})
                    choices.append(
                        ChatCompletionChoice(
                            index=c.get("index", idx),
                            message=ChatMessage(
                                role=msg.get("role", "assistant"),
                                content=msg.get("content", ""),
                            ),
                            finish_reason=c.get("finish_reason", "stop"),
                        )
                    )

                # Collect usage stats
                usage = data.get("usage", {})
                prompt_tokens = usage.get("prompt_tokens", 0)
                completion_tokens = usage.get("completion_tokens", 0)
                total_tokens = usage.get(
                    "total_tokens", prompt_tokens + completion_tokens
                )

                # Track token metrics
                if prompt_tokens > 0:
                    INFERENCE_TOKENS_TOTAL.labels(
                        model=model, provider=provider_name, token_type="prompt"
                    ).inc(prompt_tokens)
                if completion_tokens > 0:
                    INFERENCE_TOKENS_TOTAL.labels(
                        model=model, provider=provider_name, token_type="completion"
                    ).inc(completion_tokens)

                # Speed
                if completion_tokens > 0 and latency > 0:
                    tokens_per_sec = completion_tokens / latency
                    INFERENCE_TOKENS_PER_SECOND.labels(
                        model=model, provider=provider_name
                    ).set(tokens_per_sec)

                return ChatCompletionResponse(
                    id=data.get("id", f"chatcmpl-{uuid.uuid4()}"),
                    created=data.get("created", int(start_time)),
                    model=model,
                    choices=choices,
                    usage=ChatCompletionUsage(
                        prompt_tokens=prompt_tokens,
                        completion_tokens=completion_tokens,
                        total_tokens=total_tokens,
                    ),
                )

        except Exception as e:
            if isinstance(e, HTTPException):
                raise e
            INFERENCE_ERRORS_TOTAL.labels(
                model=model, provider=provider_name, error_type="connection_error"
            ).inc()
            raise HTTPException(
                status_code=503,
                detail=f"Failed to communicate with vLLM backend: {str(e)}",
            )
