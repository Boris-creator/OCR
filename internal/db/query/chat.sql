-- name: CreateOrUpdateChat :exec
INSERT INTO chats (user_id) VALUES ($1)
ON CONFLICT(user_id)
DO UPDATE SET active_at = NOW();