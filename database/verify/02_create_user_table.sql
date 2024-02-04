-- Verify sportgether:create_user_table on pg

BEGIN;

SELECT id, username, password, email, status, created_at, version
FROM sportgether_schema.users
WHERE FALSE;

ROLLBACK;
