CREATE TABLE chat_messages (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id   UUID NOT NULL,
    role         VARCHAR(20) NOT NULL,  -- user | assistant
    content      TEXT NOT NULL,
    context_data JSONB,
    tokens_used  INT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chat_user ON chat_messages(user_id, session_id, created_at DESC);
