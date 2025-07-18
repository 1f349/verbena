-- name: RegisterBotToken :execlastid
INSERT INTO bot_tokens(owner_id, zone_id)
VALUES (?, ?);

-- name: BotTokenExists :one
SELECT *
FROM bot_tokens
WHERE id = ?;
