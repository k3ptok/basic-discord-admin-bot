-- +goose Up
CREATE TABLE server_levels (
    guild_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    xp INTEGER NOT NULL DEFAULT 0,
    level INTEGER NOT NULL DEFAULT 0,
    last_xp_gain DATETIME NOT NULL,
    PRIMARY KEY (guild_id, user_id)
);

-- +goose Down
DROP TABLE server_levels;
