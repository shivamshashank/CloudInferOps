import os
from fastapi.testclient import TestClient
from unittest.mock import AsyncMock, patch
from app.main import app

client = TestClient(app)


@patch("app.providers.vllm.VLLMProvider.chat_completions", new_callable=AsyncMock)
def test_router_vllm_provider(mock_vllm_chat):
    # Set PROVIDER env var to vllm
    with patch.dict(os.environ, {"PROVIDER": "vllm"}):
        from app.schemas import (
            ChatCompletionResponse,
            ChatCompletionChoice,
            ChatCompletionUsage,
            ChatMessage,
        )

        mock_vllm_chat.return_value = ChatCompletionResponse(
            id="chatcmpl-vllm-id",
            created=1686935002,
            model="llama3-70b",
            choices=[
                ChatCompletionChoice(
                    index=0,
                    message=ChatMessage(
                        role="assistant", content="vllm hello response"
                    ),
                    finish_reason="stop",
                )
            ],
            usage=ChatCompletionUsage(
                prompt_tokens=10, completion_tokens=5, total_tokens=15
            ),
        )

        request_payload = {
            "model": "llama3-70b",
            "messages": [{"role": "user", "content": "hello"}],
        }

        response = client.post("/v1/chat/completions", json=request_payload)
        assert response.status_code == 200
        data = response.json()
        assert data["id"] == "chatcmpl-vllm-id"
        mock_vllm_chat.assert_called_once()
