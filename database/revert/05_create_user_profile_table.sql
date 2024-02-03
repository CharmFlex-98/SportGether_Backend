-- Revert sportgether:create_user_profile_table from pg

BEGIN;

DROP TABLE sportgether_schema.user_profile;

COMMIT;
