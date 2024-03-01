-- Deploy sportgether:08_create_token_table to pg

BEGIN;

CREATE TABLE IF NOT EXISTS sportgether_schema.tokens (
    hash bytea PRIMARY KEY, 
    user_id bigint NOT NULL REFERENCES sportgether_schema.users on DELETE CASCADE
    expiry timestamp(0) with time zone NOT NULL
    scope text NOT NULL
)

COMMIT;
