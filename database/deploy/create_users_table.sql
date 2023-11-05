-- Deploy sportgether:create_users_table to pg

BEGIN;

CREATE TABLE IF NOT EXISTS users(
    id bigserial PRIMARY KEY,
    username text NOT NULL,
    email text NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    is_blocked bool DEFAULT false,
    version integer NOT NULL DEFAULT 1
);

COMMIT;
