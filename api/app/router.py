import os
from fastapi import APIRouter, Depends
from typing import List, Dict, Any
from app.providers.ollama import OllamaProvider
from app.providers.vllm import VLLMProvider
from app.providers.base import BaseProvider
from app.schemas import ChatCompletionRequest, ChatCompletionResponse

router = APIRouter()


def get_provider() -> BaseProvider:
    provider_name = os.getenv("PROVIDER", "ollama").lower()
    if provider_name == "vllm":
        return VLLMProvider()
    return OllamaProvider()


@router.get("/health")
async def health_check():
    return {"status": "healthy"}


@router.get("/models")
async def list_models(
    provider: BaseProvider = Depends(get_provider),
) -> List[Dict[str, Any]]:
    # Specifically return a flat list mapping to the expected structure in client models.go
    return await provider.get_available_models()


@router.get("/v1/models")
async def list_models_openai(provider: BaseProvider = Depends(get_provider)):
    # Standard OpenAI compatible response format
    models = await provider.get_available_models()
    return {"object": "list", "data": models}


@router.post("/v1/chat/completions")
async def chat_completions(
    request: ChatCompletionRequest, provider: BaseProvider = Depends(get_provider)
) -> ChatCompletionResponse:
    return await provider.chat_completions(request)
