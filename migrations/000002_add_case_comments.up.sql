-- Case comments table
CREATE TABLE IF NOT EXISTS case_comments (
    id UUID PRIMARY KEY,
    case_id UUID NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_case_comments_case ON case_comments(case_id, created_at ASC);
CREATE INDEX idx_case_comments_user ON case_comments(user_id);

-- Trigger for updated_at
CREATE TRIGGER update_case_comments_updated_at BEFORE UPDATE ON case_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
