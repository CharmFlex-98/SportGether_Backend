-- Verify sportgether:create_user_table on pg

BEGIN;

SELECT id, username, email, created_at, is_blocked, version
FROM sportgether_schema.users
WHERE FALSE;

ROLLBACK;
