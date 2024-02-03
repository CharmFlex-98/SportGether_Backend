-- Deploy sportgether:create_user_profile_table to pg

BEGIN;

CREATE table if not exists sportgether_schema.user_profile
(
    user_id bigserial PRIMARY KEY NOT NULL REFERENCES sportgether_schema.users on DELETE CASCADE,
    preferred_name text NOT NULL,
    gender text NOT NULL,
    birthday timestamp(0) with time zone NOT NULL,
    join_date date,
    profile_icon_url text,
    signature text,
    memo text
);

COMMIT;
