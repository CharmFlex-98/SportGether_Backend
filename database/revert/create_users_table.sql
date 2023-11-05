-- Revert sportgether:create_users_table from pg

BEGIN;

DROP TABLE users;

COMMIT;
