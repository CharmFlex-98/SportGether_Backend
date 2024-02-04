-- Deploy sportgether:create_user_table to pg

BEGIN;

CREATE TABLE IF NOT EXISTS sportgether_schema.users(
    id bigserial PRIMARY KEY,
    username text NOT NULL UNIQUE,
    password text NOT NULL,
    email text NOT NULL UNIQUE,
    status text NOT NULL DEFAULT 'ACTIVE',
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version int NOT NULL DEFAULT 1
);

COMMIT;


-- Status can be ACTIVE, BLOCKED