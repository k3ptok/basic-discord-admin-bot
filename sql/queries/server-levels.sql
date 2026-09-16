-- name: UpdateUserXP :one
INSERT INTO server_levels (guild_id, user_id, xp, level, last_xp_gain)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(guild_id, user_id) DO UPDATE SET
    xp = excluded.xp,
    level = excluded.level,
    last_xp_gain = excluded.last_xp_gain
RETURNING *;

-- name: GetUserRank :one
SELECT xp, level
FROM server_levels
WHERE guild_id = ? AND user_id = ?;

-- name: GetLeaderboard :many
SELECT user_id, xp, level
FROM server_levels
WHERE guild_id = ?
ORDER BY xp DESC
LIMIT 10;