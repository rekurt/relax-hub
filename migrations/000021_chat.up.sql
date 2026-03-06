-- Chat conversations
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bathhouse_id UUID NOT NULL REFERENCES bathhouses(id),
    client_id UUID NOT NULL REFERENCES users(id),
    booking_id UUID REFERENCES bookings(id),
    last_message_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_conversations_bathhouse_client UNIQUE (bathhouse_id, client_id)
);

CREATE INDEX idx_conversations_client_id ON conversations (client_id);
CREATE INDEX idx_conversations_bathhouse_id ON conversations (bathhouse_id);
CREATE INDEX idx_conversations_last_message_at ON conversations (last_message_at DESC);

-- Chat messages
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id),
    text TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_messages_conversation_id ON messages (conversation_id, created_at DESC);
CREATE INDEX idx_messages_sender_id ON messages (sender_id);
CREATE INDEX idx_messages_unread ON messages (conversation_id, is_read) WHERE is_read = false;
