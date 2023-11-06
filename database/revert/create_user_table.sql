-- Revert sportgether:create_user_table from pg

BEGIN;

DROP TABLE sportgether_schema.users;
COMMIT;
