from abc import ABC, abstractmethod
from typing import List, Dict, Any
from app.schemas import ChatCompletionRequest, ChatCompletionResponse


class BaseProvider(ABC):
    @abstractmethod
    async def get_available_models(self) -> List[Dict[str, Any]]:
        """Retrieve available models from the backend provider."""
        pass

    @abstractmethod
    async def chat_completions(
        self, request: ChatCompletionRequest
    ) -> ChatCompletionResponse:
        """Execute chat completions on the backend provider."""
        pass
