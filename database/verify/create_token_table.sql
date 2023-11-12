-- Verify sportgether:create_token_table on pg

BEGIN;

SELECT hashed, expiry_time, scope, user_id
FROM sportgether_schema.tokens
WHERE FALSE;

ROLLBACK;
