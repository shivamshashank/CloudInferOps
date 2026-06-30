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


class OllamaProvider(BaseProvider):
    def __init__(self):
        # Default local URL or Kubernetes service URL
        self.host = os.getenv("OLLAMA_HOST", "http://localhost:11434").rstrip("/")

    async def get_available_models(self) -> List[Dict[str, Any]]:
        url = f"{self.host}/api/tags"
        try:
            async with httpx.AsyncClient() as client:
                resp = await client.get(url, timeout=5.0)
                if resp.status_code != 200:
                    raise HTTPException(
                        status_code=502,
                        detail=f"Ollama returned status {resp.status_code}",
                    )

                data = resp.json()
                models = []
                for m in data.get("models", []):
                    # We return a flat list mapping to the expected structure in client models.go
                    models.append({"id": m.get("name"), "object": "model"})
                return models
        except Exception as e:
            if isinstance(e, HTTPException):
                raise e
            raise HTTPException(
                status_code=503,
                detail=f"Could not connect to Ollama at {self.host}: {str(e)}",
            )

    async def chat_completions(
        self, request: ChatCompletionRequest
    ) -> ChatCompletionResponse:
        url = f"{self.host}/api/chat"
        model = request.model
        provider_name = "ollama"

        # Record model total requests
        MODEL_REQUESTS_TOTAL.labels(model=model).inc()
        INFERENCE_REQUESTS_TOTAL.labels(model=model, provider=provider_name).inc()

        start_time = time.time()

        # Prepare Ollama specific payload
        ollama_messages = [
            {"role": msg.role, "content": msg.content} for msg in request.messages
        ]
        payload = {
            "model": model,
            "messages": ollama_messages,
            "stream": False,
            "options": {},
        }
        if request.temperature is not None:
            payload["options"]["temperature"] = request.temperature
        if request.max_tokens is not None:
            payload["options"]["num_predict"] = request.max_tokens

        try:
            async with httpx.AsyncClient() as client:
                # Time to First Token (TTFT) simulation/measurement:
                # For non-streaming, TTFT is roughly equivalent to the response time minus generating time,
                # but to be accurate we measure the full call.
                resp = await client.post(url, json=payload, timeout=60.0)

                if resp.status_code != 200:
                    INFERENCE_ERRORS_TOTAL.labels(
                        model=model,
                        provider=provider_name,
                        error_type=f"http_{resp.status_code}",
                    ).inc()
                    raise HTTPException(
                        status_code=resp.status_code,
                        detail=f"Ollama returned error: {resp.text}",
                    )

                data = resp.json()
                latency = time.time() - start_time

                # Update Latency metric
                INFERENCE_LATENCY_SECONDS.labels(
                    model=model, provider=provider_name
                ).observe(latency)

                # Mock TTFT metric (for non-streaming, TTFT is often close to full latency / 2)
                ttft = latency * 0.4
                INFERENCE_TTFT_SECONDS.labels(
                    model=model, provider=provider_name
                ).observe(ttft)

                # Parse assistant response message
                assistant_msg = data.get("message", {})
                choice = ChatCompletionChoice(
                    index=0,
                    message=ChatMessage(
                        role=assistant_msg.get("role", "assistant"),
                        content=assistant_msg.get("content", ""),
                    ),
                    finish_reason="stop",
                )

                # Collect usage stats from Ollama response
                prompt_tokens = data.get("prompt_eval_count", 0)
                completion_tokens = data.get("eval_count", 0)
                total_tokens = prompt_tokens + completion_tokens

                # Track token metrics
                if prompt_tokens > 0:
                    INFERENCE_TOKENS_TOTAL.labels(
                        model=model, provider=provider_name, token_type="prompt"
                    ).inc(prompt_tokens)
                if completion_tokens > 0:
                    INFERENCE_TOKENS_TOTAL.labels(
                        model=model, provider=provider_name, token_type="completion"
                    ).inc(completion_tokens)

                # Speed calculation (Tokens per second)
                if completion_tokens > 0 and latency > 0:
                    tokens_per_sec = completion_tokens / latency
                    INFERENCE_TOKENS_PER_SECOND.labels(
                        model=model, provider=provider_name
                    ).set(tokens_per_sec)

                return ChatCompletionResponse(
                    id=f"chatcmpl-{uuid.uuid4()}",
                    created=int(start_time),
                    model=model,
                    choices=[choice],
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
                detail=f"Failed to communicate with Ollama backend: {str(e)}",
            )
