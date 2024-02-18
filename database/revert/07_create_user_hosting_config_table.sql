-- Revert sportgether:07_create_user_hosting_config_table from pg

BEGIN;

DROP TABLE sportgether_schema.user_hosting_config;

COMMIT;
