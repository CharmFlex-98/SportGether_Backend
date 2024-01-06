-- Verify sportgether:create_user_table on pg

BEGIN;

SELECT id, username, email, created_at, is_blocked, version, gender, profile_icon_name
FROM sportgether_schema.users
WHERE FALSE;

ROLLBACK;
