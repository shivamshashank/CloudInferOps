from fastapi.testclient import TestClient
from unittest.mock import AsyncMock, patch
from app.main import app

client = TestClient(app)


@patch(
    "app.providers.ollama.OllamaProvider.get_available_models", new_callable=AsyncMock
)
def test_list_models(mock_get_models):
    mock_get_models.return_value = [
        {"id": "llama3:latest", "object": "model"},
        {"id": "mistral:latest", "object": "model"},
    ]
    response = client.get("/models")
    assert response.status_code == 200
    assert response.json() == [
        {"id": "llama3:latest", "object": "model"},
        {"id": "mistral:latest", "object": "model"},
    ]
    mock_get_models.assert_called_once()
