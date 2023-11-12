-- Deploy sportgether:create_user_table to pg
-- requires: add_schema

BEGIN;

CREATE TABLE IF NOT EXISTS sportgether_schema.users(
    id bigserial PRIMARY KEY,
    username text NOT NULL UNIQUE,
    email text NOT NULL UNIQUE,
    password text NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    is_blocked bool NOT NULL DEFAULT false,
    version int NOT NULL DEFAULT 1
);

COMMIT;
