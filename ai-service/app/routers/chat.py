"""AI advisor chat endpoint with SSE streaming."""

import json

from fastapi import APIRouter, Depends
from fastapi.responses import StreamingResponse

from app.models.schemas import ChatRequest, ChatSyncResponse
from app.services.advisor import chat, chat_stream
from app.utils.auth import verify_service_auth

router = APIRouter()


@router.post("/chat")
async def chat_endpoint(
    request: ChatRequest,
    _: None = Depends(verify_service_auth),
) -> StreamingResponse:
    """Chat with AI financial advisor. Returns SSE stream."""

    async def event_stream():
        try:
            async for chunk in chat_stream(
                request.message, request.history, request.context
            ):
                data = json.dumps({"content": chunk})
                yield f"data: {data}\n\n"
            yield "data: [DONE]\n\n"
        except Exception:
            error = json.dumps({"error": "Stream interrupted"})
            yield f"data: {error}\n\n"
            yield "data: [DONE]\n\n"

    return StreamingResponse(
        event_stream(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
        },
    )


@router.post("/chat/sync", response_model=ChatSyncResponse)
async def chat_sync_endpoint(
    request: ChatRequest,
    _: None = Depends(verify_service_auth),
) -> ChatSyncResponse:
    """Non-streaming chat endpoint for simpler integrations."""
    response = await chat(request.message, request.history, request.context)
    return ChatSyncResponse(content=response)
