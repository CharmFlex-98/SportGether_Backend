-- Deploy sportgether:06_create_firebase_messaging_token to pg

BEGIN;

CREATE TABLE IF NOT EXISTS sportgether_schema.firebase_messaging_token_table(
    user_id bigserial PRIMARY KEY NOT NULL REFERENCES sportgether_schema.users on DELETE CASCADE,
    token text NOT NULL
);

COMMIT;
