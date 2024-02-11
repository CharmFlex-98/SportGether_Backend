-- Deploy sportgether:create_user_profile_table to pg

BEGIN;

CREATE table if not exists sportgether_schema.user_profile
(
    user_id          bigserial PRIMARY KEY NOT NULL REFERENCES sportgether_schema.users on DELETE CASCADE,
    preferred_name   text                  NOT NULL,
    gender           text                  NOT NULL,
    birth_date       date,
    join_date        date                  NOT NULL DEFAULT now(),
    status           text                  NOT NULL DEFAULT 'NOT_ONBOARDED',
    profile_icon_url text,
    signature        text,
    memo             text,
    version          int                   NOT NULL DEFAULT 1
);

COMMIT;


-- status can be NOT_ONBOARDED, ONBOARDED