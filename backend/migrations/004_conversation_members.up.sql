CREATE TABLE conversation_members (
    conversation_id UUID REFERENCES conversations(id),
    user_id UUID REFERENCES users(id),
    PRIMARY KEY (conversation_id, user_id)
)