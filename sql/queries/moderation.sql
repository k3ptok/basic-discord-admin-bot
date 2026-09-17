-- name: GetLogChannel :one
SELECT log_channel_id FROM guild_settings
WHERE guild_id = ? LIMIT 1;

-- name: SetLogChannel :exec
INSERT INTO guild_settings (guild_id, log_channel_id)
VALUES (?, ?)
ON CONFLICT(guild_id) DO UPDATE SET
    log_channel_id = excluded.log_channel_id,
    updated_at = CURRENT_TIMESTAMP;

-- name: InsertModLog :exec
INSERT INTO mod_logs (guild_id, user_id, action, reason)
VALUES (?, ?, ?, ?);

-- name: GetUserModLogs :many
SELECT action, reason, created_at FROM mod_logs
WHERE guild_id = ? AND user_id = ?
ORDER BY created_at DESC;

-- name: GetUserWarnings :many
SELECT reason, created_at 
FROM warnings 
WHERE guild_id = ? AND user_id = ?
ORDER BY created_at DESC
LIMIT 5;