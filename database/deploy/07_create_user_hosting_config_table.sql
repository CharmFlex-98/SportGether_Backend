-- Deploy sportgether:07_create_user_hosting_config_table to pg

BEGIN;

CREATE TABLE IF NOT EXISTS sportgether_schema.user_hosting_config(
    user_id bigserial PRIMARY KEY NOT NULL REFERENCES sportgether_schema.users on DELETE CASCADE,
    host_count int NOT NULL DEFAULT 0,
    last_refresh_time timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    challenge_id text
);

COMMIT;
