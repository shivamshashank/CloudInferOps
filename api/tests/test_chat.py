from fastapi.testclient import TestClient
from unittest.mock import AsyncMock, patch
from app.main import app

client = TestClient(app)


@patch("app.providers.ollama.OllamaProvider.chat_completions", new_callable=AsyncMock)
def test_chat_completions(mock_chat):
    # Mock return value matching ChatCompletionResponse structure
    from app.schemas import (
        ChatCompletionResponse,
        ChatCompletionChoice,
        ChatCompletionUsage,
        ChatMessage,
    )

    mock_chat.return_value = ChatCompletionResponse(
        id="chatcmpl-test-id",
        created=1686935002,
        model="llama3",
        choices=[
            ChatCompletionChoice(
                index=0,
                message=ChatMessage(role="assistant", content="hello response"),
                finish_reason="stop",
            )
        ],
        usage=ChatCompletionUsage(
            prompt_tokens=10, completion_tokens=5, total_tokens=15
        ),
    )

    request_payload = {
        "model": "llama3",
        "messages": [{"role": "user", "content": "hello"}],
        "temperature": 0.7,
    }

    response = client.post("/v1/chat/completions", json=request_payload)
    assert response.status_code == 200

    data = response.json()
    assert data["id"] == "chatcmpl-test-id"
    assert data["model"] == "llama3"
    assert data["choices"][0]["message"]["content"] == "hello response"
    assert data["usage"]["total_tokens"] == 15
    mock_chat.assert_called_once()
