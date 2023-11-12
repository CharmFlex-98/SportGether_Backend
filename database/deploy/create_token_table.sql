-- Deploy sportgether:create_token_table to pg
-- requires: add_schema

BEGIN;

CREATE TABLE IF NOT EXISTS sportgether_schema.tokens(
    hashed bytea,
    expiry_time timestamp(0) with time zone NOT NULL,
    scope text NOT NULL,
    user_id bigint NOT NULL REFERENCES sportgether_schema.users ON DELETE CASCADE
);

COMMIT;
