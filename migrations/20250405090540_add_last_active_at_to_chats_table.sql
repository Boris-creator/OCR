-- +goose Up
-- +goose StatementBegin
ALTER TABLE chats ADD COLUMN active_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE chats DROP COLUMN active_at;
-- +goose StatementEnd
