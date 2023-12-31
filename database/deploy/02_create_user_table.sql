-- Deploy sportgether:create_user_table to pg

BEGIN;

CREATE TABLE IF NOT EXISTS sportgether_schema.users(
    id bigserial PRIMARY KEY,
    username text NOT NULL UNIQUE,
    email text NOT NULL UNIQUE,
    gender text NOT NULL DEFAULT 'MALE',
    profile_icon_name text NOT NULL DEFAULT 'CHINESE_BOY',
    password text NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    is_blocked bool NOT NULL DEFAULT false,
    version int NOT NULL DEFAULT 1
);

COMMIT;
